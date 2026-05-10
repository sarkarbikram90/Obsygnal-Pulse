// API types matching the Go backend domain models

export interface GPUMetric {
  id: string;
  timestamp: string;
  cluster_id: string;
  node_id: string;
  gpu_id: string;
  gpu_model: string;
  gpu_index: number;
  utilization: number;
  memory_utilization: number;
  memory_used: number;
  memory_total: number;
  memory_free: number;
  temperature: number;
  power_usage: number;
  power_limit: number;
  fan_speed: number;
  pcie_tx_bytes: number;
  pcie_rx_bytes: number;
  running_processes: number;
  job_id: string;
  labels: Record<string, string>;
}

export interface NodeMetric {
  timestamp: string;
  cluster_id: string;
  node_id: string;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  network_in: number;
  network_out: number;
  gpu_count: number;
  pod_count: number;
}

export interface ClusterOverview {
  cluster_id: string;
  total_nodes: number;
  healthy_nodes: number;
  total_gpus: number;
  active_gpus: number;
  idle_gpus: number;
  avg_utilization: number;
  avg_memory_usage: number;
  active_jobs: number;
  failed_jobs_24h: number;
  inefficiency_score: number;
  estimated_waste_usd: number;
  timestamp: string;
}

export interface Inefficiency {
  id: string;
  type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  cluster_id: string;
  node_id: string;
  gpu_id: string;
  job_id: string;
  description: string;
  affected_metrics: string[];
  current_value: number;
  threshold_value: number;
  detected_at: string;
  is_active: boolean;
}

export interface Recommendation {
  id: string;
  category: string;
  status: 'active' | 'dismissed' | 'applied' | 'expired';
  title: string;
  description: string;
  explanation: string;
  confidence_score: number;
  estimated_savings: EstimatedSavings;
  affected_metrics: string[];
  cluster_id: string;
  node_id: string;
  gpu_id: string;
  job_id: string;
  inefficiency_id: string;
  created_at: string;
  expires_at: string;
}

export interface EstimatedSavings {
  cost_per_hour_usd: number;
  cost_per_month_usd: number;
  gpu_hours_per_month: number;
  memory_gb: number;
  power_watts: number;
}

export interface Alert {
  id: string;
  type: string;
  severity: string;
  title: string;
  message: string;
  cluster_id: string;
  node_id: string;
  gpu_id: string;
  fired_at: string;
  channels: string[];
}

export interface HealthReport {
  status: string;
  checks: { name: string; status: string; latency_ms: number; error?: string }[];
  version: string;
  uptime_seconds: number;
}

export interface IngestionStats {
  total_received: number;
  total_inserted: number;
  total_dropped: number;
  total_duplicates: number;
  buffer_gpu_size: number;
  buffer_node_size: number;
}
