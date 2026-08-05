import { ArrowLeftOutlined, BookOutlined } from '@ant-design/icons';
import { Button, Collapse, Empty, Skeleton, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getCourse, type Course } from '../api/course';
import { usePortal } from '../context/PortalContext';
import { portalRoutePath } from '../utils/portalRouting';
import { CourseMaterials } from '../components/CourseMaterials';

function durationLabel(seconds: number): string | null {
  if (seconds <= 0) return null;
  const minutes = Math.floor(seconds / 60).toString().padStart(2, '0');
  const remainder = Math.floor(seconds % 60).toString().padStart(2, '0');
  return `${minutes}:${remainder}`;
}

export function CourseDetailPage() {
  const { courseId = '' } = useParams();
  const { mode, tenantCode } = usePortal();
  const pathFor = (childPath: string) => portalRoutePath(mode, tenantCode, childPath);
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
        children: chapter.lessons.length ? (
          <div className="lesson-list">
            {chapter.lessons.map((lesson, lessonIndex) => {
              const duration = durationLabel(lesson.durationSeconds);
              return (
                <Link className="lesson-row" key={lesson.id} to={pathFor(`/courses/${courseId}/lessons/${lesson.id}`)}>
                  <span>{chapterIndex + 1}.{lessonIndex + 1}　{lesson.title}</span>
                  {duration && <Typography.Text type="secondary">{duration}</Typography.Text>}
                </Link>
              );
            })}
          </div>
        ) : (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="本章暂无课时" />
        ),
      })),
    [course, courseId, mode, tenantCode],
  );

  if (loading) return <div className="detail-loading"><Skeleton active paragraph={{ rows: 8 }} /></div>;

  if (!course) {
    return (
      <div className="error-state">
        <Empty description="课程不存在或暂时无法访问" />
        <Link to={pathFor('/')}><Button>返回我的课程</Button></Link>
      </div>
    );
  }

  const lessonCount = (course.chapters ?? []).reduce(
    (total, chapter) => total + chapter.lessons.length,
    0,
  );

  return (
    <section className="page-section">
      <Link className="back-link" to={pathFor('/')}><ArrowLeftOutlined /> 返回我的课程</Link>
      <div className="detail-hero">
        <div
          className="detail-cover"
          style={course.coverImage ? { backgroundImage: `url("${course.coverImage}")` } : undefined}
        >
          {!course.coverImage && <BookOutlined />}
        </div>
        <div className="detail-summary">
          <Typography.Title>{course.title}</Typography.Title>
          <Typography.Paragraph>{course.description || '暂无课程简介'}</Typography.Paragraph>
          <Typography.Text type="secondary">{lessonCount} 课时</Typography.Text>
        </div>
      </div>
      <CourseMaterials materials={course.materials ?? []} />
      <section className="chapter-card" aria-labelledby="course-outline-title">
        <Typography.Title level={2} id="course-outline-title">课程目录</Typography.Title>
        {collapseItems.length ? (
          <Collapse items={collapseItems} defaultActiveKey={[collapseItems[0].key]} />
        ) : (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无可学习课时" />
        )}
      </section>
    </section>
  );
}
