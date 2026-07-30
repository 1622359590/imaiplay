import {
  BookOutlined,
  CheckCircleFilled,
  ClockCircleOutlined,
  PlayCircleOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import { Breadcrumb, Button, Card, Collapse, Empty, Skeleton, Space, Tag, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getCourse, type Course } from '../api/course';

function durationLabel(minutes?: number): string {
  if (!minutes) return '时长待定';
  return minutes >= 60 ? `${Math.floor(minutes / 60)} 小时 ${minutes % 60} 分钟` : `${minutes} 分钟`;
}

export function CourseDetailPage() {
  const { courseId = '' } = useParams();
  const [course, setCourse] = useState<Course | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setLoading(true);
    getCourse(courseId)
      .then((result) => {
        if (active) setCourse(result);
      })
      .catch(() => {
        if (active) setCourse(null);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [courseId]);

  const collapseItems = useMemo(
    () =>
      (course?.chapters ?? []).map((chapter, chapterIndex) => ({
        key: chapter.id,
        label: (
          <div className="chapter-title">
            <span>第 {chapterIndex + 1} 章　{chapter.title}</span>
            <Typography.Text type="secondary">{chapter.lessons.length} 课时</Typography.Text>
          </div>
        ),
        children: (
          <div className="lesson-list">
            {chapter.lessons.map((lesson, lessonIndex) => (
              <div className="lesson-row" key={lesson.id}>
                <Space>
                  {lesson.learned ? (
                    <CheckCircleFilled className="lesson-complete" />
                  ) : (
                    <PlayCircleOutlined className="lesson-play" />
                  )}
                  <span>{chapterIndex + 1}.{lessonIndex + 1}　{lesson.title}</span>
                </Space>
                <Space>
                  <Typography.Text type="secondary">{durationLabel(lesson.duration)}</Typography.Text>
                  <Link to={`/courses/${courseId}/lessons/${lesson.id}`}>学习</Link>
                </Space>
              </div>
            ))}
          </div>
        ),
      })),
    [course],
  );

  if (loading) {
    return <Card className="detail-loading"><Skeleton active paragraph={{ rows: 8 }} /></Card>;
  }

  if (!course) {
    return <Empty className="page-empty" description="课程不存在或暂时无法访问" />;
  }

  return (
    <section className="page-section reveal">
      <Breadcrumb
        className="detail-breadcrumb"
        items={[
          { title: <Link to="/">首页</Link> },
          { title: <Link to="/courses">全部课程</Link> },
          { title: course.title },
        ]}
      />
      <div className="detail-hero">
        <div
          className="detail-cover"
          style={course.cover ? { backgroundImage: `url("${course.cover}")` } : undefined}
        >
          {!course.cover && <BookOutlined />}
        </div>
        <div className="detail-summary">
          {course.category && <Tag color="blue">{course.category}</Tag>}
          <Typography.Title>{course.title}</Typography.Title>
          <Typography.Paragraph>{course.description || '暂无课程简介'}</Typography.Paragraph>
          <Space size="large" wrap className="detail-meta">
            <span><BookOutlined /> {course.lesson_count ?? 0} 课时</span>
            <span><ClockCircleOutlined /> {durationLabel(course.duration)}</span>
            {course.student_count !== undefined && (
              <span><TeamOutlined /> {course.student_count} 人学习</span>
            )}
          </Space>
          <Button
            type="primary"
            size="large"
            icon={<PlayCircleOutlined />}
            disabled={!course.chapters?.some((chapter) => chapter.lessons.length)}
            onClick={() => {
              const lesson = course.chapters?.flatMap((chapter) => chapter.lessons)[0];
              if (lesson) window.location.assign(`/courses/${course.id}/lessons/${lesson.id}`);
            }}
          >
            {(course.progress ?? 0) > 0 ? '继续学习' : '开始学习'}
          </Button>
        </div>
      </div>
      <Card className="chapter-card glass-card" title="课程目录" bordered={false}>
        {collapseItems.length ? (
          <Collapse items={collapseItems} defaultActiveKey={[collapseItems[0].key]} />
        ) : (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="课程章节正在准备中" />
        )}
      </Card>
    </section>
  );
}
