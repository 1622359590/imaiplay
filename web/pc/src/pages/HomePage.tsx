import { ArrowRightOutlined, PlayCircleFilled, ReloadOutlined } from '@ant-design/icons';
import { Button, Result, Skeleton } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { getLearnerOverview, type LearnerOverview } from '../api/learner';
import { CourseGrid } from '../components/CourseGrid';
import { LearnerFilters } from '../components/LearnerFilters';
import { LearningSummary } from '../components/LearningSummary';
import { LearnerMotivationPrompt } from '../components/LearnerMotivationPrompt';
import {
  filterLearnerCourses,
  type LearnerCourseTab,
} from '../utils/learnerCourses';
import { usePortal } from '../context/PortalContext';
import { portalRoutePath } from '../utils/portalRouting';

export function HomePage() {
  const { mode, tenantCode } = usePortal();
  const [requestVersion, setRequestVersion] = useState(0);
  const [overview, setOverview] = useState<LearnerOverview>();
  const [error, setError] = useState<unknown>();
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<LearnerCourseTab>('all');
  const [categoryId, setCategoryId] = useState<string>();

  useEffect(() => {
    let active = true;
    setOverview(undefined);
    setError(undefined);
    setLoading(true);
    void getLearnerOverview()
      .then((result) => {
        if (active) setOverview(result);
      })
      .catch((requestError: unknown) => {
        if (active) setError(requestError);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => { active = false; };
  }, [requestVersion]);

  const filteredCourses = useMemo(
    () => overview ? filterLearnerCourses(overview.courses, { tab, categoryId }) : [],
    [categoryId, overview, tab],
  );
  const continueCourse = overview?.courses.find((course) => course.recentLesson && course.progressPercent < 100);
  const continuePath = continueCourse?.recentLesson
    ? portalRoutePath(mode, tenantCode, `/courses/${continueCourse.id}/lessons/${continueCourse.recentLesson.id}`)
    : undefined;

  return (
    <section className="course-home page-section" aria-label="学习首页">
      <LearnerMotivationPrompt enabled={!loading && !error && Boolean(overview)} />
      {loading && (
        <div className="learner-dashboard-loading" aria-busy="true" aria-label="正在加载学习首页">
          <div className="learning-summary">
            {[1, 2].map((key) => (
              <div className="learning-summary-card learning-summary-skeleton" key={key}>
                <Skeleton active paragraph={{ rows: 1 }} />
              </div>
            ))}
          </div>
          <div className="learner-filter-skeleton"><Skeleton.Button active block /></div>
          <CourseGrid courses={[]} loading emptyText="" />
        </div>
      )}

      {!loading && Boolean(error) && (
        <Result
          className="learner-request-state"
          status="error"
          title="学习首页加载失败"
          subTitle="请检查网络后重试"
          extra={(
            <Button type="primary" icon={<ReloadOutlined />} onClick={() => setRequestVersion((version) => version + 1)}>
              重新加载
            </Button>
          )}
        />
      )}

      {!loading && !error && overview && (
        <>
          <header className="learner-welcome">
            <h1>你好，<span>学习者</span></h1>
            <p>继续你的学习旅程，让每一次投入都成为成长的积累。</p>
          </header>
          <LearningSummary
            completed={overview.requiredCompleted}
            required={overview.requiredTotal}
            todaySeconds={overview.todayLearningSeconds}
            totalSeconds={overview.totalLearningSeconds}
          />
          {continueCourse && continuePath && (
            <section className="continue-learning-banner" aria-label="继续学习">
              <div className="continue-learning-copy">
                <span>继续上次学习</span>
                <h2>{continueCourse.title}</h2>
                <p>{continueCourse.recentLesson?.title} · 已学习 {continueCourse.progressPercent}%</p>
              </div>
              <Link className="continue-learning-action" to={continuePath}>
                <PlayCircleFilled /><span>继续学习</span><ArrowRightOutlined />
              </Link>
            </section>
          )}
          <div className="course-section-heading" id="courses">
            <div><span>LEARNING LIBRARY</span><h2>我的课程</h2></div>
            <p>按你的节奏，完成每一段学习旅程</p>
          </div>
          <LearnerFilters
            tab={tab}
            categoryId={categoryId}
            categories={overview.categories}
            onTabChange={setTab}
            onCategoryChange={setCategoryId}
          />
          <CourseGrid
            courses={filteredCourses}
            loading={false}
            emptyText="当前筛选下暂无课程，请调整筛选条件"
          />
        </>
      )}
    </section>
  );
}
