export const courseMaterialCollectionPath = (courseId: string) =>
  `/backend/v1/courses/${encodeURIComponent(courseId)}/materials`

export const courseMaterialItemPath = (courseId: string, materialId: string) =>
  `${courseMaterialCollectionPath(courseId)}/${encodeURIComponent(materialId)}`

export const tenantAttachmentUploadPath =
  '/backend/v1/resources/attachments/upload'

export const platformAttachmentUploadPath =
  '/backend/v1/admin/resources/attachments/upload'
