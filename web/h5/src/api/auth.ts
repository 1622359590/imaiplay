import { apiClient, TOKEN_KEY, unwrap, type ApiEnvelope } from './client'

export interface LoginPayload {
  identifier: string
  password: string
}

export interface LoginResult {
  token: string
  expires_at: string
}

export async function login(payload: LoginPayload): Promise<LoginResult> {
  const response = await apiClient.post<ApiEnvelope<LoginResult>>('/api/v1/auth/login', {
    identifier: payload.identifier.trim(),
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
