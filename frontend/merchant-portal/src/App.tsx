import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { App as AntdApp } from 'antd'
import { useAuthStore } from './stores/authStore'
import Layout from './components/Layout'
import WebSocketProvider from './components/WebSocketProvider'
import PWAUpdatePrompt from './components/PWAUpdatePrompt'
import ErrorBoundary from './components/ErrorBoundary'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Transactions from './pages/Transactions'
import Orders from './pages/Orders'
import Account from './pages/Account'
import CreatePayment from './pages/CreatePayment'
import Refunds from './pages/Refunds'
import Settlements from './pages/Settlements'
import ApiKeys from './pages/ApiKeys'
import CashierConfig from './pages/CashierConfig'
import CashierCheckout from './pages/CashierCheckout'
import Notifications from './pages/Notifications'

function App() {
  return (
    <ErrorBoundary>
      <AntdApp>
        <PWAUpdatePrompt />
        <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          {/* Public cashier checkout page - no auth required, uses ?token= query param */}
          <Route path="/cashier/checkout" element={<CashierCheckout />} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="create-payment" element={<CreatePayment />} />
            <Route path="transactions" element={<Transactions />} />
            <Route path="orders" element={<Orders />} />
            <Route path="refunds" element={<Refunds />} />
            <Route path="settlements" element={<Settlements />} />
            <Route path="api-keys" element={<ApiKeys />} />
            <Route path="cashier-config" element={<CashierConfig />} />
            <Route path="notifications" element={<Notifications />} />
            <Route path="account" element={<Account />} />
          </Route>
        </Routes>
        </BrowserRouter>
      </AntdApp>
    </ErrorBoundary>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token, merchant } = useAuthStore()

  console.log('ProtectedRoute check:', { token: token ? 'exists' : 'null', merchant })

  if (!token) {
    console.log('No token found, redirecting to login')
    return <Navigate to="/login" replace />
  }

  return <WebSocketProvider>{children}</WebSocketProvider>
}

export default App
