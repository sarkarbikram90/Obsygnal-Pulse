import { useState, useEffect } from 'react';

interface GPUCell {
  gpu_id: string;
  node_id: string;
  utilization: number;
  memory_utilization: number;
  temperature: number;
  status: 'active' | 'idle' | 'hot';
}

// Generate simulated heatmap data for demo
function generateHeatmapData(): GPUCell[] {
  const nodes = ['gpu-node-000', 'gpu-node-001', 'gpu-node-002', 'gpu-node-003'];
  const cells: GPUCell[] = [];
  nodes.forEach(node => {
    for (let g = 0; g < 4; g++) {
      const util = Math.random() * 100;
      cells.push({
        gpu_id: `${node}-gpu-${g}`,
        node_id: node,
        utilization: util,
        memory_utilization: Math.random() * 100,
        temperature: 30 + Math.random() * 60,
        status: util < 5 ? 'idle' : util > 90 ? 'hot' : 'active',
      });
    }
  });
  return cells;
}

function utilizationColor(util: number): string {
  if (util < 5) return 'rgba(239, 68, 68, 0.6)';
  if (util < 20) return 'rgba(249, 115, 22, 0.6)';
  if (util < 50) return 'rgba(234, 179, 8, 0.5)';
  if (util < 80) return 'rgba(34, 197, 94, 0.5)';
  return 'rgba(34, 197, 94, 0.8)';
}

export function GPUHeatmap() {
  const [cells, setCells] = useState<GPUCell[]>(generateHeatmapData());
  const [hoveredCell, setHoveredCell] = useState<GPUCell | null>(null);

  useEffect(() => {
    const interval = setInterval(() => setCells(generateHeatmapData()), 15000);
    return () => clearInterval(interval);
  }, []);

  const nodes = [...new Set(cells.map(c => c.node_id))];

  return (
    <div className="glass-card" style={{ height: '100%' }}>
      <h3 style={{ fontSize: 16, fontWeight: 600, marginBottom: 16, color: 'var(--text-primary)' }}>
        GPU Utilization Heatmap
      </h3>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {nodes.map(node => (
          <div key={node} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span style={{ fontSize: 11, color: 'var(--text-muted)', width: 100, textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }}>
              {node}
            </span>
            <div style={{ display: 'flex', gap: 6, flex: 1 }}>
              {cells.filter(c => c.node_id === node).map(cell => (
                <div
                  key={cell.gpu_id}
                  className="heatmap-cell"
                  onMouseEnter={() => setHoveredCell(cell)}
                  onMouseLeave={() => setHoveredCell(null)}
                  style={{
                    width: 64, height: 64,
                    background: utilizationColor(cell.utilization),
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: 14, fontWeight: 600, color: '#fff',
                  }}
                >
                  {cell.utilization.toFixed(0)}%
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Legend */}
      <div style={{ display: 'flex', gap: 16, marginTop: 16, justifyContent: 'center' }}>
        {[
          { label: 'Idle (<5%)', color: 'rgba(239, 68, 68, 0.6)' },
          { label: 'Low (<20%)', color: 'rgba(249, 115, 22, 0.6)' },
          { label: 'Medium', color: 'rgba(234, 179, 8, 0.5)' },
          { label: 'Good (>50%)', color: 'rgba(34, 197, 94, 0.5)' },
          { label: 'High (>80%)', color: 'rgba(34, 197, 94, 0.8)' },
        ].map(l => (
          <div key={l.label} style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--text-muted)' }}>
            <div style={{ width: 12, height: 12, borderRadius: 3, background: l.color }} />
            {l.label}
          </div>
        ))}
      </div>

      {/* Tooltip */}
      {hoveredCell && (
        <div style={{
          marginTop: 12, padding: '12px 16px',
          background: 'rgba(17, 24, 39, 0.95)', borderRadius: 8,
          border: '1px solid var(--border-subtle)', fontSize: 13,
        }}>
          <strong>{hoveredCell.gpu_id}</strong>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '4px 24px', marginTop: 8 }}>
            <span style={{ color: 'var(--text-muted)' }}>Compute:</span>
            <span style={{ fontWeight: 600 }}>{hoveredCell.utilization.toFixed(1)}%</span>
            <span style={{ color: 'var(--text-muted)' }}>Memory:</span>
            <span style={{ fontWeight: 600 }}>{hoveredCell.memory_utilization.toFixed(1)}%</span>
            <span style={{ color: 'var(--text-muted)' }}>Temp:</span>
            <span style={{ fontWeight: 600, color: hoveredCell.temperature > 85 ? 'var(--accent-rose)' : 'inherit' }}>
              {hoveredCell.temperature.toFixed(0)}°C
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
