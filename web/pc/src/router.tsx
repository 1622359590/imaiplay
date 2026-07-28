import { Navigate, Outlet, createBrowserRouter } from 'react-router-dom';
import { lazy } from 'react';
import { AppLayout } from './components/AppLayout';
import { useAuth } from './context/AuthContext';

const CourseDetailPage = lazy(() => import('./pages/CourseDetailPage').then(({ CourseDetailPage }) => ({ default: CourseDetailPage })));
const CoursesPage = lazy(() => import('./pages/CoursesPage').then(({ CoursesPage }) => ({ default: CoursesPage })));
const HomePage = lazy(() => import('./pages/HomePage').then(({ HomePage }) => ({ default: HomePage })));
const LoginPage = lazy(() => import('./pages/LoginPage').then(({ LoginPage }) => ({ default: LoginPage })));
const RecentPage = lazy(() => import('./pages/RecentPage').then(({ RecentPage }) => ({ default: RecentPage })));
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
          { path: 'courses', element: <CoursesPage /> },
          { path: 'courses/:courseId', element: <CourseDetailPage /> },
          { path: 'courses/:courseId/lessons/:lessonId', element: <LessonPlayerPage /> },
          { path: 'recent', element: <RecentPage /> },
        ],
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);
