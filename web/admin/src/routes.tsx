import { Navigate, Outlet, createBrowserRouter, useLocation } from 'react-router-dom'
import type { PropsWithChildren } from 'react'
import { useSelector } from 'react-redux'
import type { RootState } from './store'
import AdminLayout from './layout/AdminLayout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import { tokenRole } from './api/auth'
import { lazy } from 'react'

const Register = lazy(() => import('./pages/Register'))
const ForgotPassword = lazy(() => import('./pages/ForgotPassword'))
const Tenants = lazy(() => import('./pages/Tenants'))
const Users = lazy(() => import('./pages/Users'))
const Courses = lazy(() => import('./pages/Courses'))
const CourseDetail = lazy(() => import('./pages/CourseDetail'))
const Resources = lazy(() => import('./pages/Resources'))
const ResourceCategories = lazy(() => import('./pages/ResourceCategories'))
const SMSConfig = lazy(() => import('./pages/SMSConfig'))
const AuditLogs = lazy(() => import('./pages/AuditLogs'))
const ThemeSettings = lazy(() => import('./pages/ThemeSettings'))
const Plans = lazy(() => import('./pages/Plans'))
const StorageSettings = lazy(() => import('./pages/StorageSettings'))
const CreateTenant = lazy(() => import('./pages/CreateTenant'))
const OfficialCourses = lazy(() => import('./pages/OfficialCourses'))
const DomainSettings = lazy(() => import('./pages/DomainSettings'))

function ProtectedRoute() {
  const token = useSelector((state: RootState) => state.user.token)
  const location = useLocation()
  return token ? <Outlet /> : <Navigate to="/login" replace state={{ from: location.pathname }} />
}

function GuestRoute() {
  const token = useSelector((state: RootState) => state.user.token)
  return token ? <Navigate to="/" replace /> : <Login />
}

function SuperadminOnly({ children }: PropsWithChildren) {
  const role = useSelector((state: RootState) => state.user.profile?.role || tokenRole())
  return role === 'superadmin' ? <>{children}</> : <Navigate to="/" replace />
}

function TenantAdminOnly({ children }: PropsWithChildren) {
  const role = useSelector((state: RootState) => state.user.profile?.role || tokenRole())
  return role === 'tenant_admin' ? <>{children}</> : <Navigate to="/" replace />
}

export const router = createBrowserRouter([
  { path: '/login', element: <GuestRoute /> },
  { path: '/register', element: <Register /> },
  { path: '/forgot-password', element: <ForgotPassword /> },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AdminLayout />,
        children: [
          { path: '/', element: <Dashboard /> },
          { path: '/tenants', element: <SuperadminOnly><Tenants /></SuperadminOnly> },
          { path: '/tenants/create', element: <SuperadminOnly><CreateTenant /></SuperadminOnly> },
          { path: '/plans', element: <SuperadminOnly><Plans /></SuperadminOnly> },
          { path: '/storage-settings', element: <SuperadminOnly><StorageSettings /></SuperadminOnly> },
          { path: '/official-courses', element: <OfficialCourses /> },
          { path: '/domain-settings', element: <TenantAdminOnly><DomainSettings /></TenantAdminOnly> },
          { path: '/sms-config', element: <SuperadminOnly><SMSConfig /></SuperadminOnly> },
          { path: '/audit-logs', element: <AuditLogs /> },
          { path: '/theme-settings', element: <TenantAdminOnly><ThemeSettings /></TenantAdminOnly> },
          { path: '/users', element: <Users /> },
          { path: '/courses', element: <Courses /> },
          { path: '/courses/:id', element: <CourseDetail /> },
          { path: '/resources', element: <Resources /> },
          { path: '/resource-categories', element: <ResourceCategories /> },
        ],
      },
    ],
  },
  { path: '*', element: <Navigate to="/" replace /> },
], { basename: '/admin' })
