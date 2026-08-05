import { isAxiosError } from 'axios'
import { Button, Result, Spin } from 'antd'
import { useEffect, useState, type PropsWithChildren } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Navigate, Outlet, createBrowserRouter, useLocation } from 'react-router-dom'
import { getCurrentUser } from './api/auth'
import type { AdminRole } from './config/adminNavigation'
import { allowedRolesForPath } from './config/adminNavigation'
import RouteErrorPage from './components/RouteErrorPage'
import AdminLayout from './layout/AdminLayout'
import Login from './pages/Login'
import type { AppDispatch, RootState } from './store'
import { clearSession, setProfile } from './store/userSlice'
import { lazyWithReload } from './utils/lazyWithReload'

const Dashboard = lazyWithReload(() => import('./pages/Dashboard'))
const Register = lazyWithReload(() => import('./pages/Register'))
const ForgotPassword = lazyWithReload(() => import('./pages/ForgotPassword'))
const Tenants = lazyWithReload(() => import('./pages/Tenants'))
const Users = lazyWithReload(() => import('./pages/Users'))
const Courses = lazyWithReload(() => import('./pages/Courses'))
const CourseDetail = lazyWithReload(() => import('./pages/CourseDetail'))
const CourseCategories = lazyWithReload(() => import('./pages/CourseCategories'))
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
  const profile = useSelector((state: RootState) => state.user.profile)
  const dispatch = useDispatch<AppDispatch>()
  const location = useLocation()
  const [restoring, setRestoring] = useState(Boolean(token && !profile))
  const [restoreFailed, setRestoreFailed] = useState(false)
  const [retry, setRetry] = useState(0)

  useEffect(() => {
    if (!token || profile) {
      setRestoring(false)
      return
    }
    let active = true
    setRestoring(true)
    setRestoreFailed(false)
    void getCurrentUser()
      .then((user) => active && dispatch(setProfile(user)))
      .catch((error) => {
        if (!active) return
        if (isAxiosError(error) && (error.response?.status === 401 || error.response?.status === 403)) {
          dispatch(clearSession())
        } else {
          setRestoreFailed(true)
        }
      })
      .finally(() => active && setRestoring(false))
    return () => { active = false }
  }, [dispatch, profile, retry, token])

  if (!token) return <Navigate to="/login" replace state={{ from: location.pathname }} />
  if (restoreFailed) return <Result status="warning" title="暂时无法恢复登录信息" extra={<Button type="primary" onClick={() => setRetry((value) => value + 1)}>重新加载</Button>} />
  if (restoring || !profile) return <div className="profile-bootstrap"><Spin size="large" tip="正在恢复登录信息" /></div>
  return <Outlet />
}

function GuestRoute() {
  const token = useSelector((state: RootState) => state.user.token)
  return token ? <Navigate to="/" replace /> : <Login />
}

function RoleRoute({ allow, children }: PropsWithChildren<{ allow: AdminRole[] }>) {
  const role = useSelector((state: RootState) => state.user.profile?.role)
  return allow.includes(role as AdminRole) ? <>{children}</> : <Navigate to="/" replace />
}

const roleRoute = (path: string, element: React.ReactNode) => (
  <RoleRoute allow={allowedRolesForPath(path)}>{element}</RoleRoute>
)

export const router = createBrowserRouter([
  { path: '/login', element: <GuestRoute /> },
  { path: '/register', element: <Register />, errorElement: <RouteErrorPage /> },
  { path: '/forgot-password', element: <ForgotPassword />, errorElement: <RouteErrorPage /> },
  {
    element: <ProtectedRoute />,
    errorElement: <RouteErrorPage />,
    children: [{
      element: <AdminLayout />,
      children: [
        { path: '/', element: roleRoute('/', <Dashboard />) },
        { path: '/tenants', element: roleRoute('/tenants', <Tenants />) },
        { path: '/tenants/create', element: roleRoute('/tenants/create', <CreateTenant />) },
        { path: '/plans', element: roleRoute('/plans', <Plans />) },
        { path: '/storage-settings', element: roleRoute('/storage-settings', <StorageSettings />) },
        { path: '/official-courses', element: roleRoute('/official-courses', <OfficialCourses />) },
        { path: '/official-courses/:id', element: roleRoute('/official-courses/:id', <CourseDetail />) },
        { path: '/domain-settings', element: roleRoute('/domain-settings', <DomainSettings />) },
        { path: '/sms-config', element: roleRoute('/sms-config', <SMSConfig />) },
        { path: '/audit-logs', element: roleRoute('/audit-logs', <AuditLogs />) },
        { path: '/theme-settings', element: roleRoute('/theme-settings', <ThemeSettings />) },
        { path: '/users', element: roleRoute('/users', <Users />) },
        { path: '/courses', element: roleRoute('/courses', <Courses />) },
        { path: '/courses/:id', element: roleRoute('/courses/:id', <CourseDetail />) },
        { path: '/course-categories', element: roleRoute('/course-categories', <CourseCategories />) },
        { path: '/resources', element: roleRoute('/resources', <Resources />) },
        { path: '/resource-categories', element: roleRoute('/resource-categories', <ResourceCategories />) },
      ],
    }],
  },
  { path: '*', element: <Navigate to="/" replace /> },
], { basename: '/admin' })
