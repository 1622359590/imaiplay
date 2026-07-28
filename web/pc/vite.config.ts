import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          const antdMatch = id.match(/node_modules\/antd\/es\/([^/]+)/);
          if (antdMatch) return `antd-${antdMatch[1]}`;
          const iconsMatch = id.match(/node_modules\/@ant-design\/icons\/es\/([^/]+)/);
          if (iconsMatch) return `antd-icons-${iconsMatch[1]}`;
        },
      },
    },
  },
  server: {
    port: 5174,
  },
});
