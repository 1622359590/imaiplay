interface UploadedResource {
  id: string
}

interface ResourceUploadEffects {
  notifySuccess: () => void
  refreshList: () => void
}

export function completeResourceUpload(
  resource: UploadedResource | undefined,
  effects: ResourceUploadEffects,
): undefined {
  if (resource) {
    effects.notifySuccess()
    effects.refreshList()
  }
  return undefined
}
