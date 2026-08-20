export { responseMessage, responseStatus } from './api/errors.ts'
export {
  createRefreshCoordinator,
  decodeJwtPayload,
  markSessionChanged,
} from './auth/sessionCore.ts'
export type {
  RefreshCoordinatorOptions,
  RefreshSession,
  StorageLike,
} from './auth/sessionCore.ts'
export {
  contrastRatio,
  deriveClayColors,
  normalizePrimaryColor,
  normalizeSelectionColors,
  recommendedSelectionColors,
} from './theme/tenantTheme.ts'
export {
  lessonContentLabel,
  resolveLessonContent,
} from './learning/lessonContent.ts'
export type {
  LessonContentType,
  ResolvedLessonContent,
} from './learning/lessonContent.ts'
export {
  acknowledgeAndContinue,
  formatLearningDuration,
  motivationTargetPath,
  normalizeLearnerMotivation,
} from './learning/learnerMotivation.ts'
export type {
  LearnerMotivation,
  LearnerMotivationComparison,
  LearnerMotivationCourse,
  LearnerMotivationKind,
  LearnerMotivationMetrics,
} from './learning/learnerMotivation.ts'
export {
  PlaybackLifecycleController,
  WatchHeartbeat,
  restorePlaybackPosition,
} from './learning/watchHeartbeat.ts'
export type {
  PlaybackProgressReport,
  PlaybackSnapshot,
  WatchHeartbeatPayload,
} from './learning/watchHeartbeat.ts'
export type { ApiEnvelope, PageResult } from './types/api.ts'
export type {
  TenantPortalContract,
  TenantSelectionColors,
  TenantThemeContract,
} from './types/theme.ts'
