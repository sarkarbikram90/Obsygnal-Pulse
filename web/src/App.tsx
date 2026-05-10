import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from './components/Layout'
import Dashboard from './features/cluster/Dashboard'
import Recommendations from './features/recommendations/Recommendations'
import Inefficiencies from './features/cluster/Inefficiencies'
import Alerts from './features/alerts/Alerts'

import Nodes from './features/cluster/Nodes'
import Jobs from './features/jobs/Jobs'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      staleTime: 1000 * 60, // 1 minute
    },
  },
})

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="recommendations" element={<Recommendations />} />
            <Route path="inefficiencies" element={<Inefficiencies />} />
            <Route path="alerts" element={<Alerts />} />
            <Route path="nodes" element={<Nodes />} />
            <Route path="jobs" element={<Jobs />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App
