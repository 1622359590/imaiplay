import {
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  message,
} from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { tokenRole } from '../api/auth'
import type { Course } from '../api/course'
import { courseCategoryApi, type CourseCategory } from '../api/courseCategory'
import {
  officialCourseApi,
  type OfficialCourseInput,
} from '../api/officialCourse'
import { resourceApi } from '../api/resource'
import MediaUploader, {
  type UploadedMedia,
} from '../components/MediaUploader'
import PageHeader from '../components/PageHeader'
import { updateOfficialCourseEnabled } from '../utils/officialCourses'
import { categoryIDForPayload } from '../utils/adminFormValues'

type OfficialCourseRecord = Course & { enabled?: boolean }
type OfficialCourseForm = Omit<OfficialCourseInput, 'cover_image' | 'category_id'> & {
  category_id?: string
  cover?: UploadedMedia
}

const statusOptions = [
  { value: 0, label: '草稿' },
  { value: 1, label: '已发布' },
]

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

export default function OfficialCourses() {
  const superadmin = tokenRole() === 'superadmin'
  const navigate = useNavigate()
  const [items, setItems] = useState<OfficialCourseRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [savingIds, setSavingIds] = useState<Set<string>>(new Set())
  const [editing, setEditing] = useState<Course>()
  const [categories, setCategories] = useState<CourseCategory[]>([])
  const [open, setOpen] = useState(false)
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  })
  const [form] = Form.useForm<OfficialCourseForm>()

  const load = async (
    current = pagination.current,
    pageSize = pagination.pageSize,
  ) => {
    setLoading(true)
    try {
      const data = await officialCourseApi.list({
        page: current,
        page_size: pageSize,
      })
      setItems(data.items || [])
      setPagination({ current, pageSize, total: data.total || 0 })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
    if (superadmin) {
      void courseCategoryApi.list(true)
        .then(({ data }) => setCategories(data))
        .catch(() => setCategories([]))
    }
  }, [])

  const showEditor = (course?: Course) => {
    setEditing(course)
    form.setFieldsValue(course ? {
      title: course.title,
      description: course.description,
      status: course.status,
      category_id: course.category_id || undefined,
      course_type: course.course_type,
      cover: currentCover(course),
    } : {
      title: '',
      description: '',
      status: 0,
      category_id: undefined,
      course_type: undefined,
      cover: undefined,
    })
    setOpen(true)
  }

  const save = async () => {
    const values = await form.validateFields()
    setSaving(true)
    try {
      const payload: OfficialCourseInput = {
        title: values.title,
        description: values.description,
        status: values.status,
        cover_image: values.cover?.url || '',
        category_id: categoryIDForPayload(values.category_id),
        course_type: values.course_type,
      }
      if (editing) {
        await officialCourseApi.update(editing.id, payload)
        message.success('官方课程已更新')
        setOpen(false)
        form.resetFields()
        void load()
      } else {
        const created = await officialCourseApi.create(payload)
        message.success('官方课程已创建，现在可以添加章节和课时')
        setOpen(false)
        form.resetFields()
        navigate(`/official-courses/${created.id}`)
      }
    } finally {
      setSaving(false)
    }
  }

  const remove = async (id: string) => {
    await officialCourseApi.remove(id)
    message.success('官方课程已删除')
    void load()
  }

  const columns = superadmin ? [
    {
      title: '课程',
      dataIndex: 'title',
      render: (value: string, record: Course) => (
        <Space size={14}>
          <div className="course-cover">
            {record.cover_image
              ? <img src={record.cover_image} alt="" />
              : '课'}
          </div>
          <div>
            <strong>{value}</strong>
            <div className="muted">{record.description || '暂无简介'}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 110,
      render: (value: number) => (
        <Tag color={value === 1 ? 'success' : 'default'}>
          {value === 1 ? '已发布' : '草稿'}
        </Tag>
      ),
    },
    {
      title: '分类',
      dataIndex: 'category_id',
      render: (value: string | undefined) => categories.find((item) => item.id === value)?.name || '未分类',
    },
    {
      title: '课程类型',
      dataIndex: 'course_type',
      render: (value: string) => <Tag color={value === 'required' ? 'magenta' : 'blue'}>{value === 'required' ? '必修课' : '选修课'}</Tag>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 210,
      render: (value: string) => value || '-',
    },
    {
      title: '操作',
      width: 250,
      render: (_: unknown, record: Course) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/official-courses/${record.id}`)}
          >
            内容
          </Button>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => showEditor(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除该官方课程及其全部内容？"
            onConfirm={() => void remove(record.id)}
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ] : [
    { title: '课程名称', dataIndex: 'title' },
    { title: '描述', dataIndex: 'description' },
    {
      title: '启用',
      dataIndex: 'enabled',
      width: 120,
      render: (_: unknown, record: OfficialCourseRecord) => (
        <Switch
          checked={record.enabled === true}
          loading={savingIds.has(record.id)}
          onChange={(value) => {
            const previous = record.enabled === true
            setItems((current) => updateOfficialCourseEnabled(current, record.id, value))
            setSavingIds((current) => new Set(current).add(record.id))
            void officialCourseApi.enable(record.id, value)
              .then(() => message.success(value ? '官方课程已启用' : '官方课程已停用'))
              .catch(() => {
                setItems((current) => updateOfficialCourseEnabled(current, record.id, previous))
                message.error('保存失败，请稍后重试')
              })
              .finally(() => {
                setSavingIds((current) => {
                  const next = new Set(current)
                  next.delete(record.id)
                  return next
                })
              })
          }}
        />
      ),
    },
  ]

  return (
    <div className="admin-page admin-data-page official-courses-page">
      <PageHeader
        title="官方课程"
        description={superadmin
          ? '维护平台共享课程、封面、章节和课时资源。'
          : '选择平台官方课程，为本租户启用或停用。'}
        extra={superadmin ? (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => showEditor()}
          >
            新建官方课程
          </Button>
        ) : undefined}
      />
      <Card className="admin-table-card official-courses-table-card">
        <Table<OfficialCourseRecord>
          rowKey="id"
          loading={loading}
          dataSource={items}
          columns={columns}
          pagination={{
            ...pagination,
            showSizeChanger: true,
          }}
          onChange={(page) => {
            void load(page.current, page.pageSize)
          }}
        />
      </Card>
      {superadmin && (
        <Modal
          title={editing ? '编辑官方课程' : '新建官方课程'}
          open={open}
          width={680}
          confirmLoading={saving}
          onCancel={() => {
            setOpen(false)
            form.resetFields()
          }}
          onOk={() => void save()}
          destroyOnHidden
        >
          <Form form={form} className="admin-modal-form" layout="vertical" preserve={false}>
            <Form.Item
              name="title"
              label="课程名称"
              rules={[{ required: true, message: '请输入课程名称' }]}
            >
              <Input placeholder="例如：新员工入职训练营" />
            </Form.Item>
            <Form.Item name="description" label="课程描述">
              <Input.TextArea
                rows={4}
                maxLength={1000}
                showCount
                placeholder="介绍课程目标和适合人群"
              />
            </Form.Item>
            <Form.Item
              name="status"
              label="课程状态"
              rules={[{ required: true }]}
            >
              <Select options={statusOptions} />
            </Form.Item>
            <Form.Item name="category_id" label="课程分类">
              <Select
                allowClear
                showSearch
                optionFilterProp="label"
                placeholder="选择官方课程分类"
                options={categories
                  .filter((item) => item.status === 1 || item.id === editing?.category_id)
                  .map((item) => ({ value: item.id, label: item.name }))}
              />
            </Form.Item>
            <Form.Item
              name="course_type"
              label="课程类型"
              rules={[{ required: true, message: '请选择必修课或选修课' }]}
            >
              <Select
                placeholder="请选择课程类型"
                options={[
                  { value: 'required', label: '必修课' },
                  { value: 'optional', label: '选修课' },
                ]}
              />
            </Form.Item>
            <Form.Item name="cover" label="课程封面">
              <MediaUploader
                accept="image"
                upload={(file, onProgress) =>
                  resourceApi.uploadPlatform(file, onProgress)
                    .then((response) => response.data)}
              />
            </Form.Item>
          </Form>
        </Modal>
      )}
    </div>
  )
}
