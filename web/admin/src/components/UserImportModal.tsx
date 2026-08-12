import { DownloadOutlined, InboxOutlined } from '@ant-design/icons'
import { Alert, Button, Modal, Space, Typography, Upload, message } from 'antd'
import type { UploadFile } from 'antd'
import { useState } from 'react'
import { userApi } from '../api/user'
import {
  downloadUserImportCSV,
  importResultSummary,
  userImportErrorsCSV,
  userImportTemplateCSV,
  validateUserImportFile,
  type UserImportResult,
} from '../utils/userImport'

interface UserImportModalProps {
  open: boolean
  onClose: () => void
  onImported: () => void
}

export default function UserImportModal({ open, onClose, onImported }: UserImportModalProps) {
  const [file, setFile] = useState<File>()
  const [result, setResult] = useState<UserImportResult>()
  const [uploading, setUploading] = useState(false)

  const resetAndClose = () => {
    setFile(undefined)
    setResult(undefined)
    setUploading(false)
    onClose()
  }

  const submit = async () => {
    if (result) {
      resetAndClose()
      return
    }
    if (!file) {
      message.warning('请选择要导入的文件')
      return
    }
    const error = validateUserImportFile(file)
    if (error) {
      message.error(error)
      return
    }
    setUploading(true)
    try {
      const { data } = await userApi.import(file)
      setResult(data)
      if (data.succeeded > 0) onImported()
    } finally {
      setUploading(false)
    }
  }

  const fileList: UploadFile[] = file ? [{
    uid: `${file.name}-${file.lastModified}`,
    name: file.name,
    size: file.size,
    type: file.type,
    status: 'done',
  }] : []
  const summary = result ? importResultSummary(result) : undefined

  return (
    <Modal
      title="批量导入学员"
      open={open}
      onCancel={resetAndClose}
      onOk={() => void submit()}
      okText={result ? '完成' : '开始导入'}
      okButtonProps={{ disabled: !file && !result }}
      cancelButtonProps={{ disabled: uploading }}
      confirmLoading={uploading}
      closable={!uploading}
      keyboard={!uploading}
      maskClosable={!uploading}
      destroyOnHidden
    >
      <Space direction="vertical" size="middle" className="import-modal-stack">
        <Typography.Paragraph type="secondary" className="import-modal-description">
          模板中的姓名、邮箱和初始密码为必填项。角色留空时默认为学员，也可填写学员或讲师；初始密码至少 8 位。单次最多导入 1000 条。
        </Typography.Paragraph>
        <Button
          icon={<DownloadOutlined />}
          onClick={() => downloadUserImportCSV(userImportTemplateCSV(), '学员批量导入模板.csv')}
        >
          下载导入模板
        </Button>
        {!result && (
          <Upload.Dragger
            className="import-modal-dropzone"
            accept=".csv,.xlsx"
            multiple={false}
            maxCount={1}
            disabled={uploading}
            fileList={fileList}
            beforeUpload={(selected) => {
              const error = validateUserImportFile(selected)
              if (error) {
                message.error(error)
                return Upload.LIST_IGNORE
              }
              setFile(selected)
              return false
            }}
            onRemove={() => {
              setFile(undefined)
              return true
            }}
          >
            <p className="ant-upload-drag-icon"><InboxOutlined /></p>
            <p className="ant-upload-text">点击或拖拽 CSV、XLSX 文件到这里</p>
            <p className="ant-upload-hint">文件不能超过 10MB</p>
          </Upload.Dragger>
        )}
        {result && summary && (
          <Alert
            showIcon
            type={summary.status}
            message={summary.title}
            description={`共处理 ${result.total} 条，成功 ${result.succeeded} 条，失败 ${result.failed} 条。`}
            action={result.errors.length > 0 ? (
              <Button
                size="small"
                icon={<DownloadOutlined />}
                onClick={() => downloadUserImportCSV(userImportErrorsCSV(result.errors), '学员导入错误明细.csv')}
              >
                下载错误明细
              </Button>
            ) : undefined}
          />
        )}
      </Space>
    </Modal>
  )
}
