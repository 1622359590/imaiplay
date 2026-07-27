import { apiClient, unwrap, type ApiEnvelope } from './client'

export interface TenantTheme { primary_color: string; logo_url?: string; welcome_text?: string }

export async function getTenantTheme(): Promise<TenantTheme> {
  const response = await apiClient.get<ApiEnvelope<TenantTheme>>('/api/v1/theme')
  return unwrap(response)
}

export function notifyThemeChanged() { window.dispatchEvent(new Event('tenant-theme-changed')) }
