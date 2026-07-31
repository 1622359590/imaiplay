import { message } from 'antd';
import axios, { type AxiosError, type AxiosResponse } from 'axios';
import { clearAuthSession, TOKEN_KEY } from './authSession';

interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

interface ApiErrorBody {
  message?: string;
  error?: string;
}

export const apiClient = axios.create({
  baseURL: '',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY);
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
        message.error(error.message);
        return Promise.reject(error);
      }
      response.data = envelope.data;
    }
    return response;
  },
  (error: AxiosError<ApiErrorBody>) => {
    const text =
      error.response?.data?.message ??
      error.response?.data?.error ??
      (error.code === 'ECONNABORTED' ? '请求超时，请稍后重试' : '网络异常，请检查服务是否可用');
    message.error(text);

    if (error.response?.status === 401) {
      clearAuthSession();
    }
    return Promise.reject(error);
  },
);
