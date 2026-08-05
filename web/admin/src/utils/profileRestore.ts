export type ProfileRestoreOutcome = 'unauthorized' | 'retry'

export function handleProfileRestoreFailure(
  status: number | undefined,
  clearPersistentSession: () => void,
): ProfileRestoreOutcome {
  if (status === 401 || status === 403) {
    clearPersistentSession()
    return 'unauthorized'
  }
  return 'retry'
}
