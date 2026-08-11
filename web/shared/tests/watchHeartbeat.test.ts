import assert from 'node:assert/strict'
import test from 'node:test'

import { PlaybackLifecycleController } from '../src/learning/watchHeartbeat.ts'

test('shared heartbeat keeps one session id and retries the same report', async () => {
  let now = 0
  const attempts: unknown[] = []
  const controller = new PlaybackLifecycleController({
    now: () => now,
    read: () => ({ positionSeconds: 15, durationSeconds: 100 }),
    report: async (report) => {
      attempts.push(report)
      if (attempts.length === 1) throw new Error('offline')
    },
    reportIDFactory: () => 'report-1',
    sessionIDFactory: () => 'session-1',
  })

  controller.playing()
  now = 15_000
  await controller.periodicFlush()
  now = 30_000
  await controller.periodicFlush()

  assert.equal(attempts.length, 2)
  assert.deepEqual(attempts[1], attempts[0])
  assert.deepEqual(attempts[1], {
    positionSeconds: 15,
    progressPercent: 15,
    heartbeat: {
      watched_seconds_delta: 15,
      report_id: 'report-1',
      session_id: 'session-1',
    },
  })
})
