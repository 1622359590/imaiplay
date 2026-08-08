import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { message } from 'antd'
import { responseStatus, userFacingErrorMessage } from '@imaiplay/shared/api/errors'
import {
  clearAuthSession,
  createSessionRefresher,
  isAdminSessionRefreshSuperseded,
  readAdminAccessToken,
} from './authSession'

export { ADMIN_ACCESS_TOKEN_KEY } from './authSession'

interface RefreshResponse {
  token: string
  refresh_token?: string
}

interface RefreshEnvelope {
  data: RefreshResponse
}

interface RetryableRequest extends InternalAxiosRequestConfig {
  _retry?: boolean
}

const refreshSession = createSessionRefresher(async (refreshToken) => {
  const response = await axios.post<RefreshEnvelope | RefreshResponse>(
    '/api/v1/auth/refresh',
    { refresh_token: refreshToken },
  )
  const body = response.data
  return 'data' in body ? body.data : body
})

const client = axios.create({
  baseURL: '',
  timeout: 15000,
})

client.interceptors.request.use((config) => {
  const token = readAdminAccessToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

client.interceptors.response.use(
  (response) => {
    const payload = response.data
    if (payload && typeof payload === 'object' && 'code' in payload && 'data' in payload) {
      response.data = payload.data
    }
    return response
  },
  async (error: AxiosError<{ message?: string; error?: string }>) => {
    if (error.response?.data?.error === 'account_exists_multiple_tenants') {
      return Promise.reject(error)
    }
    const request = error.config as RetryableRequest | undefined
    const authEndpoint = request?.url?.startsWith('/api/v1/auth/')
    if (responseStatus(error) === 401 && request && !request._retry && !authEndpoint) {
      request._retry = true
      try {
        const token = await refreshSession()
        request.headers.Authorization = `Bearer ${token}`
        return client.request(request)
      } catch (refreshError) {
        if (isAdminSessionRefreshSuperseded(refreshError)) return Promise.reject(error)
        clearAuthSession()
        message.error('登录状态已过期，请重新登录')
        return Promise.reject(error)
      }
    }
    if (authEndpoint) {
      if (responseStatus(error) === 401) clearAuthSession()
      return Promise.reject(error)
    }
    message.error(userFacingErrorMessage(error))
    if (responseStatus(error) === 401) clearAuthSession()
    return Promise.reject(error)
  },
)

export default client
