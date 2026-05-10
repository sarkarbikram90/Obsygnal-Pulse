import { useQuery } from '@tanstack/react-query'
import { Card } from '../../components/ui/card'
import { Clock, Server, Activity } from 'lucide-react'
import { api } from '../../api/client'
import { formatDistanceToNow } from 'date-fns'

export default function Inefficiencies() {
  const { data: inefficiencies, isLoading, isError } = useQuery({
    queryKey: ['inefficiencies', 'default-cluster'],
    queryFn: () => api.getInefficiencies('default-cluster'),
    retry: false,
  })

  // Mock data for visual prototyping
  const mockIneffs = [
    {
      id: '1',
      type: 'idle_gpu',
      severity: 'critical',
      description: 'GPU has been completely idle with 0% utilization but memory is allocated.',
      node_id: 'worker-node-4',
      gpu_id: 'gpu-1',
      current_value: 0,
      threshold_value: 2,
      detected_at: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
      is_active: true,
    },
    {
      id: '2',
      type: 'cpu_starvation',
      severity: 'high',
      description: 'Node CPU is saturated at 100% while GPUs are waiting for data (20% util).',
      node_id: 'worker-node-2',
      current_value: 99.5,
      threshold_value: 90,
      detected_at: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
      is_active: true,
    },
    {
      id: '3',
      type: 'low_memory_usage',
      severity: 'medium',
      description: 'GPU compute is utilized but only 10% of VRAM is allocated. Batch size could be increased.',
      node_id: 'worker-node-1',
      gpu_id: 'gpu-3',
      current_value: 8,
      threshold_value: 15,
      detected_at: new Date(Date.now() - 1000 * 60 * 60 * 12).toISOString(),
      is_active: true,
    }
  ]

  const displayData = inefficiencies?.length ? inefficiencies : mockIneffs

  const getSeverityColor = (severity: string) => {
    switch(severity) {
      case 'critical': return 'text-rose-500 bg-rose-500/10 border-rose-500/20'
      case 'high': return 'text-orange-500 bg-orange-500/10 border-orange-500/20'
      case 'medium': return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20'
      default: return 'text-blue-500 bg-blue-500/10 border-blue-500/20'
    }
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-white mb-2">Detected Inefficiencies</h1>
        <p className="text-neutral-400">Raw mathematical inefficiencies identified by the rule engine.</p>
      </div>

      <div className="grid gap-4">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="bg-neutral-900/50 border-neutral-800 animate-pulse h-24" />
          ))
        ) : isError && !inefficiencies ? (
          <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-lg text-rose-400">
            Failed to load inefficiencies. Ensure the backend API is running.
          </div>
        ) : (
          displayData.map((ineff) => (
            <div key={ineff.id} className="flex items-center p-4 rounded-xl bg-neutral-900/50 border border-neutral-800 hover:bg-neutral-900 transition-colors">
              <div className="mr-6">
                <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${getSeverityColor(ineff.severity)} uppercase tracking-wider`}>
                  {ineff.severity}
                </span>
              </div>
              
              <div className="flex-1 min-w-0 pr-6 border-r border-neutral-800">
                <h3 className="text-base font-semibold text-white mb-1 truncate">{ineff.type.replace(/_/g, ' ').toUpperCase()}</h3>
                <p className="text-sm text-neutral-400 truncate">{ineff.description}</p>
              </div>

              <div className="w-48 px-6 border-r border-neutral-800 hidden md:block">
                <div className="flex items-center text-sm text-neutral-300 mb-1">
                  <Server className="w-4 h-4 mr-2 text-neutral-500" />
                  <span className="truncate">{ineff.node_id}</span>
                </div>
                {ineff.gpu_id && (
                  <div className="flex items-center text-sm text-neutral-300">
                    <Activity className="w-4 h-4 mr-2 text-indigo-400" />
                    <span>{ineff.gpu_id}</span>
                  </div>
                )}
              </div>

              <div className="pl-6 w-32 text-right hidden sm:block">
                <div className="flex items-center justify-end text-sm text-neutral-500 mb-1">
                  <Clock className="w-4 h-4 mr-1.5" />
                  {formatDistanceToNow(new Date(ineff.detected_at))} ago
                </div>
                <div className="text-xs text-neutral-600">
                  Val: <span className="text-neutral-300">{ineff.current_value.toFixed(1)}</span> / Thr: {ineff.threshold_value}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
