import type { AdminSessionRole } from '../api/authSession'

export type AdminRole = AdminSessionRole
export type NavigationIcon =
  | 'dashboard' | 'resource' | 'resource-category' | 'course'
  | 'course-category' | 'official-course' | 'users' | 'theme'
  | 'domain' | 'audit' | 'tenants' | 'plans' | 'sms' | 'storage'

export interface NavigationItem {
  path: string
  label: string
  icon: NavigationIcon
  roles: AdminRole[]
}

export interface NavigationGroup {
  key: string
  label?: string
  roles: AdminRole[]
  items: NavigationItem[]
}

const tenant: AdminRole[] = ['tenant_admin']
const tenantAndInstructor: AdminRole[] = ['tenant_admin', 'instructor']
const platform: AdminRole[] = ['superadmin']
const platformAndTenant: AdminRole[] = ['superadmin', 'tenant_admin']

export const adminNavigation: NavigationGroup[] = [
  { key: 'home', roles: tenant, items: [{ path: '/', label: '首页概览', icon: 'dashboard', roles: tenant }] },
  {
    key: 'resource-center', label: '资源管理', roles: tenant,
    items: [
      { path: '/resources', label: '资源列表', icon: 'resource', roles: tenantAndInstructor },
      { path: '/resource-categories', label: '资源分类', icon: 'resource-category', roles: tenant },
    ],
  },
  {
    key: 'course-center', label: '课程中心', roles: tenant,
    items: [
      { path: '/courses', label: '课程管理', icon: 'course', roles: tenantAndInstructor },
      { path: '/course-categories', label: '课程分类', icon: 'course-category', roles: platformAndTenant },
      { path: '/official-courses', label: '官方课程', icon: 'official-course', roles: platformAndTenant },
    ],
  },
  { key: 'learner-center', label: '学员管理', roles: tenant, items: [{ path: '/users', label: '学员与成员', icon: 'users', roles: platformAndTenant }] },
  {
    key: 'site-settings', label: '站点设置', roles: tenant,
    items: [
      { path: '/theme-settings', label: '主题设置', icon: 'theme', roles: tenant },
      { path: '/domain-settings', label: '域名设置', icon: 'domain', roles: tenant },
    ],
  },
  { key: 'security', label: '安全审计', roles: tenant, items: [{ path: '/audit-logs', label: '审计日志', icon: 'audit', roles: platformAndTenant }] },
  { key: 'teaching-workbench', roles: ['instructor'], items: [{ path: '/', label: '教学工作台', icon: 'dashboard', roles: ['instructor'] }] },
  { key: 'instructor-courses', label: '课程中心', roles: ['instructor'], items: [{ path: '/courses', label: '我的课程', icon: 'course', roles: tenantAndInstructor }] },
  { key: 'instructor-resources', label: '资源管理', roles: ['instructor'], items: [{ path: '/resources', label: '资源列表', icon: 'resource', roles: tenantAndInstructor }] },
  { key: 'platform-home', roles: platform, items: [{ path: '/', label: '平台概览', icon: 'dashboard', roles: platform }] },
  {
    key: 'tenant-operations', label: '租户运营', roles: platform,
    items: [
      { path: '/tenants', label: '租户管理', icon: 'tenants', roles: platform },
      { path: '/plans', label: '套餐管理', icon: 'plans', roles: platform },
    ],
  },
  {
    key: 'platform-content', label: '内容中心', roles: platform,
    items: [
      { path: '/official-courses', label: '官方课程', icon: 'official-course', roles: platformAndTenant },
      { path: '/course-categories', label: '官方课程分类', icon: 'course-category', roles: platformAndTenant },
      { path: '/users', label: '全平台账号', icon: 'users', roles: platformAndTenant },
    ],
  },
  {
    key: 'platform-settings', label: '平台配置', roles: platform,
    items: [
      { path: '/sms-config', label: '短信服务', icon: 'sms', roles: platform },
      { path: '/storage-settings', label: '存储服务', icon: 'storage', roles: platform },
      { path: '/audit-logs', label: '审计日志', icon: 'audit', roles: platformAndTenant },
    ],
  },
]

const routeRoles = new Map<string, AdminRole[]>()
for (const group of adminNavigation) {
  for (const item of group.items) {
    routeRoles.set(item.path, [...new Set([...(routeRoles.get(item.path) || []), ...item.roles])])
  }
}
routeRoles.set('/tenants/create', platform)
routeRoles.set('/courses/:id', tenantAndInstructor)
routeRoles.set('/official-courses/:id', platform)

export function navigationForRole(role: string | undefined): NavigationGroup[] {
  if (!role) return []
  return adminNavigation
    .filter((group) => group.roles.includes(role as AdminRole))
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => item.roles.includes(role as AdminRole)),
    }))
    .filter((group) => group.items.length > 0)
}

export function pathsForRole(role: string | undefined): string[] {
  return navigationForRole(role).flatMap((group) => group.items.map((item) => item.path))
}

export function allowedRolesForPath(path: string): AdminRole[] {
  return routeRoles.get(path) || []
}

export function canAccessPath(role: string | undefined, path: string): boolean {
  return Boolean(role && allowedRolesForPath(path).includes(role as AdminRole))
}

export function requiredOpenGroups(path: string, role?: string): string[] {
  const firstSegment = `/${path.split('/').filter(Boolean)[0] || ''}`
  const matches = adminNavigation.filter((group) =>
    group.label && (!role || group.roles.includes(role as AdminRole)) &&
    group.items.some((item) => item.path === firstSegment))
  return (role ? matches : matches.slice(0, 1)).map((group) => group.key)
}

export function roleLabel(role: string | undefined): string {
  return ({ superadmin: '总管理员', tenant_admin: '站长', instructor: '讲师' } as Record<string, string>)[role || ''] || '管理员'
}
