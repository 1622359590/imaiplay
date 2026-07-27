export interface Lesson {
  id: string
  title: string
  duration: number
  completed?: boolean
  free?: boolean
}

export interface Chapter {
  id: string
  title: string
  lessons: Lesson[]
}

export interface Course {
  id: string
  title: string
  description: string
  cover: string
  instructor: string
  progress: number
  lessonCount: number
  duration: number
  category: string
  chapters?: Chapter[]
}

export interface CourseList {
  items: Course[]
  total: number
}
