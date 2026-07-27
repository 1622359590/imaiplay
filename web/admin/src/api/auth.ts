import client, { TENANT_CODE_KEY, TOKEN_KEY } from './client'

export interface LoginPayload {
  tenant_code: string
  email: string
  password: string
}

export interface AuthUser {
  id: string
  name: string
  email: string
  role?: string
}

export interface LoginResponse {
  token?: string
  access_token?: string
  user?: AuthUser
  data?: {
    token?: string
    access_token?: string
    user?: AuthUser
  }
}

export async function login(payload: LoginPayload) {
  const response = await client.post<LoginResponse>('/api/v1/auth/login', payload, {
    headers: { 'X-Tenant-Code': payload.tenant_code },
  })
  const body = response.data
  const token = body.token || body.access_token
  if (!token) throw new Error('登录响应中缺少 token')
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(TENANT_CODE_KEY, payload.tenant_code)
  return { token, user: body.user }
}

export function logout() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(TENANT_CODE_KEY)
}
