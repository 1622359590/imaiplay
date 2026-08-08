import {
  AppstoreOutlined,
  AuditOutlined,
  BgColorsOutlined,
  BookOutlined,
  CloudServerOutlined,
  CreditCardOutlined,
  DashboardOutlined,
  FolderOpenOutlined,
  GlobalOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  MessageOutlined,
  TagsOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { Avatar, Button, Drawer, Dropdown, Layout, Menu, Space, Typography } from 'antd'
import type { MenuProps } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { logout } from '../api/auth'
import {
  initialOpenGroups,
  navigationForRole,
  roleLabel,
  type NavigationIcon,
} from '../config/adminNavigation'
import { useAdminTheme } from '../context/AdminThemeContext'
import type { RootState } from '../store'
import { clearSession } from '../store/userSlice'

const { Header, Sider, Content } = Layout

const iconByName: Record<NavigationIcon, React.ReactNode> = {
  dashboard: <DashboardOutlined />, resource: <FolderOpenOutlined />,
  'resource-category': <TagsOutlined />, course: <BookOutlined />,
  'course-category': <TagsOutlined />, 'official-course': <BookOutlined />,
  users: <TeamOutlined />, theme: <BgColorsOutlined />, domain: <GlobalOutlined />,
  audit: <AuditOutlined />, tenants: <TeamOutlined />, plans: <CreditCardOutlined />,
  sms: <MessageOutlined />, storage: <CloudServerOutlined />,
}

function useViewportWidth() {
  const [width, setWidth] = useState(() => window.innerWidth)
  useEffect(() => {
    const update = () => setWidth(window.innerWidth)
    window.addEventListener('resize', update)
    return () => window.removeEventListener('resize', update)
  }, [])
  return width
}

export default function AdminLayout() {
  const viewportWidth = useViewportWidth()
  const mobile = viewportWidth < 960
  const tablet = viewportWidth >= 960 && viewportWidth < 1200
  const [desktopCollapsed, setDesktopCollapsed] = useState(false)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const dispatch = useDispatch()
  const profile = useSelector((state: RootState) => state.user.profile)
  const theme = useAdminTheme()
  const role = profile?.role
  const groups = navigationForRole(role)
  const [openKeys, setOpenKeys] = useState<string[]>(initialOpenGroups)
  const collapsed = tablet || desktopCollapsed
  const active = location.pathname === '/' ? '/' : `/${location.pathname.split('/')[1]}`

  const menuItems = useMemo<MenuProps['items']>(() => {
    const result: NonNullable<MenuProps['items']> = []
    for (const group of groups) {
      const children = group.items.map((item) => ({ key: item.path, icon: iconByName[item.icon], label: item.label }))
      if (group.label) result.push({ key: group.key, label: group.label, children })
      else result.push(...children)
    }
    return result
  }, [groups])

  const changeOpenKeys = (next: string[]) => {
    setOpenKeys(next)
  }

  const signOut = () => {
    logout()
    dispatch(clearSession())
    navigate('/login', { replace: true })
  }

  const brand = (compact = false) => (
    <div className="brand">
      {theme.logoURL
        ? <img className="brand-logo-image" src={theme.logoURL} alt={`${theme.brandName} Logo`} />
        : <span className="brand-mark" aria-hidden="true"><AppstoreOutlined /></span>}
      {!compact && <span title={theme.brandName}>{theme.brandName}</span>}
    </div>
  )

  const menu = (compact = false) => (
    <Menu
      theme="light"
      mode="inline"
      inlineCollapsed={compact}
      selectedKeys={[active]}
      openKeys={compact ? undefined : openKeys}
      items={menuItems}
      onOpenChange={changeOpenKeys}
      onClick={({ key }) => {
        navigate(key)
        setDrawerOpen(false)
      }}
    />
  )

  return (
    <Layout className="app-shell">
      {!mobile && (
        <Sider width={200} collapsedWidth={76} collapsed={collapsed} trigger={null} className="app-sider">
          <div className="app-sider-inner">
            {brand(collapsed)}
            <nav className="app-menu-scroll" aria-label="后台主导航">{menu(collapsed)}</nav>
            {!collapsed && <div className="sider-version">企业学习管理平台</div>}
          </div>
        </Sider>
      )}
      <Drawer className="admin-nav-drawer" placement="left" width={260} open={mobile && drawerOpen} onClose={() => setDrawerOpen(false)} closable={false} styles={{ body: { padding: 0 } }}>
        {brand(false)}
        <nav className="app-menu-scroll" aria-label="后台主导航">{menu(false)}</nav>
      </Drawer>
      <Layout className="app-main">
        <Header className="top-header">
          <Button
            type="text"
            className="collapse-button"
            aria-label={mobile ? '打开导航' : collapsed ? '展开导航' : '收起导航'}
            icon={mobile || collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => mobile ? setDrawerOpen(true) : !tablet && setDesktopCollapsed((value) => !value)}
          />
          <Space size={12}>
            <div className="header-identity">
              <Typography.Text strong>{profile?.name}</Typography.Text>
              <Typography.Text type="secondary">{roleLabel(role)}</Typography.Text>
            </div>
            <Dropdown menu={{ items: [{ key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: signOut }] }}>
              <Button type="text" className="avatar-button" aria-label="打开账户菜单"><Avatar className="user-avatar" icon={<UserOutlined />} /></Button>
            </Dropdown>
          </Space>
        </Header>
        <Content className="app-content"><Outlet /></Content>
      </Layout>
    </Layout>
  )
}
