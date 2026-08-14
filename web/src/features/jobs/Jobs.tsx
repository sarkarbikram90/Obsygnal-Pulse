import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card } from '../../components/ui/card'
import { Cpu, RotateCcw, Search, Clock, DollarSign, Activity, Sparkles, CheckCircle2, AlertTriangle } from 'lucide-react'
import { api } from '../../api/client'
import { useToast } from '../../context/ToastContext'
import type { JobDetail } from '../../types'

export default function Jobs() {
  const queryClient = useQueryClient()
  const toast = useToast()
  const [searchTerm, setSearchTerm] = useState('')
  const [frameworkFilter, setFrameworkFilter] = useState<string>('all')

  const { data: jobs, isLoading } = useQuery({
    queryKey: ['jobs', 'sim-cluster-01'],
    queryFn: () => api.getJobs('sim-cluster-01'),
  })

  const restartMutation = useMutation({
    mutationFn: (jobId: string) => api.restartJob(jobId),
    onSuccess: (_, jobId) => {
      toast.success('Workload Restarted', `Job ${jobId} successfully restarted from latest saved state checkpoint.`)
      queryClient.invalidateQueries({ queryKey: ['jobs'] })
    },
  })

  const displayJobs = (jobs || []).filter((job) => {
    const matchesSearch = job.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      job.id.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesFramework = frameworkFilter === 'all' || job.framework.toLowerCase().includes(frameworkFilter.toLowerCase())
    return matchesSearch && matchesFramework
  })

  const totalGpusAllocated = (jobs || []).reduce((acc, j) => acc + j.gpu_count, 0)
  const runningCount = (jobs || []).filter((j) => j.status === 'running').length
  const bottleneckCount = (jobs || []).filter((j) => j.status === 'cpu_bottleneck' || j.status === 'stalled').length

  const getStatusBadge = (status: JobDetail['status']) => {
    switch (status) {
      case 'running':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 uppercase">
            <CheckCircle2 className="w-3.5 h-3.5" /> Optimal
          </span>
        )
      case 'cpu_bottleneck':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold bg-orange-500/10 text-orange-400 border border-orange-500/30 uppercase">
            <AlertTriangle className="w-3.5 h-3.5" /> CPU Starvation
          </span>
        )
      case 'stalled':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold bg-rose-500/10 text-rose-400 border border-rose-500/30 uppercase">
            <AlertTriangle className="w-3.5 h-3.5" /> Hung Deadlock
          </span>
        )
      default:
        return null
    }
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500 text-slate-100">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-orange-400 to-rose-400 bg-clip-text text-transparent mb-2">
            AI Training & Inference Workloads
          </h1>
          <p className="text-slate-400">Live PyTorch DDP, Ray, vLLM, and DeepSpeed job profiling, memory fragmentation, and rank efficiency.</p>
        </div>
      </div>

      {/* Top Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Active Workloads</span>
            <Activity className="w-4 h-4 text-orange-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">{jobs?.length || 0} Workloads</div>
          <p className="text-[11px] text-slate-500 mt-1">Multi-node training & serving pods</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Allocated GPU Sockets</span>
            <Cpu className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-emerald-300 mt-2">{totalGpusAllocated} GPUs</div>
          <p className="text-[11px] text-slate-500 mt-1">Distributed across 4 worker nodes</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Optimal Throughput</span>
            <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">{runningCount} / {jobs?.length || 0}</div>
          <p className="text-[11px] text-emerald-400 mt-1">Operating at &gt;75% GPU compute saturation</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Compute Bottlenecks</span>
            <AlertTriangle className="w-4 h-4 text-rose-400" />
          </div>
          <div className="text-2xl font-bold text-rose-400 mt-2">{bottleneckCount} Stalled/Starving</div>
          <p className="text-[11px] text-rose-400/80 mt-1">Hung training rank / CPU starvation</p>
        </Card>
      </div>

      {/* Search & Framework Filter */}
      <div className="flex flex-col sm:flex-row gap-4 justify-between items-center bg-slate-900/60 p-4 border border-slate-800 rounded-2xl">
        <div className="relative w-full sm:w-80">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search job ID, model, or namespace..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-orange-500/50"
          />
        </div>

        <div className="flex flex-wrap gap-1.5 w-full sm:w-auto">
          {['all', 'pytorch', 'deepspeed', 'vllm', 'ray'].map((fw) => (
            <button
              key={fw}
              onClick={() => setFrameworkFilter(fw)}
              className={`px-3 py-1.5 text-xs font-bold rounded-xl uppercase transition ${
                frameworkFilter === fw
                  ? 'bg-orange-500 text-white shadow-md shadow-orange-500/20'
                  : 'bg-slate-950 border border-slate-800 text-slate-400 hover:text-white'
              }`}
            >
              {fw === 'all' ? 'All Frameworks' : fw}
            </button>
          ))}
        </div>
      </div>

      {/* Jobs List */}
      <div className="grid gap-4">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="bg-slate-900 border-slate-800 p-6 rounded-2xl h-44 animate-pulse" />
          ))
        ) : displayJobs.length === 0 ? (
          <div className="p-12 text-center border border-dashed border-slate-800 rounded-2xl bg-slate-900/30 text-slate-400 text-sm">
            No active workloads match your search filters.
          </div>
        ) : (
          displayJobs.map((job) => (
            <Card
              key={job.id}
              className="bg-slate-900 border-slate-800 hover:border-orange-500/30 p-6 rounded-2xl transition-all shadow-md flex flex-col lg:flex-row lg:items-center justify-between gap-6"
            >
              {/* Left Column: Job Info */}
              <div className="space-y-3 lg:w-80">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <h3 className="text-base font-bold text-white tracking-tight">{job.name}</h3>
                    {getStatusBadge(job.status)}
                  </div>
                  <div className="flex items-center gap-2 text-xs text-slate-400">
                    <span className="font-mono text-[11px] text-slate-500">{job.id}</span>
                    <span>•</span>
                    <span className="px-2 py-0.5 bg-slate-950 border border-slate-800 rounded text-[10px] text-slate-300">
                      {job.namespace}
                    </span>
                  </div>
                </div>

                <div className="flex items-center gap-2 text-xs text-orange-400 font-semibold">
                  <Sparkles className="w-3.5 h-3.5" />
                  <span>{job.framework}</span>
                </div>
              </div>

              {/* Middle Column: Resource Efficiency Bars */}
              <div className="flex-1 grid grid-cols-1 sm:grid-cols-2 gap-4 bg-slate-950/60 p-4 rounded-xl border border-slate-800/80">
                <div>
                  <div className="flex justify-between text-xs text-slate-400 mb-1">
                    <span>Average GPU Utilization</span>
                    <span className="font-bold text-white">{job.avg_gpu_util}%</span>
                  </div>
                  <div className="w-full h-2 bg-slate-900 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${
                        job.avg_gpu_util > 70
                          ? 'bg-emerald-400'
                          : job.avg_gpu_util < 20
                          ? 'bg-rose-500'
                          : 'bg-orange-400'
                      }`}
                      style={{ width: `${job.avg_gpu_util}%` }}
                    />
                  </div>
                  <p className="text-[10px] text-slate-500 mt-1">{job.gpu_count} GPUs allocated</p>
                </div>

                <div>
                  <div className="flex justify-between text-xs text-slate-400 mb-1">
                    <span>VRAM Allocated vs Reserved</span>
                    <span className="font-bold text-purple-300">
                      {job.vram_allocated_gb}GB / {job.vram_reserved_gb}GB
                    </span>
                  </div>
                  <div className="w-full h-2 bg-slate-900 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-purple-500 rounded-full"
                      style={{ width: `${(job.vram_allocated_gb / job.vram_reserved_gb) * 100}%` }}
                    />
                  </div>
                  <p className="text-[10px] text-slate-500 mt-1">
                    {job.vram_reserved_gb - job.vram_allocated_gb > 20
                      ? `⚠️ ${job.vram_reserved_gb - job.vram_allocated_gb}GB VRAM Fragmented`
                      : 'Efficient memory packing'}
                  </p>
                </div>
              </div>

              {/* Right Column: Run Duration, Cost, & Actions */}
              <div className="lg:w-64 flex flex-col justify-between space-y-3 lg:border-l lg:border-slate-800 lg:pl-6">
                <div className="flex justify-between items-center text-xs">
                  <div className="flex items-center text-slate-400">
                    <Clock className="w-3.5 h-3.5 mr-1" />
                    <span>Run: <strong className="text-slate-200">{job.duration}</strong></span>
                  </div>
                  <div className="flex items-center text-emerald-400 font-bold">
                    <DollarSign className="w-3.5 h-3.5 mr-0.5" />
                    <span>{job.estimated_cost}</span>
                  </div>
                </div>

                <div className="flex gap-2">
                  <button
                    onClick={() => {
                      toast.action(
                        'Workload Config Downloaded',
                        `Exported PyTorch runtime configuration for ${job.name}.`,
                        'View Spec',
                        () => {}
                      )
                    }}
                    className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-xl transition"
                  >
                    Inspect Ranks
                  </button>

                  {(job.status === 'stalled' || job.status === 'cpu_bottleneck') && (
                    <button
                      onClick={() => restartMutation.mutate(job.id)}
                      className="px-3 py-2 bg-orange-500 hover:bg-orange-600 text-white text-xs font-bold rounded-xl transition shadow-lg shadow-orange-500/20 flex items-center gap-1"
                      title="Restart from latest saved checkpoint"
                    >
                      <RotateCcw className="w-3.5 h-3.5" /> Restart
                    </button>
                  )}
                </div>
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}
