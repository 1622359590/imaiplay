import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { manualChunks as sharedManualChunks } from '../build/vendorChunks';

function manualChunks(id: string) {
  const chunk = sharedManualChunks(id);

  if (
    chunk === 'antd-framework' ||
    chunk === 'antd-primitives' ||
    chunk === 'antd-icons' ||
    chunk === 'antd-styles'
  ) {
    return undefined;
  }

  return chunk;
}

export default defineConfig({
  base: '/pc/',
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
    port: 5174,
  },
});
