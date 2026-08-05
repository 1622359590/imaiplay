import { lazy, type ComponentType, type LazyExoticComponent } from 'react'

const RELOAD_MARKER = 'imaiplay:admin:chunk-reload'

export type ReloadRuntime = {
  getReloadMarker: () => string | null
  setReloadMarker: () => void
  clearReloadMarker: () => void
  reload: () => void
}

function isChunkLoadError(error: unknown): boolean {
  if (!(error instanceof Error)) return false

  const message = error.message.toLowerCase()
  return error.name === 'ChunkLoadError'
    || message.includes('failed to fetch dynamically imported module')
    || message.includes('error loading dynamically imported module')
    || message.includes('importing a module script failed')
    || message.includes('loading chunk')
}

const browserRuntime: ReloadRuntime = {
  getReloadMarker: () => window.sessionStorage.getItem(RELOAD_MARKER),
  setReloadMarker: () => window.sessionStorage.setItem(RELOAD_MARKER, '1'),
  clearReloadMarker: () => window.sessionStorage.removeItem(RELOAD_MARKER),
  reload: () => window.location.reload(),
}

export async function loadWithReload<T>(
  importer: () => Promise<T>,
  runtime: ReloadRuntime = browserRuntime,
): Promise<T> {
  try {
    const importedModule = await importer()
    runtime.clearReloadMarker()
    return importedModule
  } catch (error) {
    if (!isChunkLoadError(error)) throw error

    if (runtime.getReloadMarker()) {
      runtime.clearReloadMarker()
      throw error
    }

    runtime.setReloadMarker()
    runtime.reload()
    return new Promise<never>(() => {})
  }
}

export function lazyWithReload<T extends ComponentType<unknown>>(
  importer: () => Promise<{ default: T }>,
): LazyExoticComponent<T> {
  return lazy(() => loadWithReload(importer))
}
