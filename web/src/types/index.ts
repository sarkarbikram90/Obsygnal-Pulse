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
