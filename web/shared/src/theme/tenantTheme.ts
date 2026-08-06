export function normalizePrimaryColor(value: unknown, fallback: string): string {
  const color = typeof value === 'string' ? value.trim() : ''
  return /^#[0-9a-f]{6}$/i.test(color) ? color.toUpperCase() : fallback
}
