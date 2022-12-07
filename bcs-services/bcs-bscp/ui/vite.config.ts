import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api/c/compapi/v2/cc/': {
        target: '/',
        changeOrigin: true,
      },
      '/dev/api/v1/': {
        target: '/',
        changeOrigin: true,
      }
    }
  }
})
