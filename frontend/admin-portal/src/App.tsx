import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/authStore'
import Layout from './components/Layout'
import WebSocketProvider from './components/WebSocketProvider'
import PWAUpdatePrompt from './components/PWAUpdatePrompt'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import SystemConfigs from './pages/SystemConfigs'
import Admins from './pages/Admins'
import Roles from './pages/Roles'
import AuditLogs from './pages/AuditLogs'
import Merchants from './pages/Merchants'
import Payments from './pages/Payments'

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
          <Route path="system-configs" element={<SystemConfigs />} />
          <Route path="admins" element={<Admins />} />
          <Route path="roles" element={<Roles />} />
          <Route path="audit-logs" element={<AuditLogs />} />
          <Route path="merchants" element={<Merchants />} />
          <Route path="payments" element={<Payments />} />
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
