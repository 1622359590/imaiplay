import axios, { AxiosError } from 'axios'
import { message } from 'antd'

export const TOKEN_KEY = 'imaiplay_token'

const client = axios.create({
  baseURL: '',
  timeout: 15000,
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
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
  (error: AxiosError<{ message?: string; error?: string }>) => {
    const text =
      error.response?.data?.message ||
      error.response?.data?.error ||
      error.message ||
      '请求失败，请稍后重试'
    message.error(text)
    if (error.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
    }
    return Promise.reject(error)
  },
)

export default client
