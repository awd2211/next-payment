import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/authStore'
import Layout from './components/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import SystemConfigs from './pages/SystemConfigs'
import Admins from './pages/Admins'
import Roles from './pages/Roles'
import AuditLogs from './pages/AuditLogs'
import Merchants from './pages/Merchants'

function App() {
  return (
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
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token } = useAuthStore()

  if (!token) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

export default App
