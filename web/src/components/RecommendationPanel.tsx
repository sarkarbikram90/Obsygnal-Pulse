import type { Recommendation } from '../types/api';

const mockRecommendations: Recommendation[] = [
  {
    id: 'rec-1', category: 'reduce_idle_runtime', status: 'active',
    title: 'Idle GPU detected: gpu-node-002-gpu-3',
    description: 'GPU is completely idle with no running processes. Release it or assign a workload.',
    explanation: 'This GPU is consuming power and cloud costs without performing any computation.',
    confidence_score: 0.95,
    estimated_savings: { cost_per_hour_usd: 2.50, cost_per_month_usd: 1825, gpu_hours_per_month: 730, memory_gb: 0, power_watts: 0 },
    affected_metrics: ['gpu_utilization'], cluster_id: 'sim-cluster-01', node_id: 'gpu-node-002', gpu_id: 'gpu-node-002-gpu-3',
    job_id: '', inefficiency_id: '1', created_at: new Date().toISOString(), expires_at: new Date(Date.now() + 86400000).toISOString(),
  },
  {
    id: 'rec-2', category: 'increase_batch_size', status: 'active',
    title: 'Increase batch size for workload on gpu-node-001-gpu-2',
    description: 'GPU has low utilization (12.3%). Increasing the training batch size could improve GPU throughput.',
    explanation: 'Low GPU utilization often indicates the batch size is too small to saturate the GPU compute pipeline.',
    confidence_score: 0.72,
    estimated_savings: { cost_per_hour_usd: 0, cost_per_month_usd: 0, gpu_hours_per_month: 50, memory_gb: 0, power_watts: 0 },
    affected_metrics: ['gpu_utilization'], cluster_id: 'sim-cluster-01', node_id: 'gpu-node-001', gpu_id: 'gpu-node-001-gpu-2',
    job_id: 'job-000042', inefficiency_id: '2', created_at: new Date(Date.now() - 300000).toISOString(), expires_at: new Date(Date.now() + 86400000).toISOString(),
  },
  {
    id: 'rec-3', category: 'investigate_cpu_bottleneck', status: 'active',
    title: 'Stalled training loop on gpu-node-003-gpu-1',
    description: 'Near-zero compute with memory still allocated and processes running. Training loop may be hung.',
    explanation: 'This pattern typically indicates a deadlock or a hung I/O operation.',
    confidence_score: 0.90,
    estimated_savings: { cost_per_hour_usd: 3.67, cost_per_month_usd: 88.08, gpu_hours_per_month: 24, memory_gb: 0, power_watts: 0 },
    affected_metrics: ['gpu_utilization', 'gpu_memory_utilization'], cluster_id: 'sim-cluster-01', node_id: 'gpu-node-003', gpu_id: 'gpu-node-003-gpu-1',
    job_id: 'job-000019', inefficiency_id: '3', created_at: new Date(Date.now() - 120000).toISOString(), expires_at: new Date(Date.now() + 86400000).toISOString(),
  },
];

function confidenceBar(score: number) {
  const pct = Math.round(score * 100);
  const color = score > 0.8 ? 'var(--accent-emerald)' : score > 0.5 ? 'var(--accent-amber)' : 'var(--accent-rose)';
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 12 }}>
      <div style={{ width: 80, height: 6, borderRadius: 3, background: 'rgba(148, 163, 184, 0.15)' }}>
        <div style={{ width: `${pct}%`, height: '100%', borderRadius: 3, background: color, transition: 'width 0.5s ease' }} />
      </div>
      <span style={{ color, fontWeight: 600 }}>{pct}%</span>
    </div>
  );
}

function formatCurrency(amount: number): string {
  if (amount >= 1000) return `$${(amount / 1000).toFixed(1)}k`;
  if (amount > 0) return `$${amount.toFixed(2)}`;
  return '—';
}

