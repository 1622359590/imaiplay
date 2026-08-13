import { afterEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './client';
import { forgotPassword, resetPassword } from './auth';

vi.mock('./client', () => ({
  apiClient: { post: vi.fn() },
}));

afterEach(() => {
  vi.restoreAllMocks();
});

describe('PC learner password recovery', () => {
  it('requests a recovery code for the normalized phone number', async () => {
    const request = vi.spyOn(apiClient, 'post').mockResolvedValueOnce({ data: undefined });

    await forgotPassword(' 13800138000 ');

    expect(request).toHaveBeenCalledWith(
      '/api/v1/auth/forgot-password',
      { phone: '13800138000' },
    );
  });

  it('submits the normalized code and new password to the reset endpoint', async () => {
    const request = vi.spyOn(apiClient, 'post').mockResolvedValueOnce({ data: undefined });

    await resetPassword(' 13800138000 ', ' 123456 ', 'new-password');

    expect(request).toHaveBeenCalledWith(
      '/api/v1/auth/reset-password',
      {
        phone: '13800138000',
        code: '123456',
        new_password: 'new-password',
      },
    );
  });
});
