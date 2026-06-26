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
    allowedHosts: [ "knowsource.cn"],
    port: 3003,
    proxy: {
      '/api': {
        target: 'http://localhost:8070',
        changeOrigin: true
      }
    }
  }
})

