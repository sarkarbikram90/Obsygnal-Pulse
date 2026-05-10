import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card'
import { Server, Activity, AlertTriangle, DollarSign } from 'lucide-react'
import { api } from '../../api/client'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, ResponsiveContainer } from 'recharts'

// Mock chart data for MVP visualization
const mockChartData = Array.from({ length: 24 }).map((_, i) => ({
  time: `${i}:00`,
  utilization: Math.floor(Math.random() * 40) + 40, // 40-80%
}))

export default function Dashboard() {
  const { data: overview, isLoading, isError } = useQuery({
    queryKey: ['clusterOverview', 'default-cluster'],
    queryFn: () => api.getClusterOverview('default-cluster'),
    retry: false,
  })

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-white mb-2">Cluster Overview</h1>
        <p className="text-neutral-400">Real-time metrics and health status for your GPU infrastructure.</p>
      </div>

      {isLoading ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="bg-neutral-900/50 border-neutral-800 animate-pulse">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <div className="h-4 w-24 bg-neutral-800 rounded"></div>
              </CardHeader>
              <CardContent>
                <div className="h-8 w-16 bg-neutral-800 rounded mb-2"></div>
                <div className="h-3 w-32 bg-neutral-800 rounded"></div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : isError ? (
        <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-lg text-rose-400">
          Failed to load cluster overview data. Ensure the backend API is running.
        </div>
      ) : (
        <>
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
            <Card className="bg-neutral-900/50 border-neutral-800">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium text-neutral-400">Total GPUs</CardTitle>
                <Server className="h-4 w-4 text-neutral-500" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-white">{overview?.total_gpus || 0}</div>
                <p className="text-xs text-neutral-500 mt-1">
                  <span className="text-emerald-400 font-medium">{overview?.active_gpus || 0} active</span>, {overview?.idle_gpus || 0} idle
                </p>
              </CardContent>
            </Card>

            <Card className="bg-neutral-900/50 border-neutral-800">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium text-neutral-400">Avg Utilization</CardTitle>
                <Activity className="h-4 w-4 text-neutral-500" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-white">{overview?.avg_utilization?.toFixed(1) || '0.0'}%</div>
                <div className="w-full bg-neutral-800 rounded-full h-1.5 mt-3 overflow-hidden">
                  <div 
                    className="bg-indigo-500 h-1.5 rounded-full" 
                    style={{ width: `${overview?.avg_utilization || 0}%` }}
                  />
                </div>
              </CardContent>
            </Card>

            <Card className="bg-neutral-900/50 border-neutral-800 relative overflow-hidden">
              <div className="absolute top-0 right-0 p-4 opacity-10">
                <AlertTriangle className="w-16 h-16 text-rose-500" />
              </div>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2 relative z-10">
                <CardTitle className="text-sm font-medium text-neutral-400">Inefficiency Score</CardTitle>
                <AlertTriangle className="h-4 w-4 text-rose-500" />
              </CardHeader>
              <CardContent className="relative z-10">
                <div className="text-3xl font-bold text-white">{overview?.inefficiency_score?.toFixed(1) || '0.0'}</div>
                <p className="text-xs text-rose-400/80 mt-1">
                  Needs attention
                </p>
              </CardContent>
            </Card>

            <Card className="bg-neutral-900/50 border-neutral-800">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium text-neutral-400">Est. Waste (Mo)</CardTitle>
                <DollarSign className="h-4 w-4 text-emerald-500" />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold text-white">${overview?.estimated_waste_usd?.toLocaleString() || '0'}</div>
                <p className="text-xs text-neutral-500 mt-1">
                  Based on current run rate
                </p>
              </CardContent>
            </Card>
          </div>

          <div className="grid gap-6 grid-cols-1 lg:grid-cols-7">
            <Card className="col-span-1 lg:col-span-4 bg-neutral-900/50 border-neutral-800">
              <CardHeader>
                <CardTitle className="text-lg font-medium text-white">Cluster Utilization (24h)</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-[300px] w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={mockChartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                      <defs>
                        <linearGradient id="colorUv" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                          <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                        </linearGradient>
                      </defs>
                      <CartesianGrid strokeDasharray="3 3" stroke="#262626" vertical={false} />
                      <XAxis dataKey="time" stroke="#525252" fontSize={12} tickLine={false} axisLine={false} />
                      <YAxis stroke="#525252" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(v) => `${v}%`} />
                      <RechartsTooltip 
                        contentStyle={{ backgroundColor: '#171717', borderColor: '#262626', borderRadius: '8px', color: '#f5f5f5' }}
                        itemStyle={{ color: '#818cf8' }}
                      />
                      <Area type="monotone" dataKey="utilization" stroke="#6366f1" strokeWidth={2} fillOpacity={1} fill="url(#colorUv)" />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              </CardContent>
            </Card>
            
            <Card className="col-span-1 lg:col-span-4 bg-neutral-900/50 border-neutral-800">
              <CardHeader>
                <CardTitle className="text-lg font-medium text-white">GPU Utilization Heatmap</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {[1, 2, 3, 4].map((node) => (
                    <div key={node} className="space-y-2">
                      <div className="text-xs font-medium text-neutral-400">worker-node-{node}</div>
                      <div className="grid grid-cols-8 gap-1">
                        {Array.from({ length: 8 }).map((_, gpu) => {
                          const intensity = Math.floor(Math.random() * 5); // 0-4
                          const colors = [
                            'bg-neutral-800', // 0: Idle/Offline
                            'bg-indigo-900/50', // 1: Low
                            'bg-indigo-600/70', // 2: Medium
                            'bg-indigo-500', // 3: High
                            'bg-rose-500' // 4: Overheated/Saturated
                          ];
                          return (
                            <div 
                              key={gpu} 
                              className={`h-6 rounded-sm ${colors[intensity]} transition-colors duration-500 hover:ring-2 ring-white/20 cursor-pointer`}
                              title={`GPU ${gpu}: ${intensity * 25}%`}
                            />
                          );
                        })}
                      </div>
                    </div>
                  ))}
                  <div className="flex items-center space-x-4 pt-4 text-xs text-neutral-500">
                    <div className="flex items-center"><div className="w-3 h-3 rounded-sm bg-neutral-800 mr-2" /> Idle</div>
                    <div className="flex items-center"><div className="w-3 h-3 rounded-sm bg-indigo-900/50 mr-2" /> Low</div>
                    <div className="flex items-center"><div className="w-3 h-3 rounded-sm bg-indigo-600/70 mr-2" /> Med</div>
                    <div className="flex items-center"><div className="w-3 h-3 rounded-sm bg-indigo-500 mr-2" /> High</div>
                    <div className="flex items-center"><div className="w-3 h-3 rounded-sm bg-rose-500 mr-2" /> Sat/Hot</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </div>
  )
}
