import { AppOutline, UnorderedListOutline, UserOutline } from 'antd-mobile-icons'
import { TabBar } from 'antd-mobile'
import { useLocation, useNavigate } from 'react-router-dom'
import { useTenantTheme } from '../context/TenantThemeContext'

export function AppTabBar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { routePath } = useTenantTheme()
  const tabs = [
    { key: routePath('/'), title: '首页', icon: <AppOutline /> },
    { key: routePath('/courses'), title: '课程', icon: <UnorderedListOutline /> },
    { key: routePath('/profile'), title: '我的', icon: <UserOutline /> },
  ]
  const activeKey =
    tabs.find((tab) => tab.key !== routePath('/') && location.pathname.startsWith(tab.key))?.key ??
    routePath('/')

  return (
    <div className="app-tabbar">
      <TabBar activeKey={activeKey} onChange={navigate} safeArea>
        {tabs.map((tab) => (
          <TabBar.Item key={tab.key} icon={tab.icon} title={tab.title} />
        ))}
      </TabBar>
    </div>
  )
}
