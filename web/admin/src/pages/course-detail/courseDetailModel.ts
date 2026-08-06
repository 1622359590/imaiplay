import type { Chapter, Lesson } from '../../api/course'

export type Editor =
  | { kind: 'chapter'; chapter?: Chapter }
  | { kind: 'lesson'; chapter: Chapter; lesson?: Lesson }

export type LessonForm = Omit<Lesson, 'id'> & { title: string }

export function lessonPayload(values: LessonForm): Omit<Lesson, 'id'> {
  return {
    title: values.title,
    content_type: values.content_type,
    content_url: values.content_type === 'text' ? values.content_url || '' : '',
    resource_id: values.content_type === 'text' ? undefined : values.resource_id,
    duration_seconds: values.duration_seconds || 0,
    sort_order: values.sort_order || 0,
  }
}
