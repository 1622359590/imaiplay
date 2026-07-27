import { Empty, Row, Col, Skeleton } from 'antd';
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
      <Row gutter={[24, 24]}>
        {[1, 2, 3, 4].map((item) => (
          <Col xs={24} sm={12} lg={8} xl={6} key={item}>
            <div className="skeleton-card"><Skeleton active /></div>
          </Col>
        ))}
      </Row>
    );
  }

  if (!courses.length) {
    return <Empty className="page-empty" description={emptyText} />;
  }

  return (
    <Row gutter={[24, 24]}>
      {courses.map((course) => (
        <Col xs={24} sm={12} lg={8} xl={6} key={course.id}>
          <CourseCard course={course} />
        </Col>
      ))}
    </Row>
  );
}
