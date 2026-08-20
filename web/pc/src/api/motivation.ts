import {
  normalizeLearnerMotivation,
  type LearnerMotivation,
} from '@imaiplay/shared/learning/learnerMotivation'
import { apiClient } from './client'

export async function getLearnerMotivation(): Promise<LearnerMotivation> {
  const response = await apiClient.get<unknown>('/api/v1/learner/motivation')
  return normalizeLearnerMotivation(response.data)
}

export async function acknowledgeLearnerMotivation(promptKey: string): Promise<void> {
  await apiClient.post('/api/v1/learner/motivation/ack', { prompt_key: promptKey })
}
