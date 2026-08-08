export type LessonContentType = 'video' | 'document' | 'text'

export type ResolvedLessonContent =
  | { kind: 'video' | 'document'; source: string }
  | { kind: 'text'; body: string }
  | { kind: 'empty' }

export function resolveLessonContent(
  contentType: LessonContentType,
  storedContent = '',
  resolvedResourceURL?: string,
): ResolvedLessonContent {
  if (contentType === 'text') {
    return storedContent.trim() ? { kind: 'text', body: storedContent } : { kind: 'empty' }
  }
  const source = (resolvedResourceURL || storedContent).trim()
  return source ? { kind: contentType, source } : { kind: 'empty' }
}

export function lessonContentLabel(contentType: LessonContentType): string {
  if (contentType === 'document') return 'PDF 文档'
  if (contentType === 'text') return '图文'
  return '视频'
}
