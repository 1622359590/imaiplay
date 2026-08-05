import { Navigate, Outlet, createBrowserRouter, useLocation } from 'react-router-dom'
import type { PropsWithChildren } from 'react'
import { useSelector } from 'react-redux'
import type { RootState } from './store'
import AdminLayout from './layout/AdminLayout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import { tokenRole } from './api/auth'
import RouteErrorPage from './components/RouteErrorPage'
import { lazyWithReload } from './utils/lazyWithReload'

const Register = lazyWithReload(() => import('./pages/Register'))
const ForgotPassword = lazyWithReload(() => import('./pages/ForgotPassword'))
const Tenants = lazyWithReload(() => import('./pages/Tenants'))
const Users = lazyWithReload(() => import('./pages/Users'))
const Courses = lazyWithReload(() => import('./pages/Courses'))
const CourseDetail = lazyWithReload(() => import('./pages/CourseDetail'))
const Resources = lazyWithReload(() => import('./pages/Resources'))
const ResourceCategories = lazyWithReload(() => import('./pages/ResourceCategories'))
const SMSConfig = lazyWithReload(() => import('./pages/SMSConfig'))
const AuditLogs = lazyWithReload(() => import('./pages/AuditLogs'))
const ThemeSettings = lazyWithReload(() => import('./pages/ThemeSettings'))
const Plans = lazyWithReload(() => import('./pages/Plans'))
const StorageSettings = lazyWithReload(() => import('./pages/StorageSettings'))
const CreateTenant = lazyWithReload(() => import('./pages/CreateTenant'))
const OfficialCourses = lazyWithReload(() => import('./pages/OfficialCourses'))
const DomainSettings = lazyWithReload(() => import('./pages/DomainSettings'))

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
  { path: '/register', element: <Register />, errorElement: <RouteErrorPage /> },
  { path: '/forgot-password', element: <ForgotPassword />, errorElement: <RouteErrorPage /> },
  {
    element: <ProtectedRoute />,
    errorElement: <RouteErrorPage />,
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
