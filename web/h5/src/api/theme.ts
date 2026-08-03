import { apiClient, unwrap, type ApiEnvelope } from './client'

export interface TenantPortal {
  tenant_id: string
  code: string
  name: string
  primary_color: string
  logo_url?: string
  welcome_text?: string
  default_portal_url: string
  custom_domain_url?: string
}

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
