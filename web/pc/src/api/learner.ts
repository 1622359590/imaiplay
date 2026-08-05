import { apiClient } from './client';

export type AssignmentType = 'required' | 'optional';

export interface LearnerCategory {
  id: string;
  name: string;
}

export interface LearnerRecentLesson {
  id: string;
  title: string;
  durationSeconds: number;
  lastPositionSeconds: number;
}

export interface LearnerCourse {
  id: string;
  title: string;
  description?: string;
  coverImage?: string;
  assignmentType: AssignmentType;
  category?: LearnerCategory;
  lessonCount: number;
  completedLessonCount: number;
  progressPercent: number;
  lastLearnedAt?: string;
  recentLesson?: LearnerRecentLesson;
}

export interface LearnerOverview {
  requiredCompleted: number;
  requiredTotal: number;
  todayLearningSeconds: number;
  totalLearningSeconds: number;
  categories: LearnerCategory[];
  courses: LearnerCourse[];
}

export interface RecentLearningItem {
  course: {
    id: string;
    title: string;
    description?: string;
    coverImage?: string;
    category?: LearnerCategory;
  };
  recentLesson: LearnerRecentLesson;
  progressPercent: number;
  lastPositionSeconds: number;
  lastLearnedAt: string;
}

export interface RecentLearningPage {
  items: RecentLearningItem[];
  total: number;
}

interface RawCategory {
  id: string;
  name: string;
}

interface RawCourseSummary {
  id: string;
  title: string;
  description?: string;
  cover_image?: string;
  category?: RawCategory;
}

interface RawRecentLesson {
  id: string;
  title: string;
  duration_seconds: number;
  last_position_seconds: number;
}

interface RawLearnerCourse {
  course: RawCourseSummary;
  assignment_type: AssignmentType;
  lesson_count: number;
  completed_lesson_count: number;
  progress_percent: number;
  last_learned_at?: string;
  recent_lesson?: RawRecentLesson;
}

interface RawLearnerOverview {
  required_completed: number;
  required_total: number;
  today_learning_seconds: number;
  total_learning_seconds: number;
  categories: RawCategory[];
  courses: RawLearnerCourse[];
}

interface RawRecentLearningItem {
  course: RawCourseSummary;
  recent_lesson: RawRecentLesson;
  progress_percent: number;
  last_position_seconds: number;
  last_learned_at: string;
}

interface RawRecentLearningPage {
  items: RawRecentLearningItem[];
  total: number;
}

function nonNegativeInteger(value: number): number {
  return Number.isFinite(value) ? Math.max(0, Math.floor(value)) : 0;
}

function progressPercent(value: number, lessonCount?: number): number {
  if (lessonCount === 0) return 0;
  return Math.min(100, nonNegativeInteger(value));
}

function mapCategory(category: RawCategory): LearnerCategory {
  return { id: category.id, name: category.name };
}

function mapCourseSummary(course: RawCourseSummary): RecentLearningItem['course'] {
  return {
    id: course.id,
    title: course.title,
    description: course.description,
    coverImage: course.cover_image,
    category: course.category ? mapCategory(course.category) : undefined,
  };
}

function mapRecentLesson(lesson: RawRecentLesson): LearnerRecentLesson {
  return {
    id: lesson.id,
    title: lesson.title,
    durationSeconds: nonNegativeInteger(lesson.duration_seconds),
    lastPositionSeconds: nonNegativeInteger(lesson.last_position_seconds),
  };
}

function mapLearnerCourse(item: RawLearnerCourse): LearnerCourse {
  const lessonCount = nonNegativeInteger(item.lesson_count);
  return {
    ...mapCourseSummary(item.course),
    assignmentType: item.assignment_type,
    lessonCount,
    completedLessonCount: Math.min(
      lessonCount,
      nonNegativeInteger(item.completed_lesson_count),
    ),
    progressPercent: progressPercent(item.progress_percent, lessonCount),
    lastLearnedAt: item.last_learned_at,
    recentLesson: item.recent_lesson ? mapRecentLesson(item.recent_lesson) : undefined,
  };
}

export async function getLearnerOverview(): Promise<LearnerOverview> {
  const response = await apiClient.get<RawLearnerOverview>('/api/v1/learner/overview');
  return {
    requiredCompleted: nonNegativeInteger(response.data.required_completed),
    requiredTotal: nonNegativeInteger(response.data.required_total),
    todayLearningSeconds: nonNegativeInteger(response.data.today_learning_seconds),
    totalLearningSeconds: nonNegativeInteger(response.data.total_learning_seconds),
    categories: response.data.categories.map(mapCategory),
    courses: response.data.courses.map(mapLearnerCourse),
  };
}

export async function getRecentLearning(
  offset = 0,
  limit = 20,
): Promise<RecentLearningPage> {
  const response = await apiClient.get<RawRecentLearningPage>(
    '/api/v1/recent-learning',
    { params: { offset, limit } },
  );
  return {
    items: response.data.items.map((item) => ({
      course: mapCourseSummary(item.course),
      recentLesson: mapRecentLesson(item.recent_lesson),
      progressPercent: progressPercent(item.progress_percent),
      lastPositionSeconds: nonNegativeInteger(item.last_position_seconds),
      lastLearnedAt: item.last_learned_at,
    })),
    total: nonNegativeInteger(response.data.total),
  };
}
