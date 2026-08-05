import { ArrowLeftOutlined, PlayCircleOutlined, ReloadOutlined, TrophyFilled } from '@ant-design/icons';
import { Button, Empty, Progress, Result, Skeleton, Tabs, Tag, Typography } from 'antd';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { getCourse, type Course } from '../api/course';
import { getLearnerOverview, type LearnerCourse } from '../api/learner';
import { getLessonProgress, type LessonProgress } from '../api/progress';
import { usePortal } from '../context/PortalContext';
import { portalRoutePath } from '../utils/portalRouting';
import { CourseMaterials } from '../components/CourseMaterials';
import { courseStatus, detailLessonState, formatLessonDuration } from '../utils/learnerCourses';

export function CourseDetailPage() {
  const { courseId = '' } = useParams();
  const navigate = useNavigate();
  const { mode, tenantCode } = usePortal();
  const pathFor = (childPath: string) => portalRoutePath(mode, tenantCode, childPath);
  const [course, setCourse] = useState<Course | null>(null);
  const [learnerCourse, setLearnerCourse] = useState<LearnerCourse | null>(null);
  const [lessonProgress, setLessonProgress] = useState<Record<string, LessonProgress | null | undefined>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>();
  const [requestVersion, setRequestVersion] = useState(0);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(undefined);
    setCourse(null);
    setLearnerCourse(null);
    setLessonProgress({});
    void Promise.all([getCourse(courseId), getLearnerOverview()])
      .then(async ([courseResult, overview]) => {
        const matchedCourse = overview.courses.find((item) => item.id === courseId);
        if (!matchedCourse) throw new Error('Course is absent from learner overview');
        const lessons = (courseResult.chapters ?? []).flatMap((chapter) => chapter.lessons);
        const progressResults = await Promise.all(lessons.map(async (lesson) => {
          try {
            return [lesson.id, await getLessonProgress(lesson.id)] as const;
          } catch {
            return [lesson.id, null] as const;
          }
        }));
        if (active) {
          setCourse(courseResult);
          setLearnerCourse(matchedCourse);
          setLessonProgress(Object.fromEntries(progressResults));
        }
      })
      .catch((requestError: unknown) => {
        if (active) setError(requestError);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [courseId, requestVersion]);

  const retryLessonProgress = useCallback((lessonId: string) => {
    setLessonProgress((current) => ({ ...current, [lessonId]: undefined }));
    void getLessonProgress(lessonId)
      .then((progress) => setLessonProgress((current) => ({ ...current, [lessonId]: progress })))
      .catch(() => setLessonProgress((current) => ({ ...current, [lessonId]: null })));
  }, []);

  const catalog = useMemo(
    () =>
      (course?.chapters ?? []).map((chapter, chapterIndex) => (
        <section className="course-chapter" key={chapter.id} aria-labelledby={`chapter-${chapter.id}`}>
          <h2 id={`chapter-${chapter.id}`}>第 {chapterIndex + 1} 章：{chapter.title}</h2>
          {chapter.lessons.length ? (
            <div className="course-lesson-list">
              {chapter.lessons.map((lesson) => {
              const progress = lessonProgress[lesson.id];
              const state = progress === undefined ? null : detailLessonState(progress ?? undefined);
              const lessonPath = pathFor(`/courses/${courseId}/lessons/${lesson.id}`);
              const lessonName = <span className="lesson-name"><PlayCircleOutlined aria-hidden="true" />{lesson.title} ({formatLessonDuration(lesson.durationSeconds)})</span>;
              if (progress === null) {
                return (
                  <div className="lesson-row" key={lesson.id}>
                    <Link className="lesson-main-link" to={lessonPath}>{lessonName}</Link>
                    <span className="lesson-state">
                      <Button type="link" icon={<ReloadOutlined />} onClick={() => retryLessonProgress(lesson.id)}>进度加载失败，重试</Button>
                    </span>
                  </div>
                );
              }
              return (
                <Link className="lesson-row" key={lesson.id} to={lessonPath}>
                  {lessonName}
                  <span className="lesson-state">
                    {!state ? <span className="lesson-progress-loading">正在加载进度</span> : (
                      <><span>{state.label}</span><strong>{state.action}</strong></>
                    )}
                  </span>
                </Link>
              );
            })}
            </div>
          ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="本章暂无课时" />}
        </section>
      )),
    [course, courseId, lessonProgress, mode, retryLessonProgress, tenantCode],
  );

  if (loading) return <div className="detail-loading"><Skeleton active paragraph={{ rows: 8 }} /></div>;

  if (error || !course || !learnerCourse) {
    return (
      <Result
        className="learner-request-state"
        status="error"
        title="课程加载失败"
        subTitle="课程不存在、无访问权限或网络暂时不可用"
        extra={[
          <Button key="retry" type="primary" icon={<ReloadOutlined />} onClick={() => setRequestVersion((version) => version + 1)}>重新加载</Button>,
          <Button key="back" onClick={() => navigate(pathFor('/'))}>返回我的课程</Button>,
        ]}
      />
    );
  }

  const completed = courseStatus(learnerCourse) === 'completed';
  const catalogContent = catalog.length ? catalog : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无可学习课时" />;

  return (
    <section className="page-section course-detail-page">
      <Link className="back-link" to={pathFor('/')}><ArrowLeftOutlined /> 返回我的课程</Link>
      <div className={`detail-hero${course.coverImage ? '' : ' detail-hero-without-cover'}`}>
        {course.coverImage && <img className="detail-cover" src={course.coverImage} alt={`${course.title}课程封面`} />}
        <div className="detail-summary">
          <Typography.Title level={1}>{course.title}</Typography.Title>
          <div className="detail-status-row">
            <Tag className="course-type-tag">{learnerCourse.assignmentType === 'required' ? '必修课' : '选修课'}</Tag>
            {completed ? (
              <span className="detail-complete-copy"><TrophyFilled aria-hidden="true" />恭喜你学完此课程！</span>
            ) : (
              <span>已完成 {learnerCourse.completedLessonCount} / {learnerCourse.lessonCount} 个课时</span>
            )}
          </div>
          <Typography.Paragraph>{course.description || '暂无课程简介'}</Typography.Paragraph>
        </div>
        <div className="detail-progress" aria-label={`课程总体进度 ${learnerCourse.progressPercent}%`}>
          <Progress type="circle" percent={learnerCourse.progressPercent} strokeColor="var(--learner-accent)" trailColor="var(--learner-line)" size={124} />
          <span>课程总体进度</span>
        </div>
      </div>
      <Tabs
        className="course-experience-tabs"
        defaultActiveKey="catalog"
        items={[
          { key: 'catalog', label: '课程目录', children: <div className="course-tab-panel">{catalogContent}</div> },
          { key: 'materials', label: '课程附件', children: <div className="course-tab-panel"><CourseMaterials materials={course.materials ?? []} /></div> },
        ]}
      />
    </section>
  );
}
