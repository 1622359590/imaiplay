import { apiClient, TENANT_KEY, TOKEN_KEY } from './client';

export interface LoginValues {
  tenantCode: string;
  email: string;
  password: string;
}

export interface LoginResult {
  token: string;
  expires_at?: string;
}

export async function login(values: LoginValues): Promise<LoginResult> {
  const tenantCode = values.tenantCode.trim();
  const response = await apiClient.post<LoginResult>(
    '/api/v1/auth/login',
    {
      email: values.email.trim(),
      password: values.password,
    },
    {
      headers: {
        'X-Tenant-Code': tenantCode,
      },
    },
  );

  localStorage.setItem(TOKEN_KEY, response.data.token);
  localStorage.setItem(TENANT_KEY, tenantCode);
  return response.data;
}

export function logout(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(TENANT_KEY);
}

export function isAuthenticated(): boolean {
  return Boolean(localStorage.getItem(TOKEN_KEY));
}
