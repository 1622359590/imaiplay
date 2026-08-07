import { Form, message, type FormInstance } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useSelector } from 'react-redux'
import { useLocation, useParams } from 'react-router-dom'
import { courseApi, type AssignmentType, type Course, type CourseEnrollment } from '../../api/course'
import { resourceApi, type Resource } from '../../api/resource'
import { normalizePage } from '../../api/types'
import { userApi, type User } from '../../api/user'
import type { UploadedMedia } from '../../components/MediaUploader'
import type { RootState } from '../../store'
import { collectPaginatedItems } from '../../utils/pagination'
import { loadResourcePreview, type ResourcePreview } from '../../utils/resourcePreview'
import { lessonPayload, type Editor, type LessonForm } from './courseDetailModel'

export interface CourseDetailController {
  id: string
  course?: Course
  loading: boolean
  saving: boolean
  officialMode: boolean
  instructor: boolean
  resources: Resource[]
  matchingResources: Resource[]
  enrollments: CourseEnrollment[]
  learners: User[]
  editor?: Editor
  selectedResource?: UploadedMedia
  previewTarget?: UploadedMedia
  preview?: ResourcePreview
  previewLoading: boolean
  enrollmentOpen: boolean
  contentType?: LessonForm['content_type']
  form: FormInstance<LessonForm>
  enrollmentForm: FormInstance<{ user_id: string; assignment_type: AssignmentType }>
  reload(): Promise<void>
  edit(editor: Editor): void
  closeEditor(): void
  save(): Promise<void>
  uploadResource(file: File, onProgress: (percent: number) => void): Promise<Resource>
  previewResource(resource: UploadedMedia): Promise<void>
  closePreview(): void
  enroll(): Promise<void>
  changeAssignment(enrollmentID: string, assignmentType: AssignmentType): Promise<void>
  removeEnrollment(enrollmentID: string): Promise<void>
  removeChapter(chapterID: string): Promise<void>
  removeLesson(chapterID: string, lessonID: string): Promise<void>
  setEnrollmentOpen(open: boolean): void
  setSelectedResource(resource?: UploadedMedia): void
  updateMaterials(materials: NonNullable<Course['materials']>): void
}

