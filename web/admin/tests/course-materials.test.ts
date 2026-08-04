import assert from 'node:assert/strict'
import test from 'node:test'
import {
  courseMaterialCollectionPath,
  courseMaterialItemPath,
  platformAttachmentUploadPath,
  tenantAttachmentUploadPath,
} from '../src/api/courseMaterialRoutes.ts'
import {
  swapMaterialOrder,
  validateCourseMaterialFile,
} from '../src/utils/courseMaterials.ts'

test('course material APIs use course-scoped management routes', () => {
  assert.equal(
    courseMaterialCollectionPath('course 1'),
    '/backend/v1/courses/course%201/materials',
  )
  assert.equal(
    courseMaterialItemPath('course 1', 'material/1'),
    '/backend/v1/courses/course%201/materials/material%2F1',
  )
})

test('validates course material extension and 200MB limit', () => {
  assert.equal(validateCourseMaterialFile({ name: 'guide.PDF', size: 1024 }), undefined)
  assert.equal(validateCourseMaterialFile({ name: 'video.mp4', size: 1024 }), '仅支持 PDF、Word、Excel、PowerPoint 和 ZIP 文件')
  assert.equal(validateCourseMaterialFile({ name: 'large.zip', size: 200 * 1024 * 1024 + 1 }), '单个资料不能超过 200MB')
})

test('swaps only adjacent material sort orders', () => {
  const result = swapMaterialOrder([
    { id: 'a', sort_order: 1 },
    { id: 'b', sort_order: 2 },
    { id: 'c', sort_order: 3 },
  ], 1, -1)
  assert.deepEqual(result, [
    { id: 'a', sort_order: 2 },
    { id: 'b', sort_order: 1 },
  ])
  assert.deepEqual(swapMaterialOrder([{ id: 'a', sort_order: 1 }], 0, -1), [])
})

test('attachment uploads use dedicated tenant and platform routes', () => {
  assert.equal(
    tenantAttachmentUploadPath,
    '/backend/v1/resources/attachments/upload',
  )
  assert.equal(
    platformAttachmentUploadPath,
    '/backend/v1/admin/resources/attachments/upload',
  )
})
