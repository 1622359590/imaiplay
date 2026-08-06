import { Modal, Spin } from 'antd'
import type { CourseDetailController } from './useCourseDetail'

export default function ResourcePreviewModal({ controller }: { controller: CourseDetailController }) {
  const { previewTarget, previewLoading, preview } = controller
  return <Modal title={`预览：${previewTarget?.name || ''}`} open={Boolean(previewTarget)} width={960} footer={null} onCancel={controller.closePreview} destroyOnHidden>
    {previewLoading ? <div className="center-spin"><Spin size="large" /></div>
      : preview?.resourceType === 'video' ? <video src={preview.url} controls style={{ display: 'block', width: '100%', maxHeight: '70vh', background: '#000' }} />
      : preview ? <iframe src={preview.url} title={preview.name} style={{ display: 'block', width: '100%', height: '70vh', border: 0 }} />
      : null}
  </Modal>
}
