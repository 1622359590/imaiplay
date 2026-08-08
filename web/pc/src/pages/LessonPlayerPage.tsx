import { ArrowLeftOutlined, FileTextOutlined } from '@ant-design/icons';
import { resolveLessonContent } from '@imaiplay/shared/learning/lessonContent';
import { Button, Empty, Progress, Skeleton, Typography } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { getCourse, getResourceFile, type Lesson } from '../api/course';
import {
  getLessonProgress,
  reportLessonProgress,
  reportLessonProgressOnPagehide,
} from '../api/progress';
import { usePortal } from '../context/PortalContext';
import { portalRoutePath } from '../utils/portalRouting';
import { PlaybackLifecycleController, restorePlaybackPosition } from '../utils/watchHeartbeat';

export function LessonPlayerPage() {
  const { courseId = '', lessonId = '' } = useParams();
  const navigate = useNavigate();
  const { mode, tenantCode } = usePortal();
  const videoRef = useRef<HTMLVideoElement>(null);
  const lifecycleRef = useRef<PlaybackLifecycleController | null>(null);
  const lastReported = useRef(-1);
  const [lesson, setLesson] = useState<Lesson>();
  const [percent, setPercent] = useState(0);
  const [initialPosition, setInitialPosition] = useState(0);
  const [loading, setLoading] = useState(true);
  const [resourceURL, setResourceURL] = useState<string>();
  const [resourceLoading, setResourceLoading] = useState(false);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setLesson(undefined);
    setPercent(0);
    setInitialPosition(0);
    lastReported.current = -1;
    void Promise.all([
      getCourse(courseId),
      getLessonProgress(lessonId).catch(() => null),
    ]).then(([course, progress]) => {
      if (active) {
        setLesson(course.chapters?.flatMap((chapter) => chapter.lessons).find((item) => item.id === lessonId));
        if (progress) {
          setPercent(progress.progressPercent);
          setInitialPosition(progress.lastPositionSeconds);
          lastReported.current = progress.progressPercent;
        }
      }
    }).catch(() => {
      if (active) setLesson(undefined);
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [courseId, lessonId]);

  useEffect(() => {
    let active = true;
    setResourceURL(undefined);
    setResourceLoading(Boolean(lesson?.resourceID));
    if (!lesson?.resourceID) return;
    void getResourceFile(lesson.resourceID).then((url) => {
      if (active) setResourceURL(url);
    }).catch(() => {
      if (active) setResourceURL(undefined);
    }).finally(() => {
      if (active) setResourceLoading(false);
    });
    return () => { active = false; };
  }, [lesson]);

  useEffect(() => {
    const controller = new PlaybackLifecycleController({
      now: () => performance.now(),
      read: () => ({
        positionSeconds: videoRef.current?.currentTime ?? 0,
        durationSeconds: videoRef.current?.duration ?? lesson?.durationSeconds ?? 0,
      }),
      report: async (progressReport) => {
        setPercent(progressReport.progressPercent);
        await reportLessonProgress(
          lessonId,
          progressReport.positionSeconds,
          progressReport.progressPercent,
          progressReport.heartbeat,
        );
      },
      terminalReport: (progressReport) => {
        if (!progressReport.heartbeat) return Promise.resolve();
        return reportLessonProgressOnPagehide(
          lessonId,
          progressReport.positionSeconds,
          progressReport.progressPercent,
          progressReport.heartbeat,
        );
      },
    });
    lifecycleRef.current = controller;
    const periodic = window.setInterval(() => void controller.periodicFlush(), 15_000);
    const visibilityChanged = () => void controller.visibilityChanged(document.visibilityState === 'visible');
    const pagehide = () => void controller.pagehide();
    const pageshow = () => void controller.pageshow(Boolean(
      videoRef.current && !videoRef.current.paused && !videoRef.current.ended,
    ));
    document.addEventListener('visibilitychange', visibilityChanged);
    window.addEventListener('pagehide', pagehide);
    window.addEventListener('pageshow', pageshow);
    if (document.visibilityState !== 'visible') void controller.visibilityChanged(false);
    return () => {
      window.clearInterval(periodic);
      document.removeEventListener('visibilitychange', visibilityChanged);
      window.removeEventListener('pagehide', pagehide);
      window.removeEventListener('pageshow', pageshow);
      void controller.pagehide();
      if (lifecycleRef.current === controller) lifecycleRef.current = null;
    };
  }, [lesson?.durationSeconds, lessonId]);

  const report = (video: HTMLVideoElement, force = false) => {
    if (!Number.isFinite(video.duration) || video.duration <= 0) return;
    const next = Math.min(100, Math.floor((video.currentTime / video.duration) * 100));
    setPercent(next);
    if (!force && next < lastReported.current + 5) return;
    lastReported.current = next;
    void reportLessonProgress(lessonId, video.currentTime, next);
  };

  if (loading) return <div className="detail-loading"><Skeleton active /></div>;
  if (!lesson) return <Empty className="page-empty" description="课时不存在" />;
  const content = resolveLessonContent(lesson.contentType, lesson.contentURL, resourceURL);

  return (
    <section className="page-section player-page">
      <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(portalRoutePath(mode, tenantCode, `/courses/${courseId}`))}>返回课程目录</Button>
      <Typography.Title level={2}>{lesson.title}</Typography.Title>
      <div className="player-content">
        {resourceLoading ? (
          <Skeleton active className="lesson-resource-loading" />
        ) : content.kind === 'video' ? (
          <video
            ref={videoRef}
            className="lesson-video"
            controls
            src={content.source}
            onLoadedMetadata={(event) => restorePlaybackPosition(event.currentTarget, initialPosition)}
            onTimeUpdate={(event) => report(event.currentTarget)}
            onPlaying={() => lifecycleRef.current?.playing()}
            onPause={() => void lifecycleRef.current?.pause()}
            onWaiting={() => void lifecycleRef.current?.waiting()}
            onSeeked={() => void lifecycleRef.current?.seeked()}
            onEnded={() => {
              setPercent(100);
              void lifecycleRef.current?.ended();
            }}
          />
        ) : content.kind === 'document' ? (
          <div className="document-panel"><FileTextOutlined /><div><strong>PDF 课时文档</strong><a href={content.source} target="_blank" rel="noreferrer">打开 PDF 文档</a></div></div>
        ) : content.kind === 'text' ? (
          <article className="text-lesson-panel">{content.body}</article>
        ) : (
          <Empty description="该课时尚未配置学习资源" />
        )}
        <div className="player-progress"><span>学习进度</span><Progress percent={percent} /></div>
      </div>
    </section>
  );
}
