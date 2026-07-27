import {
  BookOutlined,
  DashboardOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  FolderOpenOutlined,
  TagsOutlined,
  TeamOutlined,
  UserOutlined,
  MessageOutlined,
  FileSearchOutlined,
  BgColorsOutlined,
  CreditCardOutlined,
  CloudServerOutlined,
} from '@ant-design/icons'
import { Avatar, Button, Dropdown, Layout, Menu, Space, Typography } from 'antd'
import { useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { logout, tokenRole } from '../api/auth'
import { useDispatch, useSelector } from 'react-redux'
import { clearSession } from '../store/userSlice'
import type { RootState } from '../store'

const { Header, Sider, Content } = Layout

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '工作台' },
  { key: '/tenants', icon: <TeamOutlined />, label: '租户管理' },
  { key: '/users', icon: <UserOutlined />, label: '用户管理' },
  { key: '/courses', icon: <BookOutlined />, label: '课程管理' },
  { key: '/resources', icon: <FolderOpenOutlined />, label: '资源管理' },
  { key: '/resource-categories', icon: <TagsOutlined />, label: '资源分类' },
  { key: '/sms-config', icon: <MessageOutlined />, label: '短信配置' },
  { key: '/audit-logs', icon: <FileSearchOutlined />, label: '操作审计' },
  { key: '/theme-settings', icon: <BgColorsOutlined />, label: '主题设置' },
  { key: '/plans', icon: <CreditCardOutlined />, label: '套餐管理' },
  { key: '/storage-settings', icon: <CloudServerOutlined />, label: '存储配置' },
]

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const dispatch = useDispatch()
  const profile = useSelector((state: RootState) => state.user.profile)
  const role = profile?.role || tokenRole()
  const active = `/${location.pathname.split('/')[1]}` || '/'
  const visibleMenuItems = role === 'superadmin'
    ? menuItems
    : menuItems.filter((item) => item.key !== '/tenants' && item.key !== '/sms-config' && item.key !== '/plans' && item.key !== '/storage-settings' && (role === 'tenant_admin' || item.key !== '/theme-settings'))

  const signOut = () => {
    logout()
    dispatch(clearSession())
    navigate('/login', { replace: true })
  }

  return (
    <Layout className="app-shell">
      <Sider
        width={240}
        collapsedWidth={76}
        collapsed={collapsed}
        trigger={null}
        className="app-sider"
      >
        <div className="brand">
          <div className="brand-mark">I</div>
          {!collapsed && <span>ImaiPlay</span>}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[active]}
          items={visibleMenuItems}
          onClick={({ key }) => navigate(key)}
        />
        {!collapsed && <div className="sider-version">企业学习管理平台</div>}
      </Sider>
      <Layout>
        <Header className="top-header">
          <Button
            type="text"
            className="collapse-button"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed((value) => !value)}
          />
          <Space size={14}>
            <div className="header-identity">
              <Typography.Text strong>{profile?.name || '管理员'}</Typography.Text>
              <Typography.Text type="secondary">{profile?.role || 'Admin'}</Typography.Text>
            </div>
            <Dropdown
              menu={{ items: [{ key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: signOut }] }}
            >
              <Avatar className="user-avatar" icon={<UserOutlined />} />
            </Dropdown>
          </Space>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
