const normalizeModuleId = (id: string) => id.replaceAll('\\', '/')

const includesPackage = (id: string, packageName: string) =>
  id.includes(`/node_modules/${packageName}/`)

export function vendorChunkFor(id: string): string | undefined {
  const normalizedId = normalizeModuleId(id)

  if (!normalizedId.includes('/node_modules/')) return undefined

  if (includesPackage(normalizedId, 'react') || includesPackage(normalizedId, 'react-dom')) {
    return 'react-vendor'
  }

  if (
    includesPackage(normalizedId, 'react-router') ||
    includesPackage(normalizedId, 'react-router-dom')
  ) {
    return 'router-vendor'
  }

  if (
    includesPackage(normalizedId, '@ant-design/icons') ||
    includesPackage(normalizedId, '@ant-design/icons-svg')
  ) {
    return 'antd-icons'
  }

  if (
    includesPackage(normalizedId, '@rc-component') ||
    /\/node_modules\/rc-[^/]+\//.test(normalizedId)
  ) {
    return 'antd-primitives'
  }

  if (includesPackage(normalizedId, 'antd') || includesPackage(normalizedId, '@ant-design')) {
    return 'antd-framework'
  }

  if (includesPackage(normalizedId, '@reduxjs') || includesPackage(normalizedId, 'react-redux')) {
    return 'state-vendor'
  }

  if (includesPackage(normalizedId, 'axios')) return 'transport-vendor'

  return undefined
}

export function manualChunks(id: string): string | undefined {
  return vendorChunkFor(id)
}
