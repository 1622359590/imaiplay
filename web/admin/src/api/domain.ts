import client from './client'
export const domainApi = {
  updateMine: (custom_domain: string) => client.put('/backend/v1/tenant/custom-domain', { custom_domain }).then((response) => response.data),
  update: (id: string, custom_domain: string) => client.put(`/backend/v1/tenants/${id}/custom-domain`, { custom_domain }).then((response) => response.data),
}
