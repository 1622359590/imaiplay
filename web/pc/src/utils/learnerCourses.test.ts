import { describe, expect, it } from 'vitest';
import type { LearnerCourse } from '../api/learner';
import * as learnerCourseModule from './learnerCourses';
import {
  courseStatus,
  filterLearnerCourses,
  formatPlaybackPosition,
  learningMinutes,
  type LearnerCourseFilter,
} from './learnerCourses';

const detailLessonState = (
  learnerCourseModule as unknown as {
    detailLessonState: (progress?: {
      progressPercent: number;
      status: number;
      lastPositionSeconds: number;
    }) => { label: string; action: string };
  }
).detailLessonState;

const formatLessonDuration = (
  learnerCourseModule as unknown as { formatLessonDuration: (seconds: number) => string }
).formatLessonDuration;

const courses: LearnerCourse[] = [
  {
    id: 'required-progress',
    title: '销售基础',
	courseType: 'required',
    category: { id: 'sales', name: '销售' },
    lessonCount: 2,
    completedLessonCount: 1,
    progressPercent: 50,
  },
  {
    id: 'optional-complete',
    title: '文化选修',
	courseType: 'optional',
    category: { id: 'culture', name: '文化' },
    lessonCount: 1,
    completedLessonCount: 1,
    progressPercent: 100,
  },
  {
    id: 'required-zero',
    title: '待完善课程',
	courseType: 'required',
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

describe('learner dashboard presentation values', () => {
  it('floors study seconds into total minutes without converting hours', () => {
    expect(learningMinutes(0)).toBe(0);
    expect(learningMinutes(59)).toBe(0);
    expect(learningMinutes(60)).toBe(1);
    expect(learningMinutes(3_480)).toBe(58);
    expect(learningMinutes(7_260)).toBe(121);
  });

  it('formats a saved playback position as total minutes and padded seconds', () => {
    expect(formatPlaybackPosition(0)).toBe('0:00');
    expect(formatPlaybackPosition(65)).toBe('1:05');
    expect(formatPlaybackPosition(3_661)).toBe('61:01');
  });
});

describe('course detail lesson presentation', () => {
  it('renders exact padded lesson durations without rounding to minutes', () => {
    expect(formatLessonDuration(0)).toBe('00:00');
    expect(formatLessonDuration(65.9)).toBe('01:05');
    expect(formatLessonDuration(3_661)).toBe('61:01');
  });

  it('derives completed, partial, and untouched labels from exact progress', () => {
    expect(detailLessonState({
      progressPercent: 100,
      status: 2,
      lastPositionSeconds: 60,
    })).toEqual({ label: '已学完', action: '已学完' });
    expect(detailLessonState({
      progressPercent: 35,
      status: 1,
      lastPositionSeconds: 65,
    })).toEqual({ label: '上次学习到 01:05', action: '继续学习' });
    expect(detailLessonState({
      progressPercent: 0,
      status: 0,
      lastPositionSeconds: 0,
    })).toEqual({ label: '', action: '开始学习' });
    expect(detailLessonState()).toEqual({ label: '', action: '开始学习' });
  });
});
