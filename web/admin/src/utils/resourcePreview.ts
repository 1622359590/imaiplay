import type { Resource } from '../api/resource'

export interface PreviewURLFactory {
  createObjectURL: (blob: Blob) => string
  revokeObjectURL: (url: string) => void
}

export interface ResourcePreview {
  name: string
  resourceType: Resource['resource_type']
  url: string
  dispose: () => void
}

export async function loadResourcePreview(
  resource: Pick<Resource, 'id' | 'name' | 'resource_type'>,
  loadFile: (id: string) => Promise<Blob>,
  urlFactory: PreviewURLFactory = URL,
): Promise<ResourcePreview> {
  const blob = await loadFile(resource.id)
  const url = urlFactory.createObjectURL(blob)
  return {
    name: resource.name,
    resourceType: resource.resource_type,
    url,
    dispose: () => urlFactory.revokeObjectURL(url),
  }
}
