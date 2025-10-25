import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/authStore'
import Layout from './components/Layout'
import WebSocketProvider from './components/WebSocketProvider'
import PWAUpdatePrompt from './components/PWAUpdatePrompt'
import PageLoading from './components/PageLoading'

// 使用 React.lazy 实现代码分割
// Login 页面不分割,因为是首次加载必需的
import Login from './pages/Login'

// 其他页面按需加载
const Dashboard = lazy(() => import('./pages/Dashboard'))
const SystemConfigs = lazy(() => import('./pages/SystemConfigs'))
const Admins = lazy(() => import('./pages/Admins'))
const Roles = lazy(() => import('./pages/Roles'))
const AuditLogs = lazy(() => import('./pages/AuditLogs'))
const Merchants = lazy(() => import('./pages/Merchants'))
const Payments = lazy(() => import('./pages/Payments'))
const Orders = lazy(() => import('./pages/Orders'))
const RiskManagement = lazy(() => import('./pages/RiskManagement'))
const Settlements = lazy(() => import('./pages/Settlements'))
const CashierManagement = lazy(() => import('./pages/CashierManagement'))
const KYC = lazy(() => import('./pages/KYC'))
const Withdrawals = lazy(() => import('./pages/Withdrawals'))
const Channels = lazy(() => import('./pages/Channels'))
const Accounting = lazy(() => import('./pages/Accounting'))
const Analytics = lazy(() => import('./pages/Analytics'))
const Notifications = lazy(() => import('./pages/Notifications'))
const Disputes = lazy(() => import('./pages/Disputes'))
const Reconciliation = lazy(() => import('./pages/Reconciliation'))
const Webhooks = lazy(() => import('./pages/Webhooks'))
const MerchantLimits = lazy(() => import('./pages/MerchantLimits'))

function App() {
  return (
    <>
      <PWAUpdatePrompt />
      <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route
            path="dashboard"
            element={
              <Suspense fallback={<PageLoading />}>
                <Dashboard />
              </Suspense>
            }
          />
          <Route
            path="system-configs"
            element={
              <Suspense fallback={<PageLoading />}>
                <SystemConfigs />
              </Suspense>
            }
          />
          <Route
            path="admins"
            element={
              <Suspense fallback={<PageLoading />}>
                <Admins />
              </Suspense>
            }
          />
          <Route
            path="roles"
            element={
              <Suspense fallback={<PageLoading />}>
                <Roles />
              </Suspense>
            }
          />
          <Route
            path="audit-logs"
            element={
              <Suspense fallback={<PageLoading />}>
                <AuditLogs />
              </Suspense>
            }
          />
          <Route
            path="merchants"
            element={
              <Suspense fallback={<PageLoading />}>
                <Merchants />
              </Suspense>
            }
          />
          <Route
            path="payments"
            element={
              <Suspense fallback={<PageLoading />}>
                <Payments />
              </Suspense>
            }
          />
          <Route
            path="orders"
            element={
              <Suspense fallback={<PageLoading />}>
                <Orders />
              </Suspense>
            }
          />
          <Route
            path="risk"
            element={
              <Suspense fallback={<PageLoading />}>
                <RiskManagement />
              </Suspense>
            }
          />
          <Route
            path="settlements"
            element={
              <Suspense fallback={<PageLoading />}>
                <Settlements />
              </Suspense>
            }
          />
          <Route
            path="cashier"
            element={
              <Suspense fallback={<PageLoading />}>
                <CashierManagement />
              </Suspense>
            }
          />
          <Route
            path="kyc"
            element={
              <Suspense fallback={<PageLoading />}>
                <KYC />
              </Suspense>
            }
          />
          <Route
            path="withdrawals"
            element={
              <Suspense fallback={<PageLoading />}>
                <Withdrawals />
              </Suspense>
            }
          />
          {/* 暂时禁用 - 后端 API 未实现 */}
          {/* <Route
            path="channels"
            element={
              <Suspense fallback={<PageLoading />}>
                <Channels />
              </Suspense>
            }
          /> */}
          <Route
            path="accounting"
            element={
              <Suspense fallback={<PageLoading />}>
                <Accounting />
              </Suspense>
            }
          />
          <Route
            path="analytics"
            element={
              <Suspense fallback={<PageLoading />}>
                <Analytics />
              </Suspense>
            }
          />
          <Route
            path="notifications"
            element={
              <Suspense fallback={<PageLoading />}>
                <Notifications />
              </Suspense>
            }
          />
          <Route
            path="disputes"
            element={
              <Suspense fallback={<PageLoading />}>
                <Disputes />
              </Suspense>
            }
          />
          <Route
            path="reconciliation"
            element={
              <Suspense fallback={<PageLoading />}>
                <Reconciliation />
              </Suspense>
            }
          />
          <Route
            path="webhooks"
            element={
              <Suspense fallback={<PageLoading />}>
                <Webhooks />
              </Suspense>
            }
          />
          <Route
            path="merchant-limits"
            element={
              <Suspense fallback={<PageLoading />}>
                <MerchantLimits />
              </Suspense>
            }
          />
        </Route>
      </Routes>
      </BrowserRouter>
    </>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token } = useAuthStore()

  if (!token) {
    return <Navigate to="/login" replace />
  }

  return <WebSocketProvider>{children}</WebSocketProvider>
}

export default App
