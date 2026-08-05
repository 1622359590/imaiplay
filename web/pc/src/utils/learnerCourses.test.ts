import { describe, expect, it } from 'vitest';
import type { LearnerCourse } from '../api/learner';
import {
  courseStatus,
  filterLearnerCourses,
  type LearnerCourseFilter,
} from './learnerCourses';

const courses: LearnerCourse[] = [
  {
    id: 'required-progress',
    title: '销售基础',
    assignmentType: 'required',
    category: { id: 'sales', name: '销售' },
    lessonCount: 2,
    completedLessonCount: 1,
    progressPercent: 50,
  },
  {
    id: 'optional-complete',
    title: '文化选修',
    assignmentType: 'optional',
    category: { id: 'culture', name: '文化' },
    lessonCount: 1,
    completedLessonCount: 1,
    progressPercent: 100,
  },
  {
    id: 'required-zero',
    title: '待完善课程',
    assignmentType: 'required',
    category: { id: 'sales', name: '销售' },
    lessonCount: 0,
    completedLessonCount: 0,
    progressPercent: 0,
  },
];

describe('learner course status', () => {
  it('treats only a non-empty fully completed course as completed', () => {
    expect(courseStatus(courses[0])).toBe('incomplete');
    expect(courseStatus(courses[1])).toBe('completed');
    expect(courseStatus(courses[2])).toBe('incomplete');
  });
});

describe('learner course filtering', () => {
  it.each<[LearnerCourseFilter['tab'], string[]]>([
    ['all', ['required-progress', 'optional-complete', 'required-zero']],
    ['required', ['required-progress', 'required-zero']],
    ['optional', ['optional-complete']],
    ['completed', ['optional-complete']],
    ['incomplete', ['required-progress', 'required-zero']],
  ])('filters the %s tab without changing order', (tab, expected) => {
    expect(filterLearnerCourses(courses, { tab }).map((course) => course.id)).toEqual(expected);
  });

  it('composes a category with the mutually exclusive tab', () => {
    expect(filterLearnerCourses(courses, {
      tab: 'required',
      categoryId: 'sales',
    })).toEqual([courses[0], courses[2]]);
    expect(filterLearnerCourses(courses, {
      tab: 'optional',
      categoryId: 'sales',
    })).toEqual([]);
  });

  it('does not mutate the source courses', () => {
    const before = structuredClone(courses);
    filterLearnerCourses(courses, { tab: 'completed', categoryId: 'culture' });
    expect(courses).toEqual(before);
  });
});
