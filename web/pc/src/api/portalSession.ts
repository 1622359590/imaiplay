const PORTAL_CODE_KEY = 'imaiplay_portal_code';
const PORTAL_SESSION_PATH = '/api/v1/portal/session';

let activePortalCode: string | undefined;
let activePortalTenantId: string | undefined;

function storage(): Storage | undefined {
  return typeof window === 'undefined' ? undefined : window.sessionStorage;
}

export function setActivePortalCode(code: string): void {
  activePortalCode = code;
  activePortalTenantId = undefined;
  storage()?.setItem(PORTAL_CODE_KEY, code);
}

export function setActivePortalIdentity(portal: { code: string; tenant_id: string }): void {
  setActivePortalCode(portal.code);
  activePortalTenantId = portal.tenant_id;
}

export function getActivePortalCode(): string | undefined {
  if (activePortalCode) return activePortalCode;
  const stored = storage()?.getItem(PORTAL_CODE_KEY) ?? undefined;
  activePortalCode = stored || undefined;
  return activePortalCode;
}

export function getActivePortalTenantId(): string | undefined {
  return activePortalTenantId;
}

export function clearActivePortalCode(): void {
  activePortalCode = undefined;
  activePortalTenantId = undefined;
  storage()?.removeItem(PORTAL_CODE_KEY);
}

export function requestSessionPortal<T>(
  request: (path: string) => Promise<T>,
): Promise<T> {
  return request(PORTAL_SESSION_PATH);
}
