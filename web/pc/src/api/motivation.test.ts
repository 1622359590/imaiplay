import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from './client'
import { acknowledgeLearnerMotivation, getLearnerMotivation } from './motivation'

vi.mock('./client', () => ({ apiClient: { get: vi.fn(), post: vi.fn() } }))

afterEach(() => vi.restoreAllMocks())

describe('PC learner motivation API', () => {
  it('gets and normalizes the server prompt', async () => {
    vi.spyOn(apiClient, 'get').mockResolvedValueOnce({ data: { kind: 'none' } })
    await expect(getLearnerMotivation()).resolves.toEqual({ kind: 'none' })
    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/learner/motivation', {
      motivationSilent: true,
    })
  })

  it('acknowledges only the opaque prompt key', async () => {
    vi.spyOn(apiClient, 'post').mockResolvedValueOnce({ data: {} })
    await acknowledgeLearnerMotivation('prompt-key')
    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/learner/motivation/ack', {
      prompt_key: 'prompt-key',
    }, { motivationSilent: true })
  })
})
