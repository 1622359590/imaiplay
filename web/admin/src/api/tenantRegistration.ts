import client from './client'
export interface TenantRegistrationInput { organization_name: string; admin_email: string; admin_name: string; phone?: string; password: string; plan_id?: string }
export const tenantRegistrationApi = { create: (data: TenantRegistrationInput) => client.post('/backend/v1/admin/tenants', data).then((response) => response.data) }
