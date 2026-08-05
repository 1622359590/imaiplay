import { ReloadOutlined } from '@ant-design/icons';
import { Button, Empty, Result, Skeleton } from 'antd';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getRecentLearning, type RecentLearningPage } from '../api/learner';
import { RecentCourseCard } from '../components/RecentCourseCard';
import { usePortal } from '../context/PortalContext';
import { portalRoutePath } from '../utils/portalRouting';

export function RecentPage() {
  const navigate = useNavigate();
  const { mode, tenantCode } = usePortal();
  const [requestVersion, setRequestVersion] = useState(0);
  const [page, setPage] = useState<RecentLearningPage>();
  const [error, setError] = useState<unknown>();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(undefined);
    void getRecentLearning()
      .then((result) => {
        if (active) setPage(result);
      })
      .catch((requestError: unknown) => {
        if (active) {
          setPage(undefined);
          setError(requestError);
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => { active = false; };
  }, [requestVersion]);

  return (
    <section className="recent-page page-section" aria-labelledby="recent-page-title">
      <header className="learner-page-heading">
        <div>
          <h1 id="recent-page-title">最近学习</h1>
          <p>继续上次的学习进度</p>
        </div>
        {!loading && page && <span>共 {page.total} 门课程</span>}
      </header>

      {loading && (
        <div className="recent-course-grid" aria-busy="true" aria-label="正在加载最近学习">
          {[1, 2].map((key) => (
            <div className="recent-course-skeleton" key={key}><Skeleton active /></div>
          ))}
        </div>
      )}

      {!loading && Boolean(error) && (
        <Result
          className="learner-request-state"
          status="error"
          title="最近学习加载失败"
          subTitle="请检查网络后重试"
          extra={(
            <Button type="primary" icon={<ReloadOutlined />} onClick={() => setRequestVersion((version) => version + 1)}>
              重新加载
            </Button>
          )}
        />
      )}

      {!loading && !error && page?.items.length === 0 && (
        <Empty
          className="page-empty"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="还没有学习记录，去首页选择课程吧"
        >
          <Button type="primary" onClick={() => navigate(portalRoutePath(mode, tenantCode, '/'))}>返回首页</Button>
        </Empty>
      )}

      {!loading && !error && page && page.items.length > 0 && (
        <div className="recent-course-grid">
          {page.items.map((item) => <RecentCourseCard item={item} key={item.course.id} />)}
        </div>
      )}
    </section>
  );
}
