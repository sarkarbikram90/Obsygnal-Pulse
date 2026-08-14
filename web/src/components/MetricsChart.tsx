import { useState, useEffect } from 'react'
import { Card } from './ui/card'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { Activity } from 'lucide-react'

interface DataPoint {
  time: string
  utilization: number
  memory: number
  temperature: number
}

function generateTimeSeriesData(): DataPoint[] {
  const data: DataPoint[] = []
  const now = Date.now()
  for (let i = 60; i >= 0; i--) {
    const t = new Date(now - i * 60000)
    data.push({
      time: t.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }),
      utilization: 40 + Math.sin(i / 5) * 25 + Math.random() * 15,
      memory: 55 + Math.cos(i / 8) * 15 + Math.random() * 10,
      temperature: 60 + Math.sin(i / 6) * 12 + Math.random() * 5,
    })
  }
  return data
}

export function MetricsChart() {
  const [data, setData] = useState<DataPoint[]>(generateTimeSeriesData())

  useEffect(() => {
    const interval = setInterval(() => setData(generateTimeSeriesData()), 30000)
    return () => clearInterval(interval)
  }, [])

  return (
    <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl space-y-4 text-slate-100">
      <div className="flex items-center gap-2">
        <Activity className="w-5 h-5 text-orange-400" />
        <h3 className="text-base font-bold text-white">Cluster Telemetry Matrix (60m)</h3>
      </div>

      <div className="h-[320px] w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
            <defs>
              <linearGradient id="gradUtil" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#f97316" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#f97316" stopOpacity={0} />
              </linearGradient>
              <linearGradient id="gradMem" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#818cf8" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#818cf8" stopOpacity={0} />
              </linearGradient>
              <linearGradient id="gradTemp" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#fb7185" stopOpacity={0.2} />
                <stop offset="95%" stopColor="#fb7185" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis dataKey="time" stroke="#64748b" fontSize={11} tickLine={false} />
            <YAxis stroke="#64748b" fontSize={11} tickLine={false} domain={[0, 100]} />
            <Tooltip
              contentStyle={{
                backgroundColor: '#0f172a',
                borderColor: '#334155',
                borderRadius: '12px',
                fontSize: '12px',
                color: '#f8fafc',
              }}
            />
            <Legend wrapperStyle={{ fontSize: '12px', color: '#94a3b8' }} />
            <Area type="monotone" dataKey="utilization" name="GPU Compute %" stroke="#f97316" fill="url(#gradUtil)" strokeWidth={2} dot={false} />
            <Area type="monotone" dataKey="memory" name="VRAM Usage %" stroke="#818cf8" fill="url(#gradMem)" strokeWidth={2} dot={false} />
            <Area type="monotone" dataKey="temperature" name="Temp (°C)" stroke="#fb7185" fill="url(#gradTemp)" strokeWidth={1.5} dot={false} />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </Card>
  )
}
