import { apiClient } from './client';
import { requestSessionPortal } from './portalSession';
import type { TenantPortalContract } from '@imaiplay/shared/types/theme';

export interface Portal extends TenantPortalContract {
  logo_url: string;
  welcome_text: string;
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
