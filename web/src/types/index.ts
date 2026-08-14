export interface GPUMetric {
  id: string
  timestamp: string
  cluster_id: string
  node_id: string
  gpu_id: string
  utilization: number
  memory_used: number
  memory_total: number
  memory_utilization: number
  temperature: number
  power_draw: number
  power_limit: number
  xid_error_code?: number
  nvlink_bandwidth_gbps?: number
  tensor_core_utilization?: number
  cuda_memory_reserved_bytes?: number
  kv_cache_utilization?: number
}

export interface NodeMetric {
  id: string
  timestamp: string
  cluster_id: string
  node_id: string
  cpu_utilization: number
  memory_used: number
  memory_total: number
  network_rx_bytes: number
  network_tx_bytes: number
}

export interface ClusterOverview {
  cluster_id: string
  total_nodes: number
  healthy_nodes: number
  total_gpus: number
  active_gpus: number
  idle_gpus: number
  avg_utilization: number
  avg_memory_usage: number
  active_jobs: number
  failed_jobs_24h: number
  inefficiency_score: number
  estimated_waste_usd: number
  timestamp: string
}

export interface Inefficiency {
  id: string
  type: string
  severity: "low" | "medium" | "high" | "critical"
  cluster_id: string
  node_id?: string
  gpu_id?: string
  job_id?: string
  description: string
  affected_metrics: string[]
  current_value: number
  threshold_value: number
  detected_at: string
  resolved_at?: string
  is_active: boolean
  labels?: Record<string, string>
}

export interface EstimatedSavings {
  cost_per_hour_usd: number
  cost_per_month_usd: number
  gpu_hours_per_month: number
  memory_gb?: number
  power_watts?: number
}

export interface Recommendation {
  id: string
  category: string
  status: "active" | "dismissed" | "applied" | "expired"
  title: string
  description: string
  explanation: string
  confidence_score: number
  estimated_savings: EstimatedSavings
  affected_metrics: string[]
  cluster_id: string
  node_id?: string
  gpu_id?: string
  job_id?: string
  inefficiency_id: string
  created_at: string
  expires_at: string
}

export interface GPUSummary {
  gpu_id: string
  index: number
  model: string
  utilization: number
  memory_used_gb: number
  memory_total_gb: number
  temperature: number
  power_watts: number
  status: 'active' | 'idle' | 'hot' | 'fault'
  xid_error?: number
}

export interface NodeDetail {
  node_id: string
  status: 'ready' | 'degraded' | 'cordoned'
  gpu_model: string
  gpu_count: number
  cpu_usage: number
  memory_gb: number
  driver_version: string
  cuda_version: string
  ip_address: string
  cluster_id: string
  last_heartbeat: string
  gpus: GPUSummary[]
}

export interface JobDetail {
  id: string
  name: string
  namespace: string
  framework: string
  gpu_count: number
  status: 'running' | 'cpu_bottleneck' | 'stalled' | 'completed' | 'failed'
  avg_gpu_util: number
  avg_vram_util: number
  vram_allocated_gb: number
  vram_reserved_gb: number
  started_at: string
  duration: string
  estimated_cost: string
  cluster_id: string
  node_ids: string[]
}

export interface AlertDetail {
  id: string
  type: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  title: string
  message: string
  cluster_id: string
  node_id?: string
  gpu_id?: string
  fired_at: string
  status: 'active' | 'resolved' | 'acknowledged'
  channels: string[]
}

export interface AlertChannelConfig {
  slack_enabled: boolean
  slack_webhook_url: string
  pagerduty_enabled: boolean
  pagerduty_service_key: string
  email_enabled: boolean
  email_recipients: string
}

export interface User {
  id: string
  email: string
  display_name: string
}

export interface Organization {
  id: string
  name: string
  tier: "free" | "pro" | "enterprise"
  role: "owner" | "admin" | "operator" | "viewer"
}

export interface UsageSummary {
  organization_id: string
  tier: "free" | "pro" | "enterprise"
  billing_cycle: string
  monitored_gpus: number
  gpu_hours_used: number
  gpu_hours_limit: number
  estimated_cost_usd: number
  estimated_savings_usd: number
  active_clusters: number
}

export interface APIKey {
  id: string
  name: string
  prefix: string
  secret_key?: string
  created_at: string
  last_used?: string
  description?: string
  status?: string
}
