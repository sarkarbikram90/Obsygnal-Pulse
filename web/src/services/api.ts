// API client for Obsygnal-Pulse backend
import type { ClusterOverview, GPUMetric, HealthReport, IngestionStats } from '../types/api';

const BASE_URL = '/api/v1';

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`);
  if (!res.ok) throw new Error(`API error: ${res.status} ${res.statusText}`);
  return res.json();
}

export const api = {
  health: () => fetchJSON<HealthReport>('/health'),

  clusterOverview: (clusterId = 'sim-cluster-01') =>
    fetchJSON<ClusterOverview>(`/cluster/overview?cluster_id=${clusterId}`),

  gpuMetrics: (params?: { cluster_id?: string; node_id?: string; gpu_id?: string; from?: string; to?: string }) => {
    const qs = new URLSearchParams(params as Record<string, string>).toString();
    return fetchJSON<{ data: GPUMetric[]; count: number }>(`/gpus/metrics?${qs}`);
  },

  ingestionStats: () => fetchJSON<IngestionStats>('/ingestion/stats'),
};
