import { apiClient } from './client';
import {
  clearAuthSession,
  bindPortalSessionToPortal,
  createPortalLogoutRequest,
  decodeClaims,
  isPortalSessionToken,
  readPortalAccessToken,
  validateAuthenticatedSession,
  writeAdminSession,
  writePortalSession,
} from './authSession';
import type { Portal } from './portal';
import { authenticatedLoginDestination, type PortalMode } from '../utils/portalRouting';

export interface LoginValues {
  identifier: string;
  password: string;
}

export interface AuthUser {
  id: string;
  name: string;
  email: string;
  phone?: string;
  role?: string;
  tenant_id?: string;
}

export interface OrganizationOption {
  code: string;
  name: string;
  logo_url?: string;
  role: string;
}

export interface TenantSelectionRequired {
  requires_tenant_selection: true;
  selection_token: string;
  organizations: OrganizationOption[];
}

export interface AuthenticatedLoginResult {
  requires_tenant_selection?: false;
  token: string;
  refresh_token?: string;
  expires_at?: string;
  user: AuthUser;
  tenant?: Portal;
}

export type LoginResult = TenantSelectionRequired | AuthenticatedLoginResult;

export async function login(values: LoginValues): Promise<LoginResult> {
  const response = await apiClient.post<LoginResult>('/api/v1/auth/login', {
    identifier: values.identifier.trim(),
    password: values.password,
  });
  return response.data;
}

export async function selectTenant(
  values: { selection_token: string; tenant_code: string },
): Promise<AuthenticatedLoginResult> {
  const response = await apiClient.post<AuthenticatedLoginResult>('/api/v1/auth/select-tenant', values);
  return response.data;
}

export function persistLogin(
  result: AuthenticatedLoginResult,
  mode: PortalMode,
  explicitTenantCode?: string,
): string {
  validateAuthenticatedSession(result);
  const claims = decodeClaims(result.token);
  if (!claims) throw new Error('登录响应中的会话或企业信息无效');
  const redirect = authenticatedLoginDestination(
    claims.role,
    result.tenant?.code,
    mode,
    explicitTenantCode,
  );

  if (claims.role === 'learner') {
    if (!isPortalSessionToken(result.token)) throw new Error('登录响应中的学员会话无效');
    writePortalSession(result, result.tenant!.code);
  } else {
    writeAdminSession(result);
  }
  return redirect;
}

export function bindCurrentPortalSession(portal: Portal): boolean {
  return bindPortalSessionToPortal(portal);
}

export function logout(): void {
  const request = createPortalLogoutRequest();
  if (request) {
    void apiClient.post(
      '/api/v1/auth/logout',
      { refresh_token: request.refreshToken },
      { headers: { Authorization: request.authorization } },
    ).catch(() => undefined);
  }
  clearAuthSession();
}

export function isAuthenticated(): boolean {
  const token = readPortalAccessToken();
  const authenticated = isPortalSessionToken(token);
  if (!authenticated && token) clearAuthSession();
  return authenticated;
}
