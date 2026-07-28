import { ArrowLeftOutlined, FileTextOutlined } from '@ant-design/icons';
import { Button, Card, Empty, Progress, Skeleton, Typography } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { getCourse, getResourceFile, type Lesson } from '../api/course';
import { getLessonProgress, reportLessonProgress } from '../api/progress';

export function LessonPlayerPage() {
  const { courseId = '', lessonId = '' } = useParams();
  const navigate = useNavigate();
  const videoRef = useRef<HTMLVideoElement>(null);
  const lastReported = useRef(-1);
  const [lesson, setLesson] = useState<Lesson>();
  const [percent, setPercent] = useState(0);
  const [initialPosition, setInitialPosition] = useState(0);
  const [loading, setLoading] = useState(true);
  const [resourceURL, setResourceURL] = useState<string>();

  useEffect(() => {
    Promise.all([
      getCourse(courseId),
      getLessonProgress(lessonId).catch(() => null),
    ]).then(([course, progress]) => {
      setLesson(course.chapters?.flatMap((chapter) => chapter.lessons).find((item) => item.id === lessonId));
      if (progress) {
        setPercent(progress.progress_percent);
        setInitialPosition(progress.last_position_seconds);
      }
    }).finally(() => setLoading(false));
  }, [courseId, lessonId]);

  useEffect(() => {
    let objectURL: string | undefined;
    setResourceURL(undefined);
    if (!lesson?.resource_id) return;
    void getResourceFile(lesson.resource_id).then((blob) => {
      objectURL = URL.createObjectURL(blob);
      setResourceURL(objectURL);
    }).catch(() => setResourceURL(undefined));
    return () => { if (objectURL) URL.revokeObjectURL(objectURL); };
  }, [lesson]);

  const report = (video: HTMLVideoElement, force = false) => {
    if (!Number.isFinite(video.duration) || video.duration <= 0) return;
    const next = Math.min(100, Math.floor((video.currentTime / video.duration) * 100));
    setPercent(next);
    if (!force && next < lastReported.current + 5) return;
    lastReported.current = next;
    void reportLessonProgress(lessonId, video.currentTime, next);
  };

  if (loading) return <Card className="detail-loading"><Skeleton active /></Card>;
  if (!lesson) return <Empty className="page-empty" description="课时不存在" />;

  return (
    <section className="page-section player-page">
      <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/courses/${courseId}`)}>返回课程</Button>
      <Typography.Title level={2}>{lesson.title}</Typography.Title>
      <Card bordered={false}>
        {lesson.content_type === 'video' && (resourceURL || lesson.content_url) ? (
          <video
            ref={videoRef}
            className="lesson-video"
            controls
            src={resourceURL || lesson.content_url}
            onLoadedMetadata={(event) => { event.currentTarget.currentTime = initialPosition }}
            onTimeUpdate={(event) => report(event.currentTarget)}
            onPause={(event) => report(event.currentTarget, true)}
            onEnded={(event) => {
              setPercent(100);
              void reportLessonProgress(lessonId, event.currentTarget.duration, 100);
            }}
          />
        ) : resourceURL || lesson.content_url ? (
          <div className="document-panel"><FileTextOutlined /><a href={resourceURL || lesson.content_url} target="_blank" rel="noreferrer">打开课时资料</a></div>
        ) : (
          <Empty description="该课时尚未配置学习资源" />
        )}
        <div className="player-progress"><span>学习进度</span><Progress percent={percent} /></div>
      </Card>
    </section>
  );
}
