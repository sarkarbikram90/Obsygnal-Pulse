import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card } from '../../components/ui/card'
import { Clock, Server, Activity, ShieldAlert, Cpu, Search, AlertTriangle } from 'lucide-react'
import { api } from '../../api/client'
import { formatDistanceToNow } from 'date-fns'

export default function Inefficiencies() {
  const [searchTerm, setSearchTerm] = useState('')
  const [severityFilter, setSeverityFilter] = useState<string>('all')

  const { data: inefficiencies, isLoading, isError } = useQuery({
    queryKey: ['inefficiencies', 'sim-cluster-01'],
    queryFn: () => api.getInefficiencies('sim-cluster-01'),
    retry: false,
  })

  const displayData = (inefficiencies || []).filter((ineff) => {
    const matchesSearch = ineff.type.toLowerCase().includes(searchTerm.toLowerCase()) ||
      ineff.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (ineff.node_id && ineff.node_id.toLowerCase().includes(searchTerm.toLowerCase()))
    const matchesSeverity = severityFilter === 'all' || ineff.severity === severityFilter
    return matchesSearch && matchesSeverity
  })

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'text-rose-400 bg-rose-500/10 border-rose-500/30 font-bold'
      case 'high':
        return 'text-orange-400 bg-orange-500/10 border-orange-500/30'
      case 'medium':
        return 'text-amber-400 bg-amber-500/10 border-amber-500/30'
      default:
        return 'text-slate-400 bg-slate-800 border-slate-700'
    }
  }

  const getTypeBadge = (type: string) => {
    switch (type) {
      case 'xid_hardware_error':
        return <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] uppercase font-extrabold bg-rose-500 text-slate-950"><ShieldAlert className="w-3 h-3" /> Hardware Fault</span>
      case 'cuda_memory_fragmentation':
        return <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] uppercase font-extrabold bg-purple-500/20 text-purple-300 border border-purple-500/40"><Cpu className="w-3 h-3" /> PyTorch VRAM Frag</span>
      case 'cpu_starvation':
        return <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] uppercase font-extrabold bg-orange-500/20 text-orange-300 border border-orange-500/40"><AlertTriangle className="w-3 h-3" /> CPU Starvation</span>
      default:
        return null
    }
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500 text-slate-100">
      <div>
        <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-orange-400 to-rose-400 bg-clip-text text-transparent mb-2">
          GPU Hardware & Compute Bottlenecks
        </h1>
        <p className="text-slate-400">Real-time hardware fault indicators, PyTorch memory fragmentation, and distributed communication skew.</p>
      </div>

      {/* Search & Severity Filter Controls */}
      <div className="flex flex-col sm:flex-row gap-4 justify-between items-center bg-slate-900/60 p-4 border border-slate-800 rounded-2xl">
        <div className="relative w-full sm:w-80">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search bottleneck, rule, or node..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-orange-500/50"
          />
        </div>

        <div className="flex gap-1.5 w-full sm:w-auto">
          {['all', 'critical', 'high', 'medium'].map((sev) => (
            <button
              key={sev}
              onClick={() => setSeverityFilter(sev)}
              className={`px-3 py-1.5 text-xs font-bold rounded-xl uppercase transition ${
                severityFilter === sev
                  ? 'bg-orange-500 text-white shadow-md shadow-orange-500/20'
                  : 'bg-slate-950 border border-slate-800 text-slate-400 hover:text-white'
              }`}
            >
              {sev}
            </button>
          ))}
        </div>
      </div>

      <div className="grid gap-4">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="bg-slate-900/50 border-slate-800 animate-pulse h-24 rounded-2xl" />
          ))
        ) : isError && !inefficiencies ? (
          <div className="p-4 bg-rose-500/10 border border-rose-500/30 rounded-xl text-rose-400 text-sm">
            Failed to load inefficiencies. Ensure the backend server is reachable.
          </div>
        ) : displayData.length === 0 ? (
          <div className="p-12 text-center border border-dashed border-slate-800 rounded-2xl bg-slate-900/30 text-slate-400 text-sm">
            No active bottlenecks found matching filter criteria.
          </div>
        ) : (
          displayData.map((ineff) => (
            <div key={ineff.id} className="flex items-center p-4 rounded-2xl bg-slate-900 border border-slate-800 hover:border-orange-500/30 transition-all shadow-md">
              <div className="mr-4">
                <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold border ${getSeverityColor(ineff.severity)} uppercase tracking-wider`}>
                  {ineff.severity}
                </span>
              </div>

              <div className="flex-1 min-w-0 pr-6 border-r border-slate-800/80">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="text-sm font-bold text-white uppercase tracking-tight">{ineff.type.replace(/_/g, ' ')}</h3>
                  {getTypeBadge(ineff.type)}
                </div>
                <p className="text-xs text-slate-400 truncate">{ineff.description}</p>
              </div>

              <div className="w-48 px-6 border-r border-slate-800/80 hidden md:block">
                <div className="flex items-center text-xs font-semibold text-slate-300 mb-1">
                  <Server className="w-3.5 h-3.5 mr-2 text-slate-500" />
                  <span className="truncate">{ineff.node_id}</span>
                </div>
                {ineff.gpu_id && (
                  <div className="flex items-center text-xs text-slate-400">
                    <Activity className="w-3.5 h-3.5 mr-2 text-orange-400" />
                    <span className="truncate">{ineff.gpu_id}</span>
                  </div>
                )}
              </div>

              <div className="pl-6 w-36 text-right hidden sm:block">
                <div className="flex items-center justify-end text-xs text-slate-400 mb-1">
                  <Clock className="w-3.5 h-3.5 mr-1" />
                  {formatDistanceToNow(new Date(ineff.detected_at))} ago
                </div>
                <div className="text-[11px] text-slate-500 font-mono">
                  Val: <span className="text-slate-200 font-bold">{ineff.current_value.toFixed(1)}</span> / Thr: {ineff.threshold_value}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
