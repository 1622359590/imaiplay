import { ArrowRightOutlined, BookOutlined, ClockCircleOutlined, TrophyOutlined } from '@ant-design/icons';
import { Button, Col, Row, Statistic, Typography } from 'antd';
import { useNavigate } from 'react-router-dom';
import { getCourses } from '../api/course';
import { CourseGrid } from '../components/CourseGrid';
import { useCourseList } from '../hooks/useCourseList';
import { useTenantTheme } from '../context/TenantThemeContext';

export function HomePage() {
  const navigate = useNavigate();
  const { courses, loading } = useCourseList(getCourses);
  const learning = courses.filter((course) => (course.progress ?? 0) > 0);
  const theme = useTenantTheme();

  return (
    <>
      <section className="hero-panel">
        <div>
          <Typography.Text className="hero-eyebrow">LEARNING FOR GROWTH</Typography.Text>
          <Typography.Title>今天，也向更好的自己迈进一步</Typography.Title>
          <Typography.Paragraph>{theme.welcome_text || '探索专业课程，按照自己的节奏学习，让知识真正转化为工作能力。'}</Typography.Paragraph>
          <Button type="primary" size="large" onClick={() => navigate('/courses')}>
            浏览全部课程 <ArrowRightOutlined />
          </Button>
        </div>
        <div className="hero-visual">
          <div className="hero-orbit orbit-one" />
          <div className="hero-orbit orbit-two" />
          <div className="hero-icon"><TrophyOutlined /></div>
        </div>
      </section>

      <Row gutter={[20, 20]} className="stat-row">
        <Col xs={24} sm={8}>
          <div className="stat-card">
            <Statistic title="可学课程" value={courses.length} prefix={<BookOutlined />} />
          </div>
        </Col>
        <Col xs={24} sm={8}>
          <div className="stat-card">
            <Statistic title="正在学习" value={learning.length} prefix={<ClockCircleOutlined />} />
          </div>
        </Col>
        <Col xs={24} sm={8}>
          <div className="stat-card">
            <Statistic
              title="已完成课程"
              value={courses.filter((course) => course.progress === 100).length}
              prefix={<TrophyOutlined />}
            />
          </div>
        </Col>
      </Row>

      <section className="content-section">
        <div className="section-heading">
          <div>
            <Typography.Title level={2}>精选课程</Typography.Title>
            <Typography.Text type="secondary">从优质内容开始今天的学习</Typography.Text>
          </div>
          <Button type="link" onClick={() => navigate('/courses')}>
            查看全部 <ArrowRightOutlined />
          </Button>
        </div>
        <CourseGrid courses={courses.slice(0, 8)} loading={loading} />
      </section>
    </>
  );
}
