export const LEGACY_TOKEN_KEY = 'imaiplay_token'
export const PORTAL_ACCESS_TOKEN_KEY = 'imaiplay_portal_access_token'
export const PORTAL_REFRESH_TOKEN_KEY = 'imaiplay_portal_refresh_token'
export const PORTAL_TENANT_CODE_KEY = 'imaiplay_portal_tenant_code'
export const PORTAL_SESSION_EXPIRED_EVENT = 'imaiplay:portal-session-expired'

const PLATFORM_HOST = 'play.imai.work'

export interface PortalIdentity {
  code: string
  tenant_id: string
}

export interface PortalSessionClaims {
  user_id: string
  tenant_id: string
  role: 'learner'
  exp: number
}

export class PortalSessionChangedError extends Error {
  constructor() {
    super('登录状态已发生变化')
    this.name = 'PortalSessionChangedError'
  }
}

export function shouldExpirePortalSessionAfterRefresh(error: unknown): boolean {
  return !(error instanceof PortalSessionChangedError)
}

interface TokenStorage {
  getItem: (key: string) => string | null
  setItem: (key: string, value: string) => void
  removeItem: (key: string) => void
}

interface SessionEventTarget {
  dispatchEvent: (event: Event) => unknown
}

let activePortalCode: string | undefined
let activePortalTenantId: string | undefined
let portalSessionGeneration = 0

function decodeClaims(token: string | null): PortalSessionClaims | null {
  const payload = token?.split('.')[1]
  if (!payload) return null

  try {
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const decoded = JSON.parse(
      atob(normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')),
    ) as Record<string, unknown>
    if (
      typeof decoded.user_id !== 'string' ||
      typeof decoded.tenant_id !== 'string' ||
      decoded.role !== 'learner' ||
      typeof decoded.exp !== 'number'
    ) {
      return null
    }
    return decoded as unknown as PortalSessionClaims
  } catch {
    return null
  }
}

export function isValidPortalSession(
  token: string | null,
  tenantId?: string,
  now = Date.now(),
): boolean {
  const claims = decodeClaims(token)
  return Boolean(
    claims &&
    claims.exp > Math.floor(now / 1000) &&
    (!tenantId || claims.tenant_id === tenantId),
  )
}

export function migrateLegacyPortalSession(
  storage: TokenStorage = localStorage,
  now = Date.now(),
): void {
  const legacy = storage.getItem(LEGACY_TOKEN_KEY)
  if (!isValidPortalSession(legacy, undefined, now)) return
  if (!storage.getItem(PORTAL_ACCESS_TOKEN_KEY)) {
    storage.setItem(PORTAL_ACCESS_TOKEN_KEY, legacy!)
  }
  storage.removeItem(LEGACY_TOKEN_KEY)
}

function removePortalTokens(storage: Pick<TokenStorage, 'removeItem'>): void {
  storage.removeItem(PORTAL_ACCESS_TOKEN_KEY)
  storage.removeItem(PORTAL_REFRESH_TOKEN_KEY)
  storage.removeItem(PORTAL_TENANT_CODE_KEY)
}

function invalidatePortalSession(storage: Pick<TokenStorage, 'removeItem'>): void {
  removePortalTokens(storage)
  portalSessionGeneration += 1
}

export function getPortalSessionGeneration(): number {
  return portalSessionGeneration
}

export function isPortalSessionCurrent(
  generation: number,
  refreshToken: string,
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
): boolean {
  return (
    portalSessionGeneration === generation &&
    storage.getItem(PORTAL_REFRESH_TOKEN_KEY) === refreshToken
  )
}

export function classifyPortalRefreshFailure(
  error: unknown,
  generation: number,
  refreshToken: string,
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
): unknown {
  return isPortalSessionCurrent(generation, refreshToken, storage)
    ? error
    : new PortalSessionChangedError()
}

export function clearPortalSession(
  storage: Pick<TokenStorage, 'removeItem'> = localStorage,
  eventTarget: SessionEventTarget = window,
): void {
  invalidatePortalSession(storage)
  eventTarget.dispatchEvent(new Event(PORTAL_SESSION_EXPIRED_EVENT))
}

export function bindPortalSession(
  portal: PortalIdentity,
  storage: TokenStorage = localStorage,
  now = Date.now(),
): boolean {
  migrateLegacyPortalSession(storage, now)
  const token = storage.getItem(PORTAL_ACCESS_TOKEN_KEY)
  if (!token) return false
  if (!isValidPortalSession(token, portal.tenant_id, now)) {
    invalidatePortalSession(storage)
    return false
  }
  storage.setItem(PORTAL_TENANT_CODE_KEY, portal.code.toLowerCase())
  return true
}

export function readPortalAccessToken(storage: TokenStorage = localStorage): string | null {
  migrateLegacyPortalSession(storage)
  return storage.getItem(PORTAL_ACCESS_TOKEN_KEY)
}

export function readPortalRefreshToken(
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
): string | null {
  return storage.getItem(PORTAL_REFRESH_TOKEN_KEY)
}

export function readPortalTenantCode(
  storage: Pick<TokenStorage, 'getItem'> = localStorage,
): string | undefined {
  const code = storage.getItem(PORTAL_TENANT_CODE_KEY)?.trim().toLowerCase()
  return code || undefined
}

export function writePortalSession(
  session: { token: string; refresh_token?: string },
  tenantCode: string,
  storage: TokenStorage = localStorage,
): void {
  if (!isValidPortalSession(session.token)) {
    throw new Error('登录响应中的学员会话无效')
  }
  storage.setItem(PORTAL_ACCESS_TOKEN_KEY, session.token)
  if (session.refresh_token) {
    storage.setItem(PORTAL_REFRESH_TOKEN_KEY, session.refresh_token)
  } else {
    storage.removeItem(PORTAL_REFRESH_TOKEN_KEY)
  }
  storage.setItem(PORTAL_TENANT_CODE_KEY, tenantCode.toLowerCase())
}

export function setActivePortalCode(code?: string): void {
  activePortalCode = code?.trim().toLowerCase() || undefined
  activePortalTenantId = undefined
}

export function getActivePortalCode(): string | undefined {
  return activePortalCode
}

export function setActivePortalIdentity(portal?: PortalIdentity): void {
  setActivePortalCode(portal?.code)
  activePortalTenantId = portal?.tenant_id || undefined
}

export function getActivePortalTenantId(): string | undefined {
  return activePortalTenantId
}

export function portalLoginPath(
  tenantCode: string | undefined,
  hostname = window.location.hostname,
): string {
  if (hostname.toLowerCase() !== PLATFORM_HOST) return '/h5/login'
  return tenantCode
    ? `/h5/t/${encodeURIComponent(tenantCode)}/login`
    : '/login'
}

export function portalRoutePath(
  tenantCode: string | undefined,
  hostname: string,
  childPath: string,
): string {
  const normalizedChild = childPath === '/'
    ? ''
    : `/${childPath.replace(/^\/+/, '')}`
  return hostname.toLowerCase() === PLATFORM_HOST && tenantCode
    ? `/t/${encodeURIComponent(tenantCode)}${normalizedChild}`
    : childPath
}

export function createSingleFlight<T>() {
  let pending: Promise<T> | undefined
  return (factory: () => Promise<T>): Promise<T> => {
    if (!pending) {
      pending = factory().finally(() => {
        pending = undefined
      })
    }
    return pending
  }
}
