import {
  BookOutlined,
  ClockCircleOutlined,
  PlayCircleFilled,
  TeamOutlined,
} from '@ant-design/icons';
import { Button, Card, Space, Tag, Typography } from 'antd';
import type { CSSProperties } from 'react';
import { useNavigate } from 'react-router-dom';
import type { Course } from '../api/course';

interface CourseCardProps {
  course: Course;
}

function formatDuration(minutes?: number): string {
  if (!minutes) return '时长待定';
  if (minutes < 60) return `${minutes} 分钟`;
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  return rest ? `${hours} 小时 ${rest} 分` : `${hours} 小时`;
}

export function CourseCard({ course }: CourseCardProps) {
  const navigate = useNavigate();
  const progress = Math.max(0, Math.min(100, course.progress ?? 0));

  return (
    <Card
      className="course-card glass-card tilt-card reveal"
      hoverable
      cover={
        <div
          className="course-cover"
          style={course.cover ? { backgroundImage: `url("${course.cover}")` } : undefined}
        >
          {!course.cover && <BookOutlined />}
          <span className="cover-overlay" />
          {course.category && <Tag color="blue">{course.category}</Tag>}
        </div>
      }
      onClick={() => navigate(`/courses/${course.id}`)}
    >
      <Typography.Title level={4} ellipsis={{ rows: 2 }}>
        {course.title}
      </Typography.Title>
      <Typography.Paragraph className="course-description" ellipsis={{ rows: 2 }}>
        {course.description || '课程内容持续更新中，点击查看课程详情。'}
      </Typography.Paragraph>
      <Space size="middle" className="course-meta">
        <span><BookOutlined /> {course.lesson_count ?? 0} 课时</span>
        <span><ClockCircleOutlined /> {formatDuration(course.duration)}</span>
        {course.student_count !== undefined && (
          <span><TeamOutlined /> {course.student_count}</span>
        )}
      </Space>
      {progress > 0 && (
        <div className="course-progress">
          <span>学习进度</span>
          <div className="progress-track"><div className="progress-fill" style={{ '--progress': `${progress}%` } as CSSProperties} /></div>
        </div>
      )}
      <Button type="primary" ghost icon={<PlayCircleFilled />} block>
        {progress > 0 ? '继续学习' : '开始学习'}
      </Button>
    </Card>
  );
}
