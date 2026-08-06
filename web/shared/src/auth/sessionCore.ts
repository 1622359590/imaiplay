export interface StorageLike {
  getItem(key: string): string | null
  setItem(key: string, value: string): void
  removeItem(key: string): void
}

export interface RefreshSession {
  token: string
  refresh_token?: string
}

export interface RefreshCoordinatorOptions {
  storage: StorageLike
  accessTokenKey: string
  refreshTokenKey: string
  request(refreshToken: string): Promise<RefreshSession>
  validateAccessToken(token: string): boolean
  supersededError(): Error
  invalidAccessTokenError?: () => Error
  identity?: () => string | undefined
  clearMissingRefreshToken?: boolean
  onCommitted?: (session: RefreshSession) => void
}

const sessionGenerations = new WeakMap<object, number>()

function sessionGeneration(storage: object): number {
  return sessionGenerations.get(storage) ?? 0
}

export function markSessionChanged(storage: object): void {
  sessionGenerations.set(storage, sessionGeneration(storage) + 1)
}

export function decodeJwtPayload(token: string | null): Record<string, unknown> | null {
  const encoded = token?.split('.')[1]
  if (!encoded) return null
  try {
    const normalized = encoded.replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(
      atob(normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')),
    ) as Record<string, unknown>
  } catch {
    return null
  }
}

export function createRefreshCoordinator(options: RefreshCoordinatorOptions): () => Promise<string> {
  let pending: {
    generation: number
    refreshToken: string
    identity?: string
    promise: Promise<string>
  } | undefined

  return () => {
    const refreshToken = options.storage.getItem(options.refreshTokenKey)
    if (!refreshToken) return Promise.reject(new Error('refresh token missing'))
    const generation = sessionGeneration(options.storage)
    const identity = options.identity?.()
    if (pending?.generation === generation &&
      pending.refreshToken === refreshToken && pending.identity === identity) {
      return pending.promise
    }

    let promise!: Promise<string>
    promise = options.request(refreshToken)
      .then((session) => {
        if (sessionGeneration(options.storage) !== generation ||
          options.storage.getItem(options.refreshTokenKey) !== refreshToken ||
          options.identity?.() !== identity) {
          throw options.supersededError()
        }
        if (!session.token || !options.validateAccessToken(session.token)) {
          throw options.invalidAccessTokenError?.() ?? new Error('access token missing')
        }
        options.storage.setItem(options.accessTokenKey, session.token)
        if (session.refresh_token) {
          options.storage.setItem(options.refreshTokenKey, session.refresh_token)
        } else if (options.clearMissingRefreshToken) {
          options.storage.removeItem(options.refreshTokenKey)
        }
        options.onCommitted?.(session)
        return session.token
      })
      .finally(() => {
        if (pending?.promise === promise) pending = undefined
      })
    pending = { generation, refreshToken, identity, promise }
    return promise
  }
}
