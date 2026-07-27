import { apiClient, TENANT_KEY, TOKEN_KEY, unwrap, type ApiEnvelope } from './client'

export interface LoginPayload {
  tenantCode: string
  email: string
  password: string
}

export interface LoginResult {
  token: string
  expires_at: string
}

export async function login(payload: LoginPayload): Promise<LoginResult> {
  localStorage.setItem(TENANT_KEY, payload.tenantCode.trim())
  const response = await apiClient.post<ApiEnvelope<LoginResult>>('/api/v1/auth/login', {
    email: payload.email.trim(),
    password: payload.password,
  })
  const result = unwrap(response)
  localStorage.setItem(TOKEN_KEY, result.token)
  return result
}

export function logout(): void {
  localStorage.removeItem(TOKEN_KEY)
}

export function isAuthenticated(): boolean {
  return Boolean(localStorage.getItem(TOKEN_KEY))
}