export function useCourseDetail(): CourseDetailController {
  const { id = '' } = useParams()
  const location = useLocation()
  const officialMode = location.pathname.startsWith('/official-courses/')
  const role = useSelector((state: RootState) => state.user.profile?.role)
  const instructor = role === 'instructor'
  const [course, setCourse] = useState<Course>()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [editor, setEditor] = useState<Editor>()
  const [resources, setResources] = useState<Resource[]>([])
  const [selectedResource, setSelectedResource] = useState<UploadedMedia>()
  const [previewTarget, setPreviewTarget] = useState<UploadedMedia>()
  const [preview, setPreview] = useState<ResourcePreview>()
  const [previewLoading, setPreviewLoading] = useState(false)
  const [form] = Form.useForm<LessonForm>()
  const [enrollmentForm] = Form.useForm<{ user_id: string; assignment_type: AssignmentType }>()
  const [enrollments, setEnrollments] = useState<CourseEnrollment[]>([])
  const [learners, setLearners] = useState<User[]>([])
  const [enrollmentOpen, setEnrollmentOpen] = useState(false)
  const contentType = Form.useWatch('content_type', form)

  useEffect(() => () => preview?.dispose(), [preview])

  const reload = async () => {
    setLoading(true)
    try {
      const { data } = await courseApi.detail(id)
      setCourse(data)
    } finally {
      setLoading(false)
    }
  }

  const loadResources = async () => {
    try {
      const { data } = await (officialMode ? resourceApi.listPlatform() : resourceApi.list())
      setResources(normalizePage(data).items)
    } catch { setResources([]) }
  }

  const loadEnrollments = async () => {
    if (officialMode || instructor) return
    try {
      const [{ data: assigned }, users] = await Promise.all([
        courseApi.listEnrollments(id),
        collectPaginatedItems(async (page, pageSize) => {
          const { data } = await userApi.list({ page, page_size: pageSize })
          return normalizePage(data)
        }),
      ])
      setEnrollments(assigned)
      setLearners(users.filter((user) => user.role === 'learner' && user.status === 1))
    } catch { setEnrollments([]); setLearners([]) }
  }

  useEffect(() => { void reload() }, [id])
  useEffect(() => { void loadResources() }, [officialMode])
  useEffect(() => { void loadEnrollments() }, [id, officialMode, instructor])

  const edit = (value: Editor) => {
    setEditor(value)
    if (value.kind === 'chapter') {
      form.setFieldsValue(value.chapter || { title: '' })
      setSelectedResource(undefined)
      return
    }
    form.setFieldsValue(value.lesson || { title: '', content_type: 'video', duration_seconds: 0, sort_order: 0 })
    const resource = value.lesson?.resource_id
      ? resources.find((item) => item.id === value.lesson?.resource_id)
      : undefined
    setSelectedResource(resource || (value.lesson?.resource_id ? {
      id: value.lesson.resource_id,
      name: '当前课时资源',
      resource_type: value.lesson.content_type === 'document' ? 'document' : 'video',
      url: '', size_bytes: 0,
    } : undefined))
  }

  const closeEditor = () => {
    setEditor(undefined); setSelectedResource(undefined); form.resetFields()
  }

  const save = async () => {
    const values = await form.validateFields()
    if (!editor) return
    setSaving(true)
    try {
      if (editor.kind === 'chapter') {
        if (editor.chapter) await courseApi.updateChapter(id, editor.chapter.id, values)
        else await courseApi.createChapter(id, values)
      } else if (editor.lesson) {
        await courseApi.updateLesson(id, editor.chapter.id, editor.lesson.id, lessonPayload(values))
      } else {
        await courseApi.createLesson(id, editor.chapter.id, lessonPayload(values))
      }
      message.success('课程内容已保存')
      closeEditor()
      void reload()
    } finally { setSaving(false) }
  }

  const uploadResource = async (file: File, onProgress: (percent: number) => void) => {
    const { data: resource } = officialMode
      ? await resourceApi.uploadPlatform(file, onProgress)
      : await resourceApi.upload(file, onProgress)
    setResources((current) => [resource, ...current.filter((item) => item.id !== resource.id)])
    form.setFieldsValue({ resource_id: resource.id, content_type: resource.resource_type === 'document' ? 'document' : 'video' })
    return resource
  }

  const previewResource = async (resource: UploadedMedia) => {
    setPreviewTarget(resource); setPreviewLoading(true)
    try {
      setPreview(await loadResourcePreview(resource, async () => {
        const response = officialMode ? await resourceApi.platformFile(resource.id) : await resourceApi.file(resource.id)
        return response.data
      }))
    } catch {
      setPreview(undefined); setPreviewTarget(undefined)
      message.error('资源预览加载失败，请稍后重试')
    } finally { setPreviewLoading(false) }
  }

  const closePreview = () => { setPreview(undefined); setPreviewTarget(undefined); setPreviewLoading(false) }
  const enroll = async () => {
    await courseApi.enroll(id, await enrollmentForm.validateFields())
    message.success('学员已分配到课程'); setEnrollmentOpen(false); enrollmentForm.resetFields(); await loadEnrollments()
  }
  const changeAssignment = async (enrollmentID: string, assignmentType: AssignmentType) => {
    await courseApi.updateAssignment(enrollmentID, assignmentType); message.success('分配类型已更新'); await loadEnrollments()
  }
  const removeEnrollment = async (enrollmentID: string) => {
    await courseApi.removeEnrollment(enrollmentID); message.success('课程分配已移除'); await loadEnrollments()
  }
  const removeChapter = async (chapterID: string) => {
    await courseApi.removeChapter(id, chapterID); message.success('章节已删除'); void reload()
  }
  const removeLesson = async (chapterID: string, lessonID: string) => {
    await courseApi.removeLesson(id, chapterID, lessonID); message.success('课时已删除'); void reload()
  }

  return {
    id, course, loading, saving, officialMode, instructor, resources,
    matchingResources: useMemo(() => resources.filter((resource) => resource.resource_type === contentType), [contentType, resources]),
    enrollments, learners, editor, selectedResource, previewTarget, preview,
    previewLoading, enrollmentOpen, contentType, form, enrollmentForm, reload,
    edit, closeEditor, save, uploadResource, previewResource, closePreview, enroll,
    changeAssignment, removeEnrollment, removeChapter, removeLesson,
    setEnrollmentOpen, setSelectedResource,
    updateMaterials: (materials) => setCourse((current) => current ? { ...current, materials } : current),
  }
}
