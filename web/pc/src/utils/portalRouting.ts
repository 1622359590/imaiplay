const PLATFORM_HOST = 'play.imai.work';

export type PortalMode = 'platform' | 'default' | 'custom-domain';

export interface PortalLocation {
  tenantCode?: string;
  mode: PortalMode;
  shouldResolvePortal: boolean;
  resolutionKey: string;
}

export interface PortalSnapshot<T> {
  resolutionKey: string;
  portal?: T;
  loading: boolean;
  error?: unknown;
}

export type LegacySessionKind = 'learner' | 'staff' | 'none';

export type LegacyPortalRedirect =
  | { action: 'route'; target: string }
  | { action: 'document'; target: string }
  | { action: 'restore'; childPath: string };

export interface PortalErrorContent {
  title: string;
  description: string;
}

interface PortalApiError {
  response?: {
    status?: number;
    data?: { error?: string };
  };
}

export function tenantCodeFromPath(pathname: string): string | undefined {
  const match = pathname.match(/^\/t\/([^/]+)(?:\/|$)/);
  if (!match) return undefined;

  try {
    return decodeURIComponent(match[1]).toLowerCase();
  } catch {
    return undefined;
  }
}

export function portalPath(code: string, child = '/'): string {
  const suffix = child === '/' ? '' : `/${child.replace(/^\/+/, '')}`;
  return `/t/${encodeURIComponent(code)}${suffix}`;
}

export function portalRoutePath(
  mode: PortalMode,
  tenantCode: string | undefined,
  childPath: string,
): string {
  return mode === 'default' && tenantCode
    ? portalPath(tenantCode, childPath)
    : childPath;
}

export function portalLoginDestination(
  role: string,
  tenantCode: string,
  mode: PortalMode,
): string {
  if (role !== 'learner') return '/admin/';
  return mode === 'custom-domain' ? '/' : portalPath(tenantCode);
}

export function authenticatedLoginDestination(
  role: string,
  tenantCode: string | undefined,
  mode: PortalMode,
  explicitTenantCode?: string,
): string {
  if (mode === 'custom-domain' && role !== 'learner') {
    throw new Error('管理人员请前往平台管理后台登录');
  }
  if (role === 'superadmin') return '/admin/';
  if (!tenantCode) throw new Error('登录响应中的会话或企业信息无效');
  if (explicitTenantCode && tenantCode.toLowerCase() !== explicitTenantCode.toLowerCase()) {
    throw new Error('登录账号不属于当前企业');
  }
  return portalLoginDestination(role, tenantCode, mode);
}

export function performLoginNavigation(
  destination: string,
  routerNavigate: (destination: string, options: { replace: boolean }) => void,
  documentNavigate: (destination: string) => void = (target) => window.location.assign(target),
): void {
  if (destination.startsWith('/admin/')) {
    documentNavigate(destination);
    return;
  }
  routerNavigate(destination, { replace: true });
}

export function boundPortalLoginPath(tenantCode?: string): string | undefined {
  const code = tenantCode?.trim().toLowerCase();
  return code ? portalPath(code) : undefined;
}

export function portalLocation(pathname: string, hostname: string): PortalLocation {
  const normalizedHost = hostname.toLowerCase().replace(/\.$/, '');
  if (normalizedHost !== PLATFORM_HOST) {
    return {
      tenantCode: undefined,
      mode: 'custom-domain',
      shouldResolvePortal: true,
      resolutionKey: `custom-domain:${normalizedHost}`,
    };
  }

  const tenantCode = tenantCodeFromPath(pathname);
  if (tenantCode) {
    return {
      tenantCode,
      mode: 'default',
      shouldResolvePortal: true,
      resolutionKey: `default:${tenantCode}`,
    };
  }

  return {
    tenantCode: undefined,
    mode: 'platform',
    shouldResolvePortal: false,
    resolutionKey: 'platform',
  };
}

export function portalSnapshotForResolution<T>(
  resolutionKey: string,
  shouldResolvePortal: boolean,
  snapshot: PortalSnapshot<T>,
): PortalSnapshot<T> {
  return snapshot.resolutionKey === resolutionKey
    ? snapshot
    : { resolutionKey, loading: shouldResolvePortal };
}

export function legacyPortalRedirect(
  pathname: string,
  mode: PortalMode,
  sessionKind: LegacySessionKind = 'none',
): LegacyPortalRedirect {
  const childPath = pathname.replace(/^\/pc(?=\/|$)/, '') || '/';
  if (mode !== 'platform') {
    return { action: 'route', target: childPath };
  }
  if (sessionKind === 'staff') {
    return { action: 'document', target: '/admin/' };
  }
  if (sessionKind === 'learner') {
    return { action: 'restore', childPath };
  }
  return { action: 'route', target: '/login' };
}

export function restoredLegacyPortalTarget(
  tenantCode: string,
  childPath: string,
): string {
  return portalPath(tenantCode.trim().toLowerCase(), childPath);
}

export function portalErrorContent(error: unknown): PortalErrorContent {
  const apiError = error as PortalApiError;
  const status = apiError?.response?.status;
  const errorCode = apiError?.response?.data?.error;
  if (errorCode === 'portal_suspended') {
    return {
      title: '租户已暂停',
      description: '该企业门户已暂停，请联系企业管理员',
    };
  }
  if (errorCode === 'portal_trial_expired') {
    return {
      title: '试用已到期',
      description: '该企业的试用期已结束，请联系企业管理员',
    };
  }
  if (errorCode === 'portal_not_found' || status === 404) {
    return {
      title: '门户不存在',
      description: '请确认门户地址是否正确，或联系企业管理员',
    };
  }
  return {
    title: '企业门户不可访问',
    description: '请稍后重试，或联系企业管理员',
  };
}
