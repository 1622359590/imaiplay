import { Empty, Skeleton } from 'antd';
import type { Course } from '../api/course';
import { CourseCard } from './CourseCard';

interface CourseGridProps {
  courses: Course[];
  loading: boolean;
  emptyText?: string;
}

export function CourseGrid({ courses, loading, emptyText = '暂无课程' }: CourseGridProps) {
  if (loading) {
    return (
      <div className="course-grid">
        {[1, 2, 3, 4].map((item) => (
          <div className="skeleton-card" key={item}><Skeleton active /></div>
        ))}
      </div>
    );
  }

  if (!courses.length) {
    return <Empty className="page-empty" description={emptyText} />;
  }

  return (
    <div className="course-grid stagger-group">
      {courses.map((course) => (
        <CourseCard course={course} key={course.id} />
      ))}
    </div>
  );
}
