const allowedExtensions = new Set([
  '.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.zip',
])

const maxMaterialBytes = 200 * 1024 * 1024

export function validateCourseMaterialFile(file: Pick<File, 'name' | 'size'>) {
  const extension = file.name.slice(file.name.lastIndexOf('.')).toLowerCase()
  if (!allowedExtensions.has(extension)) {
    return '仅支持 PDF、Word、Excel、PowerPoint 和 ZIP 文件'
  }
  if (file.size > maxMaterialBytes) return '单个资料不能超过 200MB'
  return undefined
}

export function swapMaterialOrder<T extends { id: string; sort_order: number }>(
  items: T[],
  index: number,
  direction: -1 | 1,
): Array<Pick<T, 'id' | 'sort_order'>> {
  const target = index + direction
  if (index < 0 || target < 0 || index >= items.length || target >= items.length) {
    return []
  }
  return [
    { id: items[target].id, sort_order: items[index].sort_order },
    { id: items[index].id, sort_order: items[target].sort_order },
  ]
}
