import { Card, Input, Modal, Select, Space, Table, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { auditApi, type AuditLog } from '../api/audit'
import { tokenRole } from '../api/auth'
import PageHeader from '../components/PageHeader'

export default function AuditLogs() {
  const superadmin = tokenRole() === 'superadmin'
  const [items, setItems] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(false)
  const [selected, setSelected] = useState<AuditLog>()
  const [filters, setFilters] = useState({ action: '', user_id: '', tenant_id: '', from: '', to: '' })
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const load = async (current = pagination.current, pageSize = pagination.pageSize) => {
    setLoading(true)
    try {
      const query = Object.fromEntries(Object.entries(filters).filter(([, value]) => value))
      const { data } = await auditApi.list({ ...query, offset: (current - 1) * pageSize, limit: pageSize }, superadmin)
      setItems(data.items || [])
      setPagination({ current, pageSize, total: data.total || 0 })
    } finally { setLoading(false) }
  }
  useEffect(() => { void load(1) }, [filters.action, filters.user_id, filters.tenant_id, filters.from, filters.to])
  const update = (key: keyof typeof filters, value: string) => setFilters((current) => ({ ...current, [key]: value }))
  return <>
    <PageHeader title="操作审计" description="记录关键写操作与认证事件，按权限隔离可见范围。" />
    <Card className="admin-table-card audit-table-card">
      <Space wrap className="admin-toolbar audit-filter-toolbar">
        <Select allowClear placeholder="操作类型" style={{ width: 210 }} onChange={(value) => update('action', value || '')} options={[{ value: 'auth.login_failed', label: '登录失败' }, { value: 'auth.login_success', label: '登录成功' }, { value: 'user.create', label: '创建用户' }, { value: 'course.create', label: '创建课程' }, { value: 'resource.create', label: '上传资源' }]} />
        <Input placeholder="操作人用户 ID" value={filters.user_id} onChange={(event) => update('user_id', event.target.value)} style={{ width: 180 }} />
        {superadmin && <Input placeholder="租户 ID" value={filters.tenant_id} onChange={(event) => update('tenant_id', event.target.value)} style={{ width: 180 }} />}
        <Input placeholder="开始时间 RFC3339" value={filters.from} onChange={(event) => update('from', event.target.value)} style={{ width: 210 }} />
        <Input placeholder="结束时间 RFC3339" value={filters.to} onChange={(event) => update('to', event.target.value)} style={{ width: 210 }} />
      </Space>
      <Table<AuditLog> rowKey="id" loading={loading} dataSource={items} pagination={{ ...pagination, showSizeChanger: true }} onChange={(page) => void load(page.current, page.pageSize)} columns={[
        { title: '时间', dataIndex: 'created_at', width: 190 },
        { title: '操作人', dataIndex: 'user_email', render: (value: string, record: AuditLog) => value || record.user_id || '系统' },
        { title: '动作', dataIndex: 'action' },
        { title: '资源', render: (_: unknown, record: AuditLog) => `${record.resource_type || '-'}${record.resource_id ? ` / ${record.resource_id}` : ''}` },
        { title: 'IP', dataIndex: 'ip' },
        { title: '详情', render: (_: unknown, record: AuditLog) => <a onClick={() => setSelected(record)}>查看</a> },
      ]} />
    </Card>
    <Modal title="审计详情" open={Boolean(selected)} footer={null} onCancel={() => setSelected(undefined)}><Typography.Paragraph copyable>{selected ? JSON.stringify(JSON.parse(selected.detail || '{}'), null, 2) : ''}</Typography.Paragraph></Modal>
  </>
}
