import client from './client'

export interface AuditLog {
  id: string
  tenant_id: string
  user_id: string
  user_email: string
  user_role: string
  action: string
  resource_type: string
  resource_id: string
  detail: string
  ip: string
  request_id: string
  created_at: string
}

export interface AuditQuery { offset: number; limit: number; action?: string; user_id?: string; tenant_id?: string; from?: string; to?: string }

export const auditApi = {
  list: (query: AuditQuery, superadmin: boolean) => client.get<{ items: AuditLog[]; total: number }>(superadmin ? '/backend/v1/admin/audit-logs' : '/backend/v1/audit-logs', { params: query }),
}
