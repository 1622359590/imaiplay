import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { tenantApi, type Tenant, type TenantInput } from '../api/tenant'
import { planApi, type Plan } from '../api/plan'
import { normalizePage } from '../api/types'
import PageHeader from '../components/PageHeader'
import { useNavigate } from 'react-router-dom'

export default function Tenants() {
  const [items, setItems] = useState<Tenant[]>([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Tenant>()
  const [plans, setPlans] = useState<Plan[]>([])
  const [planTenant, setPlanTenant] = useState<Tenant>()
  const [planOpen, setPlanOpen] = useState(false)
  const [selectedPlan, setSelectedPlan] = useState<string>()
  const [form] = Form.useForm<TenantInput>()
  const navigate = useNavigate()

  const load = async (current = pagination.current, pageSize = pagination.pageSize) => {
    setLoading(true)
    try {
      const { data } = await tenantApi.list({ page: current, page_size: pageSize })
      const page = normalizePage(data)
      setItems(page.items)
      setPagination({ current, pageSize, total: page.total })
    } finally {
      setLoading(false)
    }
  }
  useEffect(() => { void load() }, [])
  useEffect(() => { void planApi.list(0, 100).then(({ data }) => setPlans(data.items || [])).catch(() => setPlans([])) }, [])

  const showModal = (record?: Tenant) => {
    setEditing(record)
    form.setFieldsValue(record || { status: 1 })
    setOpen(true)
  }

  const save = async () => {
    const values = await form.validateFields()
    if (editing) await tenantApi.update(editing.id, values)
    else await tenantApi.create(values)
    message.success(editing ? '租户已更新' : '租户已创建')
    setOpen(false)
    form.resetFields()
    void load()
  }

  const remove = async (id: string) => {
    await tenantApi.remove(id)
    message.success('租户已删除')
    void load()
  }

  const openPlanModal = (tenant: Tenant) => {
    setPlanTenant(tenant)
    setSelectedPlan(tenant.plan_id)
    setPlanOpen(true)
  }

  const assignPlan = async () => {
    if (!planTenant || !selectedPlan) return
    await planApi.assign(planTenant.id, selectedPlan)
    message.success('套餐已分配')
    setPlanOpen(false)
    void load()
  }

  return (
    <>
      <PageHeader title="租户管理" description="统一管理企业租户与服务状态。" extra={<Space><Button onClick={() => navigate('/tenants/create')}>代客创建租户</Button><Button type="primary" icon={<PlusOutlined />} onClick={() => showModal()}>新增租户</Button></Space>} />
      <Card>
        <Table<Tenant> rowKey="id" loading={loading} dataSource={items}
          pagination={{ ...pagination, showSizeChanger: true }}
          onChange={(page) => void load(page.current, page.pageSize)}
          columns={[
          { title: '租户名称', dataIndex: 'name', render: (value) => <strong>{value}</strong> },
          { title: '租户编码', dataIndex: 'code' },
          { title: '状态', dataIndex: 'lifecycle_status', render: (value, record) => <Tag color={value === 'active' || (!value && record.status === 1) ? 'success' : 'warning'}>{value === 'trial' ? '试用中' : value === 'suspended' ? '已停用' : value === 'deleted' ? '已注销' : '正式'}</Tag> },
          { title: '创建时间', dataIndex: 'created_at', render: (value) => value || '-' },
          { title: '套餐', render: (_, record) => plans.find((plan) => plan.id === record.plan_id)?.name || '未分配' },
          { title: '操作', width: 220, render: (_, record) => <Space><Button type="link" onClick={() => openPlanModal(record)}>套餐</Button><Button type="link" icon={<EditOutlined />} onClick={() => showModal(record)}>编辑</Button><Popconfirm title="确认删除该租户？" onConfirm={() => remove(record.id)}><Button type="link" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm></Space> },
        ]} />
      </Card>
      <Modal title={editing ? '编辑租户' : '新增租户'} open={open} onCancel={() => setOpen(false)} onOk={save} destroyOnHidden>
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item label="租户名称" name="name" rules={[{ required: true, message: '请输入租户名称' }]}><Input /></Form.Item>
          <Form.Item label="租户编码" name="code" rules={[{ required: true, message: '请输入租户编码' }]}><Input disabled={Boolean(editing)} /></Form.Item>
          {editing ? (
            <Form.Item label="状态" name="status" rules={[{ required: true }]}><Select options={[{ value: 1, label: '启用' }, { value: 0, label: '停用' }]} /></Form.Item>
          ) : (
            <Form.Item label="状态"><Input value="启用（创建后可停用）" disabled /></Form.Item>
          )}
          {editing && <Form.Item label="生命周期" name="lifecycle_status"><Select options={[{ value: 'trial', label: '试用中' }, { value: 'active', label: '正式' }, { value: 'suspended', label: '停用' }, { value: 'deleted', label: '注销' }]} /></Form.Item>}
          {editing && <Form.Item label="自定义域名" name="custom_domain"><Input placeholder="academy.example.com，留空不修改" /></Form.Item>}
        </Form>
      </Modal>
      <Modal title={`为「${planTenant?.name || ''}」分配套餐`} open={planOpen} onCancel={() => setPlanOpen(false)} onOk={() => void assignPlan()} okButtonProps={{ disabled: !selectedPlan }} destroyOnHidden>
        <Form layout="vertical">
          <Form.Item label="套餐">
            <Select value={selectedPlan} onChange={setSelectedPlan} placeholder="请选择套餐" options={plans.filter((plan) => plan.status === 1).map((plan) => ({ value: plan.id, label: `${plan.name}（${plan.storage_quota_bytes > 0 ? `${(plan.storage_quota_bytes / 1024 ** 2).toFixed(0)} MB` : '不限额'}）` }))} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
