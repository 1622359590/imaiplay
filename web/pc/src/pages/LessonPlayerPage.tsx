import { ArrowLeftOutlined, ExpandOutlined, FileTextOutlined, PauseOutlined, PlayCircleFilled, RightOutlined, SoundOutlined } from '@ant-design/icons';
import { resolveLessonContent } from '@imaiplay/shared/learning/lessonContent';
import { Button, Empty, Progress, Skeleton } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { getCourse, getResourceFile, type Course, type Lesson } from '../api/course';
import {
  getLessonProgress,
  reportLessonProgress,
  reportLessonProgressOnPagehide,
} from '../api/progress';
import { usePortal } from '../context/PortalContext';
import { portalRoutePath } from '../utils/portalRouting';
import { PlaybackLifecycleController, restorePlaybackPosition } from '../utils/watchHeartbeat';
import { formatPlaybackPosition } from '../utils/learnerCourses';

export function LessonPlayerPage() {
  const { courseId = '', lessonId = '' } = useParams();
  const navigate = useNavigate();
  const { mode, tenantCode } = usePortal();
  const videoRef = useRef<HTMLVideoElement>(null);
  const lifecycleRef = useRef<PlaybackLifecycleController | null>(null);
  const lastReported = useRef(-1);
  const [lesson, setLesson] = useState<Lesson>();
  const [course, setCourse] = useState<Course>();
  const [percent, setPercent] = useState(0);
  const [initialPosition, setInitialPosition] = useState(0);
  const [loading, setLoading] = useState(true);
  const [resourceURL, setResourceURL] = useState<string>();
  const [resourceLoading, setResourceLoading] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [sidebarTab, setSidebarTab] = useState<'outline' | 'materials'>('outline');

  useEffect(() => {
    let active = true;
    setLoading(true);
    setLesson(undefined);
    setCourse(undefined);
    setPercent(0);
    setInitialPosition(0);
    lastReported.current = -1;
    void Promise.all([
      getCourse(courseId),
      getLessonProgress(lessonId).catch(() => null),
    ]).then(([course, progress]) => {
      if (active) {
        setCourse(course);
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
  const lessons = course?.chapters?.flatMap((chapter) => chapter.lessons) ?? [];
  const currentIndex = lessons.findIndex((item) => item.id === lesson.id);
  const previousLesson = currentIndex > 0 ? lessons[currentIndex - 1] : undefined;
  const nextLesson = currentIndex >= 0 && currentIndex < lessons.length - 1 ? lessons[currentIndex + 1] : undefined;
  const lessonPath = (id: string) => portalRoutePath(mode, tenantCode, `/courses/${courseId}/lessons/${id}`);
  const togglePlayback = () => {
    const video = videoRef.current;
    if (!video) return;
    if (video.paused) void video.play(); else video.pause();
  };

  return (
    <section className="page-section player-page">
      <div className="lesson-player-layout">
        <main className="player-workspace">
          <header className="player-topbar">
            <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate(portalRoutePath(mode, tenantCode, `/courses/${courseId}`))}>返回课程</Button>
            <strong>{lesson.title}</strong><span />
          </header>
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
                onTimeUpdate={(event) => { setCurrentTime(event.currentTarget.currentTime); report(event.currentTarget); }}
                onPlaying={() => { setPlaying(true); lifecycleRef.current?.playing(); }}
                onPause={() => { setPlaying(false); void lifecycleRef.current?.pause(); }}
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
          <div className="player-progress"><Progress percent={percent} showInfo={false} strokeColor="var(--learner-accent)" trailColor="var(--learner-player-bar)" /></div>
          <div className="player-controls">
            <div>
              <Button type="text" disabled={!previousLesson} onClick={() => previousLesson && navigate(lessonPath(previousLesson.id))}>上一课</Button>
              <Button className="player-play-button" shape="circle" icon={playing ? <PauseOutlined /> : <PlayCircleFilled />} onClick={togglePlayback} aria-label={playing ? '暂停' : '播放'} />
              <Button type="text" disabled={!nextLesson} onClick={() => nextLesson && navigate(lessonPath(nextLesson.id))}>下一课</Button>
            </div>
            <span className="player-time">{formatPlaybackPosition(currentTime)} / {formatPlaybackPosition(lesson.durationSeconds)}</span>
            <div><span className="player-speed">1.0x</span><SoundOutlined /><ExpandOutlined /></div>
          </div>
        </div>
        </main>
        <aside className="player-sidebar">
          <header><h1>{course?.title || '课程目录'}</h1><p>已完成 {Math.round((percent / 100) * lessons.length)} / {lessons.length} 个课时</p></header>
          <div className="player-sidebar-tabs" role="tablist">
            <button className={sidebarTab === 'outline' ? 'active' : ''} type="button" onClick={() => setSidebarTab('outline')}>课程目录</button>
            <button className={sidebarTab === 'materials' ? 'active' : ''} type="button" onClick={() => setSidebarTab('materials')}>资料</button>
          </div>
          {sidebarTab === 'outline' ? (
            <div className="player-lesson-list">
              {lessons.map((item, index) => (
                <Link className={`player-lesson-item${item.id === lesson.id ? ' active' : ''}`} to={lessonPath(item.id)} key={item.id}>
                  <span className={`player-lesson-number${index < currentIndex ? ' complete' : ''}`}>{index < currentIndex ? '✓' : index + 1}</span>
                  <span><strong>{item.title}</strong><small>{formatPlaybackPosition(item.durationSeconds)}</small></span><RightOutlined />
                </Link>
              ))}
            </div>
          ) : (
            <div className="player-materials">
              {(course?.materials ?? []).length ? course?.materials?.map((material) => <p key={material.id}><FileTextOutlined />{material.displayName}</p>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无课程资料" />}
            </div>
          )}
          <footer>
            <Button disabled={!previousLesson} onClick={() => previousLesson && navigate(lessonPath(previousLesson.id))}>上一课</Button>
            <Button type="primary" disabled={!nextLesson} onClick={() => nextLesson && navigate(lessonPath(nextLesson.id))}>下一课</Button>
          </footer>
        </aside>
      </div>
    </section>
  );
}
