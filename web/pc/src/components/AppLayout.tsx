import {
  LogoutOutlined,
  ReadOutlined,
} from '@ant-design/icons';
import { Button, Layout } from 'antd';
import { Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useTenantTheme } from '../context/TenantThemeContext';

const { Header, Content } = Layout;

export function AppLayout() {
  const navigate = useNavigate();
  const { logout } = useAuth();
  const theme = useTenantTheme();
  const handleLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  return (
    <Layout className="app-shell">
      <Header className="top-header">
        <button className="brand" type="button" onClick={() => navigate('/')}>
          {theme.logo_url ? <img className="brand-logo-image" src={theme.logo_url} alt="租户 Logo" /> : <span className="brand-mark"><ReadOutlined /></span>}
          <strong>{theme.welcome_text || 'iMaiPlay'}</strong>
        </button>
        <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>退出</Button>
      </Header>
      <Content className="main-content">
        <Outlet />
      </Content>
    </Layout>
  );
}
