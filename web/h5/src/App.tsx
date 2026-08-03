import { lazy, Suspense, useEffect } from 'react'
import { Navigate, Outlet, Route, Routes } from 'react-router-dom'
import { DotLoading, ErrorBlock, Skeleton } from 'antd-mobile'
import { PageShell } from './components/PageShell'
import { ProtectedRoute } from './components/ProtectedRoute'
import { useTenantTheme } from './context/TenantThemeContext'
import { portalErrorContent } from './api/portalResolution'

const CourseDetailPage = lazy(() => import('./pages/CourseDetailPage').then(({ CourseDetailPage }) => ({ default: CourseDetailPage })))
const CoursesPage = lazy(() => import('./pages/CoursesPage').then(({ CoursesPage }) => ({ default: CoursesPage })))
const ForgotPasswordPage = lazy(() => import('./pages/ForgotPasswordPage').then(({ ForgotPasswordPage }) => ({ default: ForgotPasswordPage })))
const HomePage = lazy(() => import('./pages/HomePage').then(({ HomePage }) => ({ default: HomePage })))
const LoginPage = lazy(() => import('./pages/LoginPage').then(({ LoginPage }) => ({ default: LoginPage })))
const LessonPlayerPage = lazy(() => import('./pages/LessonPlayerPage').then(({ LessonPlayerPage }) => ({ default: LessonPlayerPage })))
const ProfilePage = lazy(() => import('./pages/ProfilePage').then(({ ProfilePage }) => ({ default: ProfilePage })))

function LoadingFallback() {
  return <div className="page-loading"><Skeleton.Title animated /><Skeleton.Paragraph lineCount={4} animated /></div>
}

function PortalPublicRoute() {
  const theme = useTenantTheme()
  if (theme.loading) {
    return <div className="loading-state"><DotLoading color="primary" /> 正在加载企业门户</div>
  }
  if (theme.error || !theme.portal) {
    if (theme.mode === 'platform' && !theme.error) {
      return <PlatformLoginRedirect />
    }
    const content = portalErrorContent(theme.error)
    return (
      <ErrorBlock
        status="disconnected"
        title={content.title}
        description={content.description}
      />
    )
  }
  if (theme.mode === 'platform') {
    return <Navigate to={theme.routePath('/')} replace />
  }
  return <Outlet />
}

function PlatformLoginRedirect() {
  useEffect(() => {
    window.location.replace('/login')
  }, [])
  return <div className="loading-state"><DotLoading color="primary" /> 正在前往统一登录</div>
}

function PortalFallbackRedirect() {
  const { routePath } = useTenantTheme()
  return <Navigate to={routePath('/')} replace />
}

export default function App() {
  return (
    <Suspense fallback={<LoadingFallback />}><Routes>
      <Route element={<PortalPublicRoute />}>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/t/:tenantCode/login" element={<LoginPage />} />
        <Route path="/t/:tenantCode/forgot-password" element={<ForgotPasswordPage />} />
      </Route>
      <Route path="/t/:tenantCode" element={<ProtectedRoute />}>
        <Route element={<PageShell />}>
          <Route index element={<HomePage />} />
          <Route path="courses" element={<CoursesPage />} />
          <Route path="profile" element={<ProfilePage />} />
        </Route>
        <Route path="courses/:id" element={<CourseDetailPage />} />
        <Route path="courses/:courseId/lessons/:lessonId" element={<LessonPlayerPage />} />
      </Route>
      <Route element={<ProtectedRoute />}>
        <Route element={<PageShell />}>
          <Route index element={<HomePage />} />
          <Route path="/courses" element={<CoursesPage />} />
          <Route path="/profile" element={<ProfilePage />} />
        </Route>
        <Route path="/courses/:id" element={<CourseDetailPage />} />
        <Route path="/courses/:courseId/lessons/:lessonId" element={<LessonPlayerPage />} />
      </Route>
      <Route path="*" element={<PortalFallbackRedirect />} />
    </Routes></Suspense>
  )
}
