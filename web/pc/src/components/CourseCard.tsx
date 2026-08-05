import { BookOutlined } from '@ant-design/icons';
import { Card, Typography } from 'antd';
import { useNavigate } from 'react-router-dom';
import type { Course } from '../api/course';
import { usePortal } from '../context/PortalContext';
import { portalRoutePath } from '../utils/portalRouting';

interface CourseCardProps {
  course: Course;
}

export function CourseCard({ course }: CourseCardProps) {
  const navigate = useNavigate();
  const { mode, tenantCode } = usePortal();
  const openCourse = () => navigate(portalRoutePath(mode, tenantCode, `/courses/${course.id}`));

  return (
    <Card
      className="course-card"
      hoverable
      cover={
        <div
          className="course-cover"
          style={course.coverImage ? { backgroundImage: `url("${course.coverImage}")` } : undefined}
        >
          {!course.coverImage && <BookOutlined />}
        </div>
      }
      onClick={openCourse}
      onKeyDown={(event) => {
        if (event.key === 'Enter' || event.key === ' ') openCourse();
      }}
      role="link"
      tabIndex={0}
    >
      <Typography.Title level={4} ellipsis={{ rows: 2 }}>
        {course.title}
      </Typography.Title>
      {course.lessonCount !== undefined && (
        <Typography.Text type="secondary">{course.lessonCount} 课时</Typography.Text>
      )}
    </Card>
  );
}
