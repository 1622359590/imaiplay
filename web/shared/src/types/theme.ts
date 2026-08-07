export interface TenantThemeContract {
  primary_color: string
  selected_background_color?: string
  selected_text_color?: string
  selected_icon_color?: string
  logo_url?: string
  welcome_text?: string
  browser_title?: string
}

export interface TenantSelectionColors {
  selected_background_color: string
  selected_text_color: string
  selected_icon_color: string
}

export interface TenantPortalContract extends TenantThemeContract {
  tenant_id: string
  code: string
  name: string
  default_portal_url: string
  custom_domain_url?: string
}
