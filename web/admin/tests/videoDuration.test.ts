import assert from 'node:assert/strict'
import test from 'node:test'
import {
  readVideoDurationSeconds,
  type VideoDurationDependencies,
} from '../src/utils/videoDuration.ts'

function dependencies(
  duration: number,
  metadataError = false,
): VideoDurationDependencies {
  return {
    createObjectURL: () => 'blob:video',
    revokeObjectURL: () => undefined,
    createVideo: () => {
      const video = {
        duration,
        onloadedmetadata: null as ((event: Event) => unknown) | null,
        onerror: null as ((event: Event) => unknown) | null,
        preload: '',
        src: '',
      }
      queueMicrotask(() => {
        if (metadataError) video.onerror?.(new Event('error'))
        else video.onloadedmetadata?.(new Event('loadedmetadata'))
      })
      return video
    },
  }
}

test('rounds uploaded video duration up to a whole second', async () => {
  const duration = await readVideoDurationSeconds(
    { type: 'video/mp4' },
    dependencies(61.2),
  )

  assert.equal(duration, 62)
})

test('rejects when video metadata cannot be read', async () => {
  await assert.rejects(
    readVideoDurationSeconds(
      { type: 'video/mp4' },
      dependencies(0, true),
    ),
    /无法读取视频时长/,
  )
})
