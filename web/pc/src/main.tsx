import React from 'react';
import ReactDOM from 'react-dom/client';
import { Suspense } from 'react';
import { Spin } from 'antd';
import { App as AntdApp } from 'antd';
import { RouterProvider } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { router } from './router';
import './styles.css';
import { initLearnerAnimations } from './animations';

window.setTimeout(() => initLearnerAnimations(), 0);

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AntdApp>
      <AuthProvider><Suspense fallback={<Spin fullscreen />}><RouterProvider router={router} /></Suspense></AuthProvider>
    </AntdApp>
  </React.StrictMode>,
);
