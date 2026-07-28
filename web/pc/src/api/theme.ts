import { apiClient } from './client';

export interface TenantTheme {
  primary_color: string;
  logo_url?: string;
  welcome_text?: string;
}

export async function getTenantTheme(): Promise<TenantTheme> {
  const response = await apiClient.get<TenantTheme>('/api/v1/theme');
  return response.data;
}

export function notifyThemeChanged(): void {
  window.dispatchEvent(new Event('tenant-theme-changed'));
}
