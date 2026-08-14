import { Card } from './ui/card'
import { AlertTriangle, Clock, Server, Activity } from 'lucide-react'
import type { Inefficiency } from '../types'
import { formatDistanceToNow } from 'date-fns'

interface InefficiencyPanelProps {
  inefficiencies?: Inefficiency[]
}

const mockInefficiencies: Inefficiency[] = [
  { id: '1', type: 'idle_gpu', severity: 'high', cluster_id: 'sim-cluster-01', node_id: 'gpu-node-02', gpu_id: 'gpu-node-02-gpu-3', description: 'GPU gpu-node-02-gpu-3 is idle (0% compute utilization, no active processes)', affected_metrics: ['gpu_utilization'], current_value: 0, threshold_value: 2, detected_at: new Date().toISOString(), is_active: true },
  { id: '2', type: 'xid_hardware_error', severity: 'critical', cluster_id: 'sim-cluster-01', node_id: 'gpu-node-04', gpu_id: 'gpu-node-04-gpu-1', description: 'NVIDIA Hardware Fault XID 31 (GPU Memory Page Fault) detected', affected_metrics: ['xid_error_code'], current_value: 31, threshold_value: 1, detected_at: new Date(Date.now() - 300000).toISOString(), is_active: true },
  { id: '3', type: 'cuda_memory_fragmentation', severity: 'high', cluster_id: 'sim-cluster-01', node_id: 'gpu-node-02', gpu_id: 'gpu-node-02-gpu-1', description: 'PyTorch CUDA memory fragmentation: reserved 68 GB vs used 24 GB', affected_metrics: ['cuda_memory_reserved'], current_value: 2.8, threshold_value: 2.0, detected_at: new Date(Date.now() - 120000).toISOString(), is_active: true },
]

export function InefficiencyPanel({ inefficiencies }: InefficiencyPanelProps) {
  const data = inefficiencies && inefficiencies.length > 0 ? inefficiencies : mockInefficiencies

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <span className="px-2.5 py-0.5 rounded-full text-[10px] font-black uppercase bg-rose-500 text-slate-950">Critical</span>
      case 'high':
        return <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase bg-orange-500/20 text-orange-400 border border-orange-500/30">High</span>
      default:
        return <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase bg-amber-500/20 text-amber-400 border border-amber-500/30">Medium</span>
    }
  }

  return (
    <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl space-y-4 text-slate-100">
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-2">
          <AlertTriangle className="w-5 h-5 text-orange-400" />
          <h3 className="text-base font-bold text-white">Active Inefficiencies</h3>
        </div>
        <span className="text-xs font-bold text-rose-400 bg-rose-500/10 px-2.5 py-1 rounded-full border border-rose-500/20">
          {data.length} Detected
        </span>
      </div>

      <div className="space-y-3 max-h-[380px] overflow-y-auto pr-1">
        {data.map((ineff) => (
          <div
            key={ineff.id}
            className={`p-4 rounded-xl border transition-all ${
              ineff.severity === 'critical'
                ? 'bg-rose-500/10 border-rose-500/30 text-rose-200'
                : 'bg-slate-950 border-slate-800/80 hover:border-orange-500/30 text-slate-200'
            }`}
          >
            <div className="flex justify-between items-start mb-2">
              <div className="flex items-center gap-2">
                {getSeverityBadge(ineff.severity)}
                <span className="text-xs font-bold text-white uppercase tracking-tight">
                  {ineff.type.replace(/_/g, ' ')}
                </span>
              </div>
              <div className="flex items-center text-[10px] text-slate-400 gap-1">
                <Clock className="w-3 h-3" />
                <span>{formatDistanceToNow(new Date(ineff.detected_at))} ago</span>
              </div>
            </div>

            <p className="text-xs text-slate-300 leading-relaxed mb-3">{ineff.description}</p>

            <div className="flex flex-wrap gap-4 text-[11px] text-slate-400 border-t border-slate-800/60 pt-2">
              {ineff.gpu_id && (
                <div className="flex items-center gap-1">
                  <Activity className="w-3.5 h-3.5 text-orange-400" />
                  <span>GPU: <strong className="text-slate-200">{ineff.gpu_id}</strong></span>
                </div>
              )}
              {ineff.node_id && (
                <div className="flex items-center gap-1">
                  <Server className="w-3.5 h-3.5 text-slate-400" />
                  <span>Node: <strong className="text-slate-200">{ineff.node_id}</strong></span>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </Card>
  )
}
