import { apiClient, unwrap, type ApiEnvelope } from './client'
import {
  clearPortalSession,
  isValidPortalSession,
  readPortalAccessToken,
  readPortalRefreshToken,
  readPortalTenantCode,
  writePortalSession,
} from './authSession'
import type { TenantPortal } from './theme'

export interface LoginPayload {
  identifier: string
  password: string
}

interface AuthUser {
  id: string
  role: string
  tenant_id?: string
}

interface AuthenticatedLogin {
  requires_tenant_selection?: false
  token: string
  refresh_token?: string
  expires_at?: string
  user: AuthUser
  tenant?: TenantPortal
}

interface TenantSelectionRequired {
  requires_tenant_selection: true
}

type LoginResult = AuthenticatedLogin | TenantSelectionRequired

export async function login(
  payload: LoginPayload,
  portal: TenantPortal,
): Promise<void> {
  const response = await apiClient.post<ApiEnvelope<LoginResult>>('/api/v1/auth/login', {
    identifier: payload.identifier.trim(),
    password: payload.password,
  })
  const result = unwrap(response)
  if (result.requires_tenant_selection) {
    throw new Error('该账号属于多个企业，请前往平台统一登录选择企业')
  }
  if (
    result.user?.role !== 'learner' ||
    result.user.tenant_id !== portal.tenant_id ||
    result.tenant?.tenant_id !== portal.tenant_id ||
    result.tenant.code.toLowerCase() !== portal.code.toLowerCase() ||
    !isValidPortalSession(result.token, portal.tenant_id)
  ) {
    throw new Error(
      result.user?.role && result.user.role !== 'learner'
        ? '管理人员请使用管理后台登录'
        : '登录响应中的企业会话无效',
    )
  }
  writePortalSession(result, portal.code)
}

export function logout(): void {
  const accessToken = readPortalAccessToken()
  const refreshToken = readPortalRefreshToken()
  clearPortalSession()
  if (!accessToken) return

  void apiClient.post(
    '/api/v1/auth/logout',
    { refresh_token: refreshToken || '' },
    { headers: { Authorization: `Bearer ${accessToken}` } },
  ).catch(() => undefined)
}

export function isAuthenticated(portal?: TenantPortal): boolean {
  if (!portal) return false
  const token = readPortalAccessToken()
  return (
    isValidPortalSession(token, portal.tenant_id) &&
    readPortalTenantCode() === portal.code.toLowerCase()
  )
}

export async function forgotPassword(phone: string): Promise<void> {
  const response = await apiClient.post<ApiEnvelope<unknown>>(
    '/api/v1/auth/forgot-password',
    { phone: phone.trim() },
  )
  unwrap(response)
}

export async function resetPassword(
  phone: string,
  code: string,
  newPassword: string,
): Promise<void> {
  const response = await apiClient.post<ApiEnvelope<unknown>>(
    '/api/v1/auth/reset-password',
    {
      phone: phone.trim(),
      code: code.trim(),
      new_password: newPassword,
    },
  )
  unwrap(response)
}
