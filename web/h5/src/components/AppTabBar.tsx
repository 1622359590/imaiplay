import { AppOutline, UnorderedListOutline, UserOutline } from 'antd-mobile-icons'
import { TabBar } from 'antd-mobile'
import { useLocation, useNavigate } from 'react-router-dom'

const tabs = [
  { key: '/', title: '首页', icon: <AppOutline /> },
  { key: '/courses', title: '课程', icon: <UnorderedListOutline /> },
  { key: '/profile', title: '我的', icon: <UserOutline /> },
]

export function AppTabBar() {
  const navigate = useNavigate()
  const location = useLocation()
  const activeKey =
    tabs.find((tab) => tab.key !== '/' && location.pathname.startsWith(tab.key))?.key ?? '/'

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
