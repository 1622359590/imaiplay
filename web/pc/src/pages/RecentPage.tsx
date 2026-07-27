import { ClockCircleOutlined } from '@ant-design/icons';
import { Typography } from 'antd';
import { getRecentCourses } from '../api/course';
import { CourseGrid } from '../components/CourseGrid';
import { useCourseList } from '../hooks/useCourseList';

export function RecentPage() {
  const { courses, loading } = useCourseList(getRecentCourses);

  return (
    <section className="content-section page-section">
      <div className="section-heading">
        <div>
          <Typography.Title level={1}><ClockCircleOutlined /> 最近学习</Typography.Title>
          <Typography.Text type="secondary">从上次停下的地方继续，保持学习节奏</Typography.Text>
        </div>
      </div>
      <CourseGrid courses={courses} loading={loading} emptyText="还没有学习记录，去挑选一门课程吧" />
    </section>
  );
}
