import { Navigate, Outlet, createBrowserRouter } from 'react-router-dom';
import { lazy } from 'react';
import { AppLayout } from './components/AppLayout';
import { useAuth } from './context/AuthContext';

const CourseDetailPage = lazy(() => import('./pages/CourseDetailPage').then(({ CourseDetailPage }) => ({ default: CourseDetailPage })));
const HomePage = lazy(() => import('./pages/HomePage').then(({ HomePage }) => ({ default: HomePage })));
const LoginPage = lazy(() => import('./pages/LoginPage').then(({ LoginPage }) => ({ default: LoginPage })));
const LessonPlayerPage = lazy(() => import('./pages/LessonPlayerPage').then(({ LessonPlayerPage }) => ({ default: LessonPlayerPage })));

function ProtectedRoute() {
  const { authenticated } = useAuth();
  if (!authenticated) {
    return <Navigate to="/login" replace state={{ from: window.location.pathname }} />;
  }
  return <Outlet />;
}

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { index: true, element: <HomePage /> },
          { path: 'courses', element: <Navigate to="/" replace /> },
          { path: 'courses/:courseId', element: <CourseDetailPage /> },
          { path: 'courses/:courseId/lessons/:lessonId', element: <LessonPlayerPage /> },
          { path: 'recent', element: <Navigate to="/" replace /> },
        ],
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
], { basename: '/pc' });
