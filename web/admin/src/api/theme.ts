import client from './client'
import type { TenantThemeContract } from '@imaiplay/shared/types/theme'

export interface TenantTheme extends TenantThemeContract {}

export const themeApi = {
  get: () => client.get<TenantTheme>('/backend/v1/theme'),
  update: (data: TenantTheme) => client.put<TenantTheme>('/backend/v1/theme', data),
}
