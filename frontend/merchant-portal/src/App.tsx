import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/authStore'
import Layout from './components/Layout'
import WebSocketProvider from './components/WebSocketProvider'
import PWAUpdatePrompt from './components/PWAUpdatePrompt'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Transactions from './pages/Transactions'
import Orders from './pages/Orders'
import Account from './pages/Account'
import CreatePayment from './pages/CreatePayment'
import Refunds from './pages/Refunds'
import Settlements from './pages/Settlements'

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
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="create-payment" element={<CreatePayment />} />
          <Route path="transactions" element={<Transactions />} />
          <Route path="orders" element={<Orders />} />
          <Route path="refunds" element={<Refunds />} />
          <Route path="settlements" element={<Settlements />} />
          <Route path="account" element={<Account />} />
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
