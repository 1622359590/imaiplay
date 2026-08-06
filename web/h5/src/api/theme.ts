import { apiClient, unwrap, type ApiEnvelope } from './client'
import type { TenantPortalContract } from '@imaiplay/shared/types/theme'

export interface TenantPortal extends TenantPortalContract {}

export async function getTenantPortal(tenantCode?: string): Promise<TenantPortal> {
  const response = await apiClient.get<ApiEnvelope<TenantPortal>>(
    '/api/v1/portal',
    tenantCode ? { params: { tenant_code: tenantCode } } : undefined,
  )
  return unwrap(response)
}

export async function getSessionTenantPortal(): Promise<TenantPortal> {
  const response = await apiClient.get<ApiEnvelope<TenantPortal>>(
    '/api/v1/portal/session',
  )
  return unwrap(response)
}

export function notifyThemeChanged() {
  window.dispatchEvent(new Event('tenant-theme-changed'))
}
