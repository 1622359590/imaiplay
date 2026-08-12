import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { manualChunks } from '../build/vendorChunks'

function adminManualChunks(id: string) {
  const chunk = manualChunks(id)

  if (
    chunk === 'antd-framework' ||
    chunk === 'antd-primitives' ||
    chunk === 'antd-icons'
  ) {
    return undefined
  }

  return chunk
}

export default defineConfig({
  plugins: [react()],
  base: '/admin/',
  build: {
    chunkSizeWarningLimit: 500,
    rollupOptions: {
      output: {
        manualChunks: adminManualChunks,
      },
    },
  },
  server: {
    port: 5173,
  },
})
