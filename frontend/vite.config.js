import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // Element Plus自动导入
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
  ],
  resolve: {
    // 路径别名
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    // 跨域代理（解决后端接口跨域）
    proxy: {
      '/api': {
        target: 'http://localhost:8080', // 后端接口地址
        changeOrigin: true,
        //rewrite: (path) => path.replace(/^\/api/, '/api'),
      },
      // 静态资源（头像等）转发到后端
      '/static': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
    port: 3000, // 前端运行端口
    open: true, // 启动后自动打开浏览器
  },
})