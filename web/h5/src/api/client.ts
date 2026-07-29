import axios, { AxiosError } from 'axios'
import { Toast } from 'antd-mobile'

export const TOKEN_KEY = 'imaiplay_token'

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

export const apiClient = axios.create({
  baseURL: 'http://localhost:8080',
  timeout: 12000,
  headers: { 'Content-Type': 'application/json' },
})

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ message?: string }>) => {
    const message =
      error.response?.data?.message ||
      (error.code === 'ECONNABORTED' ? '请求超时，请稍后重试' : '网络连接异常，请检查服务')

    if (error.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
    }
    Toast.show({ icon: 'fail', content: message })
    return Promise.reject(error)
  },
)

export function unwrap<T>(response: { data: ApiEnvelope<T> }): T {
  if (response.data.code !== 0) {
    throw new Error(response.data.message)
  }
  return response.data.data
}
