import {
  ArrowLeftOutlined,
  DeleteOutlined,
  EditOutlined,
  FileTextOutlined,
  PlusOutlined,
  VideoCameraOutlined,
} from '@ant-design/icons'
import {
  Button,
  Card,
  Collapse,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
  message,
} from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { courseApi, type Chapter, type Course, type Lesson } from '../api/course'
import { resourceApi, type Resource } from '../api/resource'
import { normalizePage } from '../api/types'
import MediaUploader, {
  type UploadedMedia,
} from '../components/MediaUploader'
import PageHeader from '../components/PageHeader'
import {
  loadResourcePreview,
  type ResourcePreview,
} from '../utils/resourcePreview'

type Editor =
  | { kind: 'chapter'; chapter?: Chapter }
  | { kind: 'lesson'; chapter: Chapter; lesson?: Lesson }

type LessonForm = Omit<Lesson, 'id'> & {
  title: string
}

export default function CourseDetail() {
  const { id = '' } = useParams()
  const location = useLocation()
  const navigate = useNavigate()
  const officialMode = location.pathname.startsWith('/official-courses/')
  const [course, setCourse] = useState<Course>()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [editor, setEditor] = useState<Editor>()
  const [resources, setResources] = useState<Resource[]>([])
  const [selectedResource, setSelectedResource] = useState<UploadedMedia>()
  const [previewTarget, setPreviewTarget] = useState<UploadedMedia>()
  const [preview, setPreview] = useState<ResourcePreview>()
  const [previewLoading, setPreviewLoading] = useState(false)
  const [form] = Form.useForm<LessonForm>()
  const contentType = Form.useWatch('content_type', form)

  useEffect(() => () => preview?.dispose(), [preview])

  const load = async () => {
    setLoading(true)
    try {
      const { data } = await courseApi.detail(id)
      setCourse(data)
    } finally {
      setLoading(false)
    }
  }

  const loadResources = async () => {
    try {
      const request = officialMode
        ? resourceApi.listPlatform()
        : resourceApi.list()
      const { data } = await request
      setResources(normalizePage(data).items)
    } catch {
      setResources([])
    }
  }

  useEffect(() => {
    void load()
  }, [id])

  useEffect(() => {
    void loadResources()
  }, [officialMode])

  const matchingResources = useMemo(
    () => resources.filter((resource) => resource.resource_type === contentType),
    [contentType, resources],
  )

  const edit = (value: Editor) => {
    setEditor(value)
    if (value.kind === 'chapter') {
      form.setFieldsValue(value.chapter || { title: '' })
      setSelectedResource(undefined)
      return
    }
    const initial = value.lesson || {
      title: '',
      content_type: 'video' as const,
      duration_seconds: 0,
      sort_order: 0,
    }
    form.setFieldsValue(initial)
    const resource = value.lesson?.resource_id
      ? resources.find((item) => item.id === value.lesson?.resource_id)
      : undefined
    setSelectedResource(resource || (
      value.lesson?.resource_id
        ? {
            id: value.lesson.resource_id,
            name: '当前课时资源',
            resource_type: value.lesson.content_type === 'document'
              ? 'document'
              : 'video',
            url: '',
            size_bytes: 0,
          }
        : undefined
    ))
  }

  const closeEditor = () => {
    setEditor(undefined)
    setSelectedResource(undefined)
    form.resetFields()
  }

  const save = async () => {
    const values = await form.validateFields()
    if (!editor) return
    setSaving(true)
    try {
      if (editor.kind === 'chapter') {
        if (editor.chapter) {
          await courseApi.updateChapter(id, editor.chapter.id, values)
        } else {
          await courseApi.createChapter(id, values)
        }
      } else {
        const lessonValues: Omit<Lesson, 'id'> = {
          title: values.title,
          content_type: values.content_type,
          content_url: values.content_type === 'text'
            ? values.content_url || ''
            : '',
          resource_id: values.content_type === 'text'
            ? undefined
            : values.resource_id,
          duration_seconds: values.duration_seconds || 0,
          sort_order: values.sort_order || 0,
        }
        if (editor.lesson) {
          await courseApi.updateLesson(
            id,
            editor.chapter.id,
            editor.lesson.id,
            lessonValues,
          )
        } else {
          await courseApi.createLesson(id, editor.chapter.id, lessonValues)
        }
      }
      message.success('课程内容已保存')
      closeEditor()
      void load()
    } finally {
      setSaving(false)
    }
  }

  const uploadResource = async (
    file: File,
    onProgress: (percent: number) => void,
  ) => {
    const response = officialMode
      ? await resourceApi.uploadPlatform(file, onProgress)
      : await resourceApi.upload(file, onProgress)
    const resource = response.data
    setResources((current) => [
      resource,
      ...current.filter((item) => item.id !== resource.id),
    ])
    form.setFieldsValue({
      resource_id: resource.id,
      content_type: resource.resource_type === 'document'
        ? 'document'
        : 'video',
    })
    return resource
  }

  const previewResource = async (resource: UploadedMedia) => {
    setPreviewTarget(resource)
    setPreviewLoading(true)
    try {
      const loaded = await loadResourcePreview(resource, async () => {
        const response = officialMode
          ? await resourceApi.platformFile(resource.id)
          : await resourceApi.file(resource.id)
        return response.data
      })
      setPreview(loaded)
    } catch {
      setPreview(undefined)
      setPreviewTarget(undefined)
      message.error('资源预览加载失败，请稍后重试')
    } finally {
      setPreviewLoading(false)
    }
  }

  const closePreview = () => {
    setPreview(undefined)
    setPreviewTarget(undefined)
    setPreviewLoading(false)
  }

  const removeChapter = async (chapterID: string) => {
    await courseApi.removeChapter(id, chapterID)
    message.success('章节已删除')
    void load()
  }

  const removeLesson = async (chapterID: string, lessonID: string) => {
    await courseApi.removeLesson(id, chapterID, lessonID)
    message.success('课时已删除')
    void load()
  }

  if (loading) {
    return <div className="center-spin"><Spin size="large" /></div>
  }
  if (!course) return <Empty description="未找到课程" />

  return (
    <>
      <PageHeader
        title={course.title}
        description={officialMode
          ? '维护平台官方课程的章节、视频、PDF 和文本课时。'
          : '编辑课程章节结构与课时内容。'}
        extra={(
          <Space wrap>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate(
                officialMode ? '/official-courses' : '/courses',
              )}
            >
              返回列表
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => edit({ kind: 'chapter' })}
            >
              添加章节
            </Button>
          </Space>
        )}
      />
      <Card className="course-summary">
        <Space size={18} align="start">
          <div className="detail-cover">
            {course.cover_image
              ? <img src={course.cover_image} alt="" />
              : <FileTextOutlined />}
          </div>
          <div>
            <Tag color={course.status === 1 ? 'success' : 'default'}>
              {course.status === 1 ? '已发布' : '草稿'}
            </Tag>
            {officialMode && <Tag color="blue">官方课程</Tag>}
            <Typography.Paragraph className="course-description">
              {course.description || '暂未填写课程简介'}
            </Typography.Paragraph>
          </div>
        </Space>
      </Card>
      <div className="section-heading">
        <Typography.Title level={4}>课程目录</Typography.Title>
        <Typography.Text type="secondary">
          共 {course.chapters?.length || 0} 个章节
        </Typography.Text>
      </div>
      {course.chapters?.length ? (
        <Collapse
          className="chapter-collapse"
          defaultActiveKey={course.chapters.map((chapter) => chapter.id)}
          items={course.chapters.map((chapter, index) => ({
            key: chapter.id,
            label: <strong>第 {index + 1} 章　{chapter.title}</strong>,
            extra: (
              <Space onClick={(event) => event.stopPropagation()}>
                <Button
                  type="text"
                  size="small"
                  icon={<PlusOutlined />}
                  onClick={() => edit({ kind: 'lesson', chapter })}
                >
                  添加课时
                </Button>
                <Button
                  type="text"
                  size="small"
                  icon={<EditOutlined />}
                  onClick={() => edit({ kind: 'chapter', chapter })}
                />
                <Popconfirm
                  title="删除该章节及其课时？"
                  onConfirm={() => void removeChapter(chapter.id)}
                >
                  <Button
                    type="text"
                    danger
                    size="small"
                    icon={<DeleteOutlined />}
                  />
                </Popconfirm>
              </Space>
            ),
            children: chapter.lessons?.length ? (
              chapter.lessons.map((lesson, lessonIndex) => (
                <div className="lesson-row" key={lesson.id}>
                  <Space>
                    <span className="lesson-index">{lessonIndex + 1}</span>
                    {lesson.content_type === 'video'
                      ? <VideoCameraOutlined />
                      : <FileTextOutlined />}
                    <span>{lesson.title}</span>
                    {lesson.duration_seconds ? (
                      <Tag>{Math.ceil(lesson.duration_seconds / 60)} 分钟</Tag>
                    ) : null}
                  </Space>
                  <Space>
                    <Button
                      type="link"
                      onClick={() => edit({
                        kind: 'lesson',
                        chapter,
                        lesson,
                      })}
                    >
                      编辑
                    </Button>
                    <Popconfirm
                      title="确认删除该课时？"
                      onConfirm={() =>
                        void removeLesson(chapter.id, lesson.id)}
                    >
                      <Button type="link" danger>删除</Button>
                    </Popconfirm>
                  </Space>
                </div>
              ))
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="暂无课时"
              />
            ),
          }))}
        />
      ) : (
        <Card><Empty description="暂无章节，请先添加章节" /></Card>
      )}
      <Modal
        title={editor?.kind === 'chapter'
          ? `${editor.chapter ? '编辑' : '添加'}章节`
          : `${editor?.lesson ? '编辑' : '添加'}课时`}
        open={Boolean(editor)}
        width={720}
        confirmLoading={saving}
        onCancel={closeEditor}
        onOk={() => void save()}
        destroyOnHidden
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item
            label={editor?.kind === 'chapter' ? '章节标题' : '课时标题'}
            name="title"
            rules={[{ required: true, message: '请输入标题' }]}
          >
            <Input />
          </Form.Item>
          {editor?.kind === 'lesson' && (
            <>
              <Form.Item
                label="内容类型"
                name="content_type"
                rules={[{ required: true }]}
              >
                <Select
                  options={[
                    { value: 'video', label: '视频' },
                    { value: 'document', label: 'PDF 文档' },
                    { value: 'text', label: '文本' },
                  ]}
                  onChange={(nextType) => {
                    if (nextType === 'text' ||
                      selectedResource?.resource_type !== nextType) {
                      setSelectedResource(undefined)
                      form.setFieldValue('resource_id', undefined)
                    }
                  }}
                />
              </Form.Item>
              {contentType === 'text' ? (
                <Form.Item
                  label="课时正文"
                  name="content_url"
                  rules={[{ required: true, message: '请输入课时正文' }]}
                >
                  <Input.TextArea
                    rows={8}
                    maxLength={20_000}
                    showCount
                    placeholder="输入学员需要阅读的文本内容"
                  />
                </Form.Item>
              ) : (
                <>
                  <Form.Item
                    label={contentType === 'document'
                      ? '上传 PDF'
                      : '上传视频'}
                    extra={officialMode
                      ? '文件将保存为平台共享资源'
                      : '文件将保存到当前租户资源库'}
                  >
                    <MediaUploader
                      value={selectedResource}
                      accept={contentType === 'document'
                        ? 'document'
                        : 'video'}
                      upload={uploadResource}
                      onPreview={previewResource}
                      onVideoDuration={(seconds) => {
                        form.setFieldValue('duration_seconds', seconds)
                      }}
                      onChange={(resource) => {
                        setSelectedResource(resource)
                        form.setFieldValue(
                          'resource_id',
                          resource?.id,
                        )
                      }}
                    />
                  </Form.Item>
                  <Form.Item
                    label="或复用已有资源"
                    name="resource_id"
                    rules={[{ required: true, message: '请上传或选择资源' }]}
                  >
                    <Select
                      allowClear
                      showSearch
                      optionFilterProp="label"
                      placeholder="从资源库选择"
                      options={matchingResources.map((resource) => ({
                        value: resource.id,
                        label: `${resource.name}（${resource.resource_type}）`,
                      }))}
                      onChange={(resourceID) => {
                        const resource = resources.find(
                          (item) => item.id === resourceID,
                        )
                        setSelectedResource(resource)
                      }}
                    />
                  </Form.Item>
                  <Form.Item
                    label="时长（秒）"
                    name="duration_seconds"
                  >
                    <InputNumber min={0} style={{ width: '100%' }} />
                  </Form.Item>
                </>
              )}
              <Form.Item label="排序" name="sort_order">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>
      <Modal
        title={`预览：${previewTarget?.name || ''}`}
        open={Boolean(previewTarget)}
        width={960}
        footer={null}
        onCancel={closePreview}
        destroyOnHidden
      >
        {previewLoading ? (
          <div className="center-spin"><Spin size="large" /></div>
        ) : preview?.resourceType === 'video' ? (
          <video
            src={preview.url}
            controls
            style={{
              display: 'block',
              width: '100%',
              maxHeight: '70vh',
              background: '#000',
            }}
          />
        ) : preview ? (
          <iframe
            src={preview.url}
            title={preview.name}
            style={{
              display: 'block',
              width: '100%',
              height: '70vh',
              border: 0,
            }}
          />
        ) : null}
      </Modal>
    </>
  )
}
