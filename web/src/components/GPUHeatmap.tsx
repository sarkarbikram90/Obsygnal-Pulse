import { useState } from 'react'
import { Card } from './ui/card'
import { Cpu } from 'lucide-react'

export interface GPUHeatmapCell {
  gpu_id: string
  node_id: string
  index: number
  utilization: number
  memory_used_gb: number
  memory_total_gb: number
  temperature: number
  power_watts: number
  status: 'active' | 'idle' | 'hot' | 'fault'
  xid_error?: number
}

interface GPUHeatmapProps {
  onSelectGPU?: (gpu: GPUHeatmapCell) => void
}

export function GPUHeatmap({ onSelectGPU }: GPUHeatmapProps) {
  const [selectedCell, setSelectedCell] = useState<GPUHeatmapCell | null>(null)
  const [filter, setFilter] = useState<'all' | 'active' | 'idle' | 'fault'>('all')

  const nodes = ['gpu-node-01', 'gpu-node-02', 'gpu-node-03', 'gpu-node-04']
  
  // Realistically modeled GPU cluster cells
  const cells: GPUHeatmapCell[] = [
    // gpu-node-01: Mix of active and 1 idle
    { gpu_id: 'gpu-node-01-gpu-0', node_id: 'gpu-node-01', index: 0, utilization: 84.5, memory_used_gb: 62.4, memory_total_gb: 80, temperature: 68, power_watts: 510, status: 'active' },
    { gpu_id: 'gpu-node-01-gpu-1', node_id: 'gpu-node-01', index: 1, utilization: 79.2, memory_used_gb: 58.0, memory_total_gb: 80, temperature: 65, power_watts: 480, status: 'active' },
    { gpu_id: 'gpu-node-01-gpu-2', node_id: 'gpu-node-01', index: 2, utilization: 88.0, memory_used_gb: 64.0, memory_total_gb: 80, temperature: 71, power_watts: 530, status: 'active' },
    { gpu_id: 'gpu-node-01-gpu-3', node_id: 'gpu-node-01', index: 3, utilization: 0.0, memory_used_gb: 0.4, memory_total_gb: 80, temperature: 38, power_watts: 110, status: 'idle' },
    { gpu_id: 'gpu-node-01-gpu-4', node_id: 'gpu-node-01', index: 4, utilization: 76.4, memory_used_gb: 54.2, memory_total_gb: 80, temperature: 64, power_watts: 470, status: 'active' },
    { gpu_id: 'gpu-node-01-gpu-5', node_id: 'gpu-node-01', index: 5, utilization: 82.1, memory_used_gb: 60.1, memory_total_gb: 80, temperature: 67, power_watts: 495, status: 'active' },
    { gpu_id: 'gpu-node-01-gpu-6', node_id: 'gpu-node-01', index: 6, utilization: 85.0, memory_used_gb: 61.8, memory_total_gb: 80, temperature: 69, power_watts: 515, status: 'active' },
    { gpu_id: 'gpu-node-01-gpu-7', node_id: 'gpu-node-01', index: 7, utilization: 78.9, memory_used_gb: 57.5, memory_total_gb: 80, temperature: 65, power_watts: 485, status: 'active' },

    // gpu-node-02: High load training cluster
    { gpu_id: 'gpu-node-02-gpu-0', node_id: 'gpu-node-02', index: 0, utilization: 92.4, memory_used_gb: 68.0, memory_total_gb: 80, temperature: 74, power_watts: 560, status: 'active' },
    { gpu_id: 'gpu-node-02-gpu-1', node_id: 'gpu-node-02', index: 1, utilization: 94.0, memory_used_gb: 68.0, memory_total_gb: 80, temperature: 75, power_watts: 570, status: 'active' },
    { gpu_id: 'gpu-node-02-gpu-2', node_id: 'gpu-node-02', index: 2, utilization: 91.5, memory_used_gb: 67.5, memory_total_gb: 80, temperature: 73, power_watts: 555, status: 'active' },
    { gpu_id: 'gpu-node-02-gpu-3', node_id: 'gpu-node-02', index: 3, utilization: 93.2, memory_used_gb: 68.0, memory_total_gb: 80, temperature: 75, power_watts: 565, status: 'active' },
    { gpu_id: 'gpu-node-02-gpu-4', node_id: 'gpu-node-02', index: 4, utilization: 89.8, memory_used_gb: 66.2, memory_total_gb: 80, temperature: 72, power_watts: 545, status: 'active' },
    { gpu_id: 'gpu-node-02-gpu-5', node_id: 'gpu-node-02', index: 5, utilization: 91.0, memory_used_gb: 67.0, memory_total_gb: 80, temperature: 73, power_watts: 550, status: 'active' },
    { gpu_id: 'gpu-node-02-gpu-6', node_id: 'gpu-node-02', index: 6, utilization: 94.5, memory_used_gb: 68.0, memory_total_gb: 80, temperature: 76, power_watts: 575, status: 'hot' },
    { gpu_id: 'gpu-node-02-gpu-7', node_id: 'gpu-node-02', index: 7, utilization: 92.0, memory_used_gb: 67.8, memory_total_gb: 80, temperature: 74, power_watts: 560, status: 'active' },

    // gpu-node-03: CPU starvation
    { gpu_id: 'gpu-node-03-gpu-0', node_id: 'gpu-node-03', index: 0, utilization: 11.2, memory_used_gb: 42.0, memory_total_gb: 80, temperature: 52, power_watts: 260, status: 'active' },
    { gpu_id: 'gpu-node-03-gpu-1', node_id: 'gpu-node-03', index: 1, utilization: 10.8, memory_used_gb: 41.5, memory_total_gb: 80, temperature: 51, power_watts: 255, status: 'active' },
    { gpu_id: 'gpu-node-03-gpu-2', node_id: 'gpu-node-03', index: 2, utilization: 12.0, memory_used_gb: 42.0, memory_total_gb: 80, temperature: 53, power_watts: 265, status: 'active' },
    { gpu_id: 'gpu-node-03-gpu-3', node_id: 'gpu-node-03', index: 3, utilization: 11.5, memory_used_gb: 41.8, memory_total_gb: 80, temperature: 52, power_watts: 260, status: 'active' },
    { gpu_id: 'gpu-node-03-gpu-4', node_id: 'gpu-node-03', index: 4, utilization: 10.5, memory_used_gb: 41.0, memory_total_gb: 80, temperature: 50, power_watts: 250, status: 'active' },
    { gpu_id: 'gpu-node-03-gpu-5', node_id: 'gpu-node-03', index: 5, utilization: 11.8, memory_used_gb: 42.0, memory_total_gb: 80, temperature: 52, power_watts: 262, status: 'active' },
    { gpu_id: 'gpu-node-03-gpu-6', node_id: 'gpu-node-03', index: 6, utilization: 11.0, memory_used_gb: 41.6, memory_total_gb: 80, temperature: 51, power_watts: 258, status: 'active' },
    { gpu_id: 'gpu-node-03-gpu-7', node_id: 'gpu-node-03', index: 7, utilization: 12.2, memory_used_gb: 42.2, memory_total_gb: 80, temperature: 53, power_watts: 266, status: 'active' },

    // gpu-node-04: XID 31 Hardware fault on GPU 1
    { gpu_id: 'gpu-node-04-gpu-0', node_id: 'gpu-node-04', index: 0, utilization: 64.0, memory_used_gb: 52.0, memory_total_gb: 80, temperature: 62, power_watts: 460, status: 'active' },
    { gpu_id: 'gpu-node-04-gpu-1', node_id: 'gpu-node-04', index: 1, utilization: 0.0, memory_used_gb: 0.0, memory_total_gb: 80, temperature: 41, power_watts: 120, status: 'fault', xid_error: 31 },
    { gpu_id: 'gpu-node-04-gpu-2', node_id: 'gpu-node-04', index: 2, utilization: 66.5, memory_used_gb: 53.5, memory_total_gb: 80, temperature: 63, power_watts: 470, status: 'active' },
    { gpu_id: 'gpu-node-04-gpu-3', node_id: 'gpu-node-04', index: 3, utilization: 65.0, memory_used_gb: 52.8, memory_total_gb: 80, temperature: 62, power_watts: 465, status: 'active' },
    { gpu_id: 'gpu-node-04-gpu-4', node_id: 'gpu-node-04', index: 4, utilization: 63.8, memory_used_gb: 51.9, memory_total_gb: 80, temperature: 61, power_watts: 455, status: 'active' },
    { gpu_id: 'gpu-node-04-gpu-5', node_id: 'gpu-node-04', index: 5, utilization: 67.2, memory_used_gb: 54.1, memory_total_gb: 80, temperature: 64, power_watts: 475, status: 'active' },
    { gpu_id: 'gpu-node-04-gpu-6', node_id: 'gpu-node-04', index: 6, utilization: 64.5, memory_used_gb: 52.4, memory_total_gb: 80, temperature: 62, power_watts: 462, status: 'active' },
    { gpu_id: 'gpu-node-04-gpu-7', node_id: 'gpu-node-04', index: 7, utilization: 65.8, memory_used_gb: 53.0, memory_total_gb: 80, temperature: 63, power_watts: 468, status: 'active' },
  ]

  const getTileColor = (cell: GPUHeatmapCell) => {
    if (cell.status === 'fault') return 'bg-rose-500 ring-2 ring-rose-400 animate-pulse text-slate-950 font-black'
    if (cell.status === 'idle') return 'bg-slate-800 text-slate-400 border border-slate-700'
    if (cell.status === 'hot') return 'bg-amber-500 text-slate-950 font-bold'
    if (cell.utilization > 80) return 'bg-emerald-500 text-slate-950 font-bold'
    if (cell.utilization > 40) return 'bg-indigo-500 text-white font-bold'
    return 'bg-orange-500/80 text-white font-bold'
  }

  const handleTileClick = (cell: GPUHeatmapCell) => {
    setSelectedCell(cell)
    onSelectGPU?.(cell)
  }

  return (
    <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl space-y-5 text-slate-100">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h3 className="text-base font-bold text-white flex items-center gap-2">
            <Cpu className="w-5 h-5 text-orange-400" />
            Live Cluster GPU Heatmap (32 Sockets)
          </h3>
          <p className="text-xs text-slate-400">High-frequency NVIDIA DCGM telemetry mapped per compute node</p>
        </div>

        {/* Status Filter */}
        <div className="flex gap-1.5">
          {(['all', 'active', 'idle', 'fault'] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`px-3 py-1 text-xs font-bold rounded-xl uppercase transition ${
                filter === f
                  ? 'bg-orange-500 text-white shadow-md shadow-orange-500/20'
                  : 'bg-slate-950 border border-slate-800 text-slate-400 hover:text-white'
              }`}
            >
              {f}
            </button>
          ))}
        </div>
      </div>

      {/* Nodes and GPU Grid */}
      <div className="space-y-4">
        {nodes.map((node) => {
          const nodeCells = cells.filter((c) => c.node_id === node)
          return (
            <div key={node} className="space-y-1.5 bg-slate-950/60 p-3.5 rounded-xl border border-slate-800/80">
              <div className="flex justify-between items-center text-xs">
                <span className="font-bold text-slate-300 font-mono">{node}</span>
                <span className="text-[11px] text-slate-500">8x H100 SXM5</span>
              </div>
              <div className="grid grid-cols-4 sm:grid-cols-8 gap-2">
                {nodeCells.map((cell) => {
                  const isDimmed = filter !== 'all' && (
                    (filter === 'active' && (cell.status === 'idle' || cell.status === 'fault')) ||
                    (filter === 'idle' && cell.status !== 'idle') ||
                    (filter === 'fault' && cell.status !== 'fault')
                  )

                  return (
                    <div
                      key={cell.gpu_id}
                      onClick={() => handleTileClick(cell)}
                      className={`h-11 rounded-xl flex flex-col items-center justify-center cursor-pointer transition-all duration-200 hover:scale-105 shadow-sm ${getTileColor(cell)} ${
                        isDimmed ? 'opacity-20' : 'opacity-100'
                      }`}
                      title={`${cell.gpu_id}: ${cell.utilization}% compute, ${cell.memory_used_gb}GB VRAM, ${cell.temperature}°C`}
                    >
                      <span className="text-[9px] uppercase tracking-wider opacity-80">GPU {cell.index}</span>
                      <span className="text-[11px]">
                        {cell.status === 'fault' ? 'XID 31' : `${cell.utilization.toFixed(0)}%`}
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          )
        })}
      </div>

      {/* Legend */}
      <div className="flex flex-wrap items-center justify-between gap-4 pt-2 text-xs border-t border-slate-800 text-slate-400">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-md bg-slate-800 border border-slate-700" /> Idle (0%)</div>
          <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-md bg-orange-500/80" /> Low (&lt;40%)</div>
          <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-md bg-indigo-500" /> Med (40-80%)</div>
          <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-md bg-emerald-500" /> Optimal (&gt;80%)</div>
          <div className="flex items-center gap-1.5"><div className="w-3 h-3 rounded-md bg-rose-500 animate-pulse" /> XID Fault</div>
        </div>

        {selectedCell && (
          <div className="text-xs text-orange-400 font-semibold">
            Selected: {selectedCell.gpu_id} ({selectedCell.utilization}% Util, {selectedCell.temperature}°C)
          </div>
        )}
      </div>
    </Card>
  )
}
