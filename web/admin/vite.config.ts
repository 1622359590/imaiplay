import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/admin/',
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (
            id.indexOf('/node_modules/react/') !== -1 ||
            id.indexOf('/node_modules/react-dom/') !== -1 ||
            id.indexOf('/node_modules/react-router/') !== -1 ||
            id.indexOf('/node_modules/react-router-dom/') !== -1
          ) {
            return 'react-vendor'
          }
          if (
            id.indexOf('/node_modules/antd/') !== -1 ||
            id.indexOf('/node_modules/@ant-design/') !== -1
          ) {
            return 'antd-vendor'
          }
          if (
            id.indexOf('/node_modules/@reduxjs/') !== -1 ||
            id.indexOf('/node_modules/react-redux/') !== -1
          ) {
            return 'state-vendor'
          }
        },
      },
    },
  },
  server: {
    port: 5173,
  },
})
