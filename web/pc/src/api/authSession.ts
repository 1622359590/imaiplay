export const TOKEN_KEY = 'imaiplay_token';
export const PORTAL_ACCESS_TOKEN_KEY = 'imaiplay_portal_access_token';
export const PORTAL_REFRESH_TOKEN_KEY = 'imaiplay_portal_refresh_token';
export const PORTAL_TENANT_CODE_KEY = 'imaiplay_portal_tenant_code';
export const ADMIN_ACCESS_TOKEN_KEY = 'imaiplay_admin_access_token';
export const ADMIN_REFRESH_TOKEN_KEY = 'imaiplay_admin_refresh_token';
export const SESSION_EXPIRED_EVENT = 'imaiplay:session-expired';

export type SessionRole = 'learner' | 'instructor' | 'tenant_admin' | 'superadmin';
export type StaffSessionRole = Exclude<SessionRole, 'learner'>;

export interface SessionClaims {
  user_id: string;
  tenant_id: string;
  role: SessionRole;
  exp: number;
}

type TokenStorage = StorageLike;

interface SessionEventTarget {
  dispatchEvent: (event: Event) => unknown;
}

interface RefreshedPortalSession {
  token: string;
  refresh_token?: string;
}

interface PortalIdentity {
  code: string;
  tenant_id: string;
}

export class PortalSessionRefreshSupersededError extends Error {
  constructor() {
    super('登录状态已发生变化');
    this.name = 'PortalSessionRefreshSupersededError';
  }
}

export function isPortalSessionRefreshSuperseded(error: unknown): boolean {
  return error instanceof PortalSessionRefreshSupersededError;
}

export function decodeClaims(token: string): SessionClaims | null {
  const payload = decodeJwtPayload(token);
  if (!payload || typeof payload.user_id !== 'string' || typeof payload.tenant_id !== 'string' ||
    typeof payload.exp !== 'number' || !isSessionRole(payload.role)) {
    return null;
  }
  return payload as unknown as SessionClaims;
}

export function isSessionRole(role: unknown): role is SessionRole {
  return role === 'learner' || role === 'instructor' || role === 'tenant_admin' || role === 'superadmin';
}

export function readValidLegacyStaffRole(
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
  now = Date.now(),
): StaffSessionRole | undefined {
  const claims = decodeClaims(storage.getItem(TOKEN_KEY) ?? '');
  if (!claims || claims.role === 'learner' || claims.exp <= Math.floor(now / 1000)) {
    return undefined;
  }
  if (claims.role === 'superadmin') {
    return claims.tenant_id ? undefined : claims.role;
  }
  return claims.tenant_id ? claims.role : undefined;
}

export function isLearnerSessionToken(token: string | null, now = Date.now()): boolean {
  if (!token) return false;
  const payload = decodeJwtPayload(token);
  return payload?.role === 'learner' && typeof payload.exp === 'number' &&
    payload.exp > Math.floor(now / 1000);
}

export function isPortalSessionToken(
  token: string | null,
  tenantId?: string,
  now = Date.now(),
): boolean {
  if (!token) return false;
  const claims = decodeClaims(token);
  return claims?.role === 'learner' && claims.exp > Math.floor(now / 1000) &&
    (!tenantId || claims.tenant_id === tenantId);
}

export function migrateLegacySession(storage: TokenStorage = localStorage, now = Date.now()): void {
  const legacyToken = storage.getItem(TOKEN_KEY);
  const legacyIsLearnerSession = isPortalSessionToken(legacyToken, undefined, now);
  if (storage.getItem(PORTAL_ACCESS_TOKEN_KEY)) {
    if (legacyIsLearnerSession) storage.removeItem(TOKEN_KEY);
    return;
  }
  if (!legacyIsLearnerSession) return;
  storage.setItem(PORTAL_ACCESS_TOKEN_KEY, legacyToken!);
  storage.removeItem(TOKEN_KEY);
  markSessionChanged(storage);
}

