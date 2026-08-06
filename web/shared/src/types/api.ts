export interface ApiEnvelope<T> {
  data: T
  code?: number
  message?: string
}

export interface PageResult<T> {
  items?: T[]
  list?: T[]
  data?: T[]
  total?: number
}
