import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('./client', () => ({
  apiClient: { get: vi.fn(), post: vi.fn() },
  unwrap: vi.fn(),
}))

import { apiClient, unwrap } from './client'
import { acknowledgeLearnerMotivation, getLearnerMotivation } from './motivation'

beforeEach(() => vi.resetAllMocks())

describe('H5 learner motivation API', () => {
  it('gets and normalizes the server prompt', async () => {
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: { code: 0, message: 'ok', data: { kind: 'none' } },
    })
    vi.mocked(unwrap).mockReturnValueOnce({ kind: 'none' })
    await expect(getLearnerMotivation()).resolves.toEqual({ kind: 'none' })
    expect(apiClient.get).toHaveBeenCalledWith('/api/v1/learner/motivation')
  })

  it('acknowledges only the opaque prompt key', async () => {
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { code: 0, message: 'ok', data: {} },
    })
    vi.mocked(unwrap).mockReturnValueOnce({})
    await acknowledgeLearnerMotivation('prompt-key')
    expect(apiClient.post).toHaveBeenCalledWith('/api/v1/learner/motivation/ack', {
      prompt_key: 'prompt-key',
    })
  })
})
