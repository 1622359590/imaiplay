import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, message } from 'antd'
import { useEffect, useState } from 'react'
import { resourceCategoryApi, type ResourceCategory } from '../api/resource'
import PageHeader from '../components/PageHeader'

export default function ResourceCategories() {
  const [items, setItems] = useState<ResourceCategory[]>([])
  const [editing, setEditing] = useState<ResourceCategory | null | undefined>()
  const [form] = Form.useForm()

  const load = async () => {
    const { data } = await resourceCategoryApi.list()
    setItems(data)
  }
  useEffect(() => { void load() }, [])

  const open = (category: ResourceCategory | null) => {
    setEditing(category)
    form.setFieldsValue(category ?? {})
  }
  const save = async () => {
    const values = await form.validateFields()
    if (editing) await resourceCategoryApi.update(editing.id, values)
    else await resourceCategoryApi.create(values)
    message.success('资源分类已保存')
    setEditing(undefined)
    form.resetFields()
    void load()
  }
  const remove = async (id: string) => {
    await resourceCategoryApi.remove(id)
    message.success('资源分类已删除')
    void load()
  }

  return (
    <>
      <PageHeader title="资源分类" description="维护资源库的分类层级。" extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => open(null)}>新增分类</Button>} />
      <Card>
        <Table
          rowKey="id"
          dataSource={items}
          pagination={false}
          columns={[
            { title: '分类名称', dataIndex: 'name' },
            { title: '上级分类', dataIndex: 'parent_id', render: (id?: string) => items.find((item) => item.id === id)?.name || '—' },
            {
              title: '操作',
              render: (_value, record: ResourceCategory) => (
                <Space>
                  <Button type="link" icon={<EditOutlined />} onClick={() => open(record)}>编辑</Button>
                  <Popconfirm title="确认删除该分类？" onConfirm={() => remove(record.id)}>
                    <Button type="link" danger icon={<DeleteOutlined />}>删除</Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>
      <Modal title={editing ? '编辑分类' : '新增分类'} open={editing !== undefined} onCancel={() => setEditing(undefined)} onOk={save} destroyOnHidden>
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="name" label="分类名称" rules={[{ required: true, message: '请输入分类名称' }]}><Input /></Form.Item>
          <Form.Item name="parent_id" label="上级分类">
            <Select allowClear options={items.filter((item) => item.id !== editing?.id).map((item) => ({ value: item.id, label: item.name }))} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
