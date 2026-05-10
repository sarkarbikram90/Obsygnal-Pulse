import type { IngestionStats } from '../types/api';

interface Props {
  lastUpdate: Date;
  stats: IngestionStats | null;
  error: string | null;
}

export function StatusBar({ lastUpdate, stats, error }: Props) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 20, fontSize: 12 }}>
      {error ? (
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--accent-rose)' }}>
          <div style={{
            width: 8, height: 8, borderRadius: '50%',
            background: 'var(--accent-rose)',
            animation: 'pulse-glow 2s infinite',
          }} />
          <span>API Disconnected</span>
        </div>
      ) : (
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--accent-emerald)' }}>
          <div style={{
            width: 8, height: 8, borderRadius: '50%',
            background: 'var(--accent-emerald)',
          }} />
          <span>Connected</span>
        </div>
      )}

      {stats && (
        <>
          <div style={{ color: 'var(--text-muted)' }}>
            Ingested: <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>
              {stats.total_received.toLocaleString()}
            </span>
          </div>
          {stats.total_dropped > 0 && (
            <div style={{ color: 'var(--accent-amber)' }}>
              Dropped: {stats.total_dropped}
            </div>
          )}
        </>
      )}

      <div style={{ color: 'var(--text-muted)' }}>
        Updated: <span style={{ color: 'var(--text-secondary)' }}>
          {lastUpdate.toLocaleTimeString()}
        </span>
      </div>
    </div>
  );
}
