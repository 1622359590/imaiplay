import { BookOutlined, TrophyFilled } from '@ant-design/icons';
import { Progress, Tag } from 'antd';
import { useNavigate } from 'react-router-dom';
import type { LearnerCourse } from '../api/learner';
import { usePortal } from '../context/PortalContext';
import { courseStatus } from '../utils/learnerCourses';
import { portalRoutePath } from '../utils/portalRouting';

interface CourseCardProps {
  course: LearnerCourse;
}

export function CourseCard({ course }: CourseCardProps) {
  const navigate = useNavigate();
  const { mode, tenantCode } = usePortal();
  const completed = courseStatus(course) === 'completed';
  const openCourse = () => navigate(portalRoutePath(mode, tenantCode, `/courses/${course.id}`));

  return (
    <article
      aria-label={`查看课程：${course.title}`}
      className={`course-card learner-course-card${course.coverImage ? ' learner-course-card-with-cover' : ''}`}
      onClick={openCourse}
      onKeyDown={(event) => {
        if (event.key !== 'Enter' && event.key !== ' ') return;
        event.preventDefault();
        openCourse();
      }}
      role="link"
      tabIndex={0}
    >
      {course.coverImage && (
        <div className="learner-course-cover">
          <img src={course.coverImage} alt={`${course.title}课程封面`} />
        </div>
      )}
      {!course.coverImage && (
        <span className="learner-course-fallback-icon" aria-hidden="true"><BookOutlined /></span>
      )}
      <div className="learner-course-body">
        <h2 title={course.title}>{course.title}</h2>
		<div className="course-card-tags">
		  <Tag className="course-type-tag">
			{course.courseType === 'required' ? '必修课' : '选修课'}
		  </Tag>
		  {course.category && <Tag className="course-category-tag">{course.category.name}</Tag>}
		</div>
        {completed ? (
          <p className="course-complete-message">
            <TrophyFilled aria-hidden="true" />
            <span>恭喜你学完此课程！</span>
          </p>
        ) : (
          <div className="course-progress-row" aria-label={`学习进度 ${course.progressPercent}%`}>
            <Progress
              className="course-progress-bar"
              percent={course.progressPercent}
              showInfo={false}
              size="small"
              strokeColor="var(--learner-accent)"
              trailColor="var(--learner-line)"
            />
            <span>{course.progressPercent}%</span>
          </div>
        )}
      </div>
    </article>
  );
}
