import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { Provider } from 'react-redux'
import App from './App'
import { store } from './store'
import './styles.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <ConfigProvider
        locale={zhCN}
        theme={{
          token: {
            colorPrimary: '#1769e0',
            colorInfo: '#1769e0',
            borderRadius: 8,
            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif',
          },
          components: {
            Layout: { headerBg: '#fff', siderBg: '#0b3c8c' },
            Menu: { darkItemBg: '#0b3c8c', darkItemSelectedBg: '#1769e0', darkItemHoverBg: '#124fa6' },
          },
        }}
      >
        <App />
      </ConfigProvider>
    </Provider>
  </React.StrictMode>,
)
