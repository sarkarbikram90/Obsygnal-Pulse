import { AlertCircle } from 'lucide-react'

interface Props {
  lastUpdate: Date
  stats?: {
    total_received: number
    total_dropped: number
  } | null
  error?: string | null
}

export function StatusBar({ lastUpdate, stats, error }: Props) {
  return (
    <div className="flex items-center gap-4 text-xs text-slate-400 bg-slate-950/80 px-4 py-2 rounded-xl border border-slate-800">
      {error ? (
        <div className="flex items-center gap-1.5 text-rose-400 font-medium">
          <AlertCircle className="w-3.5 h-3.5" />
          <span>Offline Fallback Active</span>
        </div>
      ) : (
        <div className="flex items-center gap-1.5 text-emerald-400 font-medium">
          <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
          <span>Telemetry Stream Active</span>
        </div>
      )}

      {stats && (
        <div className="hidden sm:flex items-center gap-3 border-l border-slate-800 pl-3">
          <span>Ingested: <strong className="text-slate-200">{stats.total_received.toLocaleString()}</strong></span>
          {stats.total_dropped > 0 && (
            <span className="text-amber-400">Dropped: {stats.total_dropped}</span>
          )}
        </div>
      )}

      <div className="border-l border-slate-800 pl-3 text-[11px] text-slate-500">
        Updated: {lastUpdate.toLocaleTimeString()}
      </div>
    </div>
  )
}
