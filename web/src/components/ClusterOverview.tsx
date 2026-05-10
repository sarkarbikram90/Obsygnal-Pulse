import type { ClusterOverview } from '../types/api';

interface Props { overview: ClusterOverview | null }

export function ClusterOverviewPanel({ overview }: Props) {
  const cards = overview ? [
    { label: 'Total GPUs', value: overview.total_gpus, icon: '🖥️', color: 'var(--accent-indigo)' },
    { label: 'Active GPUs', value: overview.active_gpus, icon: '🟢', color: 'var(--accent-emerald)' },
    { label: 'Idle GPUs', value: overview.idle_gpus, icon: '🔴', color: 'var(--accent-rose)' },
    { label: 'Avg Utilization', value: `${overview.avg_utilization.toFixed(1)}%`, icon: '📊', color: 'var(--accent-cyan)' },
    { label: 'Avg Memory', value: `${overview.avg_memory_usage.toFixed(1)}%`, icon: '💾', color: 'var(--accent-violet)' },
    { label: 'Nodes', value: `${overview.healthy_nodes}/${overview.total_nodes}`, icon: '🖧', color: 'var(--accent-indigo)' },
    { label: 'Inefficiency Score', value: `${overview.inefficiency_score.toFixed(0)}`, icon: '⚠️', color: overview.inefficiency_score > 50 ? 'var(--accent-rose)' : 'var(--accent-amber)' },
    { label: 'Est. Waste/hr', value: `$${overview.estimated_waste_usd.toFixed(2)}`, icon: '💸', color: 'var(--accent-amber)' },
  ] : Array(8).fill(null).map((_, i) => ({ label: '—', value: '—', icon: '⏳', color: 'var(--text-muted)', key: i }));

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16 }}>
      {cards.map((c, i) => (
        <div key={i} className="stat-card" style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontSize: 13, color: 'var(--text-secondary)', fontWeight: 500 }}>{c.label}</span>
            <span style={{ fontSize: 20 }}>{c.icon}</span>
          </div>
          <span style={{ fontSize: 28, fontWeight: 700, color: c.color, letterSpacing: '-1px' }}>
            {c.value}
          </span>
        </div>
      ))}
    </div>
  );
}
