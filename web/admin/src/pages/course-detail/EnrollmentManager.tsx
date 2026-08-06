import { PlusOutlined } from '@ant-design/icons'
import { Button, Card, Form, Modal, Popconfirm, Select, Table } from 'antd'
import type { AssignmentType, CourseEnrollment } from '../../api/course'
import type { CourseDetailController } from './useCourseDetail'

export default function EnrollmentManager({ controller }: { controller: CourseDetailController }) {
  const { officialMode, instructor, enrollments, learners, enrollmentForm, enrollmentOpen } = controller
  if (officialMode || instructor) return null
  return <>
    <Card className="course-enrollment-manager" title="学员分配" extra={
      <Button type="primary" icon={<PlusOutlined />} onClick={() => {
        enrollmentForm.setFieldsValue({ assignment_type: 'required' }); controller.setEnrollmentOpen(true)
      }}>分配学员</Button>
    }>
      <Table<CourseEnrollment>
        rowKey="id" dataSource={enrollments} pagination={false} locale={{ emptyText: '暂无已分配学员' }}
        columns={[
          { title: '学员', dataIndex: 'user_id', render: (userID) => { const user = learners.find((item) => item.id === userID); return <div><strong>{user?.name || userID}</strong>{user?.email && <div className="muted">{user.email}</div>}</div> } },
          { title: '分配类型', dataIndex: 'assignment_type', render: (value: AssignmentType, record) => <Select value={value || 'required'} style={{ width: 110 }} options={[{ value: 'required', label: '必修' }, { value: 'optional', label: '选修' }]} onChange={(next) => void controller.changeAssignment(record.id, next)} /> },
          { title: '操作', width: 100, render: (_, record) => <Popconfirm title="确认移除该学员的课程分配？" onConfirm={() => void controller.removeEnrollment(record.id)}><Button type="link" danger>移除</Button></Popconfirm> },
        ]}
      />
    </Card>
    <Modal title="分配学员" open={enrollmentOpen} onCancel={() => controller.setEnrollmentOpen(false)} onOk={() => void controller.enroll()} destroyOnHidden>
      <Form form={enrollmentForm} layout="vertical" preserve={false}>
        <Form.Item name="user_id" label="学员" rules={[{ required: true, message: '请选择学员' }]}>
          <Select showSearch optionFilterProp="label" options={learners.filter((learner) => !enrollments.some((item) => item.user_id === learner.id)).map((learner) => ({ value: learner.id, label: `${learner.name}（${learner.email}）` }))} />
        </Form.Item>
        <Form.Item name="assignment_type" label="分配类型" rules={[{ required: true }]}>
          <Select options={[{ value: 'required', label: '必修' }, { value: 'optional', label: '选修' }]} />
        </Form.Item>
      </Form>
    </Modal>
  </>
}
