export const TOKEN_KEY = 'imaiplay_token'
export const REFRESH_TOKEN_KEY = 'imaiplay_refresh_token'
export const AUTH_SESSION_EXPIRED_EVENT = 'imaiplay:admin-session-expired'

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

export function createSessionRefresher(
  request: (refreshToken: string) => Promise<RefreshedSession>,
  storage: SessionStorage = localStorage,
) {
  let pending: Promise<string> | null = null

  return () => {
    if (pending) return pending

    const refreshToken = storage.getItem(REFRESH_TOKEN_KEY)
    if (!refreshToken) return Promise.reject(new Error('refresh token missing'))

    pending = request(refreshToken)
      .then((session) => {
        if (!session.token) throw new Error('access token missing')
        storage.setItem(TOKEN_KEY, session.token)
        if (session.refresh_token) {
          storage.setItem(REFRESH_TOKEN_KEY, session.refresh_token)
        }
        return session.token
      })
      .finally(() => {
        pending = null
      })
    return pending
  }
}

export function clearAuthSession(
  storage: Pick<SessionStorage, 'removeItem'> = localStorage,
  eventTarget: SessionEventTarget = window,
) {
  storage.removeItem(TOKEN_KEY)
  storage.removeItem(REFRESH_TOKEN_KEY)
  eventTarget.dispatchEvent(new Event(AUTH_SESSION_EXPIRED_EVENT))
}
