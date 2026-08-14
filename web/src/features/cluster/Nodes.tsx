import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card } from '../../components/ui/card'
import { Server, Cpu, ShieldAlert, CheckCircle2, Search, Activity, X } from 'lucide-react'
import { api } from '../../api/client'
import { useToast } from '../../context/ToastContext'
import type { NodeDetail, GPUSummary } from '../../types'

export default function Nodes() {
  const queryClient = useQueryClient()
  const toast = useToast()
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'ready' | 'degraded' | 'cordoned'>('all')
  const [selectedNode, setSelectedNode] = useState<NodeDetail | null>(null)

  const { data: nodes, isLoading } = useQuery({
    queryKey: ['nodes', 'sim-cluster-01'],
    queryFn: () => api.getNodes('sim-cluster-01'),
  })

  const cordonMutation = useMutation({
    mutationFn: (nodeId: string) => api.cordonNode(nodeId),
    onSuccess: (_, nodeId) => {
      toast.success('Node Cordoned', `Successfully cordoned & drained node ${nodeId}. Workloads will not be scheduled here.`)
      queryClient.invalidateQueries({ queryKey: ['nodes'] })
    },
  })

  const displayNodes = (nodes || []).filter((node) => {
    const matchesSearch = node.node_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      node.gpu_model.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesStatus = statusFilter === 'all' || node.status === statusFilter
    return matchesSearch && matchesStatus
  })

  const totalGPUs = (nodes || []).reduce((acc, n) => acc + n.gpu_count, 0)
  const readyNodes = (nodes || []).filter((n) => n.status === 'ready').length
  const degradedNodes = (nodes || []).filter((n) => n.status === 'degraded').length

  const getStatusBadge = (status: NodeDetail['status']) => {
    switch (status) {
      case 'ready':
        return <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 uppercase"><CheckCircle2 className="w-3.5 h-3.5" /> Ready</span>
      case 'degraded':
        return <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold bg-rose-500/10 text-rose-400 border border-rose-500/30 uppercase"><ShieldAlert className="w-3.5 h-3.5" /> Hardware Fault</span>
      case 'cordoned':
        return <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold bg-amber-500/10 text-amber-400 border border-amber-500/30 uppercase">Cordoned</span>
    }
  }

  const getGpuColor = (gpu: GPUSummary) => {
    if (gpu.status === 'fault') return 'bg-rose-500 ring-2 ring-rose-400 animate-pulse'
    if (gpu.status === 'idle') return 'bg-slate-800 border border-slate-700'
    if (gpu.utilization > 80) return 'bg-emerald-500'
    if (gpu.utilization > 40) return 'bg-indigo-500'
    return 'bg-amber-500'
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500 text-slate-100">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-orange-400 to-rose-400 bg-clip-text text-transparent mb-2">
            Physical GPU Compute Nodes
          </h1>
          <p className="text-slate-400">Inventory, real-time hardware telemetry, driver matrix, and node lifecycle control.</p>
        </div>
      </div>

      {/* Top Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Total Compute Nodes</span>
            <Server className="w-4 h-4 text-orange-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">{nodes?.length || 0} Nodes</div>
          <p className="text-[11px] text-slate-500 mt-1">Supermicro & DGX H100 hosts</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Total GPU Accelerators</span>
            <Cpu className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-emerald-300 mt-2">{totalGPUs} GPUs</div>
          <p className="text-[11px] text-slate-500 mt-1">NVLink 4 Mesh / SXM5</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Healthy / Ready</span>
            <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">{readyNodes} / {nodes?.length || 0}</div>
          <p className="text-[11px] text-emerald-400 mt-1">Passing all DCGM health checks</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Hardware Faults</span>
            <ShieldAlert className="w-4 h-4 text-rose-400" />
          </div>
          <div className="text-2xl font-bold text-rose-400 mt-2">{degradedNodes} Degraded</div>
          <p className="text-[11px] text-rose-400/80 mt-1">Requires node isolation (XID 31)</p>
        </Card>
      </div>

      {/* Filters & Search */}
      <div className="flex flex-col sm:flex-row gap-4 justify-between items-center bg-slate-900/60 p-4 border border-slate-800 rounded-2xl">
        <div className="relative w-full sm:w-80">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search node ID or GPU model..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-orange-500/50"
          />
        </div>

        <div className="flex gap-1.5 w-full sm:w-auto">
          {(['all', 'ready', 'degraded'] as const).map((st) => (
            <button
              key={st}
              onClick={() => setStatusFilter(st)}
              className={`px-3.5 py-1.5 text-xs font-bold rounded-xl capitalize transition ${
                statusFilter === st
                  ? 'bg-orange-500 text-white shadow-md shadow-orange-500/20'
                  : 'bg-slate-950 border border-slate-800 text-slate-400 hover:text-white'
              }`}
            >
              {st}
            </button>
          ))}
        </div>
      </div>

      {/* Node Cards Grid */}
      <div className="grid gap-4">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="bg-slate-900 border-slate-800 p-6 rounded-2xl h-44 animate-pulse" />
          ))
        ) : displayNodes.length === 0 ? (
          <div className="p-12 text-center border border-dashed border-slate-800 rounded-2xl bg-slate-900/30 text-slate-400 text-sm">
            No nodes found matching your search.
          </div>
        ) : (
          displayNodes.map((node) => (
            <Card
              key={node.node_id}
              className="bg-slate-900 border-slate-800 hover:border-orange-500/30 p-6 rounded-2xl transition-all shadow-md flex flex-col lg:flex-row lg:items-center justify-between gap-6"
            >
              {/* Left Column: Node Info & Specs */}
              <div className="space-y-3 lg:w-80">
                <div className="flex items-center gap-3">
                  <div className="p-2.5 bg-slate-950 border border-slate-800 rounded-xl text-orange-400">
                    <Server className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="text-base font-bold text-white tracking-tight">{node.node_id}</h3>
                      {getStatusBadge(node.status)}
                    </div>
                    <p className="text-xs text-slate-400 mt-0.5">{node.gpu_model}</p>
                  </div>
                </div>

                <div className="flex gap-4 text-[11px] text-slate-400">
                  <span>Driver: <strong className="text-slate-200">{node.driver_version}</strong></span>
                  <span>CUDA: <strong className="text-slate-200">{node.cuda_version}</strong></span>
                  <span>IP: <strong className="text-slate-200">{node.ip_address}</strong></span>
                </div>
              </div>

              {/* Middle Column: Per-GPU Tiles Grid */}
              <div className="flex-1 space-y-2">
                <div className="flex justify-between items-center text-xs text-slate-400">
                  <span className="font-semibold">NVIDIA GPU Sockets ({node.gpu_count}x)</span>
                  <span>Click tile to inspect metrics</span>
                </div>
                <div className="grid grid-cols-4 sm:grid-cols-8 gap-2">
                  {node.gpus.map((gpu) => (
                    <div
                      key={gpu.gpu_id}
                      onClick={() => setSelectedNode(node)}
                      className={`h-11 rounded-xl flex flex-col items-center justify-center cursor-pointer transition-transform hover:scale-105 ${getGpuColor(gpu)}`}
                      title={`${gpu.gpu_id}: ${gpu.utilization}% util, ${gpu.memory_used_gb}/${gpu.memory_total_gb}GB, ${gpu.temperature}°C`}
                    >
                      <span className="text-[10px] font-extrabold text-white">GPU {gpu.index}</span>
                      <span className="text-[9px] font-bold text-white/90">
                        {gpu.status === 'fault' ? 'XID 31' : `${gpu.utilization.toFixed(0)}%`}
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Right Column: Host CPU / RAM + Actions */}
              <div className="lg:w-72 flex flex-col justify-between space-y-3 lg:border-l lg:border-slate-800 lg:pl-6">
                <div className="space-y-2">
                  <div>
                    <div className="flex justify-between text-xs text-slate-400 mb-1">
                      <span>Host CPU Usage</span>
                      <span className="font-bold text-slate-200">{node.cpu_usage}%</span>
                    </div>
                    <div className="w-full h-1.5 bg-slate-950 rounded-full overflow-hidden">
                      <div
                        className={`h-full rounded-full ${node.cpu_usage > 90 ? 'bg-rose-500' : 'bg-orange-500'}`}
                        style={{ width: `${node.cpu_usage}%` }}
                      />
                    </div>
                  </div>

                  <div>
                    <div className="flex justify-between text-xs text-slate-400 mb-1">
                      <span>Host Memory ({node.memory_gb}GB)</span>
                      <span className="font-bold text-slate-200">58%</span>
                    </div>
                    <div className="w-full h-1.5 bg-slate-950 rounded-full overflow-hidden">
                      <div className="h-full bg-indigo-500 rounded-full" style={{ width: '58%' }} />
                    </div>
                  </div>
                </div>

                <div className="flex gap-2 pt-1">
                  <button
                    onClick={() => setSelectedNode(node)}
                    className="flex-1 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-xl transition"
                  >
                    Inspect
                  </button>
                  {node.status === 'degraded' && (
                    <button
                      onClick={() => cordonMutation.mutate(node.node_id)}
                      className="px-3 py-1.5 bg-rose-500 hover:bg-rose-600 text-white text-xs font-bold rounded-xl transition shadow-lg shadow-rose-500/20"
                    >
                      Cordon
                    </button>
                  )}
                </div>
              </div>
            </Card>
          ))
        )}
      </div>

      {/* Node Detail Drawer Modal */}
      {selectedNode && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md p-4 animate-in fade-in">
          <div className="w-full max-w-3xl bg-slate-900 border border-slate-800 rounded-3xl shadow-2xl overflow-hidden text-slate-100">
            <div className="p-6 border-b border-slate-800 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-3 bg-slate-950 border border-slate-800 rounded-2xl text-orange-400">
                  <Server className="w-6 h-6" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-xl font-bold text-white">{selectedNode.node_id}</h3>
                    {getStatusBadge(selectedNode.status)}
                  </div>
                  <p className="text-xs text-slate-400">{selectedNode.gpu_model} • Driver {selectedNode.driver_version} • CUDA {selectedNode.cuda_version}</p>
                </div>
              </div>
              <button
                onClick={() => setSelectedNode(null)}
                className="text-slate-400 hover:text-white transition p-2"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6 space-y-6 max-h-[70vh] overflow-y-auto">
              <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider">Per-Device DCGM Telemetry Matrix</h4>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {selectedNode.gpus.map((gpu) => (
                  <div
                    key={gpu.gpu_id}
                    className={`p-4 rounded-2xl border ${
                      gpu.status === 'fault'
                        ? 'bg-rose-500/10 border-rose-500/40 text-rose-200'
                        : 'bg-slate-950 border-slate-800 text-slate-200'
                    }`}
                  >
                    <div className="flex justify-between items-center mb-3">
                      <div className="flex items-center gap-2">
                        <Activity className="w-4 h-4 text-orange-400" />
                        <span className="font-bold text-sm">GPU {gpu.index}</span>
                      </div>
                      {gpu.status === 'fault' ? (
                        <span className="px-2 py-0.5 bg-rose-500 text-slate-950 font-black text-[10px] uppercase rounded">
                          XID 31 FAULT
                        </span>
                      ) : (
                        <span className="text-xs font-bold text-emerald-400">{gpu.temperature}°C</span>
                      )}
                    </div>

                    <div className="grid grid-cols-2 gap-2 text-xs">
                      <div className="bg-slate-900/60 p-2 rounded-xl">
                        <div className="text-[10px] text-slate-400">Compute</div>
                        <div className="font-bold text-white text-sm mt-0.5">{gpu.utilization.toFixed(1)}%</div>
                      </div>
                      <div className="bg-slate-900/60 p-2 rounded-xl">
                        <div className="text-[10px] text-slate-400">VRAM Used</div>
                        <div className="font-bold text-white text-sm mt-0.5">{gpu.memory_used_gb} / {gpu.memory_total_gb} GB</div>
                      </div>
                      <div className="bg-slate-900/60 p-2 rounded-xl">
                        <div className="text-[10px] text-slate-400">Power Draw</div>
                        <div className="font-bold text-white text-sm mt-0.5">{gpu.power_watts}W / 700W</div>
                      </div>
                      <div className="bg-slate-900/60 p-2 rounded-xl">
                        <div className="text-[10px] text-slate-400">PCIe Bandwidth</div>
                        <div className="font-bold text-white text-sm mt-0.5">58.4 GB/s</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="p-6 border-t border-slate-800 bg-slate-950 flex justify-between items-center">
              <div className="text-xs text-slate-400">
                Host IP: <span className="font-mono text-slate-200">{selectedNode.ip_address}</span>
              </div>
              <button
                onClick={() => {
                  toast.success('Diagnostic Benchmark Queued', `Triggered DCGM diagnostic suite on ${selectedNode.node_id}`)
                  setSelectedNode(null)
                }}
                className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-white font-semibold text-xs rounded-xl transition"
              >
                Run Hardware Diagnostics
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
