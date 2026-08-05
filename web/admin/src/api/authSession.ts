export const LEGACY_ACCESS_TOKEN_KEY = 'imaiplay_token'
export const LEGACY_REFRESH_TOKEN_KEY = 'imaiplay_refresh_token'
export const ADMIN_ACCESS_TOKEN_KEY = 'imaiplay_admin_access_token'
export const ADMIN_REFRESH_TOKEN_KEY = 'imaiplay_admin_refresh_token'
export const AUTH_SESSION_EXPIRED_EVENT = 'imaiplay:admin-session-expired'

export type AdminSessionRole = 'instructor' | 'tenant_admin' | 'superadmin'

export interface SessionClaims {
  user_id: string
  tenant_id: string
  role: AdminSessionRole | 'learner'
  exp: number
}

interface SessionStorage {
  getItem: (key: string) => string | null
  setItem: (key: string, value: string) => void
  removeItem: (key: string) => void
}

interface SessionEventTarget {
  dispatchEvent: (event: Event) => unknown
}

interface RefreshedSession {
  token: string
  refresh_token?: string
}

const sessionGenerations = new WeakMap<object, number>()

function sessionGeneration(storage: object): number {
  return sessionGenerations.get(storage) ?? 0
}

function advanceSessionGeneration(storage: object): void {
  sessionGenerations.set(storage, sessionGeneration(storage) + 1)
}

export class AdminSessionRefreshSupersededError extends Error {
  constructor() {
    super('refresh superseded')
    this.name = 'AdminSessionRefreshSupersededError'
  }
}

export function isAdminSessionRefreshSuperseded(error: unknown): boolean {
  return error instanceof AdminSessionRefreshSupersededError
}

function decodePayload(token: string): Record<string, unknown> | null {
  const encoded = token.split('.')[1]
  if (!encoded) return null
  try {
    const normalized = encoded.replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '='))) as Record<string, unknown>
  } catch {
    return null
  }
}

export function decodeSessionClaims(token: string | null): SessionClaims | null {
  if (!token) return null
  const payload = decodePayload(token)
  if (!payload || typeof payload.user_id !== 'string' || typeof payload.tenant_id !== 'string' ||
    typeof payload.role !== 'string' || typeof payload.exp !== 'number') return null
  if (!['learner', 'instructor', 'tenant_admin', 'superadmin'].includes(payload.role)) return null
  return payload as unknown as SessionClaims
}

export function isAdminRole(role: unknown): role is AdminSessionRole {
  return role === 'instructor' || role === 'tenant_admin' || role === 'superadmin'
}

export function isValidAdminSession(token: string | null, now = Date.now()): boolean {
  const claims = decodeSessionClaims(token)
  if (!claims || !isAdminRole(claims.role) || claims.exp <= Math.floor(now / 1000)) return false
  return claims.role === 'superadmin' ? claims.tenant_id === '' : claims.tenant_id !== ''
}

export function migrateLegacyAdminSession(storage: SessionStorage = localStorage, now = Date.now()): void {
  const legacyToken = storage.getItem(LEGACY_ACCESS_TOKEN_KEY)
  const legacyIsAdminSession = isValidAdminSession(legacyToken, now)
  if (storage.getItem(ADMIN_ACCESS_TOKEN_KEY)) {
    if (legacyIsAdminSession) {
      storage.removeItem(LEGACY_ACCESS_TOKEN_KEY)
      storage.removeItem(LEGACY_REFRESH_TOKEN_KEY)
    }
    return
  }
  if (!legacyIsAdminSession) return

  storage.setItem(ADMIN_ACCESS_TOKEN_KEY, legacyToken!)
  storage.removeItem(LEGACY_ACCESS_TOKEN_KEY)
  const legacyRefreshToken = storage.getItem(LEGACY_REFRESH_TOKEN_KEY)
  if (legacyRefreshToken) {
    storage.setItem(ADMIN_REFRESH_TOKEN_KEY, legacyRefreshToken)
    storage.removeItem(LEGACY_REFRESH_TOKEN_KEY)
  }
  advanceSessionGeneration(storage)
}

export function readAdminAccessToken(storage: SessionStorage = localStorage): string | null {
  migrateLegacyAdminSession(storage)
  // Keep an expired access token until the Axios interceptor has a chance to
  // exchange the paired refresh token. Login writes and legacy migration are
  // still strictly validated before they reach these scoped keys.
  return storage.getItem(ADMIN_ACCESS_TOKEN_KEY)
}

export function writeAdminSession(
  session: { token: string; refresh_token?: string },
  storage: SessionStorage = localStorage,
): void {
  if (!isValidAdminSession(session.token)) throw new Error('登录响应中的管理会话无效')
  storage.setItem(ADMIN_ACCESS_TOKEN_KEY, session.token)
  if (session.refresh_token) storage.setItem(ADMIN_REFRESH_TOKEN_KEY, session.refresh_token)
  else storage.removeItem(ADMIN_REFRESH_TOKEN_KEY)
  advanceSessionGeneration(storage)
}

export function createAdminLogoutRequest(
  storage: Pick<SessionStorage, 'getItem'> = localStorage,
): { refreshToken: string; authorization: string } | undefined {
  const accessToken = storage.getItem(ADMIN_ACCESS_TOKEN_KEY)
  const refreshToken = storage.getItem(ADMIN_REFRESH_TOKEN_KEY)
  if (!accessToken || !refreshToken) return undefined
  return { refreshToken, authorization: `Bearer ${accessToken}` }
}

export function createSessionRefresher(
  request: (refreshToken: string) => Promise<RefreshedSession>,
  storage: SessionStorage = localStorage,
) {
  let pending: {
    generation: number
    refreshToken: string
    promise: Promise<string>
  } | null = null

  return () => {
    const refreshToken = storage.getItem(ADMIN_REFRESH_TOKEN_KEY)
    if (!refreshToken) return Promise.reject(new Error('refresh token missing'))
    const generation = sessionGeneration(storage)
    if (pending?.generation === generation && pending.refreshToken === refreshToken) {
      return pending.promise
    }

    let promise!: Promise<string>
    promise = request(refreshToken)
      .then((session) => {
        if (sessionGeneration(storage) !== generation ||
          storage.getItem(ADMIN_REFRESH_TOKEN_KEY) !== refreshToken) {
          throw new AdminSessionRefreshSupersededError()
        }
        if (!session.token || !isValidAdminSession(session.token)) {
          throw new Error('access token missing')
        }
        storage.setItem(ADMIN_ACCESS_TOKEN_KEY, session.token)
        if (session.refresh_token) storage.setItem(ADMIN_REFRESH_TOKEN_KEY, session.refresh_token)
        return session.token
      })
      .finally(() => {
        if (pending?.promise === promise) pending = null
      })
    pending = { generation, refreshToken, promise }
    return promise
  }
}

export function clearAuthSession(
  storage: Pick<SessionStorage, 'removeItem'> = localStorage,
  eventTarget: SessionEventTarget = window,
) {
  storage.removeItem(ADMIN_ACCESS_TOKEN_KEY)
  storage.removeItem(ADMIN_REFRESH_TOKEN_KEY)
  advanceSessionGeneration(storage)
  eventTarget.dispatchEvent(new Event(AUTH_SESSION_EXPIRED_EVENT))
}
