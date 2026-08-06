import { ArrowLeftOutlined, PlusOutlined } from '@ant-design/icons'
import { Button, Empty, Space, Spin } from 'antd'
import { useNavigate } from 'react-router-dom'
import PageHeader from '../../components/PageHeader'
import CourseOutline from './CourseOutline'
import CourseSummary from './CourseSummary'
import EnrollmentManager from './EnrollmentManager'
import LessonEditor from './LessonEditor'
import ResourcePreviewModal from './ResourcePreviewModal'
import { useCourseDetail } from './useCourseDetail'

export default function CourseDetailPage() {
  const controller = useCourseDetail()
  const navigate = useNavigate()
  const { course, loading, officialMode } = controller
  if (loading) return <div className="center-spin"><Spin size="large" /></div>
  if (!course) return <Empty description="未找到课程" />

  return <>
    <PageHeader
      title={course.title}
      description={officialMode ? '维护平台官方课程的章节、视频、PDF 和文本课时。' : '编辑课程章节结构与课时内容。'}
      extra={<Space wrap>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(officialMode ? '/official-courses' : '/courses')}>返回列表</Button>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => controller.edit({ kind: 'chapter' })}>添加章节</Button>
      </Space>}
    />
    <CourseSummary controller={controller} />
    <EnrollmentManager controller={controller} />
    <CourseOutline controller={controller} />
    <LessonEditor controller={controller} />
    <ResourcePreviewModal controller={controller} />
  </>
}
