import type { DomainBindState, DomainBindStatus } from '../api/domain'

export function shouldGuardDomainBinding(state?: DomainBindState) {
  return state === 'creating_site' || state === 'configuring'
}

export function domainBindingProgress(
  status?: Pick<DomainBindStatus, 'current_step' | 'total_steps'>,
) {
  const total = Math.max(1, status?.total_steps || 5)
  const current = Math.max(0, Math.min(total, status?.current_step || 0))
  return Math.round((current / total) * 100)
}
