import { BookOutlined, DeleteOutlined, EditOutlined, EyeOutlined, PlusOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { courseApi, type Course, type CourseInput } from '../api/course'
import { courseCategoryApi, type CourseCategory } from '../api/courseCategory'
import { normalizePage } from '../api/types'
import PageHeader from '../components/PageHeader'
import { resourceApi } from '../api/resource'
import MediaUploader, { type UploadedMedia } from '../components/MediaUploader'
import OfficialCoursePicker from '../components/OfficialCoursePicker'
import type { RootState } from '../store'
import { categoryIDForPayload, normalizeCourseEditValues } from '../utils/adminFormValues'
import { consumeOneShotAction } from '../utils/oneShotAction'

type CourseForm = Omit<CourseInput, 'cover_image' | 'is_official' | 'category_id'> & {
  category_id?: string
  cover?: UploadedMedia
}

function currentCover(course: Course): UploadedMedia | undefined {
  if (!course.cover_image) return undefined
  return {
    id: `cover-${course.id}`,
    name: `${course.title}封面`,
    resource_type: 'image',
    url: course.cover_image,
    size_bytes: 0,
  }
}

const statusOptions = [
  { value: 0, label: '草稿' },
  { value: 1, label: '已发布' },
]

export default function Courses() {
  const [items, setItems] = useState<Course[]>([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Course>()
  const [officialPickerOpen, setOfficialPickerOpen] = useState(false)
  const [categories, setCategories] = useState<CourseCategory[]>([])
  const [form] = Form.useForm<CourseForm>()
  const navigate = useNavigate()
  const location = useLocation()
  const instructor = useSelector((state: RootState) => state.user.profile?.role === 'instructor')

  const load = async (current = pagination.current, pageSize = pagination.pageSize) => {
    setLoading(true)
    try {
      const { data } = await courseApi.list({ page: current, page_size: pageSize })
      const page = normalizePage(data)
      setItems(page.items)
      setPagination({ current, pageSize, total: page.total })
    } finally { setLoading(false) }
  }
  useEffect(() => {
    void load()
    void courseCategoryApi.list().then(({ data }) => setCategories(data)).catch(() => setCategories([]))
  }, [])

  const showModal = (record?: Course) => {
    setEditing(record)
    form.setFieldsValue(record ? {
      ...normalizeCourseEditValues(record),
      cover: currentCover(record),
    } : { status: 0, cover: undefined })
    setOpen(true)
  }

  useEffect(() => {
    const action = consumeOneShotAction(location.search, 'create')
    if (!action.active) return
    navigate({ pathname: location.pathname, search: action.remainingSearch }, { replace: true })
    showModal()
  }, [location.pathname, location.search])
  const save = async () => {
    const values = await form.validateFields()
    const payload: CourseInput = {
      title: values.title,
      description: values.description,
      status: values.status,
      cover_image: values.cover?.url || '',
      category_id: categoryIDForPayload(values.category_id),
    }
    if (editing) await courseApi.update(editing.id, payload)
    else await courseApi.create(payload)
    message.success(editing ? '课程已更新' : '课程已创建')
    setOpen(false)
    form.resetFields()
    void load()
  }
  const remove = async (id: string) => {
    await courseApi.remove(id)
    message.success('课程已删除')
    void load()
  }

  return (
    <>
      <PageHeader
        title={instructor ? '我的课程' : '课程管理'}
        description={instructor ? '创建并维护由你负责的课程。' : '创建课程内容并维护章节与课时。'}
        extra={(
          <Space>
            {!instructor && <Button icon={<BookOutlined />} onClick={() => setOfficialPickerOpen(true)}>添加官方课程</Button>}
            <Button type="primary" icon={<PlusOutlined />} onClick={() => showModal()}>
              新建课程
            </Button>
          </Space>
        )}
      />
      <Card>
        <Table<Course> rowKey="id" loading={loading} dataSource={items}
          pagination={{ ...pagination, showSizeChanger: true }}
          onChange={(page) => void load(page.current, page.pageSize)}
          columns={[
          { title: '课程', dataIndex: 'title', render: (value, record) => <Space><div className="course-cover">{record.cover_image ? <img src={record.cover_image} alt="" /> : '课'}</div><div><strong>{value}</strong><div className="muted">{record.description || '暂无简介'}</div></div></Space> },
          { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === 1 ? 'success' : 'default'}>{value === 1 ? '已发布' : '草稿'}</Tag> },
          { title: '分类', dataIndex: 'category_id', render: (value) => categories.find((item) => item.id === value)?.name || '未分类' },
          { title: '学习人数', dataIndex: 'student_count', render: (value) => value ?? 0 },
          { title: '创建时间', dataIndex: 'created_at', render: (value) => value || '-' },
          { title: '操作', width: 230, render: (_, record) => <Space><Button type="link" icon={<EyeOutlined />} onClick={() => navigate(`/courses/${record.id}`)}>内容</Button><Button type="link" icon={<EditOutlined />} onClick={() => showModal(record)}>编辑</Button><Popconfirm title="确认删除该课程？" onConfirm={() => remove(record.id)}><Button type="link" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm></Space> },
        ]} />
      </Card>
      <Modal title={editing ? '编辑课程' : '新建课程'} open={open} onCancel={() => setOpen(false)} onOk={save} width={620} destroyOnHidden>
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item label="课程标题" name="title" rules={[{ required: true, message: '请输入课程标题' }]}><Input /></Form.Item>
          <Form.Item label="课程简介" name="description"><Input.TextArea rows={4} /></Form.Item>
          <Form.Item label="课程分类" name="category_id"><Select allowClear showSearch optionFilterProp="label" placeholder="选择课程分类" options={categories.filter((item) => item.status === 1 || item.id === editing?.category_id).map((item) => ({ value: item.id, label: item.name }))} /></Form.Item>
          <Form.Item label="课程封面" name="cover">
            <MediaUploader
              accept="image"
              upload={(file, onProgress) =>
                resourceApi.upload(file, onProgress)
                  .then((response) => response.data)}
            />
          </Form.Item>
          {editing ? (
            <Form.Item label="状态" name="status" rules={[{ required: true }]}><Select options={statusOptions} /></Form.Item>
          ) : (
            <Form.Item label="状态"><Input value="草稿（创建后可发布）" disabled /></Form.Item>
          )}
        </Form>
      </Modal>
      <OfficialCoursePicker
        open={officialPickerOpen}
        onClose={() => setOfficialPickerOpen(false)}
      />
    </>
  )
}
