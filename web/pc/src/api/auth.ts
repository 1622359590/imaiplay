import { apiClient, TOKEN_KEY } from './client';

export interface LoginValues {
  identifier: string;
  password: string;
}

export interface LoginResult {
  token: string;
  expires_at?: string;
}

export async function login(values: LoginValues): Promise<LoginResult> {
  const response = await apiClient.post<LoginResult>(
    '/api/v1/auth/login',
    {
      identifier: values.identifier.trim(),
      password: values.password,
    },
  );

  localStorage.setItem(TOKEN_KEY, response.data.token);
  return response.data;
}

export function logout(): void {
  localStorage.removeItem(TOKEN_KEY);
}

export function isAuthenticated(): boolean {
  return Boolean(localStorage.getItem(TOKEN_KEY));
}
