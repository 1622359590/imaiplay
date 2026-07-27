import client from './client'

export interface TenantTheme { primary_color: string; logo_url?: string; welcome_text?: string }

export const themeApi = {
  get: () => client.get<TenantTheme>('/backend/v1/theme'),
  update: (data: TenantTheme) => client.put<TenantTheme>('/backend/v1/theme', data),
}
