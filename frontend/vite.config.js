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
        changeOrigin: true
      },
      '/api/v1/product': {
        target: 'http://localhost:8082',
        changeOrigin: true
      }
    }
  }
})

