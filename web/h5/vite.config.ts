import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { manualChunks } from '../build/vendorChunks'

export default defineConfig({
  base: '/h5/',
  plugins: [react()],
  build: {
    chunkSizeWarningLimit: 500,
    rollupOptions: {
      output: {
        manualChunks,
      },
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5175,
  },
})
