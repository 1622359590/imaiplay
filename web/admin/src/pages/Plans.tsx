import { Button, Card, Form, Input, InputNumber, Modal, Popconfirm, Space, Table, Tag, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useEffect, useState } from 'react'
import PageHeader from '../components/PageHeader'
import { planApi, type Plan, type PlanInput } from '../api/plan'

const formatBytes = (bytes: number) => bytes >= 1024 ** 3 ? `${(bytes / 1024 ** 3).toFixed(1)} GB` : `${(bytes / 1024 ** 2).toFixed(0)} MB`
type PlanFormValues = Omit<PlanInput, 'storage_quota_bytes'> & { storage_quota_mb: number }

export default function Plans() {
  const [items, setItems] = useState<Plan[]>([]); const [editing, setEditing] = useState<Plan>(); const [open, setOpen] = useState(false); const [form] = Form.useForm<PlanFormValues>()
  const load = async () => { const { data } = await planApi.list(); setItems(data.items || []) }
  useEffect(() => { void load() }, [])
  const save = async () => { const values = await form.validateFields(); const { storage_quota_mb, ...rest } = values; const payload: PlanInput = { ...rest, storage_quota_bytes: storage_quota_mb * 1024 ** 2 }; if (editing) await planApi.update(editing.id, payload); else await planApi.create(payload); message.success('套餐已保存'); setOpen(false); form.resetFields(); void load() }
  return <><PageHeader title="套餐管理" description="定义存储配额和预留的产品能力字段。" extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(undefined); form.resetFields(); setOpen(true) }}>新建套餐</Button>} /><Card><Table<Plan> rowKey="id" dataSource={items} columns={[{ title: '名称', dataIndex: 'name' }, { title: '存储配额', dataIndex: 'storage_quota_bytes', render: (value) => value > 0 ? formatBytes(value) : '不限额' }, { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === 1 ? 'success' : 'default'}>{value === 1 ? '启用' : '停用'}</Tag> }, { title: '默认', dataIndex: 'is_default', render: (value) => value ? '是' : '-' }, { title: '操作', render: (_, record) => <Space><Button type="link" onClick={() => { setEditing(record); form.setFieldsValue({ ...record, storage_quota_mb: record.storage_quota_bytes / 1024 ** 2 }); setOpen(true) }}>编辑</Button><Popconfirm title="确认删除套餐？" onConfirm={async () => { await planApi.remove(record.id); void load() }}><Button type="link" danger>删除</Button></Popconfirm></Space> }]} /></Card><Modal title={editing ? '编辑套餐' : '新建套餐'} open={open} onCancel={() => setOpen(false)} onOk={() => void save()}><Form form={form} layout="vertical" initialValues={{ storage_quota_mb: 1024, max_users: 0, max_courses: 0, status: 1 }}><Form.Item name="name" label="套餐名称" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="storage_quota_mb" label="存储配额（MB，0=不限额）"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item><Form.Item name="max_users" label="学员数预留"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item><Form.Item name="max_courses" label="课程数预留"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Form></Modal></>
}
