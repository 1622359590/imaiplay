import { ReloadOutlined } from '@ant-design/icons';
import { Button, Result, Skeleton } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { getLearnerOverview, type LearnerOverview } from '../api/learner';
import { CourseGrid } from '../components/CourseGrid';
import { LearnerFilters } from '../components/LearnerFilters';
import { LearningSummary } from '../components/LearningSummary';
import {
  filterLearnerCourses,
  type LearnerCourseTab,
} from '../utils/learnerCourses';

export function HomePage() {
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

  return (
    <section className="course-home page-section" aria-label="学习首页">
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
          <LearningSummary
            completed={overview.requiredCompleted}
            required={overview.requiredTotal}
            todaySeconds={overview.todayLearningSeconds}
            totalSeconds={overview.totalLearningSeconds}
          />
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
