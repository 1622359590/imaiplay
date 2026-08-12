import { Modal, Spin } from 'antd'
import type { CourseDetailController } from './useCourseDetail'

export default function ResourcePreviewModal({ controller }: { controller: CourseDetailController }) {
  const { previewTarget, previewLoading, preview } = controller
  return <Modal title={`预览：${previewTarget?.name || ''}`} open={Boolean(previewTarget)} width={960} footer={null} onCancel={controller.closePreview} destroyOnHidden>
    {previewLoading ? <div className="center-spin"><Spin size="large" /></div>
      : preview?.resourceType === 'video' ? <video className="resource-preview-video" src={preview.url} controls />
      : preview ? <iframe className="resource-preview-frame" src={preview.url} title={preview.name} />
      : null}
  </Modal>
}
