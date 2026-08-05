import { apiClient } from './client';
import { requestSessionPortal } from './portalSession';

export interface Portal {
  tenant_id: string;
  code: string;
  name: string;
  logo_url: string;
  primary_color: string;
  welcome_text: string;
  browser_title?: string;
  default_portal_url: string;
  custom_domain_url?: string;
}

export async function resolvePortal(code?: string): Promise<Portal> {
  const response = await apiClient.get<Portal>(
    '/api/v1/portal',
    code ? { params: { tenant_code: code } } : undefined,
  );
  return response.data;
}

export function resolveSessionPortal(): Promise<Portal> {
  return requestSessionPortal(async (path) => {
    const response = await apiClient.get<Portal>(path);
    return response.data;
  });
}
