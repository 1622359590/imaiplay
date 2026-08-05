import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './styles.css'
import { TenantThemeProvider } from './context/TenantThemeContext'
import { initLearnerAnimations } from './animations'

window.setTimeout(() => initLearnerAnimations(), 0)

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter basename="/h5"><TenantThemeProvider><App /></TenantThemeProvider></BrowserRouter>
  </React.StrictMode>,
)
