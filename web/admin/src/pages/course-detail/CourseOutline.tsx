import { DeleteOutlined, EditOutlined, FileTextOutlined, PlusOutlined, VideoCameraOutlined } from '@ant-design/icons'
import { Button, Card, Collapse, Empty, Popconfirm, Space, Tag, Typography } from 'antd'
import type { CourseDetailController } from './useCourseDetail'

export default function CourseOutline({ controller }: { controller: CourseDetailController }) {
  const { course } = controller
  if (!course) return null
  return <>
    <div className="section-heading">
      <Typography.Title level={4}>课程目录</Typography.Title>
      <Typography.Text type="secondary">共 {course.chapters?.length || 0} 个章节</Typography.Text>
    </div>
    {course.chapters?.length ? <Collapse
      className="chapter-collapse"
      defaultActiveKey={course.chapters.map((chapter) => chapter.id)}
      items={course.chapters.map((chapter, index) => ({
        key: chapter.id,
        label: <strong>第 {index + 1} 章　{chapter.title}</strong>,
        extra: <Space onClick={(event) => event.stopPropagation()}>
          <Button type="text" size="small" icon={<PlusOutlined />} onClick={() => controller.edit({ kind: 'lesson', chapter })}>添加课时</Button>
          <Button type="text" size="small" icon={<EditOutlined />} onClick={() => controller.edit({ kind: 'chapter', chapter })} />
          <Popconfirm title="删除该章节及其课时？" onConfirm={() => void controller.removeChapter(chapter.id)}>
            <Button type="text" danger size="small" icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>,
        children: chapter.lessons?.length ? chapter.lessons.map((lesson, lessonIndex) => <div className="lesson-row" key={lesson.id}>
          <Space>
            <span className="lesson-index">{lessonIndex + 1}</span>
            {lesson.content_type === 'video' ? <VideoCameraOutlined /> : <FileTextOutlined />}
            <span>{lesson.title}</span>
            {lesson.duration_seconds ? <Tag>{Math.ceil(lesson.duration_seconds / 60)} 分钟</Tag> : null}
          </Space>
          <Space>
            <Button type="link" onClick={() => controller.edit({ kind: 'lesson', chapter, lesson })}>编辑</Button>
            <Popconfirm title="确认删除该课时？" onConfirm={() => void controller.removeLesson(chapter.id, lesson.id)}><Button type="link" danger>删除</Button></Popconfirm>
          </Space>
        </div>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无课时" />,
      }))}
    /> : <Card><Empty description="暂无章节，请先添加章节" /></Card>}
  </>
}
