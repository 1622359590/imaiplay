import { BookOutlined } from '@ant-design/icons'
import { Alert, Empty, List, Modal, Space, Switch, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import type { Course } from '../api/course'
import { officialCourseApi } from '../api/officialCourse'
import { updateOfficialCourseEnabled } from '../utils/officialCourses'

interface OfficialCoursePickerProps {
  open: boolean
  onClose: () => void
}

export default function OfficialCoursePicker({
  open,
  onClose,
}: OfficialCoursePickerProps) {
  const [items, setItems] = useState<Course[]>([])
  const [loading, setLoading] = useState(false)
  const [loadFailed, setLoadFailed] = useState(false)
  const [savingIds, setSavingIds] = useState<Set<string>>(new Set())

  const load = async () => {
    setLoading(true)
    setLoadFailed(false)
    try {
      const data = await officialCourseApi.list({ page: 1, page_size: 100 })
      setItems(data.items || [])
    } catch {
      setLoadFailed(true)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (open) void load()
  }, [open])

  const toggle = async (course: Course, enabled: boolean) => {
    const previous = course.enabled === true
    setItems((current) => updateOfficialCourseEnabled(current, course.id, enabled))
    setSavingIds((current) => new Set(current).add(course.id))
    try {
      await officialCourseApi.enable(course.id, enabled)
      message.success(enabled ? '官方课程已添加' : '官方课程已移除')
    } catch {
      setItems((current) => updateOfficialCourseEnabled(current, course.id, previous))
      message.error('保存失败，请稍后重试')
    } finally {
      setSavingIds((current) => {
        const next = new Set(current)
        next.delete(course.id)
        return next
      })
    }
  }

  return (
    <Modal
      title={(
        <Space>
          <BookOutlined />
          添加官方课程
        </Space>
      )}
      open={open}
      width={780}
      footer={null}
      onCancel={onClose}
      destroyOnHidden
    >
      <p className="muted official-course-picker-description">
        启用平台维护的官方课程后，学院学员即可开始学习；课程内容会随平台更新。
      </p>
      {loadFailed && (
        <Alert
          type="error"
          showIcon
          message="官方课程加载失败"
          description="请关闭弹窗后重试。"
          className="official-course-picker-alert"
        />
      )}
      <List<Course>
        className="official-course-picker-list"
        loading={loading}
        dataSource={items}
        locale={{ emptyText: <Empty description="暂无已发布的官方课程" /> }}
        pagination={items.length > 8 ? { pageSize: 8, hideOnSinglePage: true } : false}
        renderItem={(course) => (
          <List.Item className="official-course-picker-item">
            <div className="official-course-picker-row">
              <div className="official-course-picker-copy">
                <Space size={8} wrap>
                  <strong>{course.title}</strong>
                  <Tag color="success">已发布</Tag>
                </Space>
                <div className="muted">{course.description || '暂无简介'}</div>
              </div>
              <Space size={8}>
                <span className="muted">添加到学院</span>
              <Switch
                checked={course.enabled === true}
                loading={savingIds.has(course.id)}
                checkedChildren="已添加"
                unCheckedChildren="未添加"
                onChange={(enabled) => void toggle(course, enabled)}
              />
              </Space>
            </div>
          </List.Item>
        )}
      />
    </Modal>
  )
}
