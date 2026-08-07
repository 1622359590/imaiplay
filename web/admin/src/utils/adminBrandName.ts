export function resolveAdminBrandName(brandName: unknown, tenantName: unknown): string {
  const configured = typeof brandName === 'string' ? brandName.trim() : ''
  if (configured) return configured
  const tenant = typeof tenantName === 'string' ? tenantName.trim() : ''
  return tenant || 'ImaiPlay'
}
