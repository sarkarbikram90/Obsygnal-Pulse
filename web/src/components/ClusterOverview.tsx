import type { ClusterOverview } from '../types'
import { Card } from './ui/card'
import { Activity, AlertTriangle, DollarSign, Cpu, CheckCircle2, HardDrive } from 'lucide-react'

interface Props {
  overview: ClusterOverview | null
}

export function ClusterOverviewPanel({ overview }: Props) {
  const cards = overview
    ? [
        { label: 'Total GPUs', value: `${overview.total_gpus} Sockets`, icon: Cpu, color: 'text-orange-400', sub: 'NVLink 4 Mesh' },
        { label: 'Active Compute', value: `${overview.active_gpus} GPUs`, icon: CheckCircle2, color: 'text-emerald-400', sub: `${overview.avg_utilization.toFixed(1)}% Avg Util` },
        { label: 'Idle GPUs', value: `${overview.idle_gpus} GPUs`, icon: AlertTriangle, color: 'text-rose-400', sub: '0% compute saturation' },
        { label: 'Compute Nodes', value: `${overview.healthy_nodes} / ${overview.total_nodes}`, icon: HardDrive, color: 'text-indigo-400', sub: 'Supermicro H100' },
        { label: 'Inefficiency Score', value: `${overview.inefficiency_score.toFixed(0)}`, icon: Activity, color: overview.inefficiency_score > 50 ? 'text-rose-400' : 'text-amber-400', sub: 'High compute waste' },
        { label: 'Monthly Waste Rate', value: `$${overview.estimated_waste_usd.toLocaleString()}`, icon: DollarSign, color: 'text-emerald-400', sub: 'Recoverable via auto-fix' },
      ]
    : []

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {cards.map((c, i) => {
        const Icon = c.icon
        return (
          <Card key={i} className="bg-slate-900 border-slate-800 p-5 rounded-2xl flex flex-col justify-between">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold uppercase tracking-wider text-slate-400">{c.label}</span>
              <Icon className={`w-4 h-4 ${c.color}`} />
            </div>
            <div className={`text-2xl font-black mt-2 ${c.color}`}>{c.value}</div>
            <p className="text-[11px] text-slate-500 mt-1">{c.sub}</p>
          </Card>
        )
      })}
    </div>
  )
}
