import { BookOutlined, CheckCircleFilled } from '@ant-design/icons';
import { Progress } from 'antd';
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
      <div className="learner-course-cover">
        {course.coverImage
          ? <img src={course.coverImage} alt={`${course.title}课程封面`} />
          : <BookOutlined className="learner-course-fallback-icon" aria-hidden="true" />}
        <span className="learner-course-pattern" />
        <span className="course-type-badge">{course.courseType === 'required' ? '必修' : '选修'}</span>
        <span className="course-duration-badge">{course.lessonCount} 课时</span>
      </div>
      <div className="learner-course-body">
        <h2 title={course.title}>{course.title}</h2>
        <p className="course-card-category">{course.category?.name || '通用课程'}</p>
        <div className="course-progress-row" aria-label={`学习进度 ${course.progressPercent}%`}>
          <Progress className="course-progress-bar" percent={course.progressPercent} showInfo={false} size="small" strokeColor="var(--learner-accent)" trailColor="var(--learner-line)" />
          <span className={completed ? 'course-progress-complete' : ''}>{completed && <CheckCircleFilled />} {completed ? '已完成' : `${course.progressPercent}%`}</span>
        </div>
      </div>
    </article>
  );
}
