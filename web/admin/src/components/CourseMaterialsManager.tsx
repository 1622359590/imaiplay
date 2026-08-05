import {
  ArrowDownOutlined,
  ArrowUpOutlined,
  DeleteOutlined,
  FileOutlined,
  ReloadOutlined,
  UploadOutlined,
} from '@ant-design/icons'
import {
  Button,
  Card,
  Input,
  Popconfirm,
  Progress,
  Space,
  Typography,
  Upload,
  message,
} from 'antd'
import { useMemo, useState } from 'react'
import {
  courseApi,
  type CourseMaterial,
} from '../api/course'
import { resourceApi } from '../api/resource'
import {
  swapMaterialOrder,
  validateCourseMaterialFile,
} from '../utils/courseMaterials'

interface CourseMaterialsManagerProps {
  courseId: string
  officialMode: boolean
  initialMaterials: CourseMaterial[]
  onChange?: (materials: CourseMaterial[]) => void
}

interface UploadQueueItem {
  key: string
  file: File
  progress: number
  status: 'waiting' | 'uploading' | 'error'
  error?: string
}

const extensionOf = (name: string) => {
  const value = name.split('.').pop()
  return value && value !== name ? value.toUpperCase() : 'FILE'
}

const formatBytes = (size = 0) => {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${Math.ceil(size / 1024)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

export default function CourseMaterialsManager({
  courseId,
  officialMode,
  initialMaterials,
  onChange,
}: CourseMaterialsManagerProps) {
  const [materials, setMaterials] = useState<CourseMaterial[]>(initialMaterials)
  const [queue, setQueue] = useState<UploadQueueItem[]>([])
  const [editing, setEditing] = useState<Record<string, string>>({})
  const sorted = useMemo(
    () => [...materials].sort((a, b) => a.sort_order - b.sort_order),
    [materials],
  )

  const updateMaterials = (next: CourseMaterial[]) => {
    setMaterials(next)
    onChange?.(next)
  }

  const uploadFile = async (item: UploadQueueItem) => {
    setQueue((current) => current.map((row) => row.key === item.key
      ? { ...row, status: 'uploading', progress: 0, error: undefined }
      : row))
    try {
      const upload = officialMode
        ? resourceApi.uploadPlatformAttachment
        : resourceApi.uploadAttachment
      const { data: resource } = await upload(item.file, (progress) => {
        setQueue((current) => current.map((row) => row.key === item.key
          ? { ...row, progress }
          : row))
      })
      const { data: material } = await courseApi.addMaterial(courseId, {
        resource_id: resource.id,
        display_name: item.file.name,
        sort_order: sorted.length + 1,
      })
      setMaterials((current) => {
        const next = [...current.filter((row) => row.id !== material.id), material]
        onChange?.(next)
        return next
      })
      setQueue((current) => current.filter((row) => row.key !== item.key))
      message.success(`${item.file.name} 已添加`)
    } catch {
      setQueue((current) => current.map((row) => row.key === item.key
        ? { ...row, status: 'error', error: '上传或关联失败，请重试' }
        : row))
    }
  }

  const selectFile = (file: File) => {
    const error = validateCourseMaterialFile(file)
    if (error) {
      message.error(error)
      return Upload.LIST_IGNORE
    }
    const item: UploadQueueItem = {
      key: `${file.name}-${file.size}-${file.lastModified}`,
      file,
      progress: 0,
      status: 'waiting',
    }
    setQueue((current) => [...current.filter((row) => row.key !== item.key), item])
    void uploadFile(item)
    return false
  }

  const saveName = async (material: CourseMaterial) => {
    const displayName = (editing[material.id] ?? material.display_name).trim()
    if (!displayName) {
      message.error('资料名称不能为空')
      return
    }
    const { data } = await courseApi.updateMaterial(courseId, material.id, {
      resource_id: material.resource_id,
      display_name: displayName,
      sort_order: material.sort_order,
    })
    updateMaterials(materials.map((item) => item.id === data.id ? data : item))
    setEditing((current) => {
      const next = { ...current }
      delete next[material.id]
      return next
    })
  }

  const move = async (index: number, direction: -1 | 1) => {
    const changes = swapMaterialOrder(sorted, index, direction)
    if (changes.length !== 2) return
    const next = [...materials]
    for (const change of changes) {
      const material = materials.find((item) => item.id === change.id)
      if (!material) continue
      const { data } = await courseApi.updateMaterial(courseId, material.id, {
        resource_id: material.resource_id,
        display_name: material.display_name,
        sort_order: change.sort_order,
      })
      const position = next.findIndex((item) => item.id === data.id)
      next[position] = data
    }
    updateMaterials(next)
  }

  const replace = async (material: CourseMaterial, file: File) => {
    const error = validateCourseMaterialFile(file)
    if (error) {
      message.error(error)
      return Upload.LIST_IGNORE
    }
    const upload = officialMode
      ? resourceApi.uploadPlatformAttachment
      : resourceApi.uploadAttachment
    const { data: resource } = await upload(file)
    const { data } = await courseApi.updateMaterial(courseId, material.id, {
      resource_id: resource.id,
      display_name: material.display_name,
      sort_order: material.sort_order,
    })
    updateMaterials(materials.map((item) => item.id === data.id ? data : item))
    message.success('资料文件已替换')
    return false
  }

  const remove = async (material: CourseMaterial) => {
    await courseApi.removeMaterial(courseId, material.id)
    updateMaterials(materials.filter((item) => item.id !== material.id))
    message.success('资料已移除')
  }

  return (
    <Card className="course-material-manager" title="学习资料">
      <div className="course-material-toolbar">
        <Typography.Text type="secondary">
          学员可在课程详情中下载；支持 PDF、Office 和 ZIP，单文件不超过 200MB。
        </Typography.Text>
        <Upload
          multiple
          showUploadList={false}
          beforeUpload={selectFile}
          accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.zip"
        >
          <Button type="primary" icon={<UploadOutlined />}>上传资料</Button>
        </Upload>
      </div>

      {queue.map((item) => (
        <div className="course-material-queue" key={item.key} aria-live="polite">
          <div><strong>{item.file.name}</strong><small>{item.error || '正在上传'}</small></div>
          <Progress percent={item.progress} status={item.status === 'error' ? 'exception' : 'active'} size="small" />
          {item.status === 'error' && (
            <Button size="small" icon={<ReloadOutlined />} onClick={() => void uploadFile(item)}>重试</Button>
          )}
        </div>
      ))}

      <div className="course-material-list">
        {sorted.length === 0 && queue.length === 0 ? (
          <div className="course-material-empty">暂无学习资料，上传后学员可在课程内下载。</div>
        ) : sorted.map((material, index) => (
          <div className="course-material-admin-row" key={material.id}>
            <span className="course-material-file-icon"><FileOutlined /></span>
            <div className="course-material-meta">
              <Input
                value={editing[material.id] ?? material.display_name}
                onChange={(event) => setEditing((current) => ({ ...current, [material.id]: event.target.value }))}
                onPressEnter={() => void saveName(material)}
                onBlur={() => editing[material.id] !== undefined && void saveName(material)}
                aria-label="资料显示名称"
              />
              <small>{extensionOf(material.display_name)} · {formatBytes(material.resource?.size_bytes)}</small>
            </div>
            <Space className="course-material-actions" wrap size={4}>
              <Button type="text" size="small" icon={<ArrowUpOutlined />} disabled={index === 0} aria-label="上移" onClick={() => void move(index, -1)} />
              <Button type="text" size="small" icon={<ArrowDownOutlined />} disabled={index === sorted.length - 1} aria-label="下移" onClick={() => void move(index, 1)} />
              <Upload showUploadList={false} beforeUpload={(file) => replace(material, file)} accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.zip">
                <Button type="link" size="small">替换</Button>
              </Upload>
              <Popconfirm title="仅移除课程关联，确认继续？" onConfirm={() => void remove(material)}>
                <Button type="text" danger size="small" icon={<DeleteOutlined />}>删除</Button>
              </Popconfirm>
            </Space>
          </div>
        ))}
      </div>
    </Card>
  )
}
