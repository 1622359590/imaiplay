export interface Lesson {
  id: string
  title: string
  contentType?: 'video' | 'document' | 'text'
  contentUrl?: string
  resourceId?: string
  duration: number
  completed?: boolean
  free?: boolean
}

export interface Chapter {
  id: string
  title: string
  lessons: Lesson[]
}

export interface CourseMaterial {
  id: string
  displayName: string
  sizeBytes: number
  resourceType: 'attachment'
}

export interface Course {
  id: string
  title: string
  description: string
  cover?: string
  instructor: string
  progress: number
  lessonCount?: number
  duration: number
  category: string
	courseType: 'required' | 'optional'
  chapters?: Chapter[]
  materials?: CourseMaterial[]
}

export interface CourseList {
  items: Course[]
  total: number
}
