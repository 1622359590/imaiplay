import {
  DeleteOutlined,
  EyeOutlined,
  FileImageOutlined,
  FilePdfOutlined,
  InboxOutlined,
  ReloadOutlined,
  SwapOutlined,
  VideoCameraOutlined,
} from '@ant-design/icons'
import { Alert, Button, Image, Progress, Space, Typography, Upload } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import type { Resource } from '../api/resource'
import { readVideoDurationSeconds } from '../utils/videoDuration'

const { Dragger } = Upload

export type UploadedMedia = Pick<
  Resource,
  'id' | 'name' | 'resource_type' | 'url' | 'size_bytes'
>

export interface MediaUploaderProps {
  value?: UploadedMedia
  onChange?: (value?: UploadedMedia) => void
  accept: 'image' | 'video' | 'document' | 'all'
  upload: (
    file: File,
    onProgress: (percent: number) => void,
    durationSeconds?: number,
  ) => Promise<Resource>
  onPreview?: (value: UploadedMedia) => void | Promise<void>
  onVideoDuration?: (seconds: number) => void
  disabled?: boolean
}

const acceptByKind = {
  image: 'image/jpeg,image/png,image/webp',
  video: 'video/mp4,video/webm',
  document: 'application/pdf',
  all: 'image/jpeg,image/png,image/webp,video/mp4,video/webm,application/pdf',
}

const labelByKind = {
  image: 'JPEG、PNG 或 WebP 图片',
  video: 'MP4 或 WebM 视频',
  document: 'PDF 文档',
  all: '图片、MP4/WebM 视频或 PDF 文档',
}

function formatBytes(value?: number) {
  if (!value) return ''
  if (value >= 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MB`
  return `${Math.max(1, Math.round(value / 1024))} KB`
}

function mediaIcon(type?: Resource['resource_type']) {
  if (type === 'image') return <FileImageOutlined />
  if (type === 'video') return <VideoCameraOutlined />
  return <FilePdfOutlined />
}

function fileType(file: File): Resource['resource_type'] | undefined {
  if (file.type.startsWith('image/')) return 'image'
  if (file.type.startsWith('video/')) return 'video'
  if (file.type === 'application/pdf') return 'document'
  return undefined
}

export default function MediaUploader({
  value,
  onChange,
  accept,
  upload,
  onPreview,
  onVideoDuration,
  disabled,
}: MediaUploaderProps) {
  const [candidate, setCandidate] = useState<File>()
  const [progress, setProgress] = useState(0)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const [previewOpen, setPreviewOpen] = useState(false)
  const candidateURL = useMemo(
    () => candidate && candidate.type.startsWith('image/')
      ? URL.createObjectURL(candidate)
      : undefined,
    [candidate],
  )

  useEffect(() => () => {
    if (candidateURL) URL.revokeObjectURL(candidateURL)
  }, [candidateURL])

  const uploadFile = async (file: File) => {
    const type = fileType(file)
    const allowed = accept === 'all' || type === accept
    if (!type || !allowed) {
      setCandidate(file)
      setError(`请选择${labelByKind[accept]}`)
      return
    }
    setCandidate(file)
    setUploading(true)
    setProgress(0)
    setError('')
    try {
      const duration = type === 'video'
        ? await readVideoDurationSeconds(file).catch(() => undefined)
        : undefined
      const resource = await upload(file, setProgress, duration)
      onChange?.(resource)
      if (duration) onVideoDuration?.(duration)
      setCandidate(undefined)
      setProgress(100)
    } catch {
      setError('上传失败，文件已保留，请重试')
    } finally {
      setUploading(false)
    }
  }

  const picker = (compact = false) => (
    <Upload
      accept={acceptByKind[accept]}
      showUploadList={false}
      disabled={disabled || uploading}
      beforeUpload={(file) => {
        void uploadFile(file)
        return false
      }}
    >
      <Button
        icon={compact ? <SwapOutlined /> : <InboxOutlined />}
        disabled={disabled}
        loading={uploading}
      >
        {compact ? '替换文件' : '选择文件'}
      </Button>
    </Upload>
  )

  return (
    <div className="media-uploader">
      {!value ? (
        <Dragger
          accept={acceptByKind[accept]}
          showUploadList={false}
          disabled={disabled || uploading}
          beforeUpload={(file) => {
            void uploadFile(file)
            return false
          }}
        >
          <p className="ant-upload-drag-icon"><InboxOutlined /></p>
          <p className="media-uploader-title">拖拽文件到这里，或点击选择</p>
          <p className="media-uploader-hint">
            支持 {labelByKind[accept]}，单个文件最大 1024 MB
          </p>
        </Dragger>
      ) : (
        <div className="media-uploader-preview">
          <div className="media-uploader-thumb">
            {value.resource_type === 'image' && value.url ? (
              <img src={value.url} alt={value.name} />
            ) : (
              mediaIcon(value.resource_type)
            )}
          </div>
          <div className="media-uploader-meta">
            <Typography.Text strong ellipsis>{value.name}</Typography.Text>
            <Typography.Text type="secondary">
              {formatBytes(value.size_bytes) || '已上传'}
            </Typography.Text>
          </div>
          <Space wrap className="media-uploader-actions">
            {(value.resource_type === 'image' || onPreview) && (
              <Button
                icon={<EyeOutlined />}
                onClick={() => {
                  if (value.resource_type === 'image' && value.url) {
                    setPreviewOpen(true)
                  } else {
                    void onPreview?.(value)
                  }
                }}
              >
                预览
              </Button>
            )}
            {picker(true)}
            <Button
              danger
              icon={<DeleteOutlined />}
              disabled={disabled || uploading}
              onClick={() => {
                onChange?.(undefined)
                setCandidate(undefined)
                setError('')
              }}
            >
              移除
            </Button>
          </Space>
        </div>
      )}

      {(uploading || candidate) && (
        <div className="media-uploader-status">
          <Space>
            {mediaIcon(candidate ? fileType(candidate) : value?.resource_type)}
            <Typography.Text ellipsis>{candidate?.name}</Typography.Text>
          </Space>
          {uploading && <Progress percent={progress} size="small" />}
        </div>
      )}
      {error && (
        <Alert
          type="error"
          showIcon
          message={error}
          action={candidate ? (
            <Button
              size="small"
              icon={<ReloadOutlined />}
              disabled={uploading}
              onClick={() => void uploadFile(candidate)}
            >
              重试
            </Button>
          ) : undefined}
        />
      )}
      {value?.resource_type === 'image' && value.url && (
        <Image
          src={value.url}
          alt={value.name}
          style={{ display: 'none' }}
          preview={{
            visible: previewOpen,
            onVisibleChange: setPreviewOpen,
          }}
        />
      )}
    </div>
  )
}
