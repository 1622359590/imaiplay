import { ArrowLeftOutlined, DeleteOutlined, EditOutlined, FileTextOutlined, PlusOutlined, VideoCameraOutlined } from '@ant-design/icons'
import { Button, Card, Collapse, Empty, Form, Input, InputNumber, Modal, Popconfirm, Select, Space, Spin, Tag, Typography, message } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { courseApi, type Chapter, type Course, type Lesson } from '../api/course'
import PageHeader from '../components/PageHeader'

type Editor =
  | { kind: 'chapter'; chapter?: Chapter }
  | { kind: 'lesson'; chapter: Chapter; lesson?: Lesson }

export default function CourseDetail() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const [course, setCourse] = useState<Course>()
  const [loading, setLoading] = useState(true)
  const [editor, setEditor] = useState<Editor>()
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const { data } = await courseApi.detail(id)
      setCourse(data)
    } finally { setLoading(false) }
  }
  useEffect(() => { void load() }, [id])

  const edit = (value: Editor) => {
    setEditor(value)
    form.setFieldsValue(value.kind === 'chapter' ? value.chapter : value.lesson || { content_type: 'video' })
  }
  const save = async () => {
    const values = await form.validateFields()
    if (!editor) return
    if (editor.kind === 'chapter') {
      if (editor.chapter) await courseApi.updateChapter(id, editor.chapter.id, values)
      else await courseApi.createChapter(id, values)
    } else if (editor.lesson) {
      await courseApi.updateLesson(id, editor.chapter.id, editor.lesson.id, values)
    } else {
      await courseApi.createLesson(id, editor.chapter.id, values)
    }
    message.success('课程内容已保存')
    setEditor(undefined)
    form.resetFields()
    void load()
  }
  const removeChapter = async (chapterId: string) => {
    await courseApi.removeChapter(id, chapterId)
    message.success('章节已删除')
    void load()
  }
  const removeLesson = async (chapterId: string, lessonId: string) => {
    await courseApi.removeLesson(id, chapterId, lessonId)
    message.success('课时已删除')
    void load()
  }

  if (loading) return <div className="center-spin"><Spin size="large" /></div>
  if (!course) return <Empty description="未找到课程" />

  return (
    <>
      <PageHeader
        title={course.title}
        description="编辑课程章节结构与课时内容。"
        extra={<Space><Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/courses')}>返回列表</Button><Button type="primary" icon={<PlusOutlined />} onClick={() => edit({ kind: 'chapter' })}>添加章节</Button></Space>}
      />
      <Card className="course-summary">
        <Space size={18}>
          <div className="detail-cover">{course.cover_image ? <img src={course.cover_image} alt="" /> : <FileTextOutlined />}</div>
          <div>
            <Tag color={course.status === 1 ? 'success' : 'default'}>{course.status === 1 ? '已发布' : '编辑中'}</Tag>
            <Typography.Paragraph className="course-description">{course.description || '暂未填写课程简介'}</Typography.Paragraph>
          </div>
        </Space>
      </Card>
      <div className="section-heading"><Typography.Title level={4}>课程目录</Typography.Title><Typography.Text type="secondary">共 {course.chapters?.length || 0} 个章节</Typography.Text></div>
      {course.chapters?.length ? (
        <Collapse
          className="chapter-collapse"
          defaultActiveKey={course.chapters.map((chapter) => chapter.id)}
          items={course.chapters.map((chapter, index) => ({
            key: chapter.id,
            label: <strong>第 {index + 1} 章　{chapter.title}</strong>,
            extra: <Space onClick={(event) => event.stopPropagation()}><Button type="text" size="small" icon={<PlusOutlined />} onClick={() => edit({ kind: 'lesson', chapter })}>添加课时</Button><Button type="text" size="small" icon={<EditOutlined />} onClick={() => edit({ kind: 'chapter', chapter })} /><Popconfirm title="删除该章节及其课时？" onConfirm={() => removeChapter(chapter.id)}><Button type="text" danger size="small" icon={<DeleteOutlined />} /></Popconfirm></Space>,
            children: chapter.lessons?.length ? chapter.lessons.map((lesson, lessonIndex) => (
              <div className="lesson-row" key={lesson.id}>
                <Space><span className="lesson-index">{lessonIndex + 1}</span>{lesson.content_type === 'video' ? <VideoCameraOutlined /> : <FileTextOutlined />}<span>{lesson.title}</span>{lesson.duration_seconds ? <Tag>{Math.ceil(lesson.duration_seconds / 60)} 分钟</Tag> : null}</Space>
                <Space><Button type="link" onClick={() => edit({ kind: 'lesson', chapter, lesson })}>编辑</Button><Popconfirm title="确认删除该课时？" onConfirm={() => removeLesson(chapter.id, lesson.id)}><Button type="link" danger>删除</Button></Popconfirm></Space>
              </div>
            )) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无课时" />,
          }))}
        />
      ) : <Card><Empty description="暂无章节，请先添加章节" /></Card>}
      <Modal title={editor?.kind === 'chapter' ? `${editor.chapter ? '编辑' : '添加'}章节` : `${editor?.lesson ? '编辑' : '添加'}课时`} open={Boolean(editor)} onCancel={() => setEditor(undefined)} onOk={save} destroyOnHidden>
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item label={editor?.kind === 'chapter' ? '章节标题' : '课时标题'} name="title" rules={[{ required: true, message: '请输入标题' }]}><Input /></Form.Item>
          {editor?.kind === 'lesson' && <>
            <Form.Item label="内容类型" name="content_type" rules={[{ required: true }]}><Select options={[{ value: 'video', label: '视频' }, { value: 'document', label: '文档' }, { value: 'text', label: '文本' }]} /></Form.Item>
            <Form.Item label="时长（秒）" name="duration_seconds"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          </>}
        </Form>
      </Modal>
    </>
  )
}
