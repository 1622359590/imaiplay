import axios, {
  AxiosError,
  type InternalAxiosRequestConfig,
} from 'axios'
import { Toast } from 'antd-mobile'
import {
  classifyPortalRefreshFailure,
  clearPortalSession,
  createSingleFlight,
  getActivePortalCode,
  getActivePortalTenantId,
  getPortalSessionGeneration,
  isPortalSessionCurrent,
  isValidPortalSession,
  PortalSessionChangedError,
  portalLoginPath,
  readPortalAccessToken,
  readPortalRefreshToken,
  readPortalTenantCode,
  shouldExpirePortalSessionAfterRefresh,
  writePortalSession,
} from './authSession'

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
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
const runRefresh = createSingleFlight<string>()

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

async function refreshAccessToken(): Promise<string> {
  const sessionGeneration = getPortalSessionGeneration()
  const refreshToken = readPortalRefreshToken()
  const tenantCode = requestTenantCode() ?? readPortalTenantCode()
  if (!refreshToken || !tenantCode) throw new Error('登录状态已失效')

  try {
    const response = await refreshClient.post<ApiEnvelope<RefreshResult>>(
      '/api/v1/auth/refresh',
      { refresh_token: refreshToken },
      { headers: { 'X-Tenant-Code': tenantCode } },
    )
    if (!isPortalSessionCurrent(sessionGeneration, refreshToken)) {
      throw new PortalSessionChangedError()
    }
    const result = unwrap(response)
    const tenantId = getActivePortalTenantId()
    if (!tenantId || !isValidPortalSession(result.token, tenantId)) {
      throw new Error('刷新后的企业会话无效')
    }
    writePortalSession(result, tenantCode)
    return result.token
  } catch (error) {
    throw classifyPortalRefreshFailure(error, sessionGeneration, refreshToken)
  }
}

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
      error.response?.status === 401 &&
      request &&
      !request.portalRetry &&
      readPortalRefreshToken()
    ) {
      request.portalRetry = true
      try {
        const token = await runRefresh(refreshAccessToken)
        request.headers.Authorization = `Bearer ${token}`
        return await apiClient(request)
      } catch (refreshError) {
        if (shouldExpirePortalSessionAfterRefresh(refreshError)) {
          expirePortalSession()
        }
        return Promise.reject(error)
      }
    }

    if (error.response?.status === 401) {
      expirePortalSession()
    } else {
      const message =
        error.response?.data?.message ||
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
