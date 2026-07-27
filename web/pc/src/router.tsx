import { Navigate, Outlet, createBrowserRouter } from 'react-router-dom';
import { AppLayout } from './components/AppLayout';
import { useAuth } from './context/AuthContext';
import { CourseDetailPage } from './pages/CourseDetailPage';
import { CoursesPage } from './pages/CoursesPage';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/LoginPage';
import { RecentPage } from './pages/RecentPage';
import { LessonPlayerPage } from './pages/LessonPlayerPage';

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
