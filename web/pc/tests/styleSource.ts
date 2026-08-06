import { readFileSync } from 'node:fs'

export function readStyleBundle(entry: URL, seen = new Set<string>()): string {
  if (seen.has(entry.href)) return ''
  seen.add(entry.href)
  return readFileSync(entry, 'utf8').replace(
    /@import\s+["']([^"']+)["'];?/g,
    (_, path: string) => readStyleBundle(new URL(path, entry), seen),
  )
}
