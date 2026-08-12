import { DeleteOutlined, ExportOutlined } from '@ant-design/icons'
import { Button, Card, Popconfirm, Segmented, Space, Table, Tag, message } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import { useLocation, useNavigate } from 'react-router-dom'
import { resourceApi, type Resource } from '../api/resource'
import { normalizePage } from '../api/types'
import MediaUploader from '../components/MediaUploader'
import PageHeader from '../components/PageHeader'
import type { RootState } from '../store'
import { consumeOneShotAction } from '../utils/oneShotAction'
import { completeResourceUpload } from '../utils/resourceUploadFlow'

const typeLabels: Record<Resource['resource_type'], string> = {
  image: '图片',
  video: '视频',
  document: '文档',
  attachment: '附件',
}

type ResourceFilter = 'all' | Resource['resource_type']

export default function Resources() {
  const [items, setItems] = useState<Resource[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<ResourceFilter>('all')
  const uploadCard = useRef<HTMLDivElement>(null)
  const role = useSelector((state: RootState) => state.user.profile?.role)
  const instructor = role === 'instructor'
  const location = useLocation()
  const navigate = useNavigate()
  const filteredItems = useMemo(() => filter === 'all' ? items : items.filter((item) => item.resource_type === filter), [filter, items])

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

  useEffect(() => {
    const action = consumeOneShotAction(location.search, 'upload')
    if (!action.active) return
    navigate({ pathname: location.pathname, search: action.remainingSearch }, { replace: true })
    requestAnimationFrame(() => {
      const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
      uploadCard.current?.scrollIntoView({ block: 'start', behavior: reducedMotion ? 'auto' : 'smooth' })
      uploadCard.current?.focus({ preventScroll: true })
    })
  }, [location.pathname, location.search])

  const remove = async (id: string) => {
    await resourceApi.remove(id)
    message.success('资源已删除')
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
        title="资源列表"
        description={instructor ? '上传、查看并预览本站教学资源。' : '上传并管理课时使用的图片、视频、文档和附件。'}
      />
      <Card ref={uploadCard} tabIndex={-1} className="admin-section-card resource-upload-card">
        <div className="resource-upload-heading">
          <div>
            <strong>上传新资源</strong>
            <p>文件上传成功后会自动进入下方资源列表。</p>
          </div>
        </div>
        <MediaUploader
          accept="all"
          upload={(file, onProgress) =>
            resourceApi.upload(file, onProgress)
              .then((response) => response.data)}
          onPreview={openResource}
          onChange={(resource) => {
            completeResourceUpload(resource, {
              notifySuccess: () => message.success('资源上传成功'),
              refreshList: () => void load(),
            })
          }}
        />
      </Card>
      <Card className="admin-table-card resources-table-card">
        <Segmented<ResourceFilter>
          className="admin-toolbar resource-type-filter"
          value={filter}
          onChange={setFilter}
          options={[
            { value: 'all', label: '全部' },
            { value: 'video', label: '视频' },
            { value: 'image', label: '图片' },
            { value: 'document', label: '文档' },
            { value: 'attachment', label: '附件' },
          ]}
        />
        <Table<Resource>
          rowKey="id"
          loading={loading}
          dataSource={filteredItems}
          pagination={false}
          columns={[
            { title: '名称', dataIndex: 'name', render: (value: string, record: Resource) => <div className="resource-identity"><span className="resource-identity-icon" aria-hidden="true">{typeLabels[record.resource_type].slice(0, 1)}</span><strong>{value}</strong></div> },
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
            ...(!instructor ? [{
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
            }] : []),
          ]}
        />
      </Card>
    </>
  )
}
