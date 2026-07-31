export const TOKEN_KEY = 'imaiplay_token';
export const SESSION_EXPIRED_EVENT = 'imaiplay:session-expired';

interface TokenPayload {
  exp?: number;
  role?: string;
}

interface TokenStorage {
  removeItem: (key: string) => void;
}

interface SessionEventTarget {
  dispatchEvent: (event: Event) => unknown;
}

function decodePayload(token: string): TokenPayload | null {
  const payload = token.split('.')[1];
  if (!payload) return null;

  try {
    return JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/'))) as TokenPayload;
  } catch {
    return null;
  }
}

export function isLearnerSessionToken(token: string | null, now = Date.now()): boolean {
  if (!token) return false;

  const payload = decodePayload(token);
  return payload?.role === 'learner' &&
    typeof payload.exp === 'number' &&
    payload.exp > Math.floor(now / 1000);
}

export function clearAuthSession(
  storage: TokenStorage = localStorage,
  eventTarget: SessionEventTarget = window,
): void {
  storage.removeItem(TOKEN_KEY);
  eventTarget.dispatchEvent(new Event(SESSION_EXPIRED_EVENT));
}
