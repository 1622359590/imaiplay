import { Typography } from 'antd';
import { getCourses } from '../api/course';
import { CourseGrid } from '../components/CourseGrid';
import { useCourseList } from '../hooks/useCourseList';

export function HomePage() {
  const { courses, loading } = useCourseList(getCourses);

  return (
    <section className="course-home page-section">
      <div className="course-home-heading">
        <div>
          <Typography.Title level={1}>我的课程</Typography.Title>
          <Typography.Text type="secondary">选择课程开始学习</Typography.Text>
        </div>
        {!loading && <Typography.Text type="secondary">共 {courses.length} 门</Typography.Text>}
      </div>
      <CourseGrid courses={courses} loading={loading} emptyText="暂无可学习课程" />
    </section>
  );
}
