import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api/v1/users': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        rewrite: (path) => path // 保持路径不变
      },
      '/api/v1/product': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        rewrite: (path) => path
      },
      '/api/v1/cart': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        rewrite: (path) => path
      }
    }
  }
})

