import {
  LogoutOutlined,
  ReadOutlined,
} from '@ant-design/icons';
import { Button, Layout } from 'antd';
import { Outlet, useNavigate } from 'react-router-dom';
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
    document.title = `${portal?.name || 'iMaiPlay'} | 企业学习中心`;
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
        <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>退出</Button>
      </Header>
      <Content className="main-content">
        <Outlet />
      </Content>
    </Layout>
  );
}
