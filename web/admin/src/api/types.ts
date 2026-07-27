export interface PageResult<T> {
  items?: T[]
  list?: T[]
  data?: T[]
  total?: number
}

export interface ListParams {
  page?: number
  page_size?: number
  keyword?: string
}

export function normalizePage<T>(value: PageResult<T> | T[]) {
  if (Array.isArray(value)) return { items: value, total: value.length }
  const items = value.items || value.list || value.data || []
  return { items, total: value.total ?? items.length }
}
