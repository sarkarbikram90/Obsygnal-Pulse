import type { ClusterOverview, Inefficiency, Recommendation, GPUMetric } from '../types'

const API_BASE = '/api/v1'

export const api = {
  // Dashboard & Overview
  getClusterOverview: async (clusterId: string): Promise<ClusterOverview> => {
    const res = await fetch(`${API_BASE}/clusters/${clusterId}/overview`)
    if (!res.ok) throw new Error('Failed to fetch cluster overview')
    return res.json()
  },

  // Inefficiencies
  getInefficiencies: async (clusterId: string): Promise<Inefficiency[]> => {
    const res = await fetch(`${API_BASE}/clusters/${clusterId}/inefficiencies`)
    if (!res.ok) throw new Error('Failed to fetch inefficiencies')
    return res.json()
  },

  // Recommendations
  getRecommendations: async (clusterId: string): Promise<Recommendation[]> => {
    const res = await fetch(`${API_BASE}/clusters/${clusterId}/recommendations`)
    if (!res.ok) throw new Error('Failed to fetch recommendations')
    return res.json()
  },

  // Metrics (Mocked for now since specific route might vary)
  getGPUMetrics: async (clusterId: string): Promise<GPUMetric[]> => {
    // This is mocked as I need to check the actual route.
    const res = await fetch(`${API_BASE}/clusters/${clusterId}/metrics/gpu`)
    if (!res.ok) throw new Error('Failed to fetch GPU metrics')
    return res.json()
  }
}
