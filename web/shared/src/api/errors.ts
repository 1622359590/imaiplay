type ResponseLike = {
  response?: {
    status?: unknown
    data?: { message?: unknown }
  }
}

export function responseStatus(error: unknown): number | undefined {
  const value = (error as ResponseLike | null)?.response?.status
  return typeof value === 'number' ? value : undefined
}

export function responseMessage(error: unknown): string | undefined {
  const value = (error as ResponseLike | null)?.response?.data?.message
  return typeof value === 'string' && value.trim() ? value : undefined
}
