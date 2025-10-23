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
    port: 40200,
    proxy: {
      // Merchant Service (商户相关)
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

      // Accounting Service (账务)
      '/api/v1/accounts': {
        target: 'http://localhost:8007',
        changeOrigin: true,
      },
      '/api/v1/transactions': {
        target: 'http://localhost:8007',
        changeOrigin: true,
      },
      '/api/v1/settlements': {
        target: 'http://localhost:8007',
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
    },
  },
})
