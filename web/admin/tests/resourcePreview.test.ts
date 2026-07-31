import assert from 'node:assert/strict'
import test from 'node:test'
import type { Resource } from '../src/api/resource.ts'
import { loadResourcePreview } from '../src/utils/resourcePreview.ts'

const videoResource: Resource = {
  id: 'resource-1',
  name: 'lesson.mp4',
  resource_type: 'video',
  url: '',
  size_bytes: 1024,
  created_at: '2026-07-31T00:00:00Z',
}

test('loads an authenticated resource into a disposable preview URL', async () => {
  const blob = new Blob(['video'], { type: 'video/mp4' })
  const revoked: string[] = []

  const preview = await loadResourcePreview(
    videoResource,
    async (id) => {
      if (id !== 'resource-1') throw new Error(`unexpected id ${id}`)
      return blob
    },
    {
      createObjectURL: (value) => {
        assert.equal(value, blob)
        return 'blob:preview'
      },
      revokeObjectURL: (url) => revoked.push(url),
    },
  )

  assert.deepEqual(
    {
      name: preview.name,
      resourceType: preview.resourceType,
      url: preview.url,
    },
    {
      name: 'lesson.mp4',
      resourceType: 'video',
      url: 'blob:preview',
    },
  )

  preview.dispose()
  assert.deepEqual(revoked, ['blob:preview'])
})
