import { message } from 'antd';
import { responseStatus, userFacingErrorMessage } from '@imaiplay/shared/api/errors';
import type { ApiEnvelope as SharedApiEnvelope } from '@imaiplay/shared/types/api';
import axios, {
  type AxiosError,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from 'axios';
import {
  clearAuthSession,
  createPortalSessionRefresher,
  isPortalSessionRefreshSuperseded,
  readPortalAccessToken,
  readPortalRefreshToken,
  readPortalTenantCode,
  shouldRefreshPortalRequest,
} from './authSession';
import {
  getActivePortalCode,
  getActivePortalTenantId,
} from './portalSession';

declare module 'axios' {
  interface AxiosRequestConfig {
    motivationSilent?: boolean;
  }
}

interface ApiEnvelope<T> extends SharedApiEnvelope<T> {
  code: number;
  message: string;
}

interface ApiErrorBody {
  message?: string;
  error?: string;
}

interface RefreshResult {
  token: string;
  refresh_token?: string;
}

interface RetryableRequest extends InternalAxiosRequestConfig {
  portalRetry?: boolean;
  motivationSilent?: boolean;
}

export const apiClient = axios.create({
  baseURL: '',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
});

const refreshClient = axios.create({
  baseURL: '',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

const refreshPortalSession = createPortalSessionRefresher(
  async (refreshToken, portal) => {
    const response = await refreshClient.post<ApiEnvelope<RefreshResult> | RefreshResult>(
      '/api/v1/auth/refresh',
      { refresh_token: refreshToken },
      { headers: { 'X-Tenant-Code': portal.code } },
    );
    const body = response.data;
    if (typeof body === 'object' && body !== null && 'code' in body) {
      if (body.code !== 0) throw new Error(body.message || '刷新登录状态失败');
      return body.data;
    }
    return body;
  },
  () => {
    const code = getActivePortalCode() ?? readPortalTenantCode();
    const tenantId = getActivePortalTenantId();
    return code && tenantId ? { code, tenant_id: tenantId } : undefined;
  },
);

apiClient.interceptors.request.use((config) => {
  const tenantCode = getActivePortalCode();
  if (tenantCode) {
    config.headers['X-Tenant-Code'] = tenantCode;
  }
  const token = readPortalAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response: AxiosResponse<ApiEnvelope<unknown> | unknown>) => {
    const payload = response.data;
    if (
      typeof payload === 'object' &&
      payload !== null &&
      'code' in payload &&
      typeof (payload as ApiEnvelope<unknown>).code === 'number'
    ) {
      const envelope = payload as ApiEnvelope<unknown>;
      if (envelope.code !== 0) {
        const error = new Error(envelope.message || '请求失败');
        if (!(response.config as RetryableRequest).motivationSilent) {
          message.error(userFacingErrorMessage(error));
        }
        return Promise.reject(error);
      }
      response.data = envelope.data;
    }
    return response;
  },
  async (error: AxiosError<ApiErrorBody>) => {
    const request = error.config as RetryableRequest | undefined;
    if (request && shouldRefreshPortalRequest({
      status: responseStatus(error),
      url: request.url,
      retried: request.portalRetry,
      hasRefreshToken: Boolean(readPortalRefreshToken()),
    })) {
      request.portalRetry = true;
      try {
        const token = await refreshPortalSession();
        request.headers.Authorization = `Bearer ${token}`;
        return await apiClient.request(request);
      } catch (refreshError) {
        if (isPortalSessionRefreshSuperseded(refreshError)) {
          return Promise.reject(error);
        }
        clearAuthSession();
        message.error('登录状态已过期，请重新登录');
        return Promise.reject(error);
      }
    }

    const authEndpoint = request?.url?.startsWith('/api/v1/auth/');
    if (responseStatus(error) === 401 && !authEndpoint) {
      clearAuthSession();
      message.error('登录状态已过期，请重新登录');
      return Promise.reject(error);
    }
    if (!request?.motivationSilent) message.error(userFacingErrorMessage(error));
    return Promise.reject(error);
  },
);