function categoryIcon(category: string): string {
  const icons: Record<string, string> = {
    reduce_gpu_allocation: '📉',
    increase_batch_size: '📦',
    increase_dataloader_workers: '⚡',
    reduce_idle_runtime: '🔌',
    investigate_cpu_bottleneck: '🔍',
    right_size_memory: '💾',
  };
  return icons[category] || '💡';
}

export function RecommendationPanel() {
  return (
    <div className="glass-card">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h3 style={{ fontSize: 16, fontWeight: 600 }}>Optimization Recommendations</h3>
        <span style={{ fontSize: 13, color: 'var(--accent-emerald)', fontWeight: 600 }}>
          {mockRecommendations.length} active
        </span>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12, maxHeight: 400, overflowY: 'auto' }}>
        {mockRecommendations.map(rec => (
          <div key={rec.id} style={{
            padding: '14px 16px', borderRadius: 10,
            background: 'rgba(17, 24, 39, 0.6)',
            border: '1px solid var(--border-subtle)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 6 }}>
              <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <span style={{ fontSize: 18 }}>{categoryIcon(rec.category)}</span>
                <span style={{ fontSize: 14, fontWeight: 600 }}>{rec.title}</span>
              </div>
            </div>

            <p style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.5, marginBottom: 10 }}>
              {rec.description}
            </p>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div style={{ display: 'flex', gap: 20, alignItems: 'center', fontSize: 12 }}>
                <div>
                  <span style={{ color: 'var(--text-muted)' }}>Confidence: </span>
                  {confidenceBar(rec.confidence_score)}
                </div>
                <div>
                  <span style={{ color: 'var(--text-muted)' }}>Savings: </span>
                  <span style={{ color: 'var(--accent-amber)', fontWeight: 600 }}>
                    {formatCurrency(rec.estimated_savings.cost_per_month_usd)}/mo
                  </span>
                </div>
              </div>

              <div style={{ display: 'flex', gap: 8 }}>
                <button style={{
                  padding: '6px 12px', borderRadius: 6, fontSize: 12, fontWeight: 500,
                  background: 'rgba(99, 102, 241, 0.15)', color: 'var(--accent-indigo)',
                  border: '1px solid rgba(99, 102, 241, 0.3)', cursor: 'pointer',
                  transition: 'all 0.2s ease',
                }}
                onMouseEnter={e => { (e.target as HTMLButtonElement).style.background = 'rgba(99, 102, 241, 0.3)'; }}
                onMouseLeave={e => { (e.target as HTMLButtonElement).style.background = 'rgba(99, 102, 241, 0.15)'; }}
                >
                  Apply
                </button>
                <button style={{
                  padding: '6px 12px', borderRadius: 6, fontSize: 12, fontWeight: 500,
                  background: 'transparent', color: 'var(--text-muted)',
                  border: '1px solid rgba(148, 163, 184, 0.15)', cursor: 'pointer',
                  transition: 'all 0.2s ease',
                }}
                onMouseEnter={e => { (e.target as HTMLButtonElement).style.borderColor = 'rgba(148, 163, 184, 0.3)'; }}
                onMouseLeave={e => { (e.target as HTMLButtonElement).style.borderColor = 'rgba(148, 163, 184, 0.15)'; }}
                >
                  Dismiss
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Total Savings Banner */}
      <div style={{
        marginTop: 16, padding: '12px 16px', borderRadius: 10,
        background: 'linear-gradient(135deg, rgba(34, 197, 94, 0.1), rgba(99, 102, 241, 0.1))',
        border: '1px solid rgba(34, 197, 94, 0.2)',
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
      }}>
        <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>💰 Total potential savings</span>
        <span style={{ fontSize: 18, fontWeight: 700, color: 'var(--accent-emerald)' }}>
          {formatCurrency(mockRecommendations.reduce((s, r) => s + r.estimated_savings.cost_per_month_usd, 0))}/mo
        </span>
      </div>
    </div>
  );
}
