export function consumeOneShotAction(search: string, key: string) {
  const params = new URLSearchParams(search)
  const active = params.get(key) === '1'
  if (active) params.delete(key)
  const remaining = params.toString()
  return {
    active,
    remainingSearch: remaining ? `?${remaining}` : '',
  }
}
