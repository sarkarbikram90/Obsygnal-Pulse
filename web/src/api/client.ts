import type { ClusterOverview, Inefficiency, Recommendation, UsageSummary, APIKey } from '../types'

const API_BASE = '/api/v1'

const getHeaders = () => {
  const token = localStorage.getItem('obsygnal_token')
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export const api = {
  // Dashboard & Overview
  getClusterOverview: async (clusterId: string): Promise<ClusterOverview> => {
    try {
      const res = await fetch(`${API_BASE}/cluster/overview?cluster_id=${clusterId}`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to fetch cluster overview')
      return await res.json()
    } catch {
      return {
        cluster_id: clusterId,
        total_nodes: 4,
        healthy_nodes: 4,
        total_gpus: 32,
        active_gpus: 24,
        idle_gpus: 8,
        avg_utilization: 44.8,
        avg_memory_usage: 58.2,
        active_jobs: 6,
        failed_jobs_24h: 0,
        inefficiency_score: 55.2,
        estimated_waste_usd: 6820.00,
        timestamp: new Date().toISOString(),
      }
    }
  },

  // Inefficiencies
  getInefficiencies: async (clusterId: string): Promise<Inefficiency[]> => {
    try {
      const res = await fetch(`${API_BASE}/inefficiencies?cluster_id=${clusterId}`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to fetch inefficiencies')
      const body = await res.json()
      return body.data || []
    } catch {
      return [
        {
          id: 'ineff-001',
          type: 'idle_gpu',
          severity: 'high',
          cluster_id: clusterId,
          node_id: 'gpu-node-01',
          gpu_id: 'gpu-node-01-gpu-3',
          description: 'GPU gpu-node-01-gpu-3 is completely idle (0% compute utilization, 0 active processes)',
          affected_metrics: ['gpu_utilization', 'running_processes'],
          current_value: 0.0,
          threshold_value: 2.0,
          detected_at: new Date(Date.now() - 25 * 60000).toISOString(),
          is_active: true,
        },
        {
          id: 'ineff-003',
          type: 'xid_hardware_error',
          severity: 'critical',
          cluster_id: clusterId,
          node_id: 'gpu-node-04',
          gpu_id: 'gpu-node-04-gpu-1',
          description: 'NVIDIA Critical Hardware Fault XID 31 (GPU Memory Page Fault) detected on gpu-node-04-gpu-1',
          affected_metrics: ['xid_error_code', 'gpu_status'],
          current_value: 31.0,
          threshold_value: 1.0,
          detected_at: new Date(Date.now() - 5 * 60000).toISOString(),
          is_active: true,
        },
        {
          id: 'ineff-004',
          type: 'cuda_memory_fragmentation',
          severity: 'high',
          cluster_id: clusterId,
          node_id: 'gpu-node-02',
          gpu_id: 'gpu-node-02-gpu-1',
          job_id: 'job-llama3-finetune-891',
          description: 'PyTorch CUDA memory fragmentation on gpu-node-02-gpu-1: reserved 68 GB vs used 24 GB (ratio 2.8x)',
          affected_metrics: ['cuda_memory_reserved', 'memory_used'],
          current_value: 2.8,
          threshold_value: 2.0,
          detected_at: new Date(Date.now() - 8 * 60000).toISOString(),
          is_active: true,
        },
      ]
    }
  },

  // Recommendations
  getRecommendations: async (clusterId: string): Promise<Recommendation[]> => {
    try {
      const res = await fetch(`${API_BASE}/recommendations?cluster_id=${clusterId}`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to fetch recommendations')
      const body = await res.json()
      return body.data || []
    } catch {
      return [
        {
          id: 'rec-001',
          category: 'reduce_idle_time',
          status: 'active',
          title: 'Isolate Hardware Fault (XID 31)',
          description: 'Cordon & drain node gpu-node-04 immediately to isolate GPU Memory Page Fault XID 31.',
          explanation: 'Hardware memory error detected on GPU 1. Immediate node isolation prevents recurring training crashes across distributed jobs.',
          confidence_score: 0.99,
          estimated_savings: {
            cost_per_hour_usd: 5.12,
            cost_per_month_usd: 3737.60,
            gpu_hours_per_month: 730,
          },
          affected_metrics: ['xid_error_code', 'gpu_status'],
          cluster_id: clusterId,
          node_id: 'gpu-node-04',
          gpu_id: 'gpu-node-04-gpu-1',
          inefficiency_id: 'ineff-003',
          created_at: new Date(Date.now() - 4 * 60000).toISOString(),
          expires_at: new Date(Date.now() + 24 * 3600000).toISOString(),
        },
        {
          id: 'rec-002',
          category: 'right_size_memory',
          status: 'active',
          title: 'Defragment PyTorch CUDA VRAM Allocator',
          description: 'Set PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True on job-llama3-finetune-891.',
          explanation: 'PyTorch has reserved 68GB out of 80GB VRAM but only 24GB is actively used. Enabling expandable segments eliminates CUDA allocator fragmentation OOM risk.',
          confidence_score: 0.92,
          estimated_savings: {
            cost_per_hour_usd: 2.40,
            cost_per_month_usd: 1752.00,
            gpu_hours_per_month: 210,
          },
          affected_metrics: ['cuda_memory_reserved', 'memory_used'],
          cluster_id: clusterId,
          node_id: 'gpu-node-02',
          gpu_id: 'gpu-node-02-gpu-1',
          job_id: 'job-llama3-finetune-891',
          inefficiency_id: 'ineff-004',
          created_at: new Date(Date.now() - 7 * 60000).toISOString(),
          expires_at: new Date(Date.now() + 24 * 3600000).toISOString(),
        },
          affected_metrics: ['gpu_utilization', 'gpu_memory_utilization'],
          cluster_id: clusterId,
          node_id: 'gpu-node-02',
          gpu_id: 'gpu-node-02-gpu-1',
          job_id: 'job-llama3-finetune-891',
          inefficiency_id: 'ineff-002',
          created_at: new Date(Date.now() - 10 * 60000).toISOString(),
          expires_at: new Date(Date.now() + 24 * 3600000).toISOString(),
        },
      ]
    }
  },

  applyRecommendation: async (recId: string): Promise<void> => {
    const res = await fetch(`${API_BASE}/recommendations/${recId}/apply`, {
      method: 'POST',
      headers: getHeaders(),
    })
    if (!res.ok) throw new Error('Failed to apply recommendation')
  },

  dismissRecommendation: async (recId: string, reason?: string): Promise<void> => {
    const res = await fetch(`${API_BASE}/recommendations/${recId}/dismiss?reason=${encodeURIComponent(reason || '')}`, {
      method: 'POST',
      headers: getHeaders(),
    })
    if (!res.ok) throw new Error('Failed to dismiss recommendation')
  },

  // Usage & Metering
  getUsageSummary: async (): Promise<UsageSummary> => {
    try {
      const res = await fetch(`${API_BASE}/billing/usage`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to fetch usage summary')
      return await res.json()
    } catch {
      return {
        organization_id: 'org_demo_101',
        tier: 'pro',
        billing_cycle: 'August 2026',
        monitored_gpus: 32,
        gpu_hours_used: 3420.5,
        gpu_hours_limit: 5000.0,
        estimated_cost_usd: 12450.00,
        estimated_savings_usd: 5489.60,
        active_clusters: 2,
      }
    }
  },

  // API Keys
  getAPIKeys: async (): Promise<APIKey[]> => {
    try {
      const res = await fetch(`${API_BASE}/api-keys`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to fetch API keys')
      const body = await res.json()
      return body.api_keys || []
    } catch {
      return [
        {
          id: 'key-01',
          name: 'Production-K8s-Collector',
          prefix: 'obsygnal_sec_a8b9...',
          created_at: new Date(Date.now() - 72 * 3600000).toISOString(),
          last_used: '2 minutes ago',
          status: 'active',
        },
      ]
    }
  },

  createAPIKey: async (name: string, description: string): Promise<APIKey> => {
    const res = await fetch(`${API_BASE}/api-keys`, {
      method: 'POST',
      headers: getHeaders(),
      body: JSON.stringify({ name, description }),
    })
    if (!res.ok) throw new Error('Failed to create API key')
    return await res.json()
  },
}
