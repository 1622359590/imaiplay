import client from './client'

export interface SMSConfig { provider: string; access_key_id: string; sign_name: string; template_code: string }
export const smsApi = {
  get: () => client.get<SMSConfig>('/backend/v1/sms-config'),
  save: (data: SMSConfig & { access_key_secret?: string }) => client.put<SMSConfig>('/backend/v1/sms-config', data),
  test: (phone: string) => client.post('/backend/v1/sms-config/test', { phone }),
}