function removePortalSession(storage: Pick<TokenStorage, 'removeItem'>): void {
  storage.removeItem(PORTAL_ACCESS_TOKEN_KEY);
  storage.removeItem(PORTAL_REFRESH_TOKEN_KEY);
  storage.removeItem(PORTAL_TENANT_CODE_KEY);
}

function invalidatePortalSession(storage: Pick<TokenStorage, 'removeItem'> & object): void {
  removePortalSession(storage);
  markSessionChanged(storage);
}

// A legacy learner session is not trusted for a portal until the public portal
// response confirms that its tenant ID is the tenant encoded in the JWT.
export function bindPortalSessionToPortal(
  portal: { code: string; tenant_id: string },
  storage: TokenStorage = localStorage,
  now = Date.now(),
): boolean {
  migrateLegacySession(storage, now);
  const token = storage.getItem(PORTAL_ACCESS_TOKEN_KEY);
  if (!token) return false;
  if (!isPortalSessionToken(token, portal.tenant_id, now)) {
    invalidatePortalSession(storage);
    return false;
  }
  storage.setItem(PORTAL_TENANT_CODE_KEY, portal.code.toLowerCase());
  return true;
}

export function readPortalAccessToken(storage: TokenStorage = localStorage): string | null {
  migrateLegacySession(storage);
  return storage.getItem(PORTAL_ACCESS_TOKEN_KEY);
}

export function readPortalRefreshToken(
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
): string | null {
  return storage.getItem(PORTAL_REFRESH_TOKEN_KEY);
}

export function readPortalClaims(storage: TokenStorage = localStorage): SessionClaims | null {
  return decodeClaims(readPortalAccessToken(storage) ?? '');
}

export function portalSessionMatchesTenantCode(
  tenantCode: string,
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
): boolean {
  return storage.getItem(PORTAL_TENANT_CODE_KEY) === tenantCode.toLowerCase();
}

export function readPortalTenantCode(
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
): string | undefined {
  const code = storage.getItem(PORTAL_TENANT_CODE_KEY)?.trim().toLowerCase();
  return code || undefined;
}

export function portalSessionMatchesPortal(
  portal: { code: string; tenant_id: string },
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
  now = Date.now(),
): boolean {
  return portalSessionMatchesTenantCode(portal.code, storage) &&
    isPortalSessionToken(storage.getItem(PORTAL_ACCESS_TOKEN_KEY), portal.tenant_id, now);
}

export function writePortalSession(
  session: { token: string; refresh_token?: string },
  tenantCode: string,
  storage: TokenStorage = localStorage,
): void {
  if (!isPortalSessionToken(session.token)) throw new Error('登录响应中的学员会话无效');
  storage.setItem(PORTAL_ACCESS_TOKEN_KEY, session.token);
  if (session.refresh_token) storage.setItem(PORTAL_REFRESH_TOKEN_KEY, session.refresh_token);
  else storage.removeItem(PORTAL_REFRESH_TOKEN_KEY);
  storage.setItem(PORTAL_TENANT_CODE_KEY, tenantCode.toLowerCase());
  markSessionChanged(storage);
}

export function writeAdminSession(
  session: { token: string; refresh_token?: string },
  storage: TokenStorage = localStorage,
): void {
  const claims = decodeClaims(session.token);
  if (!claims || claims.exp <= Math.floor(Date.now() / 1000) || claims.role === 'learner' ||
    (claims.role !== 'superadmin' && !claims.tenant_id) ||
    (claims.role === 'superadmin' && claims.tenant_id)) {
    throw new Error('登录响应中的管理会话无效');
  }
  storage.setItem(ADMIN_ACCESS_TOKEN_KEY, session.token);
  if (session.refresh_token) storage.setItem(ADMIN_REFRESH_TOKEN_KEY, session.refresh_token);
  else storage.removeItem(ADMIN_REFRESH_TOKEN_KEY);
}

export function createPortalLogoutRequest(
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
): { refreshToken: string; authorization: string } | undefined {
  const accessToken = storage.getItem(PORTAL_ACCESS_TOKEN_KEY);
  const refreshToken = storage.getItem(PORTAL_REFRESH_TOKEN_KEY);
  if (!accessToken || !refreshToken) return undefined;
  return { refreshToken, authorization: `Bearer ${accessToken}` };
}

