import { apiClient } from './client';
import { isLearnerSessionToken, TOKEN_KEY } from './authSession';

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

  if (!isLearnerSessionToken(response.data.token)) {
    throw new Error('请使用学员账号登录');
  }
  localStorage.setItem(TOKEN_KEY, response.data.token);
  return response.data;
}

export function logout(): void {
  localStorage.removeItem(TOKEN_KEY);
}

export function isAuthenticated(): boolean {
  const token = localStorage.getItem(TOKEN_KEY);
  const authenticated = isLearnerSessionToken(token);
  if (!authenticated && token) {
    localStorage.removeItem(TOKEN_KEY);
  }
  return authenticated;
}
