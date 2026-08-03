import {
  BookOutlined,
  ClockCircleOutlined,
  HomeOutlined,
  LogoutOutlined,
  ReadOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { Avatar, Button, Layout, Menu, Space, Typography } from 'antd';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { usePortal } from '../context/PortalContext';
import { useTenantTheme } from '../context/TenantThemeContext';
import { portalRoutePath } from '../utils/portalRouting';

const { Header, Content, Footer } = Layout;

const navItems = [
  { key: '/', icon: <HomeOutlined />, label: '学习首页' },
  { key: '/courses', icon: <BookOutlined />, label: '全部课程' },
  { key: '/recent', icon: <ClockCircleOutlined />, label: '最近学习' },
];

export function AppLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { logout } = useAuth();
  const { mode, tenantCode, portal } = usePortal();
  const theme = useTenantTheme();
  const pathFor = (childPath: string) => portalRoutePath(mode, tenantCode, childPath);
  const selectedKey =
    navItems.find((item) => item.key !== '/' && location.pathname.startsWith(pathFor(item.key)))?.key ?? '/';

  useEffect(() => {
    document.title = `${portal?.name || 'iMaiPlay'} | 企业学习中心`;
  }, [portal]);

  const handleLogout = () => {
    logout();
    navigate(pathFor('/login'), { replace: true });
  };

  return (
    <Layout className="app-shell">
      <Header className="top-header app-header">
        <button className="brand" type="button" onClick={() => navigate(pathFor('/'))}>
          {theme.logo_url ? <img className="brand-logo-image" src={theme.logo_url} alt="租户 logo" /> : <span className="brand-mark"><ReadOutlined /></span>}
          <span className="brand-copy">
            <strong>{theme.welcome_text || portal?.name || 'iMaiPlay'}</strong>
            <small>企业学习中心</small>
          </span>
        </button>
        <Menu
          className="top-nav"
          mode="horizontal"
          selectedKeys={[selectedKey]}
          items={navItems}
          onClick={({ key }) => navigate(pathFor(key))}
        />
        <Space className="user-actions">
          <Avatar icon={<UserOutlined />} />
          <Typography.Text>学员</Typography.Text>
          <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>
            退出
          </Button>
        </Space>
      </Header>
      <Content className="main-content page-content">
        <Outlet />
      </Content>
      <Footer className="app-footer">iMaiPlay 企业学习平台 · 让成长持续发生</Footer>
    </Layout>
  );
}
