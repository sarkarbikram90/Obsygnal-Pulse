import { useState, useEffect, useCallback } from 'react';
import './index.css';
import type { ClusterOverview, IngestionStats } from './types/api';
import { api } from './services/api';
import { ClusterOverviewPanel } from './components/ClusterOverview';
import { GPUHeatmap } from './components/GPUHeatmap';
import { MetricsChart } from './components/MetricsChart';
import { InefficiencyPanel } from './components/InefficiencyPanel';
import { RecommendationPanel } from './components/RecommendationPanel';
import { StatusBar } from './components/StatusBar';

function App() {
  const [overview, setOverview] = useState<ClusterOverview | null>(null);
  const [stats, setStats] = useState<IngestionStats | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  const refresh = useCallback(async () => {
    try {
      const [ov, st] = await Promise.all([
        api.clusterOverview().catch(() => null),
        api.ingestionStats().catch(() => null),
      ]);
      if (ov) setOverview(ov);
      if (st) setStats(st);
      setLastUpdate(new Date());
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to connect to API');
    }
  }, []);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, 15000);
    return () => clearInterval(interval);
  }, [refresh]);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg-primary)' }}>
      {/* Header */}
      <header style={{
        padding: '16px 32px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        borderBottom: '1px solid var(--border-subtle)',
        background: 'rgba(10, 14, 26, 0.9)',
        backdropFilter: 'blur(12px)',
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{
            width: 36, height: 36, borderRadius: 10,
            background: 'linear-gradient(135deg, var(--gradient-start), var(--gradient-end))',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 18,
          }}>⚡</div>
          <div>
            <h1 style={{ fontSize: 20, fontWeight: 700, letterSpacing: '-0.5px' }}>Obsygnal-Pulse</h1>
            <p style={{ fontSize: 12, color: 'var(--text-muted)' }}>GPU Infrastructure Intelligence</p>
          </div>
        </div>
        <StatusBar lastUpdate={lastUpdate} stats={stats} error={error} />
      </header>

      {/* Main Content */}
      <main style={{ padding: '24px 32px', maxWidth: 1600, margin: '0 auto' }}>
        {/* Cluster Overview Cards */}
        <section className="animate-in" style={{ marginBottom: 28 }}>
          <ClusterOverviewPanel overview={overview} />
        </section>

        {/* GPU Heatmap + Metrics Chart */}
        <section style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20, marginBottom: 28 }}>
          <div className="animate-in" style={{ animationDelay: '0.1s' }}>
            <GPUHeatmap />
          </div>
          <div className="animate-in" style={{ animationDelay: '0.15s' }}>
            <MetricsChart />
          </div>
        </section>

        {/* Inefficiencies + Recommendations */}
        <section style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
          <div className="animate-in" style={{ animationDelay: '0.2s' }}>
            <InefficiencyPanel />
          </div>
          <div className="animate-in" style={{ animationDelay: '0.25s' }}>
            <RecommendationPanel />
          </div>
        </section>
      </main>
    </div>
  );
}

export default App;
