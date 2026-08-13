import { ReloadOutlined } from '@ant-design/icons';
import { Button, Result, Skeleton } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { getLearnerOverview, type LearnerOverview } from '../api/learner';
import { CourseGrid } from '../components/CourseGrid';
import { LearnerFilters } from '../components/LearnerFilters';
import {
  filterLearnerCourses,
  type LearnerCourseTab,
} from '../utils/learnerCourses';

export function CourseListPage() {
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
    <section className="course-catalog-page page-section" aria-labelledby="course-catalog-title">
      <header className="learner-page-heading">
        <div>
          <span>LEARNING LIBRARY</span>
          <h1 id="course-catalog-title">全部课程</h1>
          <p>浏览企业课程，按类型和分类快速找到学习内容。</p>
        </div>
        {overview && <span>{overview.courses.length} 门课程</span>}
      </header>

      {loading && (
        <div aria-busy="true" aria-label="正在加载全部课程">
          <div className="learner-filter-skeleton"><Skeleton.Button active block /></div>
          <CourseGrid courses={[]} loading emptyText="" />
        </div>
      )}

      {!loading && Boolean(error) && (
        <Result
          className="learner-request-state"
          status="error"
          title="全部课程加载失败"
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
