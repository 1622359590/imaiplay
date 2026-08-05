import { isValidPortalSession } from './authSession.ts'

const PLATFORM_HOST = 'play.imai.work'

export type PortalMode = 'platform' | 'default' | 'custom-domain'

export interface PortalLocation {
  tenantCode?: string
  mode: PortalMode
  shouldResolve: boolean
  key: string
}

export interface PortalErrorContent {
  title: string
  description: string
}

interface PortalApiError {
  response?: {
    status?: number
    data?: { error?: string; message?: string }
  }
}

function tenantCodeFromPath(pathname: string): string | undefined {
  const match = pathname.match(/^\/t\/([^/]+)(?:\/|$)/)
  if (!match) return undefined
  try {
    return decodeURIComponent(match[1]).trim().toLowerCase() || undefined
  } catch {
    return undefined
  }
}

export function resolvePortalLocation(
  pathname: string,
  hostname: string,
): PortalLocation {
  const normalizedHost = hostname.trim().toLowerCase().replace(/\.$/, '')
  if (normalizedHost !== PLATFORM_HOST) {
    return {
      tenantCode: undefined,
      mode: 'custom-domain',
      shouldResolve: true,
      key: `custom-domain:${normalizedHost}`,
    }
  }

  const tenantCode = tenantCodeFromPath(pathname)
  if (tenantCode) {
    return {
      tenantCode,
      mode: 'default',
      shouldResolve: true,
      key: `default:${tenantCode}`,
    }
  }
  return {
    tenantCode: undefined,
    mode: 'platform',
    shouldResolve: false,
    key: 'platform',
  }
}

export function shouldRestoreSessionPortal(
  location: PortalLocation,
  accessToken: string | null,
  now = Date.now(),
): boolean {
  return (
    location.mode === 'platform' &&
    isValidPortalSession(accessToken, undefined, now)
  )
}

export function portalErrorContent(error: unknown): PortalErrorContent {
  const apiError = error as PortalApiError
  const status = apiError?.response?.status
  const errorCode = apiError?.response?.data?.error
  if (errorCode === 'portal_suspended') {
    return {
      title: '租户已暂停',
      description: '该企业门户已暂停，请联系企业管理员',
    }
  }
  if (errorCode === 'portal_trial_expired') {
    return {
      title: '试用已到期',
      description: '该企业的试用期已结束，请联系企业管理员',
    }
  }
  if (errorCode === 'portal_not_found' || status === 404) {
    return {
      title: '门户不存在',
      description: '请确认门户地址是否正确，或联系企业管理员',
    }
  }
  return {
    title: '企业门户不可访问',
    description: '请稍后重试，或联系企业管理员',
  }
}
