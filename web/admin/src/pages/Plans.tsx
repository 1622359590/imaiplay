import { PlusOutlined } from '@ant-design/icons'
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Radio,
  Space,
  Table,
  Tag,
  message,
} from 'antd'
import { useEffect, useState } from 'react'
import { planApi, type Plan } from '../api/plan'
import PageHeader from '../components/PageHeader'
import {
  buildPlanInput,
  createPlanFormValues,
  type PlanFormValues,
} from '../utils/planForm'

const formatBytes = (bytes: number) => bytes >= 1024 ** 3
  ? `${(bytes / 1024 ** 3).toFixed(1)} GB`
  : `${(bytes / 1024 ** 2).toFixed(0)} MB`

export default function Plans() {
  const [items, setItems] = useState<Plan[]>([])
  const [editing, setEditing] = useState<Plan>()
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm<PlanFormValues>()

  const load = async () => {
    const { data } = await planApi.list()
    setItems(data.items || [])
  }

  useEffect(() => {
    void load()
  }, [])

  const openCreate = () => {
    setEditing(undefined)
    form.setFieldsValue(createPlanFormValues())
    setOpen(true)
  }

  const openEdit = (plan: Plan) => {
    setEditing(plan)
    form.setFieldsValue(createPlanFormValues(plan))
    setOpen(true)
  }

  const save = async () => {
    const values = await form.validateFields()
    const payload = buildPlanInput(values)
    if (editing) await planApi.update(editing.id, payload)
    else await planApi.create(payload)
    message.success('套餐已保存')
    setOpen(false)
    form.resetFields()
    void load()
  }

  return (
    <div className="admin-page admin-data-page plans-page">
      <PageHeader
        title="套餐管理"
        description="定义存储配额和预留的产品能力字段。"
        extra={(
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
            新建套餐
          </Button>
        )}
      />
      <Card className="admin-table-card plans-table-card">
        <Table<Plan>
          rowKey="id"
          dataSource={items}
          columns={[
            { title: '名称', dataIndex: 'name' },
            {
              title: '存储配额',
              dataIndex: 'storage_quota_bytes',
              render: (value) => value > 0 ? formatBytes(value) : '不限额',
            },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value) => (
                <Tag color={value === 1 ? 'success' : 'default'}>
                  {value === 1 ? '启用' : '停用'}
                </Tag>
              ),
            },
            {
              title: '默认',
              dataIndex: 'is_default',
              render: (value) => value ? '是' : '-',
            },
            {
              title: '操作',
              render: (_, record) => (
                <Space>
                  <Button type="link" onClick={() => openEdit(record)}>
                    编辑
                  </Button>
                  <Popconfirm
                    title="确认删除套餐？"
                    onConfirm={async () => {
                      await planApi.remove(record.id)
                      void load()
                    }}
                  >
                    <Button type="link" danger>删除</Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>
      <Modal
        title={editing ? '编辑套餐' : '新建套餐'}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={() => void save()}
        width={680}
        destroyOnHidden
      >
        <Form form={form} className="admin-modal-form plan-editor-form" layout="vertical" preserve={false}>
          <div className="form-grid form-grid-two">
            <Form.Item name="name" label="套餐名称" rules={[{ required: true, message: '请输入套餐名称' }]}>
              <Input />
            </Form.Item>
            <Form.Item name="storage_quota_mb" label="存储配额（MB，0=不限额）">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="max_users" label="学员数预留">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="max_courses" label="课程数预留">
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </div>
          <Form.Item
            name="status"
            label="套餐状态"
            rules={[{ required: true, message: '请选择套餐状态' }]}
          >
            <Radio.Group
              optionType="button"
              buttonStyle="solid"
              options={[
                { label: '启用', value: 1 },
                { label: '停用', value: 0 },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
