import type { Inefficiency } from '../types/api';

const mockInefficiencies: Inefficiency[] = [
  { id: '1', type: 'idle_gpu', severity: 'high', cluster_id: 'sim-cluster-01', node_id: 'gpu-node-002', gpu_id: 'gpu-node-002-gpu-3', job_id: '', description: 'GPU gpu-node-002-gpu-3 is idle (0% utilization, no running processes)', affected_metrics: ['gpu_utilization'], current_value: 0, threshold_value: 2, detected_at: new Date().toISOString(), is_active: true },
  { id: '2', type: 'underutilized_gpu', severity: 'medium', cluster_id: 'sim-cluster-01', node_id: 'gpu-node-001', gpu_id: 'gpu-node-001-gpu-2', job_id: 'job-000042', description: 'GPU gpu-node-001-gpu-2 average utilization is 12.3% (threshold: 20%)', affected_metrics: ['gpu_utilization'], current_value: 12.3, threshold_value: 20, detected_at: new Date(Date.now() - 300000).toISOString(), is_active: true },
  { id: '3', type: 'stalled_training', severity: 'critical', cluster_id: 'sim-cluster-01', node_id: 'gpu-node-003', gpu_id: 'gpu-node-003-gpu-1', job_id: 'job-000019', description: 'GPU appears stalled: 2.1% compute but 65% memory allocated with 1 process', affected_metrics: ['gpu_utilization', 'gpu_memory_utilization'], current_value: 2.1, threshold_value: 5, detected_at: new Date(Date.now() - 120000).toISOString(), is_active: true },
  { id: '4', type: 'cpu_starvation', severity: 'high', cluster_id: 'sim-cluster-01', node_id: 'gpu-node-000', gpu_id: '', job_id: '', description: 'Node gpu-node-000 CPU at 94.2% while GPU utilization averages 35.1%', affected_metrics: ['cpu_usage', 'gpu_utilization'], current_value: 94.2, threshold_value: 90, detected_at: new Date(Date.now() - 600000).toISOString(), is_active: true },
];

function severityBadge(severity: string) {
  return <span className={`badge badge-${severity}`}>{severity}</span>;
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  return `${Math.floor(mins / 60)}h ago`;
}

function typeLabel(type: string): string {
  return type.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
}

export function InefficiencyPanel() {
  return (
    <div className="glass-card">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h3 style={{ fontSize: 16, fontWeight: 600 }}>Active Inefficiencies</h3>
        <span style={{ fontSize: 13, color: 'var(--accent-rose)', fontWeight: 600 }}>
          {mockInefficiencies.length} detected
        </span>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12, maxHeight: 400, overflowY: 'auto' }}>
        {mockInefficiencies.map(ineff => (
          <div key={ineff.id} className={ineff.severity === 'critical' ? 'pulse-alert' : ''} style={{
            padding: '14px 16px', borderRadius: 10,
            background: 'rgba(17, 24, 39, 0.6)',
            border: `1px solid ${ineff.severity === 'critical' ? 'rgba(239,68,68,0.3)' : 'var(--border-subtle)'}`,
            transition: 'all 0.2s ease',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 6 }}>
              <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                {severityBadge(ineff.severity)}
                <span style={{ fontSize: 12, color: 'var(--accent-indigo)', fontWeight: 500 }}>
                  {typeLabel(ineff.type)}
                </span>
              </div>
              <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>{timeAgo(ineff.detected_at)}</span>
            </div>
            <p style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.5 }}>
              {ineff.description}
            </p>
            <div style={{ display: 'flex', gap: 16, marginTop: 8, fontSize: 11, color: 'var(--text-muted)' }}>
              {ineff.gpu_id && <span>GPU: <strong style={{ color: 'var(--text-secondary)' }}>{ineff.gpu_id}</strong></span>}
              {ineff.node_id && <span>Node: <strong style={{ color: 'var(--text-secondary)' }}>{ineff.node_id}</strong></span>}
              <span>Value: <strong style={{ color: 'var(--accent-amber)' }}>{ineff.current_value.toFixed(1)}</strong></span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
