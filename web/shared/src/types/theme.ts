export interface TenantThemeContract {
  primary_color: string
  brand_name?: string
  logo_url?: string
  welcome_text?: string
  browser_title?: string
}

export interface TenantPortalContract extends TenantThemeContract {
  tenant_id: string
  code: string
  name: string
  default_portal_url: string
  custom_domain_url?: string
}
