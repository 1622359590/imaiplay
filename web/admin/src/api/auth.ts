import client from './client'
import {
  ADMIN_ACCESS_TOKEN_KEY,
  clearAuthSession,
  createAdminLogoutRequest,
  decodeSessionClaims,
  isAdminRole,
  isValidAdminSession,
  writeAdminSession,
} from './authSession'

export interface LoginPayload {
  identifier: string
  password: string
}
export interface LoginCodePayload { phone: string; code: string }

export interface AuthUser {
  id: string
  name: string
  email: string
  phone?: string
  role?: string
  tenant_id?: string
}

export interface OrganizationOption {
  code: string
  name: string
  logo_url?: string
  role: string
}

export interface TenantSelectionRequired {
  requires_tenant_selection: true
  selection_token: string
  organizations: OrganizationOption[]
}

export interface AuthenticatedLoginResult {
  requires_tenant_selection?: false
  token: string
  refresh_token?: string
  expires_at?: string
  user: AuthUser
  tenant?: { id?: string; tenant_id?: string; code: string; name?: string }
}

export type LoginResult = TenantSelectionRequired | AuthenticatedLoginResult

export function tokenRole(token: string | null = localStorage.getItem(ADMIN_ACCESS_TOKEN_KEY)) {
  return decodeSessionClaims(token)?.role
}

export async function login(payload: LoginPayload): Promise<LoginResult> {
  const response = await client.post<LoginResult>('/api/v1/auth/login', {
    identifier: payload.identifier.trim(),
    password: payload.password,
  })
  return response.data
}

export async function selectTenant(selectionToken: string, tenantCode: string): Promise<AuthenticatedLoginResult> {
  const response = await client.post<AuthenticatedLoginResult>('/api/v1/auth/select-tenant', {
    selection_token: selectionToken,
    tenant_code: tenantCode,
  })
  return response.data
}

export function persistAdminLogin(result: AuthenticatedLoginResult): { token: string; user: AuthUser } {
  const claims = decodeSessionClaims(result.token)
  const tenantID = result.tenant?.tenant_id || result.tenant?.id
  if (!claims || !isValidAdminSession(result.token) || !isAdminRole(result.user.role) ||
    claims.user_id !== result.user.id || claims.role !== result.user.role ||
    (claims.role === 'superadmin'
      ? Boolean(claims.tenant_id || result.user.tenant_id || result.tenant)
      : !tenantID || claims.tenant_id !== tenantID || claims.tenant_id !== result.user.tenant_id)) {
    throw new Error('登录响应中的会话或企业信息无效')
  }
  writeAdminSession(result)
  return { token: result.token, user: result.user }
}

export async function sendLoginCode(phone: string) { return client.post('/api/v1/auth/login-code/send', { phone }) }
export async function loginWithCode(payload: LoginCodePayload) {
  const response = await client.post<AuthenticatedLoginResult>('/api/v1/auth/login-code', { phone: payload.phone, code: payload.code })
  return persistAdminLogin(response.data)
}

export function forgotPassword(phone: string) { return client.post('/api/v1/auth/forgot-password', { phone }) }
export function resetPassword(phone: string, code: string, new_password: string) { return client.post('/api/v1/auth/reset-password', { phone, code, new_password }) }

export function logout() {
  const request = createAdminLogoutRequest()
  if (request) {
    void client.post(
      '/api/v1/auth/logout',
      { refresh_token: request.refreshToken },
      { headers: { Authorization: request.authorization } },
    ).catch(() => undefined)
  }
  clearAuthSession()
}
