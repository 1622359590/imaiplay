import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { Avatar, Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { userApi, type User, type UserInput } from '../api/user'
import { normalizePage } from '../api/types'
import PageHeader from '../components/PageHeader'

export default function Users() {
  const [items, setItems] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<User>()
  const [form] = Form.useForm<UserInput>()

  const load = async (current = pagination.current, pageSize = pagination.pageSize) => {
    setLoading(true)
    try {
      const { data } = await userApi.list({ page: current, page_size: pageSize })
      const page = normalizePage(data)
      setItems(page.items)
      setPagination({ current, pageSize, total: page.total })
    } finally { setLoading(false) }
  }
  useEffect(() => { void load() }, [])

  const showModal = (record?: User) => {
    setEditing(record)
    form.setFieldsValue(record || { role: 'learner', status: 1 })
    setOpen(true)
  }
  const save = async () => {
    const values = await form.validateFields()
    if (editing) await userApi.update(editing.id, values)
    else await userApi.create(values)
    message.success(editing ? '用户已更新' : '用户已创建')
    setOpen(false)
    form.resetFields()
    void load()
  }
  const remove = async (id: string) => {
    await userApi.remove(id)
    message.success('用户已删除')
    void load()
  }

  return (
    <>
      <PageHeader title="用户管理" description="管理成员账号、角色与账号状态。" extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => showModal()}>新增用户</Button>} />
      <Card>
        <Table<User> rowKey="id" loading={loading} dataSource={items}
          pagination={{ ...pagination, showSizeChanger: true }}
          onChange={(page) => void load(page.current, page.pageSize)}
          columns={[
          { title: '用户', dataIndex: 'name', render: (value, record) => <Space><Avatar>{String(value || 'U').slice(0, 1)}</Avatar><div><strong>{value}</strong><div className="muted">{record.email}</div></div></Space> },
          { title: '角色', dataIndex: 'role', render: (value) => ({ tenant_admin: '租户管理员', instructor: '讲师', learner: '学员' }[value as string] || value) },
          { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === 1 ? 'success' : 'default'}>{value === 1 ? '正常' : '停用'}</Tag> },
          { title: '创建时间', dataIndex: 'created_at', render: (value) => value || '-' },
          { title: '操作', width: 150, render: (_, record) => <Space><Button type="link" icon={<EditOutlined />} onClick={() => showModal(record)}>编辑</Button><Popconfirm title="确认删除该用户？" onConfirm={() => remove(record.id)}><Button type="link" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm></Space> },
        ]} />
      </Card>
      <Modal title={editing ? '编辑用户' : '新增用户'} open={open} onCancel={() => setOpen(false)} onOk={save} destroyOnHidden>
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item label="姓名" name="name" rules={[{ required: true, message: '请输入姓名' }]}><Input /></Form.Item>
          <Form.Item label="邮箱" name="email" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}><Input disabled={Boolean(editing)} /></Form.Item>
          {!editing && <Form.Item label="初始密码" name="password" rules={[{ required: true, min: 6, message: '密码至少 6 位' }]}><Input.Password /></Form.Item>}
          <Form.Item label="角色" name="role" rules={[{ required: true }]}><Select disabled={Boolean(editing)} options={[{ value: 'tenant_admin', label: '租户管理员' }, { value: 'instructor', label: '讲师' }, { value: 'learner', label: '学员' }]} /></Form.Item>
          {editing ? (
            <Form.Item label="状态" name="status" rules={[{ required: true }]}><Select options={[{ value: 1, label: '正常' }, { value: 0, label: '停用' }]} /></Form.Item>
          ) : (
            <Form.Item label="状态"><Input value="正常（创建后可停用）" disabled /></Form.Item>
          )}
        </Form>
      </Modal>
    </>
  )
}
