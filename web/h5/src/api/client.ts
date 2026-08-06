import axios, {
  AxiosError,
  type InternalAxiosRequestConfig,
} from 'axios'
import { Toast } from 'antd-mobile'
import { createRefreshCoordinator } from '@imaiplay/shared/auth/sessionCore'
import { responseMessage, responseStatus } from '@imaiplay/shared/api/errors'
import type { ApiEnvelope as SharedApiEnvelope } from '@imaiplay/shared/types/api'
import {
  clearPortalSession,
  getActivePortalCode,
  getActivePortalTenantId,
  isValidPortalSession,
  PORTAL_ACCESS_TOKEN_KEY,
  PORTAL_REFRESH_TOKEN_KEY,
  PORTAL_TENANT_CODE_KEY,
  PortalSessionChangedError,
  portalLoginPath,
  readPortalAccessToken,
  readPortalRefreshToken,
  readPortalTenantCode,
  shouldExpirePortalSessionAfterRefresh,
} from './authSession'

export interface ApiEnvelope<T> extends SharedApiEnvelope<T> {
  code: number
  message: string
}

interface RefreshResult {
  token: string
  refresh_token?: string
}

interface RetriableRequest extends InternalAxiosRequestConfig {
  portalRetry?: boolean
}

export const apiClient = axios.create({
  baseURL: '',
  timeout: 12000,
  headers: { 'Content-Type': 'application/json' },
})

const refreshClient = axios.create({
  baseURL: '',
  timeout: 12000,
  headers: { 'Content-Type': 'application/json' },
})
function requestTenantCode(): string | undefined {
  return getActivePortalCode()
}

apiClient.interceptors.request.use((config) => {
  const tenantCode = requestTenantCode()
  if (tenantCode) {
    config.headers['X-Tenant-Code'] = tenantCode
  }
  const token = readPortalAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

const refreshAccessToken = createRefreshCoordinator({
  storage: localStorage,
  accessTokenKey: PORTAL_ACCESS_TOKEN_KEY,
  refreshTokenKey: PORTAL_REFRESH_TOKEN_KEY,
  identity: () => {
    const tenantCode = requestTenantCode() ?? readPortalTenantCode()
    const tenantId = getActivePortalTenantId()
    return tenantCode && tenantId ? `${tenantCode}:${tenantId}` : undefined
  },
  request: async (refreshToken) => {
    const tenantCode = requestTenantCode() ?? readPortalTenantCode()
    if (!tenantCode) throw new Error('登录状态已失效')
    const response = await refreshClient.post<ApiEnvelope<RefreshResult>>(
      '/api/v1/auth/refresh',
      { refresh_token: refreshToken },
      { headers: { 'X-Tenant-Code': tenantCode } },
    )
    return unwrap(response)
  },
  validateAccessToken: (token) => {
    const tenantId = getActivePortalTenantId()
    return Boolean(tenantId && isValidPortalSession(token, tenantId))
  },
  supersededError: () => new PortalSessionChangedError(),
  invalidAccessTokenError: () => new Error('刷新后的企业会话无效'),
  clearMissingRefreshToken: true,
  onCommitted: () => {
    const tenantCode = requestTenantCode() ?? readPortalTenantCode()
    if (tenantCode) localStorage.setItem(PORTAL_TENANT_CODE_KEY, tenantCode.toLowerCase())
  },
})

function expirePortalSession(): void {
  const hadSession = Boolean(readPortalAccessToken() || readPortalRefreshToken())
  if (!hadSession) return

  const tenantCode = requestTenantCode() ?? readPortalTenantCode()
  clearPortalSession()
  Toast.show({ icon: 'fail', content: '登录状态已失效，请重新登录' })
  const target = portalLoginPath(tenantCode)
  if (window.location.pathname !== target) {
    window.location.replace(target)
  }
}

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<{ message?: string }>) => {
    const request = error.config as RetriableRequest | undefined
    if (
      responseStatus(error) === 401 &&
      request &&
      !request.portalRetry &&
      readPortalRefreshToken()
    ) {
      request.portalRetry = true
      try {
        const token = await refreshAccessToken()
        request.headers.Authorization = `Bearer ${token}`
        return await apiClient(request)
      } catch (refreshError) {
        if (shouldExpirePortalSessionAfterRefresh(refreshError)) {
          expirePortalSession()
        }
        return Promise.reject(error)
      }
    }

    if (responseStatus(error) === 401) {
      expirePortalSession()
    } else {
      const message =
        responseMessage(error) ||
        (error.code === 'ECONNABORTED' ? '请求超时，请稍后重试' : '网络连接异常，请检查服务')
      Toast.show({ icon: 'fail', content: message })
    }
    return Promise.reject(error)
  },
)

export function unwrap<T>(response: { data: ApiEnvelope<T> }): T {
  if (response.data.code !== 0) {
    throw new Error(response.data.message)
  }
  return response.data.data
}
