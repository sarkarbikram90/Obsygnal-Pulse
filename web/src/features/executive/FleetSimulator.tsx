import React, { useState } from 'react'
import { Card } from '../../components/ui/card'
import { DollarSign, Cpu, TrendingUp, ShieldCheck, Zap, Download } from 'lucide-react'

export const FleetSimulator: React.FC = () => {
  const [gpuCount, setGpuCount] = useState<number>(64)
  const [gpuModel, setGpuModel] = useState<string>('H100')
  const [idlePercent, setIdlePercent] = useState<number>(25)
  const [hoursPerDay, setHoursPerDay] = useState<number>(24)

  const hourlyRates: Record<string, number> = {
    B200: 9.50,
    H200: 6.85,
    H100: 5.12,
    A100: 3.67,
    L40S: 1.85,
  }

  const rate = hourlyRates[gpuModel] || 5.12
  const totalHoursMonth = gpuCount * hoursPerDay * 30.5
  const monthlySpend = totalHoursMonth * rate
  const wastedMonthly = monthlySpend * (idlePercent / 100)
  const recoverableSavings = wastedMonthly * 0.85 // 85% recovery rate
  const annualSavings = recoverableSavings * 12

  return (
    <div className="space-y-8 p-2 text-slate-100 animate-in fade-in duration-500">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <span className="px-3 py-1 bg-orange-500/10 text-orange-400 border border-orange-500/30 text-xs font-bold rounded-full uppercase tracking-wider">
            Investor & VP Infrastructure Tool
          </span>
          <h1 className="text-3xl font-extrabold bg-gradient-to-r from-orange-400 via-rose-400 to-amber-300 bg-clip-text text-transparent mt-2">
            AI Fleet ROI & Savings Simulator
          </h1>
          <p className="text-sm text-slate-400 mt-1">Model custom GPU topologies, compute instant annual cost recovery, and project unit economics.</p>
        </div>
        <button
          onClick={() => alert('Executive PDF Investment Brief downloaded.')}
          className="flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-orange-500 to-rose-500 hover:from-orange-600 hover:to-rose-600 text-white font-bold text-xs rounded-xl shadow-lg shadow-orange-500/20 transition"
        >
          <Download className="w-4 h-4" /> Export Executive PDF Brief
        </button>
      </div>

      {/* Simulator Inputs & Output Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Controls Card */}
        <Card className="lg:col-span-1 bg-slate-900 border-slate-800 p-6 rounded-2xl space-y-5">
          <h3 className="text-sm font-bold text-slate-200 uppercase tracking-wider">Cluster Configuration</h3>

          <div>
            <label className="block text-xs font-semibold text-slate-400 mb-2">GPU Model Selection</label>
            <div className="grid grid-cols-3 gap-2">
              {['B200', 'H200', 'H100', 'A100', 'L40S'].map((m) => (
                <button
                  key={m}
                  onClick={() => setGpuModel(m)}
                  className={`py-2 text-xs font-bold rounded-xl border transition ${
                    gpuModel === m
                      ? 'bg-orange-500 text-white border-orange-500 shadow-md shadow-orange-500/20'
                      : 'bg-slate-950 border-slate-800 text-slate-400 hover:text-white'
                  }`}
                >
                  {m} (${hourlyRates[m]}/hr)
                </button>
              ))}
            </div>
          </div>

          <div>
            <div className="flex justify-between text-xs font-semibold text-slate-300 mb-1">
              <span>GPU Fleet Size</span>
              <span className="text-orange-400 font-bold">{gpuCount} GPUs</span>
            </div>
            <input
              type="range"
              min="8"
              max="1024"
              step="8"
              value={gpuCount}
              onChange={(e) => setGpuCount(Number(e.target.value))}
              className="w-full accent-orange-500 bg-slate-950 rounded-lg cursor-pointer"
            />
          </div>

          <div>
            <div className="flex justify-between text-xs font-semibold text-slate-300 mb-1">
              <span>Estimated Compute Idle / Waste %</span>
              <span className="text-rose-400 font-bold">{idlePercent}%</span>
            </div>
            <input
              type="range"
              min="5"
              max="60"
              step="1"
              value={idlePercent}
              onChange={(e) => setIdlePercent(Number(e.target.value))}
              className="w-full accent-rose-500 bg-slate-950 rounded-lg cursor-pointer"
            />
          </div>
        </Card>

        {/* Financial Metrics Cards */}
        <div className="lg:col-span-2 grid grid-cols-1 sm:grid-cols-2 gap-6">
          <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl flex flex-col justify-between">
            <div className="space-y-2">
              <span className="text-xs uppercase font-semibold text-slate-400">Total Monthly Cloud GPU Spend</span>
              <div className="text-3xl font-extrabold text-white">${monthlySpend.toLocaleString(undefined, { maximumFractionDigits: 0 })}</div>
              <p className="text-xs text-slate-400">Based on {gpuCount}x {gpuModel} running {hoursPerDay}h/day</p>
            </div>
            <div className="pt-4 border-t border-slate-800 flex items-center text-xs text-slate-400 gap-1.5">
              <Cpu className="w-4 h-4 text-orange-400" /> Monitored Rate: ${rate}/gpu-hr
            </div>
          </Card>

          <Card className="bg-gradient-to-br from-rose-500/10 to-slate-900 border-rose-500/30 p-6 rounded-2xl flex flex-col justify-between">
            <div className="space-y-2">
              <span className="text-xs uppercase font-semibold text-rose-400">Unoptimized Annual Waste</span>
              <div className="text-3xl font-extrabold text-rose-400">${(wastedMonthly * 12).toLocaleString(undefined, { maximumFractionDigits: 0 })}</div>
              <p className="text-xs text-slate-400">Capital burned on idle VRAM & stalled loops</p>
            </div>
            <div className="pt-4 border-t border-slate-800/80 flex items-center text-xs text-rose-300 gap-1.5">
              <Zap className="w-4 h-4 text-rose-400" /> {idlePercent}% fleet underutilization
            </div>
          </Card>

          <Card className="sm:col-span-2 bg-gradient-to-br from-slate-900 via-slate-900 to-emerald-950/40 border border-emerald-500/40 p-6 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-6">
            <div>
              <span className="text-xs uppercase tracking-wider font-extrabold text-emerald-400">Projected Obsygnal Annual Net Recovery</span>
              <div className="text-4xl font-black text-emerald-300 mt-1">${annualSavings.toLocaleString(undefined, { maximumFractionDigits: 0 })} <span className="text-sm font-normal text-emerald-500">/ year</span></div>
              <p className="text-xs text-slate-300 mt-2">
                Delivers an estimated <span className="text-emerald-400 font-bold">5.8x Net ROI</span> on SaaS license fee while increasing PyTorch training throughput by 3.2x.
              </p>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
