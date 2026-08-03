import type { DomainBindStatus } from '../api/domain'

export function mergeDomainStatus(
  current: DomainBindStatus | undefined,
  next: DomainBindStatus,
): DomainBindStatus {
  return {
    ...current,
    ...next,
    tenant_code: next.tenant_code || current?.tenant_code,
    default_portal_url: next.default_portal_url || current?.default_portal_url,
  }
}

export function portalURLAfterRegistration(
  status: Pick<DomainBindStatus, 'default_portal_url'> | undefined,
  tenantCode: string,
): string {
  return status?.default_portal_url ||
    `https://play.imai.work/t/${encodeURIComponent(tenantCode)}`
}
