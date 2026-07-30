import client, { TOKEN_KEY } from './client'

export interface LoginPayload {
  identifier: string
  password: string
  tenant_code?: string
}
export interface LoginCodePayload { phone: string; code: string }

export interface AuthUser {
  id: string
  name: string
  email: string
  phone?: string
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
  const tenantCode = payload.tenant_code?.trim()
  const response = await client.post<LoginResponse>(
    '/api/v1/auth/login',
    {
      identifier: payload.identifier.trim(),
      password: payload.password,
    },
    tenantCode ? { headers: { 'X-Tenant-Code': tenantCode } } : undefined,
  )
  const body = response.data
  const token = body.token || body.access_token
  if (!token) throw new Error('登录响应中缺少 token')
  localStorage.setItem(TOKEN_KEY, token)
  if (body.refresh_token) localStorage.setItem(REFRESH_TOKEN_KEY, body.refresh_token)
  return { token, user: body.user }
}
export async function sendLoginCode(phone: string) { return client.post('/api/v1/auth/login-code/send', { phone }) }
export async function loginWithCode(payload: LoginCodePayload) {
  const response = await client.post<LoginResponse>('/api/v1/auth/login-code', { phone: payload.phone, code: payload.code })
  const body = response.data; const token = body.token || body.access_token; if (!token) throw new Error('登录响应中缺少 token'); localStorage.setItem(TOKEN_KEY, token); if (body.refresh_token) localStorage.setItem(REFRESH_TOKEN_KEY, body.refresh_token); return { token, user: body.user }
}

export function forgotPassword(phone: string) { return client.post('/api/v1/auth/forgot-password', { phone }) }
export function resetPassword(phone: string, code: string, new_password: string) { return client.post('/api/v1/auth/reset-password', { phone, code, new_password }) }

export function logout() {
	const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
	if (refreshToken) void client.post('/api/v1/auth/logout', { refresh_token: refreshToken }).catch(() => undefined)
	localStorage.removeItem(TOKEN_KEY)
	localStorage.removeItem(REFRESH_TOKEN_KEY)
}
