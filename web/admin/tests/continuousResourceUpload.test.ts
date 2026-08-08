import assert from 'node:assert/strict'
import test from 'node:test'
import { completeResourceUpload } from '../src/utils/resourceUploadFlow.ts'

test('a completed resource upload reports success and refreshes the list without retaining a selection', () => {
  const effects: string[] = []
  const retained = completeResourceUpload(
    { id: 'resource-1' },
    {
      notifySuccess: () => effects.push('success'),
      refreshList: () => effects.push('refresh'),
    },
  )

  assert.equal(retained, undefined)
  assert.deepEqual(effects, ['success', 'refresh'])
})

test('clearing the uploader does not report success or refresh the list', () => {
  const effects: string[] = []
  const retained = completeResourceUpload(
    undefined,
    {
      notifySuccess: () => effects.push('success'),
      refreshList: () => effects.push('refresh'),
    },
  )

  assert.equal(retained, undefined)
  assert.deepEqual(effects, [])
})
