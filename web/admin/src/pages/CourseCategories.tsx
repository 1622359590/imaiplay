import { ArrowDownOutlined, ArrowUpOutlined, DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { isAxiosError } from 'axios'
import { Button, Card, Form, Input, InputNumber, Modal, Popconfirm, Space, Switch, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { courseCategoryApi, type CourseCategory, type CourseCategoryInput } from '../api/courseCategory'
import PageHeader from '../components/PageHeader'
import type { RootState } from '../store'

export default function CourseCategories() {
  const platform = useSelector((state: RootState) => state.user.profile?.role === 'superadmin')
  const [items, setItems] = useState<CourseCategory[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<CourseCategory>()
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm<CourseCategoryInput>()

  const load = async () => {
    setLoading(true)
    try {
      const { data } = await courseCategoryApi.list(platform)
      setItems([...data].sort((left, right) => left.sort_order - right.sort_order || left.name.localeCompare(right.name)))
    } finally { setLoading(false) }
  }
  useEffect(() => { void load() }, [platform])

  const edit = (record?: CourseCategory) => {
    setEditing(record)
    form.setFieldsValue(record || { name: '', sort_order: items.length, status: 1 })
    setOpen(true)
  }

  const save = async () => {
    const values = await form.validateFields()
    try {
      if (editing) await courseCategoryApi.update(editing.id, values, platform)
      else await courseCategoryApi.create(values, platform)
      message.success(editing ? '课程分类已更新' : '课程分类已创建')
      setOpen(false)
      form.resetFields()
      await load()
    } catch (error) {
      if (isAxiosError(error) && error.response?.status === 409) {
        message.error('分类名称重复，或该分类仍被课程引用')
      }
    }
  }

  const update = async (record: CourseCategory, changes: Partial<CourseCategoryInput>) => {
    await courseCategoryApi.update(record.id, {
      name: record.name,
      sort_order: record.sort_order,
      status: record.status,
      ...changes,
    }, platform)
    await load()
  }

  const move = async (record: CourseCategory, direction: -1 | 1) => {
    const index = items.findIndex((item) => item.id === record.id)
    const sibling = items[index + direction]
    if (!sibling) return
    await Promise.all([
      courseCategoryApi.update(record.id, { name: record.name, status: record.status, sort_order: sibling.sort_order }, platform),
      courseCategoryApi.update(sibling.id, { name: sibling.name, status: sibling.status, sort_order: record.sort_order }, platform),
    ])
    await load()
  }

  const remove = async (record: CourseCategory) => {
    try {
      await courseCategoryApi.remove(record.id, platform)
      message.success('课程分类已删除')
      await load()
    } catch (error) {
      if (isAxiosError(error) && error.response?.status === 409) message.error('该分类仍被课程引用，无法删除')
    }
  }

  return (
    <>
      <PageHeader
        title={platform ? '官方课程分类' : '课程分类'}
        description={platform ? '维护平台官方课程使用的分类。' : '维护当前站点课程的单级分类与显示顺序。'}
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => edit()}>新增分类</Button>}
      />
      <Card className="admin-table-card course-categories-table-card">
        <Table<CourseCategory>
          rowKey="id"
          loading={loading}
          dataSource={items}
          pagination={false}
          columns={[
            { title: '分类名称', dataIndex: 'name' },
            { title: '状态', dataIndex: 'status', render: (status, record) => <Space><Switch checked={status === 1} onChange={(checked) => void update(record, { status: checked ? 1 : 0 })} /><Tag color={status === 1 ? 'success' : 'default'}>{status === 1 ? '启用' : '停用'}</Tag></Space> },
            { title: '排序', dataIndex: 'sort_order' },
            { title: '操作', width: 260, render: (_, record, index) => <Space size={2}><Button type="text" icon={<ArrowUpOutlined />} disabled={index === 0} aria-label="上移" onClick={() => void move(record, -1)} /><Button type="text" icon={<ArrowDownOutlined />} disabled={index === items.length - 1} aria-label="下移" onClick={() => void move(record, 1)} /><Button type="link" icon={<EditOutlined />} onClick={() => edit(record)}>编辑</Button><Popconfirm title="确认删除该分类？" onConfirm={() => void remove(record)}><Button type="link" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm></Space> },
          ]}
        />
      </Card>
      <Modal title={editing ? '编辑课程分类' : '新增课程分类'} open={open} onCancel={() => setOpen(false)} onOk={() => void save()} destroyOnHidden>
        <Form form={form} className="admin-modal-form" layout="vertical" preserve={false}>
          <Form.Item name="name" label="分类名称" rules={[{ required: true, whitespace: true, message: '请输入分类名称' }]}><Input maxLength={64} /></Form.Item>
          <Form.Item name="sort_order" label="排序" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="status" label="状态" valuePropName="checked" getValueFromEvent={(checked: boolean) => checked ? 1 : 0} getValueProps={(value: number) => ({ checked: value === 1 })} rules={[{ required: true }]}><Switch checkedChildren="启用" unCheckedChildren="停用" /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
