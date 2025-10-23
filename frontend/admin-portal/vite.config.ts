import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 40101,
    proxy: {
      // Admin Service (管理员、角色、权限、审计、系统配置)
      '/api/v1/admins': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },
      '/api/v1/roles': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },
      '/api/v1/permissions': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },
      '/api/v1/audit-logs': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },
      '/api/v1/system-configs': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },
      '/api/v1/email-templates': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },
      '/api/v1/preferences': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },
      '/api/v1/security': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },

      // Merchant Service (商户管理)
      '/api/v1/merchants': {
        target: 'http://localhost:8002',
        changeOrigin: true,
      },
      '/api/v1/api-keys': {
        target: 'http://localhost:8002',
        changeOrigin: true,
      },
      '/api/v1/webhooks': {
        target: 'http://localhost:8002',
        changeOrigin: true,
      },
      '/api/v1/channels': {
        target: 'http://localhost:8002',
        changeOrigin: true,
      },

      // Payment Gateway (支付)
      '/api/v1/payments': {
        target: 'http://localhost:8003',
        changeOrigin: true,
      },

      // Order Service (订单)
      '/api/v1/orders': {
        target: 'http://localhost:8004',
        changeOrigin: true,
      },

      // Analytics Service (数据分析)
      '/api/v1/analytics': {
        target: 'http://localhost:8009',
        changeOrigin: true,
      },
      '/api/v1/metrics': {
        target: 'http://localhost:8009',
        changeOrigin: true,
      },

      // Config Service (配置中心)
      '/api/v1/configs': {
        target: 'http://localhost:8010',
        changeOrigin: true,
      },
      '/api/v1/feature-flags': {
        target: 'http://localhost:8010',
        changeOrigin: true,
      },
    },
  },
})
