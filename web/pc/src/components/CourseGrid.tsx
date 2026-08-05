import { Empty, Skeleton } from 'antd';
import type { LearnerCourse } from '../api/learner';
import { CourseCard } from './CourseCard';

interface CourseGridProps {
  courses: LearnerCourse[];
  loading: boolean;
  emptyText?: string;
}

export function CourseGrid({ courses, loading, emptyText = '暂无课程' }: CourseGridProps) {
  if (loading) {
    return (
      <div className="course-grid" aria-busy="true" aria-label="正在加载课程">
        {[1, 2, 3, 4].map((item) => (
          <div className="skeleton-card learner-course-skeleton" key={item}><Skeleton active /></div>
        ))}
      </div>
    );
  }

  if (!courses.length) {
    return <Empty className="page-empty" image={Empty.PRESENTED_IMAGE_SIMPLE} description={emptyText} />;
  }

  return (
    <div className="course-grid">
      {courses.map((course) => (
        <CourseCard course={course} key={course.id} />
      ))}
    </div>
  );
}
