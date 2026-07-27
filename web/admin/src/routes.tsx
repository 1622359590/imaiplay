import { Navigate, Outlet, createBrowserRouter, useLocation } from 'react-router-dom'
import { useSelector } from 'react-redux'
import type { RootState } from './store'
import AdminLayout from './layout/AdminLayout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Tenants from './pages/Tenants'
import Users from './pages/Users'
import Courses from './pages/Courses'
import CourseDetail from './pages/CourseDetail'
import Resources from './pages/Resources'
import ResourceCategories from './pages/ResourceCategories'

function ProtectedRoute() {
  const token = useSelector((state: RootState) => state.user.token)
  const location = useLocation()
  return token ? <Outlet /> : <Navigate to="/login" replace state={{ from: location.pathname }} />
}

function GuestRoute() {
  const token = useSelector((state: RootState) => state.user.token)
  return token ? <Navigate to="/" replace /> : <Login />
}

export const router = createBrowserRouter([
  { path: '/login', element: <GuestRoute /> },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AdminLayout />,
        children: [
          { path: '/', element: <Dashboard /> },
          { path: '/tenants', element: <Tenants /> },
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
