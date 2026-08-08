import { Form, Input, InputNumber, Modal, Select } from 'antd'
import MediaUploader from '../../components/MediaUploader'
import type { CourseDetailController } from './useCourseDetail'
import { resourceDurationSeconds } from './courseDetailModel'

export default function LessonEditor({ controller }: { controller: CourseDetailController }) {
  const { editor, form, saving, contentType, selectedResource, matchingResources, resources, officialMode } = controller
  return <Modal
    title={editor?.kind === 'chapter' ? `${editor.chapter ? '编辑' : '添加'}章节` : `${editor?.lesson ? '编辑' : '添加'}课时`}
    open={Boolean(editor)} width={720} confirmLoading={saving}
    onCancel={controller.closeEditor} onOk={() => void controller.save()} destroyOnHidden
  >
    <Form form={form} layout="vertical" preserve={false}>
      <Form.Item label={editor?.kind === 'chapter' ? '章节标题' : '课时标题'} name="title" rules={[{ required: true, message: '请输入标题' }]}><Input /></Form.Item>
      {editor?.kind === 'lesson' && <>
        <Form.Item label="内容类型" name="content_type" rules={[{ required: true }]}>
          <Select options={[{ value: 'video', label: '视频' }, { value: 'document', label: 'PDF 文档' }, { value: 'text', label: '文本' }]}
            onChange={(nextType) => {
              if (nextType === 'text' || selectedResource?.resource_type !== nextType) {
                controller.setSelectedResource(undefined); form.setFieldValue('resource_id', undefined)
              }
            }}
          />
        </Form.Item>
        {contentType === 'text' ? <Form.Item label="课时正文" name="content_url" rules={[{ required: true, message: '请输入课时正文' }]}>
          <Input.TextArea rows={8} maxLength={20_000} showCount placeholder="输入学员需要阅读的文本内容" />
        </Form.Item> : <>
          <Form.Item label={contentType === 'document' ? '上传 PDF' : '上传视频'} extra={officialMode ? '文件将保存为平台共享资源' : '文件将保存到当前租户资源库'}>
            <MediaUploader value={selectedResource} accept={contentType === 'document' ? 'document' : 'video'}
              upload={controller.uploadResource} onPreview={controller.previewResource}
              onVideoDuration={(seconds) => form.setFieldValue('duration_seconds', seconds)}
              onChange={(resource) => {
                controller.setSelectedResource(resource)
                form.setFieldValue('resource_id', resource?.id)
                const duration = resourceDurationSeconds(resource)
                if (duration) form.setFieldValue('duration_seconds', duration)
              }}
            />
          </Form.Item>
          <Form.Item label="或复用已有资源" name="resource_id" rules={[{ required: true, message: '请上传或选择资源' }]}>
            <Select allowClear showSearch optionFilterProp="label" placeholder="从资源库选择"
              options={matchingResources.map((resource) => ({ value: resource.id, label: `${resource.name}（${resource.resource_type}）` }))}
              onChange={(resourceID) => {
                const resource = resources.find((item) => item.id === resourceID)
                controller.setSelectedResource(resource)
                const duration = resourceDurationSeconds(resource)
                if (duration) form.setFieldValue('duration_seconds', duration)
              }}
            />
          </Form.Item>
          <Form.Item label="时长（秒）" name="duration_seconds"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
        </>}
        <Form.Item label="排序" name="sort_order"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
      </>}
    </Form>
  </Modal>
}
