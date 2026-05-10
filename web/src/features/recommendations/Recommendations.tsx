import { useQuery } from '@tanstack/react-query'
import { Card } from '../../components/ui/card'
import { Button } from '../../components/ui/button'
import { Lightbulb, DollarSign, CheckCircle2 } from 'lucide-react'
import { api } from '../../api/client'

export default function Recommendations() {
  const { data: recommendations, isLoading, isError } = useQuery({
    queryKey: ['recommendations', 'default-cluster'],
    queryFn: () => api.getRecommendations('default-cluster'),
    retry: false,
  })

  // For visual prototyping when API is empty
  const mockRecs = [
    {
      id: '1',
      category: 'reduce_gpu_allocation',
      title: 'Downsize underutilized GPU instances',
      description: 'Nodes worker-4 and worker-5 have < 15% GPU utilization over the last 72 hours.',
      explanation: 'These nodes are likely over-provisioned for their current workloads. Consider migrating these workloads to smaller GPU nodes (e.g. T4 instead of A100).',
      confidence_score: 0.92,
      estimated_savings: { cost_per_month_usd: 1450.50 },
      status: 'active'
    },
    {
      id: '2',
      category: 'increase_batch_size',
      title: 'Increase batch size for Training Job #892',
      description: 'GPU compute is underutilized while memory has headroom.',
      explanation: 'Training job #892 on node-2 is running with low GPU utilization (32%) but only using 40% of VRAM. Increasing batch size will improve throughput and efficiency.',
      confidence_score: 0.85,
      estimated_savings: { cost_per_month_usd: 800.00 },
      status: 'active'
    }
  ]

  const displayRecs = recommendations?.length ? recommendations : mockRecs

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div className="flex justify-between items-end">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white mb-2">Recommendation Center</h1>
          <p className="text-neutral-400">Actionable insights to optimize your infrastructure and reduce costs.</p>
        </div>
        <div className="flex gap-4">
          <Card className="bg-emerald-500/10 border-emerald-500/20 py-2 px-6">
            <div className="flex items-center space-x-2">
              <DollarSign className="h-5 w-5 text-emerald-400" />
              <div>
                <p className="text-sm font-medium text-emerald-400">Potential Savings</p>
                <p className="text-2xl font-bold text-emerald-300">
                  ${displayRecs.reduce((acc, r) => acc + (r.estimated_savings?.cost_per_month_usd || 0), 0).toLocaleString()} <span className="text-sm font-normal">/mo</span>
                </p>
              </div>
            </div>
          </Card>
        </div>
      </div>

      <div className="grid gap-6">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="bg-neutral-900/50 border-neutral-800 animate-pulse h-48" />
          ))
        ) : isError && !recommendations ? (
          <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-lg text-rose-400">
            Failed to load recommendations. Ensure the backend API is running.
          </div>
        ) : displayRecs.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center border border-dashed border-neutral-800 rounded-xl bg-neutral-900/20">
            <CheckCircle2 className="h-12 w-12 text-emerald-500 mb-4" />
            <h3 className="text-lg font-medium text-white">All Optimized!</h3>
            <p className="text-neutral-400 mt-2 max-w-md">Your infrastructure is running efficiently. We haven't detected any significant optimization opportunities at this time.</p>
          </div>
        ) : (
          displayRecs.map((rec) => (
            <Card key={rec.id} className="bg-neutral-900/50 border-neutral-800 overflow-hidden hover:border-indigo-500/50 transition-colors group">
              <div className="flex flex-col md:flex-row">
                <div className="flex-1 p-6">
                  <div className="flex items-center space-x-3 mb-3">
                    <div className="p-2 bg-indigo-500/10 rounded-lg text-indigo-400">
                      <Lightbulb className="h-5 w-5" />
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-white group-hover:text-indigo-400 transition-colors">{rec.title}</h3>
                      <p className="text-sm text-neutral-400">{rec.description}</p>
                    </div>
                  </div>
                  
                  <div className="mt-4 bg-neutral-950/50 p-4 rounded-lg border border-neutral-800/50">
                    <p className="text-sm text-neutral-300 leading-relaxed">{rec.explanation}</p>
                  </div>
                </div>
                
                <div className="md:w-72 bg-neutral-950/30 p-6 border-t md:border-t-0 md:border-l border-neutral-800 flex flex-col justify-between">
                  <div className="space-y-4">
                    <div>
                      <p className="text-xs text-neutral-500 font-medium uppercase tracking-wider mb-1">Confidence</p>
                      <div className="flex items-center">
                        <div className="flex-1 h-2 bg-neutral-800 rounded-full overflow-hidden mr-3">
                          <div 
                            className="h-full bg-emerald-500 rounded-full" 
                            style={{ width: `${(rec.confidence_score * 100)}%` }}
                          />
                        </div>
                        <span className="text-sm font-medium text-white">{Math.round(rec.confidence_score * 100)}%</span>
                      </div>
                    </div>
                    
                    <div>
                      <p className="text-xs text-neutral-500 font-medium uppercase tracking-wider mb-1">Est. Savings</p>
                      <p className="text-2xl font-bold text-emerald-400">
                        ${rec.estimated_savings.cost_per_month_usd.toLocaleString()} <span className="text-sm text-emerald-500/70 font-normal">/mo</span>
                      </p>
                    </div>
                  </div>
                  
                  <div className="flex gap-3 mt-6">
                    <Button variant="outline" className="flex-1 border-neutral-700 hover:bg-neutral-800">Dismiss</Button>
                    <Button className="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white">Apply</Button>
                  </div>
                </div>
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}
