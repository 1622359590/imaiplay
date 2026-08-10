export function resolveAdminBrandName(
  welcomeText: unknown,
  brandName: unknown,
  tenantName: unknown,
): string {
  const welcome = typeof welcomeText === 'string' ? welcomeText.trim() : ''
  if (welcome) return welcome
  const configured = typeof brandName === 'string' ? brandName.trim() : ''
  if (configured) return configured
  const tenant = typeof tenantName === 'string' ? tenantName.trim() : ''
  return tenant || 'ImaiPlay'
}
