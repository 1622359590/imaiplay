import {
  normalizeLearnerMotivation,
  type LearnerMotivation,
} from '@imaiplay/shared/learning/learnerMotivation'
import { apiClient, unwrap, type ApiEnvelope } from './client'

const motivationRequestConfig = { motivationSilent: true }

export async function getLearnerMotivation(): Promise<LearnerMotivation> {
  const response = await apiClient.get<ApiEnvelope<unknown>>('/api/v1/learner/motivation', motivationRequestConfig)
  return normalizeLearnerMotivation(unwrap(response))
}

export async function acknowledgeLearnerMotivation(promptKey: string): Promise<void> {
  const response = await apiClient.post<ApiEnvelope<unknown>>(
    '/api/v1/learner/motivation/ack',
    { prompt_key: promptKey },
    motivationRequestConfig,
  )
  unwrap(response)
}
