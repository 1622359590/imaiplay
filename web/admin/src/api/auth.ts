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
  refresh_token?: string
  user?: AuthUser
  data?: {
    token?: string
    access_token?: string
    user?: AuthUser
  }
}

export const REFRESH_TOKEN_KEY = 'imaiplay_refresh_token'

export function tokenRole(token: string | null = localStorage.getItem(TOKEN_KEY)) {
  if (!token) return undefined
  try {
    const encoded = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    const payload = JSON.parse(atob(encoded.padEnd(Math.ceil(encoded.length / 4) * 4, '='))) as { role?: string }
    return payload.role
  } catch {
    return undefined
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
  if (body.refresh_token) localStorage.setItem(REFRESH_TOKEN_KEY, body.refresh_token)
  localStorage.setItem(TENANT_CODE_KEY, payload.tenant_code)
  return { token, user: body.user }
}

export function logout() {
	const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
	if (refreshToken) void client.post('/api/v1/auth/logout', { refresh_token: refreshToken }).catch(() => undefined)
	localStorage.removeItem(TOKEN_KEY)
	localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(TENANT_CODE_KEY)
}