export function createPortalSessionRefresher(
  request: (
    refreshToken: string,
    portal: PortalIdentity,
  ) => Promise<RefreshedPortalSession>,
  currentPortal: () => PortalIdentity | undefined,
  storage: TokenStorage = localStorage,
) {
  const refresh = createRefreshCoordinator({
    storage,
    accessTokenKey: PORTAL_ACCESS_TOKEN_KEY,
    refreshTokenKey: PORTAL_REFRESH_TOKEN_KEY,
    identity: () => {
      const portal = currentPortal();
      return portal?.code && portal.tenant_id
        ? `${portal.code.toLowerCase()}:${portal.tenant_id}`
        : undefined;
    },
    request: async (refreshToken) => {
      const portal = currentPortal();
      if (!portal?.code || !portal.tenant_id) throw new Error('登录状态已失效');
      return request(refreshToken, portal);
    },
    validateAccessToken: (token) => {
      const portal = currentPortal();
      return Boolean(portal?.tenant_id && isPortalSessionToken(token, portal.tenant_id));
    },
    invalidAccessTokenError: () => new Error('刷新后的企业会话无效'),
    supersededError: () => new PortalSessionRefreshSupersededError(),
    onCommitted: () => {
      const portal = currentPortal();
      if (portal?.code) storage.setItem(PORTAL_TENANT_CODE_KEY, portal.code.toLowerCase());
      markSessionChanged(storage);
    },
  });

  return () => {
    const refreshToken = storage.getItem(PORTAL_REFRESH_TOKEN_KEY);
    const portal = currentPortal();
    if (!refreshToken || !portal?.code || !portal.tenant_id) {
      return Promise.reject(new Error('登录状态已失效'));
    }
    return refresh();
  };
}

export function shouldRefreshPortalRequest(input: {
  status?: number;
  url?: string;
  retried?: boolean;
  hasRefreshToken: boolean;
}): boolean {
  return input.status === 401 && !input.retried && input.hasRefreshToken &&
    !input.url?.startsWith('/api/v1/auth/');
}

export function validateAuthenticatedSession(result: unknown, now = Date.now()): asserts result is {
  token: string;
  refresh_token?: string;
  user: { id: string; role?: string; tenant_id?: string };
  tenant?: { code: string; tenant_id: string };
} {
  if (!result || typeof result !== 'object') throw new Error('登录响应中的会话无效');
  const response = result as {
    token?: unknown;
    user?: { id?: unknown; role?: unknown; tenant_id?: unknown };
    tenant?: { code?: unknown; tenant_id?: unknown };
  };
  if (typeof response.token !== 'string' || !response.user ||
    typeof response.user.id !== 'string') {
    throw new Error('登录响应中的会话无效');
  }
  const claims = decodeClaims(response.token);
  if (!claims || claims.exp <= Math.floor(now / 1000) ||
    claims.user_id !== response.user.id || claims.role !== response.user.role ||
    claims.tenant_id !== response.user.tenant_id) {
    throw new Error('登录响应中的会话无效');
  }
  if (claims.role === 'superadmin') {
    if (claims.tenant_id || response.tenant) throw new Error('登录响应中的会话无效');
    return;
  }
  if (!claims.tenant_id || !response.tenant ||
    claims.tenant_id !== response.tenant.tenant_id ||
    typeof response.tenant.code !== 'string' || !response.tenant.code) {
    throw new Error('登录响应中的会话无效');
  }
}

export function clearAuthSession(
  storage: Pick<TokenStorage, 'removeItem'> & object = localStorage,
  eventTarget: SessionEventTarget = window,
): void {
  invalidatePortalSession(storage);
  eventTarget.dispatchEvent(new Event(SESSION_EXPIRED_EVENT));
}
import {
  createRefreshCoordinator,
  decodeJwtPayload,
  markSessionChanged,
  type StorageLike,
} from '@imaiplay/shared/auth/sessionCore';
