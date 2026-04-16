import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@wailsjs': resolve(__dirname, 'wailsjs')
    }
  },
  server: {
    host: '127.0.0.1',
    port: 33001,
    strictPort: true
  },
  build: {
    outDir: 'dist',
    assetsDir: 'static'
  }
})
