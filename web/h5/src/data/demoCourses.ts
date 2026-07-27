import type { Course } from '../types/course'

export const demoCourses: Course[] = [
  {
    id: 'leadership-101',
    title: '卓越管理者的领导力进阶',
    description: '从目标管理到团队激励，构建适合新任管理者的系统领导力。',
    cover: 'linear-gradient(135deg, #0e55ce 0%, #47a2ff 100%)',
    instructor: '陈欣 · 组织发展顾问',
    progress: 68,
    lessonCount: 12,
    duration: 186,
    category: '领导力',
    chapters: [
      {
        id: 'chapter-1',
        title: '第一章 · 从业务骨干到管理者',
        lessons: [
          { id: 'lesson-1', title: '管理角色与关键转变', duration: 16, completed: true },
          { id: 'lesson-2', title: '建立你的管理操作系统', duration: 22, completed: true },
          { id: 'lesson-3', title: '管理者的时间分配', duration: 18 },
        ],
      },
      {
        id: 'chapter-2',
        title: '第二章 · 打造高绩效团队',
        lessons: [
          { id: 'lesson-4', title: '将战略拆解为团队目标', duration: 24 },
          { id: 'lesson-5', title: '高质量一对一沟通', duration: 20 },
          { id: 'lesson-6', title: '反馈、认可与激励', duration: 19 },
        ],
      },
    ],
  },
  {
    id: 'ai-office',
    title: 'AI 时代的高效办公实践',
    description: '掌握生成式 AI 在文档、分析与协作中的实用工作流。',
    cover: 'linear-gradient(135deg, #4935cb 0%, #8676ff 100%)',
    instructor: '林一 · 数字化学习专家',
    progress: 25,
    lessonCount: 9,
    duration: 124,
    category: '数字技能',
  },
  {
    id: 'compliance',
    title: '企业信息安全与合规必修课',
    description: '识别常见安全风险，建立清晰的数据与隐私保护意识。',
    cover: 'linear-gradient(135deg, #007b79 0%, #36c6ae 100%)',
    instructor: '企业安全学院',
    progress: 0,
    lessonCount: 6,
    duration: 72,
    category: '合规必修',
  },
]

export function findDemoCourse(id: string): Course {
  return demoCourses.find((course) => course.id === id) ?? demoCourses[0]
}
