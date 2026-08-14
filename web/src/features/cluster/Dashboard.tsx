import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card'
import { Server, Activity, AlertTriangle, DollarSign, Sparkles } from 'lucide-react'
import { api } from '../../api/client'
import { GPUHeatmap } from '../../components/GPUHeatmap'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, ResponsiveContainer } from 'recharts'

// 24-hour realistic cluster telemetry area chart
const clusterTelemetryHistory = [
  { time: '00:00', utilization: 68.2, vram: 58.4 },
  { time: '02:00', utilization: 71.0, vram: 60.1 },
  { time: '04:00', utilization: 74.5, vram: 62.0 },
  { time: '06:00', utilization: 62.0, vram: 55.4 },
  { time: '08:00', utilization: 84.1, vram: 69.2 },
  { time: '10:00', utilization: 89.5, vram: 74.0 },
  { time: '12:00', utilization: 92.1, vram: 76.8 },
  { time: '14:00', utilization: 88.4, vram: 73.5 },
  { time: '16:00', utilization: 85.0, vram: 71.2 },
  { time: '18:00', utilization: 79.8, vram: 68.0 },
  { time: '20:00', utilization: 83.2, vram: 70.4 },
  { time: '22:00', utilization: 78.5, vram: 64.2 },
]

export default function Dashboard() {
  const { data: overview, isLoading, isError } = useQuery({
    queryKey: ['clusterOverview', 'sim-cluster-01'],
    queryFn: () => api.getClusterOverview('sim-cluster-01'),
  })

  return (
    <div className="space-y-8 animate-in fade-in duration-500 text-slate-100">
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-orange-400 to-rose-400 bg-clip-text text-transparent mb-2">
            Cluster Telemetry & Efficiency Hub
          </h1>
          <p className="text-slate-400">High-frequency NVIDIA DCGM telemetry, PyTorch memory profiling, and cluster cost health.</p>
        </div>
      </div>

      {isLoading ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="bg-slate-900 border-slate-800 p-6 rounded-2xl h-32 animate-pulse" />
          ))}
        </div>
      ) : isError && !overview ? (
        <div className="p-4 bg-rose-500/10 border border-rose-500/30 rounded-2xl text-rose-400 text-sm">
          Failed to load cluster overview data. Ensure the backend API is running.
        </div>
      ) : (
        <>
          {/* Top 4 KPI Metrics */}
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
            <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 p-0 pb-2">
                <CardTitle className="text-xs font-bold uppercase tracking-wider text-slate-400">Total GPU Fleet</CardTitle>
                <Server className="h-4 w-4 text-orange-400" />
              </CardHeader>
              <CardContent className="p-0">
                <div className="text-3xl font-black text-white">{overview?.total_gpus || 32} GPUs</div>
                <p className="text-xs text-slate-400 mt-1">
                  <span className="text-emerald-400 font-bold">{overview?.active_gpus || 24} active</span> • <span className="text-rose-400 font-bold">{overview?.idle_gpus || 8} idle</span>
                </p>
              </CardContent>
            </Card>

            <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 p-0 pb-2">
                <CardTitle className="text-xs font-bold uppercase tracking-wider text-slate-400">Avg Utilization</CardTitle>
                <Activity className="h-4 w-4 text-emerald-400" />
              </CardHeader>
              <CardContent className="p-0">
                <div className="text-3xl font-black text-white">{overview?.avg_utilization?.toFixed(1) || '78.5'}%</div>
                <div className="w-full bg-slate-950 rounded-full h-2 mt-3 overflow-hidden border border-slate-800">
                  <div
                    className="bg-gradient-to-r from-orange-400 to-emerald-400 h-full rounded-full"
                    style={{ width: `${overview?.avg_utilization || 78.5}%` }}
                  />
                </div>
              </CardContent>
            </Card>

            <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl relative overflow-hidden">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 p-0 pb-2">
                <CardTitle className="text-xs font-bold uppercase tracking-wider text-slate-400">Inefficiency Score</CardTitle>
                <AlertTriangle className="h-4 w-4 text-rose-400" />
              </CardHeader>
              <CardContent className="p-0">
                <div className="text-3xl font-black text-rose-400">{overview?.inefficiency_score?.toFixed(1) || '55.2'}</div>
                <p className="text-xs text-rose-400/80 mt-1">
                  High compute waste detected
                </p>
              </CardContent>
            </Card>

            <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 p-0 pb-2">
                <CardTitle className="text-xs font-bold uppercase tracking-wider text-slate-400">Monthly Waste Rate</CardTitle>
                <DollarSign className="h-4 w-4 text-emerald-400" />
              </CardHeader>
              <CardContent className="p-0">
                <div className="text-3xl font-black text-emerald-400">${overview?.estimated_waste_usd?.toLocaleString() || '6,820'}</div>
                <p className="text-xs text-slate-400 mt-1">
                  Recoverable via 1-click auto-fix
                </p>
              </CardContent>
            </Card>
          </div>

          {/* GPU Heatmap & 24h Area Chart Grid */}
          <div className="grid gap-6 grid-cols-1 lg:grid-cols-7">
            {/* Live GPU Heatmap */}
            <div className="col-span-1 lg:col-span-4">
              <GPUHeatmap />
            </div>

            {/* 24h Compute Saturation Chart */}
            <Card className="col-span-1 lg:col-span-3 bg-slate-900 border-slate-800 p-6 rounded-2xl flex flex-col justify-between">
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <Sparkles className="w-4 h-4 text-orange-400" />
                  <h3 className="text-base font-bold text-white">Cluster Saturation (24h)</h3>
                </div>
                <p className="text-xs text-slate-400 mb-6">Historical compute vs VRAM allocation curve</p>
              </div>

              <div className="h-[260px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={clusterTelemetryHistory} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <defs>
                      <linearGradient id="utilGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#f97316" stopOpacity={0.4} />
                        <stop offset="95%" stopColor="#f97316" stopOpacity={0} />
                      </linearGradient>
                      <linearGradient id="vramGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#818cf8" stopOpacity={0.4} />
                        <stop offset="95%" stopColor="#818cf8" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                    <XAxis dataKey="time" stroke="#64748b" fontSize={11} tickLine={false} axisLine={false} />
                    <YAxis stroke="#64748b" fontSize={11} tickLine={false} axisLine={false} tickFormatter={(v) => `${v}%`} />
                    <RechartsTooltip
                      contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '12px', color: '#f8fafc' }}
                      itemStyle={{ color: '#fb923c' }}
                    />
                    <Area type="monotone" dataKey="utilization" stroke="#f97316" strokeWidth={2} fillOpacity={1} fill="url(#utilGradient)" name="Compute Util %" />
                    <Area type="monotone" dataKey="vram" stroke="#818cf8" strokeWidth={2} fillOpacity={1} fill="url(#vramGradient)" name="VRAM Util %" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>

              <div className="flex justify-center gap-6 pt-4 border-t border-slate-800 text-xs text-slate-400">
                <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-full bg-orange-500" /> GPU Compute %</div>
                <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-full bg-indigo-400" /> VRAM Usage %</div>
              </div>
            </Card>
          </div>
        </>
      )}
    </div>
  )
}
