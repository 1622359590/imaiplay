import type { LearnerCourse } from '../api/learner';

export type LearnerCourseTab =
  | 'all'
  | 'required'
  | 'optional'
  | 'completed'
  | 'incomplete';

export interface LearnerCourseFilter {
  tab: LearnerCourseTab;
  categoryId?: string;
}

export type LearnerCourseStatus = 'completed' | 'incomplete';

export function courseStatus(course: LearnerCourse): LearnerCourseStatus {
  return course.lessonCount > 0 && course.completedLessonCount >= course.lessonCount
    ? 'completed'
    : 'incomplete';
}

export function filterLearnerCourses(
  courses: readonly LearnerCourse[],
  filter: LearnerCourseFilter,
): LearnerCourse[] {
  return courses.filter((course) => {
    if (filter.categoryId && course.category?.id !== filter.categoryId) return false;
    switch (filter.tab) {
      case 'required':
      case 'optional':
        return course.assignmentType === filter.tab;
      case 'completed':
      case 'incomplete':
        return courseStatus(course) === filter.tab;
      case 'all':
        return true;
    }
  });
}

export function learningMinutes(seconds: number): number {
  if (!Number.isFinite(seconds)) return 0;
  return Math.floor(Math.max(0, seconds) / 60);
}

export function formatPlaybackPosition(seconds: number): string {
  const normalized = Number.isFinite(seconds) ? Math.floor(Math.max(0, seconds)) : 0;
  const minutes = Math.floor(normalized / 60);
  return `${minutes}:${String(normalized % 60).padStart(2, '0')}`;
}
