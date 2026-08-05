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
