import { Card } from './ui/card'
import { Lightbulb, Zap, ArrowUpRight } from 'lucide-react'
import type { Recommendation } from '../types'

interface RecommendationPanelProps {
  recommendations?: Recommendation[]
  onSelect?: (rec: Recommendation) => void
}

const mockRecommendations: Recommendation[] = [
  {
    id: 'rec-1',
    category: 'reduce_idle_runtime',
    status: 'active',
    title: 'Isolate Hardware Fault (XID 31)',
    description: 'Cordon & drain node gpu-node-04 immediately to isolate GPU Memory Page Fault.',
    explanation: 'Hardware memory error detected on GPU 1. Immediate node isolation prevents cascading training failures.',
    confidence_score: 0.99,
    estimated_savings: { cost_per_hour_usd: 5.12, cost_per_month_usd: 3737.6, gpu_hours_per_month: 730 },
    affected_metrics: ['xid_error_code'],
    cluster_id: 'sim-cluster-01',
    node_id: 'gpu-node-04',
    gpu_id: 'gpu-node-04-gpu-1',
    inefficiency_id: '1',
    created_at: new Date().toISOString(),
    expires_at: new Date(Date.now() + 86400000).toISOString(),
  },
  {
    id: 'rec-2',
    category: 'right_size_memory',
    status: 'active',
    title: 'Defragment PyTorch CUDA VRAM Allocator',
    description: 'Set PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True to recover 44GB of fragmented VRAM.',
    explanation: 'PyTorch has reserved 68GB out of 80GB VRAM but only 24GB is actively used.',
    confidence_score: 0.92,
    estimated_savings: { cost_per_hour_usd: 2.4, cost_per_month_usd: 1752.0, gpu_hours_per_month: 210 },
    affected_metrics: ['cuda_memory_reserved'],
    cluster_id: 'sim-cluster-01',
    node_id: 'gpu-node-02',
    gpu_id: 'gpu-node-02-gpu-1',
    inefficiency_id: '2',
    created_at: new Date(Date.now() - 300000).toISOString(),
    expires_at: new Date(Date.now() + 86400000).toISOString(),
  },
]

export function RecommendationPanel({ recommendations, onSelect }: RecommendationPanelProps) {
  const data = recommendations && recommendations.length > 0 ? recommendations : mockRecommendations
  const totalMonthly = data.reduce((s, r) => s + (r.estimated_savings?.cost_per_month_usd || 0), 0)

  return (
    <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl space-y-4 text-slate-100">
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-2">
          <Lightbulb className="w-5 h-5 text-orange-400" />
          <h3 className="text-base font-bold text-white">Actionable Recommendations</h3>
        </div>
        <span className="text-xs font-bold text-emerald-400 bg-emerald-500/10 px-2.5 py-1 rounded-full border border-emerald-500/20">
          ${totalMonthly.toLocaleString()}/mo savings
        </span>
      </div>

      <div className="space-y-3 max-h-[380px] overflow-y-auto pr-1">
        {data.map((rec) => (
          <div
            key={rec.id}
            onClick={() => onSelect?.(rec)}
            className="p-4 rounded-xl bg-slate-950 border border-slate-800/80 hover:border-orange-500/40 transition-all cursor-pointer space-y-3 shadow-sm"
          >
            <div className="flex justify-between items-start">
              <h4 className="text-sm font-bold text-white">{rec.title}</h4>
              <span className="text-xs font-bold text-emerald-400 flex items-center">
                +${rec.estimated_savings.cost_per_month_usd.toLocaleString()}/mo
              </span>
            </div>

            <p className="text-xs text-slate-400 leading-relaxed">{rec.description}</p>

            <div className="flex justify-between items-center pt-2 border-t border-slate-800/60 text-xs">
              <div className="flex items-center gap-2">
                <span className="text-slate-500">Confidence:</span>
                <span className="font-bold text-slate-200">{Math.round(rec.confidence_score * 100)}%</span>
              </div>

              <span className="text-orange-400 font-bold flex items-center gap-1 hover:underline">
                <Zap className="w-3.5 h-3.5" /> 1-Click Auto Fix <ArrowUpRight className="w-3.5 h-3.5" />
              </span>
            </div>
          </div>
        ))}
      </div>
    </Card>
  )
}
