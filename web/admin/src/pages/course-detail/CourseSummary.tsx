import { FileTextOutlined } from '@ant-design/icons'
import { Card, Space, Tag, Typography } from 'antd'
import CourseMaterialsManager from '../../components/CourseMaterialsManager'
import type { CourseDetailController } from './useCourseDetail'

export default function CourseSummary({ controller }: { controller: CourseDetailController }) {
  const { course, officialMode, instructor, updateMaterials } = controller
  if (!course) return null
  return <>
    <Card className="course-summary course-summary-clay">
      <Space size={18} align="start">
        <div className="detail-cover">{course.cover_image ? <img src={course.cover_image} alt="" /> : <FileTextOutlined />}</div>
        <div className="course-summary-copy">
          <Tag color={course.status === 1 ? 'success' : 'default'}>{course.status === 1 ? '已发布' : '草稿'}</Tag>
          {officialMode && <Tag color="blue">官方课程</Tag>}
          <Typography.Paragraph className="course-description">{course.description || '暂未填写课程简介'}</Typography.Paragraph>
        </div>
      </Space>
    </Card>
    <CourseMaterialsManager
      key={course.id}
      courseId={course.id}
      officialMode={officialMode}
      initialMaterials={course.materials || []}
      readOnly={instructor}
      onChange={updateMaterials}
    />
  </>
}
