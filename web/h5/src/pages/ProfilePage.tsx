import {
  CheckShieldOutline,
  FileOutline,
  MessageOutline,
  RightOutline,
  SetOutline,
} from 'antd-mobile-icons'
import { Button, List } from 'antd-mobile'
import { useNavigate } from 'react-router-dom'
import { logout } from '../api/auth'

export function ProfilePage() {
  const navigate = useNavigate()
  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="profile-page">
      <header className="profile-header">
        <div className="avatar">学</div>
        <div><span>企业学习者</span><h1>欢迎回来</h1><p>持续学习 · 持续进步</p></div>
      </header>
      <section className="profile-achievement">
        <div><strong>18</strong><span>累计课时</span></div>
        <div><strong>32.5h</strong><span>学习时长</span></div>
        <div><strong>3</strong><span>我的证书</span></div>
      </section>
      <section className="profile-card">
        <h2>学习服务</h2>
        <List>
          <List.Item prefix={<FileOutline />} arrow={<RightOutline />}>学习记录</List.Item>
          <List.Item prefix={<CheckShieldOutline />} arrow={<RightOutline />}>我的证书</List.Item>
          <List.Item prefix={<MessageOutline />} arrow={<RightOutline />}>意见反馈</List.Item>
          <List.Item prefix={<SetOutline />} arrow={<RightOutline />}>账号设置</List.Item>
        </List>
      </section>
      <Button className="logout-button" block onClick={handleLogout}>退出登录</Button>
      <p className="app-version">IMAI Play 学员端 · v0.1.0</p>
    </div>
  )
}
