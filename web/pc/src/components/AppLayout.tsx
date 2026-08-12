import {
  BellOutlined,
  LogoutOutlined,
  ReadOutlined,
} from '@ant-design/icons';
import { Button, Layout, Tooltip } from 'antd';
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { usePortal } from '../context/PortalContext';
import { useTenantTheme } from '../context/TenantThemeContext';
import { portalRoutePath } from '../utils/portalRouting';

const { Header, Content } = Layout;

export function AppLayout() {
  const navigate = useNavigate();
  const { logout } = useAuth();
  const { mode, tenantCode, portal } = usePortal();
  const theme = useTenantTheme();
  const pathFor = (childPath: string) => portalRoutePath(mode, tenantCode, childPath);

  useEffect(() => {
    document.title = portal?.browser_title?.trim() || `${portal?.name || 'iMaiPlay'} | 企业学习中心`;
  }, [portal]);

  const handleLogout = () => {
    logout();
    navigate(pathFor('/login'), { replace: true });
  };

  return (
    <Layout className="app-shell">
      <Header className="top-header">
        <button className="brand" type="button" onClick={() => navigate(pathFor('/'))}>
          {theme.logo_url ? <img className="brand-logo-image" src={theme.logo_url} alt="租户 Logo" /> : <span className="brand-mark"><ReadOutlined /></span>}
          <strong>{theme.welcome_text || portal?.name || 'iMaiPlay'}</strong>
        </button>
        <nav className="learner-top-nav" aria-label="学习中心导航">
          <NavLink className="learner-top-nav-link" end to={pathFor('/')}>学习首页</NavLink>
          <Link className="learner-top-nav-link" to={`${pathFor('/')}#courses`}>全部课程</Link>
          <NavLink className="learner-top-nav-link" to={pathFor('/recent')}>学习记录</NavLink>
        </nav>
        <div className="learner-user-actions">
          <Tooltip title="通知">
            <Button className="learner-notification-button" type="text" aria-label="通知" icon={<BellOutlined />}>
              <span className="learner-notification-dot" />
            </Button>
          </Tooltip>
          <Tooltip title="退出登录">
            <button className="learner-avatar" type="button" onClick={handleLogout} aria-label="退出登录">
              <span>学</span><LogoutOutlined />
            </button>
          </Tooltip>
        </div>
      </Header>
      <Content className="main-content">
        <Outlet />
      </Content>
    </Layout>
  );
}
