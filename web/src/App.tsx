import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { ToastProvider } from './context/ToastContext'
import { Layout } from './components/Layout'
import Dashboard from './features/cluster/Dashboard'
import Recommendations from './features/recommendations/Recommendations'
import Inefficiencies from './features/cluster/Inefficiencies'
import Alerts from './features/alerts/Alerts'
import Nodes from './features/cluster/Nodes'
import Jobs from './features/jobs/Jobs'
import { BillingView } from './features/billing/BillingView'
import { FleetSimulator } from './features/executive/FleetSimulator'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      staleTime: 1000 * 30, // 30 seconds
    },
  },
})

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <ToastProvider>
          <BrowserRouter>
            <Routes>
              <Route path="/" element={<Layout />}>
                <Route index element={<Navigate to="/dashboard" replace />} />
                <Route path="dashboard" element={<Dashboard />} />
                <Route path="recommendations" element={<Recommendations />} />
                <Route path="inefficiencies" element={<Inefficiencies />} />
                <Route path="simulator" element={<FleetSimulator />} />
                <Route path="nodes" element={<Nodes />} />
                <Route path="jobs" element={<Jobs />} />
                <Route path="alerts" element={<Alerts />} />
                <Route path="billing" element={<BillingView />} />
              </Route>
            </Routes>
          </BrowserRouter>
        </ToastProvider>
      </AuthProvider>
    </QueryClientProvider>
  )
}

export default App
