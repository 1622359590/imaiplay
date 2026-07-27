import { Button, Card, Form, Input, Modal, Space, Switch, Table, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useEffect, useState } from 'react'
import { officialCourseApi } from '../api/officialCourse'
import type { Course } from '../api/course'
import PageHeader from '../components/PageHeader'
import { tokenRole } from '../api/auth'

export default function OfficialCourses() {
  const [items, setItems] = useState<Course[]>([]); const [open, setOpen] = useState(false); const [form] = Form.useForm(); const superadmin = tokenRole() === 'superadmin'
  const load = async () => { const data = await officialCourseApi.list(); setItems(data.items) }
  useEffect(() => { void load() }, [])
  const create = async () => { await officialCourseApi.create(await form.validateFields()); message.success('官方课程已创建'); setOpen(false); form.resetFields(); void load() }
  return <><PageHeader title="官方课程" description={superadmin ? '维护平台官方课程，租户启用后以引用方式使用。' : '选择平台官方课程，为本租户启用或停用。'} extra={superadmin ? <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>新建官方课</Button> : undefined} /><Card><Table<Course> rowKey="id" dataSource={items} columns={[{ title: '课程名称', dataIndex: 'title' }, { title: '描述', dataIndex: 'description' }, ...(!superadmin ? [{ title: '启用', dataIndex: 'enabled', render: (_: unknown, record: Course & { enabled?: boolean }) => <Switch checked={record.enabled} onChange={(value) => void officialCourseApi.enable(record.id, value).then(load)} /> }] : [])]} /></Card><Modal title="新建官方课程" open={open} onCancel={() => setOpen(false)} onOk={() => void create()}><Form form={form} layout="vertical"><Form.Item name="title" label="课程名称" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="description" label="课程描述"><Input.TextArea /></Form.Item><Form.Item name="cover_image" label="封面地址"><Input /></Form.Item></Form></Modal></>
}
