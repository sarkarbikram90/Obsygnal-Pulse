import type {
  ClusterOverview,
  Inefficiency,
  Recommendation,
  UsageSummary,
  APIKey,
  NodeDetail,
  JobDetail,
  AlertDetail,
} from '../types'

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
        healthy_nodes: 3,
        total_gpus: 32,
        active_gpus: 24,
        idle_gpus: 8,
        avg_utilization: 44.8,
        avg_memory_usage: 58.2,
        active_jobs: 6,
        failed_jobs_24h: 0,
        inefficiency_score: 55.2,
        estimated_waste_usd: 6820.0,
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
        {
          id: 'ineff-005',
          type: 'cpu_starvation',
          severity: 'high',
          cluster_id: clusterId,
          node_id: 'gpu-node-03',
          gpu_id: 'gpu-node-03-gpu-0',
          job_id: 'job-deepspeed-pretrain-102',
          description: 'Host CPU is pegged at 96% while GPU compute stalls at 11% due to DataLoader thread bottleneck',
          affected_metrics: ['cpu_utilization', 'gpu_utilization'],
          current_value: 96.2,
          threshold_value: 90.0,
          detected_at: new Date(Date.now() - 14 * 60000).toISOString(),
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
            cost_per_month_usd: 3737.6,
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
            cost_per_hour_usd: 2.4,
            cost_per_month_usd: 1752.0,
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
        {
          id: 'rec-003',
          category: 'increase_batch_size',
          status: 'active',
          title: 'Scale PyTorch DataLoader Workers',
          description: 'Increase DataLoader num_workers from 2 to 8 and enable pin_memory=True on job-deepspeed-pretrain-102.',
          explanation: 'GPUs are stalling at 11% compute due to CPU data bottleneck. Scaling workers increases GPU compute saturation to >85%.',
          confidence_score: 0.91,
          estimated_savings: {
            cost_per_hour_usd: 1.82,
            cost_per_month_usd: 1330.4,
            gpu_hours_per_month: 300,
          },
          affected_metrics: ['cpu_utilization', 'gpu_utilization'],
          cluster_id: clusterId,
          node_id: 'gpu-node-03',
          gpu_id: 'gpu-node-03-gpu-0',
          job_id: 'job-deepspeed-pretrain-102',
          inefficiency_id: 'ineff-005',
          created_at: new Date(Date.now() - 12 * 60000).toISOString(),
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

  // Nodes Explorer
  getNodes: async (clusterId = 'sim-cluster-01'): Promise<NodeDetail[]> => {
    try {
      const res = await fetch(`${API_BASE}/nodes?cluster_id=${clusterId}`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to fetch nodes')
      const body = await res.json()
      if (body.data && Array.isArray(body.data)) {
        return body.data.map((n: any) => ({
          node_id: n.node_id,
          status: n.status || 'ready',
          gpu_model: n.gpu_model || 'NVIDIA H100 SXM5 80GB',
          gpu_count: n.gpu_count || 8,
          cpu_usage: n.cpu_usage || 45.0,
          memory_gb: n.memory_gb || 512,
          driver_version: n.driver_version || '535.129.03',
          cuda_version: n.cuda_version || '12.4',
          ip_address: n.ip_address || '10.244.1.45',
          cluster_id: clusterId,
          last_heartbeat: 'Just now',
          gpus: Array.from({ length: n.gpu_count || 8 }).map((_, i) => ({
            gpu_id: `${n.node_id}-gpu-${i}`,
            index: i,
            model: n.gpu_model || 'NVIDIA H100 SXM5',
            utilization: n.node_id === 'gpu-node-04' && i === 1 ? 0 : Math.floor(Math.random() * 50) + 40,
            memory_used_gb: Math.floor(Math.random() * 40) + 20,
            memory_total_gb: 80,
            temperature: 58 + i * 2,
            power_watts: 420 + i * 15,
            status: n.node_id === 'gpu-node-04' && i === 1 ? 'fault' : 'active',
            xid_error: n.node_id === 'gpu-node-04' && i === 1 ? 31 : undefined,
          })),
        }))
      }
      throw new Error('Fallback')
    } catch {
      return [
        {
          node_id: 'gpu-node-01',
          status: 'ready',
          gpu_model: 'NVIDIA H100 SXM5 80GB',
          gpu_count: 8,
          cpu_usage: 42.5,
          memory_gb: 512,
          driver_version: '535.129.03',
          cuda_version: '12.4',
          ip_address: '10.244.0.12',
          cluster_id: clusterId,
          last_heartbeat: '3s ago',
          gpus: Array.from({ length: 8 }).map((_, i) => ({
            gpu_id: `gpu-node-01-gpu-${i}`,
            index: i,
            model: 'NVIDIA H100 SXM5',
            utilization: i === 3 ? 0.0 : 75.4 - i * 3,
            memory_used_gb: i === 3 ? 0.4 : 58.2,
            memory_total_gb: 80,
            temperature: 62 + i,
            power_watts: 480,
            status: i === 3 ? 'idle' : 'active',
          })),
        },
        {
          node_id: 'gpu-node-02',
          status: 'ready',
          gpu_model: 'NVIDIA H100 SXM5 80GB',
          gpu_count: 8,
          cpu_usage: 55.8,
          memory_gb: 512,
          driver_version: '535.129.03',
          cuda_version: '12.4',
          ip_address: '10.244.0.13',
          cluster_id: clusterId,
          last_heartbeat: '1s ago',
          gpus: Array.from({ length: 8 }).map((_, i) => ({
            gpu_id: `gpu-node-02-gpu-${i}`,
            index: i,
            model: 'NVIDIA H100 SXM5',
            utilization: 82.0,
            memory_used_gb: 68.0,
            memory_total_gb: 80,
            temperature: 67 + i,
            power_watts: 530,
            status: 'active',
          })),
        },
        {
          node_id: 'gpu-node-03',
          status: 'ready',
          gpu_model: 'NVIDIA A100 SXM4 80GB',
          gpu_count: 8,
          cpu_usage: 96.2,
          memory_gb: 512,
          driver_version: '535.104.05',
          cuda_version: '12.2',
          ip_address: '10.244.0.14',
          cluster_id: clusterId,
          last_heartbeat: '5s ago',
          gpus: Array.from({ length: 8 }).map((_, i) => ({
            gpu_id: `gpu-node-03-gpu-${i}`,
            index: i,
            model: 'NVIDIA A100 SXM4',
            utilization: 11.4,
            memory_used_gb: 42.0,
            memory_total_gb: 80,
            temperature: 54 + i,
            power_watts: 280,
            status: 'active',
          })),
        },
        {
          node_id: 'gpu-node-04',
          status: 'degraded',
          gpu_model: 'NVIDIA H100 SXM5 80GB',
          gpu_count: 8,
          cpu_usage: 31.0,
          memory_gb: 512,
          driver_version: '535.129.03',
          cuda_version: '12.4',
          ip_address: '10.244.0.15',
          cluster_id: clusterId,
          last_heartbeat: '12s ago',
          gpus: Array.from({ length: 8 }).map((_, i) => ({
            gpu_id: `gpu-node-04-gpu-${i}`,
            index: i,
            model: 'NVIDIA H100 SXM5',
            utilization: i === 1 ? 0.0 : 64.0,
            memory_used_gb: i === 1 ? 0.0 : 52.0,
            memory_total_gb: 80,
            temperature: i === 1 ? 41 : 65,
            power_watts: i === 1 ? 120 : 490,
            status: i === 1 ? 'fault' : 'active',
            xid_error: i === 1 ? 31 : undefined,
          })),
        },
      ]
    }
  },

  cordonNode: async (_nodeId: string): Promise<void> => {
    // Simulates K8s node cordon
    await new Promise((resolve) => setTimeout(resolve, 400))
  },

  // Jobs Explorer
  getJobs: async (clusterId = 'sim-cluster-01'): Promise<JobDetail[]> => {
    try {
      const res = await fetch(`${API_BASE}/jobs?cluster_id=${clusterId}`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to fetch jobs')
      const body = await res.json()
      if (body.data && Array.isArray(body.data)) {
        return body.data.map((j: any) => ({
          id: j.id,
          name: j.name || j.id,
          namespace: j.namespace || 'ml-workloads',
          framework: j.framework || 'PyTorch DDP',
          gpu_count: j.gpu_count || 8,
          status: j.status || 'running',
          avg_gpu_util: j.avg_gpu_util || 88.5,
          avg_vram_util: j.avg_vram_util || 75.0,
          vram_allocated_gb: j.vram_allocated_gb || 640,
          vram_reserved_gb: j.vram_reserved_gb || 640,
          started_at: j.started_at || new Date().toISOString(),
          duration: j.duration || '4h 12m',
          estimated_cost: j.estimated_cost || '$164.80',
          cluster_id: j.cluster_id || clusterId,
          node_ids: j.node_ids || ['gpu-node-01'],
        }))
      }
      return []
    } catch {
      return [
        {
          id: 'job-llm-pretrain-01',
          name: 'llama3-70b-pretrain-run4',
          namespace: 'core-ai',
          framework: 'Megatron-LM + DeepSpeed',
          gpu_count: 16,
          status: 'running',
          avg_gpu_util: 93.4,
          avg_vram_util: 88.0,
          vram_allocated_gb: 1280,
          vram_reserved_gb: 1280,
          started_at: new Date(Date.now() - 14 * 3600000).toISOString(),
          duration: '14h 22m',
          estimated_cost: '$1,146.88',
          cluster_id: clusterId,
          node_ids: ['gpu-node-01', 'gpu-node-02'],
        },
        {
          id: 'job-deepspeed-pretrain-102',
          name: 'deepseek-moe-pretrain-stage1',
          namespace: 'pretraining-core',
          framework: 'DeepSpeed ZeRO-3',
          gpu_count: 8,
          status: 'cpu_bottleneck',
          avg_gpu_util: 11.2,
          avg_vram_util: 52.5,
          vram_allocated_gb: 42.0,
          vram_reserved_gb: 48.0,
          started_at: new Date(Date.now() - 6 * 3600000).toISOString(),
          duration: '6h 05m',
          estimated_cost: '$245.76',
          cluster_id: clusterId,
          node_ids: ['gpu-node-03'],
        },
        {
          id: 'job-vllm-inference-02',
          name: 'vllm-mistral-large-serving',
          namespace: 'production-serving',
          framework: 'vLLM / TensorRT-LLM',
          gpu_count: 4,
          status: 'running',
          avg_gpu_util: 89.2,
          avg_vram_util: 91.0,
          vram_allocated_gb: 72.0,
          vram_reserved_gb: 78.0,
          started_at: new Date(Date.now() - 48 * 3600000).toISOString(),
          duration: '2d 0h',
          estimated_cost: '$492.00',
          cluster_id: clusterId,
          node_ids: ['gpu-node-01'],
        },
        {
          id: 'job-ray-tune-embedding-33',
          name: 'ray-tune-clip-embedding',
          namespace: 'multimodal-ai',
          framework: 'Ray Train / TorchVision',
          gpu_count: 4,
          status: 'stalled',
          avg_gpu_util: 0.5,
          avg_vram_util: 45.0,
          vram_allocated_gb: 36.0,
          vram_reserved_gb: 40.0,
          started_at: new Date(Date.now() - 3 * 3600000).toISOString(),
          duration: '3h 18m',
          estimated_cost: '$61.44',
          cluster_id: clusterId,
          node_ids: ['gpu-node-04'],
        },
      ]
    }
  },

  restartJob: async (_jobId: string): Promise<void> => {
    await new Promise((resolve) => setTimeout(resolve, 500))
  },

  // Alerts Management
  getAlerts: async (clusterId = 'sim-cluster-01'): Promise<AlertDetail[]> => {
    try {
      const res = await fetch(`${API_BASE}/alerts/recent?cluster_id=${clusterId}`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to fetch alerts')
      const body = await res.json()
      return body.data || []
    } catch {
      return [
        {
          id: 'alert-001',
          type: 'xid_hardware_error',
          severity: 'critical',
          title: 'NVIDIA Hardware Fault XID 31',
          message: 'Critical memory page fault on gpu-node-04 GPU 1. Immediate node isolation recommended.',
          cluster_id: clusterId,
          node_id: 'gpu-node-04',
          gpu_id: 'gpu-node-04-gpu-1',
          fired_at: new Date(Date.now() - 5 * 60000).toISOString(),
          status: 'active',
          channels: ['Slack #infra-alerts', 'PagerDuty: Urgent'],
        },
        {
          id: 'alert-002',
          type: 'cpu_starvation',
          severity: 'high',
          title: 'Host CPU Data-Loading Bottleneck',
          message: 'CPU usage >95% while GPUs are starving at 11% compute on gpu-node-03.',
          cluster_id: clusterId,
          node_id: 'gpu-node-03',
          fired_at: new Date(Date.now() - 14 * 60000).toISOString(),
          status: 'active',
          channels: ['Slack #infra-alerts'],
        },
        {
          id: 'alert-003',
          type: 'idle_gpu',
          severity: 'high',
          title: 'Unallocated Idle GPU Fleet Compute',
          message: 'GPU gpu-node-01-gpu-3 has been idle at 0% compute for over 25 minutes.',
          cluster_id: clusterId,
          node_id: 'gpu-node-01',
          gpu_id: 'gpu-node-01-gpu-3',
          fired_at: new Date(Date.now() - 25 * 60000).toISOString(),
          status: 'acknowledged',
          channels: ['Slack #infra-alerts'],
        },
      ]
    }
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
        estimated_cost_usd: 12450.0,
        estimated_savings_usd: 5489.6,
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
