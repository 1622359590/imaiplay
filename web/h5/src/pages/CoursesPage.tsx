import { useEffect, useState } from 'react'
import { SearchBar, Segmented } from 'antd-mobile'
import { getCourses } from '../api/course'
import { CourseCard } from '../components/CourseCard'
import type { Course } from '../types/course'

export function CoursesPage() {
  const [courses, setCourses] = useState<Course[]>([])
  const [keyword, setKeyword] = useState('')
  const [segment, setSegment] = useState<string | number>('全部')

  useEffect(() => {
    getCourses().then((result) => setCourses(result.items)).catch(() => setCourses([]))
  }, [])

  const visibleCourses = courses.filter((course) => {
    const matchesKeyword = course.title.toLowerCase().includes(keyword.toLowerCase())
    const matchesSegment =
      segment === '全部' || (segment === '进行中' ? course.progress > 0 && course.progress < 100 : course.progress === 100)
    return matchesKeyword && matchesSegment
  })

  return (
    <div className="standard-page">
      <header className="page-header reveal">
        <span>LEARNING CENTER</span>
        <h1>我的课程</h1>
        <p>聚焦岗位能力，构建专属成长路径</p>
      </header>
      <SearchBar placeholder="搜索课程名称" value={keyword} onChange={setKeyword} />
      <Segmented
        block
        options={['全部', '进行中', '已完成']}
        value={segment}
        onChange={setSegment}
        className="course-segment"
      />
      <div className="course-list stagger-group">
        {visibleCourses.map((course) => <CourseCard key={course.id} course={course} />)}
        {!visibleCourses.length && <div className="empty-state">没有找到匹配的课程</div>}
      </div>
    </div>
  )
}
