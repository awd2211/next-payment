import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'masked-icon.svg'],
      manifest: {
        name: '支付平台商户中心',
        short_name: '商户中心',
        description: '全球支付平台商户管理系统',
        theme_color: '#1890ff',
        background_color: '#ffffff',
        display: 'standalone',
        scope: '/',
        start_url: '/',
        icons: [
          {
            src: 'pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/fonts\.googleapis\.com\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'google-fonts-cache',
              expiration: {
                maxEntries: 10,
                maxAgeSeconds: 60 * 60 * 24 * 365, // 365 days
              },
              cacheableResponse: {
                statuses: [0, 200],
              },
            },
          },
          {
            urlPattern: /^https:\/\/fonts\.gstatic\.com\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'gstatic-fonts-cache',
              expiration: {
                maxEntries: 10,
                maxAgeSeconds: 60 * 60 * 24 * 365, // 365 days
              },
              cacheableResponse: {
                statuses: [0, 200],
              },
            },
          },
          {
            urlPattern: /^https:\/\/.*\.cdninstagram\.com\/.*/i,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'cdn-cache',
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 60 * 60 * 24 * 7, // 7 days
              },
            },
          },
          {
            urlPattern: /\/api\/v1\/.*/i,
            handler: 'NetworkFirst',
            method: 'GET',
            options: {
              cacheName: 'api-cache',
              expiration: {
                maxEntries: 100,
                maxAgeSeconds: 60 * 5, // 5 minutes
              },
              networkTimeoutSeconds: 10,
            },
          },
        ],
      },
      devOptions: {
        enabled: true,
      },
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'antd-vendor': ['antd', '@ant-design/icons'],
          'chart-vendor': ['@ant-design/charts'],
          'utils': ['axios', 'dayjs', 'zustand'],
        },
      },
    },
    chunkSizeWarningLimit: 1000,
  },
  server: {
    port: 5174, // 使用标准端口
    proxy: {
      // Merchant Service (商户相关)
      '/api/v1/merchants': {
        target: 'http://localhost:40002',
        changeOrigin: true,
      },
      '/api/v1/api-keys': {
        target: 'http://localhost:40002',
        changeOrigin: true,
      },
      '/api/v1/webhooks': {
        target: 'http://localhost:40002',
        changeOrigin: true,
      },
      '/api/v1/channels': {
        target: 'http://localhost:40002',
        changeOrigin: true,
      },

      // Payment Gateway (支付)
      '/api/v1/payments': {
        target: 'http://localhost:40003',
        changeOrigin: true,
      },

      // Order Service (订单)
      '/api/v1/orders': {
        target: 'http://localhost:40004',
        changeOrigin: true,
      },

      // Accounting Service (账务)
      '/api/v1/accounts': {
        target: 'http://localhost:40007',
        changeOrigin: true,
      },
      '/api/v1/transactions': {
        target: 'http://localhost:40007',
        changeOrigin: true,
      },
      '/api/v1/settlements': {
        target: 'http://localhost:40007',
        changeOrigin: true,
      },

      // Analytics Service (数据分析)
      '/api/v1/analytics': {
        target: 'http://localhost:40009',
        changeOrigin: true,
      },
      '/api/v1/metrics': {
        target: 'http://localhost:40009',
        changeOrigin: true,
      },
    },
  },
})
