import { lazy, Suspense } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { Skeleton } from 'antd-mobile'
import { PageShell } from './components/PageShell'
import { ProtectedRoute } from './components/ProtectedRoute'

const CourseDetailPage = lazy(() => import('./pages/CourseDetailPage').then(({ CourseDetailPage }) => ({ default: CourseDetailPage })))
const HomePage = lazy(() => import('./pages/HomePage').then(({ HomePage }) => ({ default: HomePage })))
const LoginPage = lazy(() => import('./pages/LoginPage').then(({ LoginPage }) => ({ default: LoginPage })))
const LessonPlayerPage = lazy(() => import('./pages/LessonPlayerPage').then(({ LessonPlayerPage }) => ({ default: LessonPlayerPage })))

function LoadingFallback() {
  return <div className="page-loading"><Skeleton.Title animated /><Skeleton.Paragraph lineCount={4} animated /></div>
}

export default function App() {
  return (
    <Suspense fallback={<LoadingFallback />}><Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<PageShell />}>
          <Route index element={<HomePage />} />
          <Route path="/courses" element={<Navigate to="/" replace />} />
          <Route path="/profile" element={<Navigate to="/" replace />} />
        </Route>
        <Route path="/courses/:id" element={<CourseDetailPage />} />
        <Route path="/courses/:courseId/lessons/:lessonId" element={<LessonPlayerPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes></Suspense>
  )
}
