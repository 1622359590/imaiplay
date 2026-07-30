import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './styles.css'
import { TenantThemeProvider } from './context/TenantThemeContext'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <TenantThemeProvider><BrowserRouter basename="/h5"><App /></BrowserRouter></TenantThemeProvider>
  </React.StrictMode>,
)
