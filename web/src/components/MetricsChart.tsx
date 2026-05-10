import { useState, useEffect } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';

interface DataPoint { time: string; utilization: number; memory: number; temperature: number; }

function generateTimeSeriesData(): DataPoint[] {
  const data: DataPoint[] = [];
  const now = Date.now();
  for (let i = 60; i >= 0; i--) {
    const t = new Date(now - i * 60000);
    data.push({
      time: t.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
      utilization: 40 + Math.sin(i / 5) * 25 + Math.random() * 15,
      memory: 55 + Math.cos(i / 8) * 15 + Math.random() * 10,
      temperature: 60 + Math.sin(i / 6) * 12 + Math.random() * 5,
    });
  }
  return data;
}

export function MetricsChart() {
  const [data, setData] = useState<DataPoint[]>(generateTimeSeriesData());

  useEffect(() => {
    const interval = setInterval(() => setData(generateTimeSeriesData()), 30000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="glass-card" style={{ height: '100%' }}>
      <h3 style={{ fontSize: 16, fontWeight: 600, marginBottom: 16 }}>
        Cluster Metrics — Last 60min
      </h3>
      <ResponsiveContainer width="100%" height={320}>
        <AreaChart data={data} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
          <defs>
            <linearGradient id="gradUtil" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#818cf8" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#818cf8" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="gradMem" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#22d3ee" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#22d3ee" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="gradTemp" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#fb7185" stopOpacity={0.2} />
              <stop offset="95%" stopColor="#fb7185" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="rgba(148, 163, 184, 0.08)" />
          <XAxis dataKey="time" stroke="#64748b" fontSize={11} tickLine={false} />
          <YAxis stroke="#64748b" fontSize={11} tickLine={false} domain={[0, 100]} />
          <Tooltip
            contentStyle={{
              background: 'rgba(17, 24, 39, 0.95)', border: '1px solid rgba(99, 102, 241, 0.2)',
              borderRadius: 8, fontSize: 13, color: '#f1f5f9',
            }}
          />
          <Legend wrapperStyle={{ fontSize: 12, color: '#94a3b8' }} />
          <Area type="monotone" dataKey="utilization" name="GPU Util %" stroke="#818cf8" fill="url(#gradUtil)" strokeWidth={2} dot={false} />
          <Area type="monotone" dataKey="memory" name="Memory %" stroke="#22d3ee" fill="url(#gradMem)" strokeWidth={2} dot={false} />
          <Area type="monotone" dataKey="temperature" name="Temp °C" stroke="#fb7185" fill="url(#gradTemp)" strokeWidth={1.5} dot={false} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
