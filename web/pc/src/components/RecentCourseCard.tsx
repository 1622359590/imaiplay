import {
  BookOutlined,
  ClockCircleOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import { Progress, Tag } from 'antd';
import dayjs from 'dayjs';
import { Link } from 'react-router-dom';
import type { RecentLearningItem } from '../api/learner';
import { usePortal } from '../context/PortalContext';
import { formatPlaybackPosition } from '../utils/learnerCourses';
import { portalRoutePath } from '../utils/portalRouting';

interface RecentCourseCardProps {
  item: RecentLearningItem;
}

export function RecentCourseCard({ item }: RecentCourseCardProps) {
  const { mode, tenantCode } = usePortal();
  const continuePath = portalRoutePath(
    mode,
    tenantCode,
    `/courses/${item.course.id}/lessons/${item.recentLesson.id}`,
  );
  const learnedAt = dayjs(item.lastLearnedAt);

  return (
    <article className={`recent-course-card${item.course.coverImage ? ' recent-course-card-with-cover' : ''}`}>
      {item.course.coverImage && (
        <div className="recent-course-cover">
          <img src={item.course.coverImage} alt={`${item.course.title}课程封面`} />
        </div>
      )}
      {!item.course.coverImage && (
        <span className="recent-course-fallback-icon" aria-hidden="true"><BookOutlined /></span>
      )}
      <div className="recent-course-content">
        <div className="recent-course-heading">
          <div>
            <h2 title={item.course.title}>{item.course.title}</h2>
            {item.course.category && <Tag className="course-type-tag">{item.course.category.name}</Tag>}
          </div>
          <span className="recent-learned-at">
            <ClockCircleOutlined aria-hidden="true" />
            {learnedAt.isValid() ? learnedAt.format('YYYY-MM-DD HH:mm') : '最近学习'}
          </span>
        </div>
        <p className="recent-lesson-title">最近学习：{item.recentLesson.title}</p>
        <p className="recent-position">上次学习到 {formatPlaybackPosition(item.lastPositionSeconds)}</p>
        <div className="recent-course-footer">
          <Progress
            className="recent-course-progress"
            percent={item.progressPercent}
            size="small"
            strokeColor="var(--learner-accent)"
            trailColor="var(--learner-line)"
          />
          <Link className="recent-continue-link" to={continuePath}>
            <PlayCircleOutlined aria-hidden="true" />
            <span>继续学习</span>
          </Link>
        </div>
      </div>
    </article>
  );
}
