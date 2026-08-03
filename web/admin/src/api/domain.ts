import client from './client'

export type DomainBindState =
  | 'none'
  | 'pending_verification'
  | 'verified'
  | 'creating_site'
  | 'configuring'
  | 'ready'
  | 'verification_failed'
  | 'setup_failed'

export interface DomainBindStatus {
  state: DomainBindState
  domain?: string
  message?: string
  current_step: number
  total_steps: number
  cname_target: string
  tenant_code?: string
  default_portal_url?: string
  updated_at?: string
}

export const domainApi = {
  verify: (domain: string) => client.post<DomainBindStatus>('/backend/v1/domain-bind/verify', { domain }).then((response) => response.data),
  bind: (domain: string) => client.post<DomainBindStatus>('/backend/v1/domain-bind', { domain }).then((response) => response.data),
  status: () => client.get<DomainBindStatus>('/backend/v1/domain-bind/status').then((response) => response.data),
  unbind: () => client.delete<DomainBindStatus>('/backend/v1/domain-bind').then((response) => response.data),
}

export const tenantDomainApi = {
  verify: (tenantID: string, domain: string) => client.post<DomainBindStatus>(`/backend/v1/tenants/${tenantID}/domain-bind/verify`, { domain }).then((response) => response.data),
  bind: (tenantID: string, domain: string) => client.post<DomainBindStatus>(`/backend/v1/tenants/${tenantID}/domain-bind`, { domain }).then((response) => response.data),
  status: (tenantID: string) => client.get<DomainBindStatus>(`/backend/v1/tenants/${tenantID}/domain-bind/status`).then((response) => response.data),
  unbind: (tenantID: string) => client.delete<DomainBindStatus>(`/backend/v1/tenants/${tenantID}/domain-bind`).then((response) => response.data),
}
