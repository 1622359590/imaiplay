import { Navigate, Outlet, createBrowserRouter, useLocation } from 'react-router-dom'
import type { PropsWithChildren } from 'react'
import { useSelector } from 'react-redux'
import type { RootState } from './store'
import AdminLayout from './layout/AdminLayout'
import Login from './pages/Login'
import Register from './pages/Register'
import Dashboard from './pages/Dashboard'
import Tenants from './pages/Tenants'
import Users from './pages/Users'
import Courses from './pages/Courses'
import CourseDetail from './pages/CourseDetail'
import Resources from './pages/Resources'
import ResourceCategories from './pages/ResourceCategories'
import { tokenRole } from './api/auth'
import ForgotPassword from './pages/ForgotPassword'
import SMSConfig from './pages/SMSConfig'
import AuditLogs from './pages/AuditLogs'
import ThemeSettings from './pages/ThemeSettings'
import Plans from './pages/Plans'
import StorageSettings from './pages/StorageSettings'
import CreateTenant from './pages/CreateTenant'
import OfficialCourses from './pages/OfficialCourses'

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
])
