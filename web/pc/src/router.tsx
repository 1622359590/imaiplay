import { Navigate, Outlet, createBrowserRouter, useLocation } from 'react-router-dom';
import { lazy, useEffect, useState } from 'react';
import { Spin } from 'antd';
import { AppLayout } from './components/AppLayout';
import { useAuth } from './context/AuthContext';
import { PortalProvider, usePortal } from './context/PortalContext';
import { TenantThemeProvider } from './context/TenantThemeContext';
import { PortalErrorPage } from './pages/PortalErrorPage';
import {
  legacyPortalRedirect,
  restoredLegacyPortalTarget,
  type LegacySessionKind,
} from './utils/portalRouting';
import {
  bindPortalSessionToPortal,
  isPortalSessionToken,
  portalSessionMatchesPortal,
  readPortalAccessToken,
  readValidLegacyStaffRole,
} from './api/authSession';
import { resolveSessionPortal } from './api/portal';
import { setActivePortalIdentity } from './api/portalSession';

const CourseDetailPage = lazy(() => import('./pages/CourseDetailPage').then(({ CourseDetailPage }) => ({ default: CourseDetailPage })));
const CourseListPage = lazy(() => import('./pages/CourseListPage').then(({ CourseListPage }) => ({ default: CourseListPage })));
const HomePage = lazy(() => import('./pages/HomePage').then(({ HomePage }) => ({ default: HomePage })));
const LoginPage = lazy(() => import('./pages/LoginPage').then(({ LoginPage }) => ({ default: LoginPage })));
const OrganizationSelectPage = lazy(() => import('./pages/OrganizationSelectPage').then(({ OrganizationSelectPage }) => ({ default: OrganizationSelectPage })));
const LessonPlayerPage = lazy(() => import('./pages/LessonPlayerPage').then(({ LessonPlayerPage }) => ({ default: LessonPlayerPage })));
const RecentPage = lazy(() => import('./pages/RecentPage').then(({ RecentPage }) => ({ default: RecentPage })));

function ProtectedRoute() {
  const { authenticated } = useAuth();
  const { loading, error, mode, portal } = usePortal();
  if (loading) return <Spin fullscreen />;
  if (error || !portal) return <PortalErrorPage error={error} />;
  if (!authenticated || !portalSessionMatchesPortal(portal)) {
    const loginPath = mode === 'default' ? `/t/${encodeURIComponent(portal.code)}/login` : '/login';
    return <Navigate to={loginPath} replace state={{ from: window.location.pathname }} />;
  }
  return <Outlet />;
}

function PortalApplication() {
  return (
    <PortalProvider>
      <TenantThemeProvider><Outlet /></TenantThemeProvider>
    </PortalProvider>
  );
}

function CustomDomainOrPlatformEntry() {
  const { mode } = usePortal();
  return mode === 'platform' ? <Navigate to="/login" replace /> : <ProtectedRoute />;
}

function PortalLoginRoute() {
  const { loading, error } = usePortal();
  if (loading) return <Spin fullscreen />;
  if (error) return <PortalErrorPage error={error} />;
  return <LoginPage />;
}

function LegacyPortalRedirect() {
  const location = useLocation();
  const { mode } = usePortal();
  const sessionKind: LegacySessionKind = readValidLegacyStaffRole()
    ? 'staff'
    : isPortalSessionToken(readPortalAccessToken()) ? 'learner' : 'none';
  const redirect = legacyPortalRedirect(location.pathname, mode, sessionKind);

  if (redirect.action === 'route') return <Navigate to={redirect.target} replace />;
  if (redirect.action === 'document') return <DocumentNavigation target={redirect.target} />;
  return <LegacyPortalSessionRestore childPath={redirect.childPath} />;
}

function DocumentNavigation({ target }: { target: string }) {
  useEffect(() => {
    window.location.assign(target);
  }, [target]);
  return <Spin fullscreen />;
}

function LegacyPortalSessionRestore({ childPath }: { childPath: string }) {
  const [state, setState] = useState<{ target?: string; error?: unknown }>({});

  useEffect(() => {
    let cancelled = false;
    setState({});
    void resolveSessionPortal()
      .then((portal) => {
        if (cancelled) return;
        if (!bindPortalSessionToPortal(portal)) {
          setState({ target: '/login' });
          return;
        }
        setActivePortalIdentity(portal);
        setState({ target: restoredLegacyPortalTarget(portal.code, childPath) });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        const status = typeof error === 'object' && error !== null && 'response' in error
          ? (error as { response?: { status?: number } }).response?.status
          : undefined;
        setState(status === 401 ? { target: '/login' } : { error });
      });

    return () => { cancelled = true; };
  }, [childPath]);

  if (state.error) return <PortalErrorPage error={state.error} />;
  if (state.target) return <Navigate to={state.target} replace />;
  return <Spin fullscreen />;
}

const portalChildren = [
  {
    element: <AppLayout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: 'courses', element: <CourseListPage /> },
      { path: 'courses/:courseId', element: <CourseDetailPage /> },
      { path: 'courses/:courseId/lessons/:lessonId', element: <LessonPlayerPage /> },
      { path: 'recent', element: <RecentPage /> },
    ],
  },
];

export const router = createBrowserRouter([
  {
    element: <PortalApplication />,
    children: [
      { path: '/login', element: <PortalLoginRoute /> },
      { path: '/select-organization', element: <OrganizationSelectPage /> },
      { path: '/t/:tenantCode/login', element: <PortalLoginRoute /> },
      { path: '/t/:tenantCode', element: <ProtectedRoute />, children: portalChildren },
      { path: '/pc/login', element: <LegacyPortalRedirect /> },
      { path: '/pc/*', element: <LegacyPortalRedirect /> },
      { path: '/', element: <CustomDomainOrPlatformEntry />, children: portalChildren },
      { path: '*', element: <Navigate to="/login" replace /> },
    ],
  },
]);
