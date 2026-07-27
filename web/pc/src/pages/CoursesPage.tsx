import { SearchOutlined } from '@ant-design/icons';
import { Input, Typography } from 'antd';
import { useMemo, useState } from 'react';
import { getCourses } from '../api/course';
import { CourseGrid } from '../components/CourseGrid';
import { useCourseList } from '../hooks/useCourseList';

export function CoursesPage() {
  const [keyword, setKeyword] = useState('');
  const { courses, loading } = useCourseList(getCourses);
  const filtered = useMemo(() => {
    const query = keyword.trim().toLowerCase();
    if (!query) return courses;
    return courses.filter(
      (course) =>
        course.title.toLowerCase().includes(query) ||
        course.description?.toLowerCase().includes(query) ||
        course.category?.toLowerCase().includes(query),
    );
  }, [courses, keyword]);

  return (
    <section className="content-section page-section">
      <div className="section-heading course-list-heading">
        <div>
          <Typography.Title level={1}>全部课程</Typography.Title>
          <Typography.Text type="secondary">找到适合你的课程，开启新的学习计划</Typography.Text>
        </div>
        <Input
          allowClear
          size="large"
          prefix={<SearchOutlined />}
          placeholder="搜索课程"
          value={keyword}
          onChange={(event) => setKeyword(event.target.value)}
        />
      </div>
      <CourseGrid
        courses={filtered}
        loading={loading}
        emptyText={keyword ? '没有找到匹配的课程' : '暂无课程'}
      />
    </section>
  );
}
