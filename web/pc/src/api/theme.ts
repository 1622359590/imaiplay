import { apiClient } from './client';
import type { TenantThemeContract } from '@imaiplay/shared/types/theme';

export interface TenantTheme extends TenantThemeContract {}

export async function getTenantTheme(): Promise<TenantTheme> {
  const response = await apiClient.get<TenantTheme>('/api/v1/theme');
  return response.data;
}

export function notifyThemeChanged(): void {
  window.dispatchEvent(new Event('tenant-theme-changed'));
}
