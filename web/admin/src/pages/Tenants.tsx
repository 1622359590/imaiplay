import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { tenantApi, type Tenant, type TenantInput } from '../api/tenant'
import { tenantDomainApi, type DomainBindStatus } from '../api/domain'
import { planApi, type Plan } from '../api/plan'
import { userApi, type User } from '../api/user'
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
  const [tenantAdmins, setTenantAdmins] = useState<User[]>([])
  const [passwordUser, setPasswordUser] = useState<User>()
  const [passwordOpen, setPasswordOpen] = useState(false)
  const [domainTenant, setDomainTenant] = useState<Tenant>()
  const [domainOpen, setDomainOpen] = useState(false)
  const [domainStatus, setDomainStatus] = useState<DomainBindStatus>()
  const [domainLoading, setDomainLoading] = useState(false)
  const [form] = Form.useForm<TenantInput>()
  const [passwordForm] = Form.useForm<{ password: string }>()
  const [domainForm] = Form.useForm<{ domain: string }>()
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
  useEffect(() => { void userApi.list({ page: 1, page_size: 100 }).then(({ data }) => setTenantAdmins(normalizePage(data).items.filter((user) => user.role === 'tenant_admin'))).catch(() => setTenantAdmins([])) }, [])
  useEffect(() => {
    if (!domainOpen || !domainTenant) return
    let mounted = true
    let timer: ReturnType<typeof setInterval> | undefined
    const loadStatus = async () => {
      try {
        const status = await tenantDomainApi.status(domainTenant.id)
        if (!mounted) return
        setDomainStatus(status)
        if (status.state === 'ready' || status.state === 'verification_failed' || status.state === 'setup_failed') {
          if (timer) clearInterval(timer)
        }
      } catch {
        if (mounted) setDomainStatus(undefined)
      }
    }
    void loadStatus()
    timer = setInterval(() => void loadStatus(), 2500)
    return () => { mounted = false; if (timer) clearInterval(timer) }
  }, [domainOpen, domainTenant])

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

  const openPasswordModal = (user: User) => {
    setPasswordUser(user)
    passwordForm.resetFields()
    setPasswordOpen(true)
  }

  const resetTenantAdminPassword = async () => {
    if (!passwordUser) return
    const { password } = await passwordForm.validateFields()
    await userApi.resetTenantAdminPassword(passwordUser.id, password)
    message.success('租户管理员密码已重置')
    setPasswordOpen(false)
    setPasswordUser(undefined)
  }

  const openDomainModal = (tenant: Tenant) => {
    setDomainTenant(tenant)
    setDomainStatus(undefined)
    domainForm.setFieldsValue({ domain: tenant.custom_domain || '' })
    setDomainOpen(true)
  }

  const verifyTenantDomain = async () => {
    if (!domainTenant) return
    const { domain } = await domainForm.validateFields()
    setDomainLoading(true)
    try {
      const status = await tenantDomainApi.verify(domainTenant.id, domain)
      setDomainStatus(status)
      message.success('DNS 验证通过，可以开始绑定')
    } finally {
      setDomainLoading(false)
    }
  }

  const bindTenantDomain = async () => {
    if (!domainTenant) return
    const { domain } = await domainForm.validateFields()
    setDomainLoading(true)
    try {
      const status = await tenantDomainApi.bind(domainTenant.id, domain)
      setDomainStatus(status)
      message.success('域名绑定任务已启动')
    } finally {
      setDomainLoading(false)
    }
  }

  const unbindTenantDomain = async () => {
    if (!domainTenant) return
    setDomainLoading(true)
    try {
      const status = await tenantDomainApi.unbind(domainTenant.id)
      setDomainStatus(status)
      domainForm.resetFields()
      message.success('域名已解绑')
      void load()
    } finally {
      setDomainLoading(false)
    }
  }

  const domainStateLabel: Record<DomainBindStatus['state'], string> = {
    none: '未配置', pending_verification: '待验证', verified: 'DNS 已验证', creating_site: '正在创建站点', configuring: '正在配置 HTTPS', ready: '已就绪', verification_failed: '验证失败', setup_failed: '配置失败',
  }

  return (
    <div className="admin-page admin-data-page tenants-page">
      <PageHeader title="租户管理" description="统一管理企业租户与服务状态。" extra={<Space><Button onClick={() => navigate('/tenants/create')}>代客创建租户</Button><Button type="primary" icon={<PlusOutlined />} onClick={() => showModal()}>新增租户</Button></Space>} />
      <Card className="admin-table-card tenants-table-card">
        <Table<Tenant> rowKey="id" loading={loading} dataSource={items}
          pagination={{ ...pagination, showSizeChanger: true }}
          onChange={(page) => void load(page.current, page.pageSize)}
          columns={[
          { title: '租户名称', dataIndex: 'name', render: (value) => <strong>{value}</strong> },
          { title: '租户编码', dataIndex: 'code' },
          { title: '域名', dataIndex: 'custom_domain', render: (value) => value || '未配置' },
          { title: '状态', dataIndex: 'lifecycle_status', render: (value, record) => <Tag color={value === 'active' || (!value && record.status === 1) ? 'success' : 'warning'}>{value === 'trial' ? '试用中' : value === 'suspended' ? '已停用' : value === 'deleted' ? '已注销' : '正式'}</Tag> },
          { title: '创建时间', dataIndex: 'created_at', render: (value) => value || '-' },
          { title: '套餐', render: (_, record) => plans.find((plan) => plan.id === record.plan_id)?.name || '未分配' },
          { title: '租户管理员', render: (_, record) => { const admin = tenantAdmins.find((user) => user.tenant_id === record.id); return admin ? <div><div>{admin.email}</div>{admin.phone && <div className="muted">{admin.phone}</div>}</div> : '未设置' } },
          { title: '操作', width: 180, render: (_, record) => <Space><Button type="link" icon={<EditOutlined />} onClick={() => showModal(record)}>编辑</Button><Popconfirm title="确认删除该租户？" onConfirm={() => remove(record.id)}><Button type="link" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm></Space> },
        ]} />
      </Card>
      <Modal title={editing ? '编辑租户' : '新增租户'} open={open} onCancel={() => setOpen(false)} onOk={save} destroyOnHidden>
        <Form form={form} className="admin-modal-form" layout="vertical" preserve={false}>
          <Form.Item label="租户名称" name="name" rules={[{ required: true, message: '请输入租户名称' }]}><Input /></Form.Item>
          <Form.Item label="租户编码" name="code" rules={[{ required: true, message: '请输入租户编码' }]}><Input disabled={Boolean(editing)} /></Form.Item>
          {editing ? (
            <Form.Item label="状态" name="status" rules={[{ required: true }]}><Select options={[{ value: 1, label: '启用' }, { value: 0, label: '停用' }]} /></Form.Item>
          ) : (
            <Form.Item label="状态"><Input value="启用（创建后可停用）" disabled /></Form.Item>
          )}
          {editing && <Form.Item label="生命周期" name="lifecycle_status"><Select options={[{ value: 'trial', label: '试用中' }, { value: 'active', label: '正式' }, { value: 'suspended', label: '停用' }, { value: 'deleted', label: '注销' }]} /></Form.Item>}
          {editing && <Form.Item label="租户操作"><Space wrap>
            <Button onClick={() => openPlanModal(editing)}>套餐设置</Button>
            <Button onClick={() => openDomainModal(editing)}>域名设置</Button>
            {tenantAdmins.find((user) => user.tenant_id === editing.id) && <Button onClick={() => openPasswordModal(tenantAdmins.find((user) => user.tenant_id === editing.id)!)}>设置密码</Button>}
          </Space></Form.Item>}
        </Form>
      </Modal>
      <Modal title={`为「${planTenant?.name || ''}」分配套餐`} open={planOpen} onCancel={() => setPlanOpen(false)} onOk={() => void assignPlan()} okButtonProps={{ disabled: !selectedPlan }} destroyOnHidden>
        <Form layout="vertical">
          <Form.Item label="套餐">
            <Select value={selectedPlan} onChange={setSelectedPlan} placeholder="请选择套餐" options={plans.filter((plan) => plan.status === 1).map((plan) => ({ value: plan.id, label: `${plan.name}（${plan.storage_quota_bytes > 0 ? `${(plan.storage_quota_bytes / 1024 ** 2).toFixed(0)} MB` : '不限额'}）` }))} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal title={`重置「${passwordUser?.email || ''}」的密码`} open={passwordOpen} onCancel={() => setPasswordOpen(false)} onOk={() => void resetTenantAdminPassword()} destroyOnHidden>
        <Form form={passwordForm} layout="vertical">
          <Form.Item label="新密码" name="password" rules={[{ required: true, min: 8, message: '密码至少 8 位' }]}>
            <Input.Password placeholder="请输入新密码" autoComplete="new-password" />
          </Form.Item>
        </Form>
      </Modal>
      <Modal title={`设置「${domainTenant?.name || ''}」的域名`} open={domainOpen} onCancel={() => setDomainOpen(false)} footer={null} destroyOnHidden>
        <Form form={domainForm} layout="vertical">
          <Form.Item label="自定义域名" name="domain" rules={[{ required: true, message: '请输入租户域名' }]}>
            <Input placeholder="例如 academy.example.com" disabled={domainStatus?.state === 'ready'} />
          </Form.Item>
          <div className="domain-helper-text">
            请让租户将该域名配置 CNAME 到 <strong>{domainStatus?.cname_target || '平台分配的接入地址'}</strong>。
          </div>
          {domainStatus && <div className="domain-status-line"><Tag color={domainStatus.state === 'ready' ? 'success' : domainStatus.state.endsWith('failed') ? 'error' : 'processing'}>{domainStateLabel[domainStatus.state]}</Tag><span>{domainStatus.message}</span></div>}
          <Space>
            <Button onClick={() => void verifyTenantDomain()} loading={domainLoading} disabled={domainStatus?.state === 'ready'}>验证 DNS</Button>
            <Button type="primary" onClick={() => void bindTenantDomain()} loading={domainLoading} disabled={domainStatus?.state !== 'verified'}>自动绑定</Button>
            <Popconfirm title="确认解绑该租户域名？" onConfirm={() => void unbindTenantDomain()} disabled={!domainStatus || domainStatus.state === 'none'}><Button danger loading={domainLoading} disabled={!domainStatus || domainStatus.state === 'none'}>解绑</Button></Popconfirm>
          </Space>
        </Form>
      </Modal>
    </div>
  )
}
