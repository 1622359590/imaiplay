import { DeleteOutlined, UploadOutlined } from '@ant-design/icons'
import { Button, Card, Popconfirm, Space, Table, Tag, Upload, message } from 'antd'
import type { UploadProps } from 'antd'
import { useEffect, useState } from 'react'
import { resourceApi, type Resource } from '../api/resource'
import { normalizePage } from '../api/types'
import PageHeader from '../components/PageHeader'

const typeLabels = { image: '图片', video: '视频', document: '文档' }

export default function Resources() {
  const [items, setItems] = useState<Resource[]>([])
  const [loading, setLoading] = useState(true)

  const load = async () => {
    setLoading(true)
    try {
      const { data } = await resourceApi.list()
      setItems(normalizePage(data).items)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  const upload: UploadProps['customRequest'] = async ({ file, onSuccess, onError }) => {
    try {
      await resourceApi.upload(file as File)
      message.success('资源上传成功')
      onSuccess?.({})
      void load()
    } catch (error) {
      onError?.(error as Error)
    }
  }

  const remove = async (id: string) => {
    await resourceApi.remove(id)
    message.success('资源已删除')
    void load()
  }

  return (
    <>
      <PageHeader
        title="资源管理"
        description="上传并管理课时使用的图片、视频和 PDF 文档。"
        extra={
          <Upload accept="image/jpeg,image/png,image/webp,video/mp4,video/webm,application/pdf" showUploadList={false} customRequest={upload}>
            <Button type="primary" icon={<UploadOutlined />}>上传资源</Button>
          </Upload>
        }
      />
      <Card>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={items}
          pagination={false}
          columns={[
            { title: '名称', dataIndex: 'name' },
            { title: '类型', dataIndex: 'resource_type', render: (value: Resource['resource_type']) => <Tag>{typeLabels[value]}</Tag> },
            { title: '大小', dataIndex: 'size_bytes', render: (value: number) => `${(value / 1024 / 1024).toFixed(2)} MB` },
            { title: '访问地址', dataIndex: 'url', render: (url: string) => <a href={url} target="_blank" rel="noreferrer">打开</a> },
            {
              title: '操作',
              render: (_value, record: Resource) => (
                <Space>
                  <Popconfirm title="确认删除该资源？" onConfirm={() => remove(record.id)}>
                    <Button type="link" danger icon={<DeleteOutlined />}>删除</Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>
    </>
  )
}
