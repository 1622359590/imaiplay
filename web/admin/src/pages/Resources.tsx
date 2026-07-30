import { DeleteOutlined, ExportOutlined } from '@ant-design/icons'
import { Button, Card, Popconfirm, Space, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { resourceApi, type Resource } from '../api/resource'
import { normalizePage } from '../api/types'
import MediaUploader, {
  type UploadedMedia,
} from '../components/MediaUploader'
import PageHeader from '../components/PageHeader'

const typeLabels = { image: '图片', video: '视频', document: '文档' }

export default function Resources() {
  const [items, setItems] = useState<Resource[]>([])
  const [loading, setLoading] = useState(true)
  const [uploaded, setUploaded] = useState<UploadedMedia>()

  const load = async () => {
    setLoading(true)
    try {
      const { data } = await resourceApi.list()
      setItems(normalizePage(data).items)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const remove = async (id: string) => {
    await resourceApi.remove(id)
    message.success('资源已删除')
    if (uploaded?.id === id) setUploaded(undefined)
    void load()
  }

  const openResource = async (resource: Pick<Resource, 'id'>) => {
    const { data } = await resourceApi.file(resource.id)
    const url = URL.createObjectURL(data)
    window.open(url, '_blank', 'noopener,noreferrer')
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
  }

  return (
    <>
      <PageHeader
        title="资源管理"
        description="上传并管理课时使用的图片、视频和 PDF 文档。"
      />
      <Card className="resource-upload-card">
        <div className="resource-upload-heading">
          <div>
            <strong>上传新资源</strong>
            <p>文件上传成功后会自动进入下方资源列表。</p>
          </div>
        </div>
        <MediaUploader
          value={uploaded}
          accept="all"
          upload={(file, onProgress) =>
            resourceApi.upload(file, onProgress)
              .then((response) => response.data)}
          onPreview={openResource}
          onChange={(resource) => {
            setUploaded(resource)
            if (resource) {
              message.success('资源上传成功')
              void load()
            }
          }}
        />
      </Card>
      <Card>
        <Table<Resource>
          rowKey="id"
          loading={loading}
          dataSource={items}
          pagination={false}
          columns={[
            { title: '名称', dataIndex: 'name' },
            {
              title: '类型',
              dataIndex: 'resource_type',
              render: (value: Resource['resource_type']) => (
                <Tag>{typeLabels[value]}</Tag>
              ),
            },
            {
              title: '大小',
              dataIndex: 'size_bytes',
              render: (value: number) =>
                `${(value / 1024 / 1024).toFixed(2)} MB`,
            },
            {
              title: '预览',
              render: (_: unknown, record: Resource) => (
                <Button
                  type="link"
                  icon={<ExportOutlined />}
                  onClick={() => void openResource(record)}
                >
                  打开
                </Button>
              ),
            },
            {
              title: '操作',
              render: (_: unknown, record: Resource) => (
                <Space>
                  <Popconfirm
                    title="确认删除该资源？"
                    onConfirm={() => void remove(record.id)}
                  >
                    <Button type="link" danger icon={<DeleteOutlined />}>
                      删除
                    </Button>
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
