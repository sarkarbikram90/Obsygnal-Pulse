import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card } from '../../components/ui/card'
import { Bell, ShieldAlert, AlertTriangle, CheckCircle2, Clock, Send, MessageSquare, Webhook, Settings2 } from 'lucide-react'
import { api } from '../../api/client'
import { useToast } from '../../context/ToastContext'

export default function Alerts() {
  const toast = useToast()
  const [severityFilter, setSeverityFilter] = useState<string>('all')
  const [slackWebhook, setSlackWebhook] = useState('https://hooks.slack.com/services/T00/B00/XXXXX')
  const [pagerdutyKey, setPagerdutyKey] = useState('pd_live_9f8a31b2c4e5f6')
  const [slackEnabled, setSlackEnabled] = useState(true)
  const [pagerdutyEnabled, setPagerdutyEnabled] = useState(true)
  const [acknowledgedIds, setAcknowledgedIds] = useState<string[]>([])

  const { data: alerts, isLoading } = useQuery({
    queryKey: ['alerts', 'sim-cluster-01'],
    queryFn: () => api.getAlerts('sim-cluster-01'),
  })

  const displayAlerts = (alerts || []).filter((alert) => {
    return severityFilter === 'all' || alert.severity === severityFilter
  })

  const activeCount = (alerts || []).filter((a) => a.status === 'active' && !acknowledgedIds.includes(a.id)).length
  const criticalCount = (alerts || []).filter((a) => a.severity === 'critical').length

  const handleAcknowledge = (alertId: string, title: string) => {
    setAcknowledgedIds((prev) => [...prev, alertId])
    toast.success('Incident Acknowledged', `Alert '${title}' marked as acknowledged by operator.`)
  }

  const handleSaveChannels = () => {
    toast.success('Channels Updated', 'Notification channels & incident routing rules saved successfully.')
  }

  return (
    <div className="space-y-8 animate-in fade-in duration-500 text-slate-100">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-orange-400 to-rose-400 bg-clip-text text-transparent mb-2">
            Incident Management & Alert Routing
          </h1>
          <p className="text-slate-400">Real-time hardware alerts, threshold breach dispatches, Slack notifications, and PagerDuty integration.</p>
        </div>
      </div>

      {/* Top Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Active Incidents</span>
            <Bell className="w-4 h-4 text-orange-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">{activeCount} Unresolved</div>
          <p className="text-[11px] text-slate-500 mt-1">Requiring MLOps intervention</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Critical Hardware Faults</span>
            <ShieldAlert className="w-4 h-4 text-rose-400" />
          </div>
          <div className="text-2xl font-bold text-rose-400 mt-2">{criticalCount} Critical</div>
          <p className="text-[11px] text-rose-400/80 mt-1">XID 31 GPU Memory Page Fault</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Connected Channels</span>
            <Send className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">2 Channels Active</div>
          <p className="text-[11px] text-emerald-400 mt-1">Slack & PagerDuty Webhooks</p>
        </Card>

        <Card className="bg-slate-900 border-slate-800 p-5 rounded-2xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Mean Time to Fix (MTTR)</span>
            <Clock className="w-4 h-4 text-cyan-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-2">4.2 Minutes</div>
          <p className="text-[11px] text-slate-500 mt-1">With 1-click K8s auto-remediation</p>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left 2 Cols: Alert Incident Stream */}
        <div className="lg:col-span-2 space-y-4">
          <div className="flex justify-between items-center bg-slate-900/60 p-4 border border-slate-800 rounded-2xl">
            <h3 className="text-sm font-bold text-slate-200">Incident Event Stream</h3>
            <div className="flex gap-1.5">
              {['all', 'critical', 'high', 'medium'].map((sev) => (
                <button
                  key={sev}
                  onClick={() => setSeverityFilter(sev)}
                  className={`px-3 py-1 text-xs font-bold rounded-xl uppercase transition ${
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

          <div className="space-y-3">
            {isLoading ? (
              Array.from({ length: 3 }).map((_, i) => (
                <Card key={i} className="bg-slate-900 border-slate-800 p-5 rounded-2xl h-28 animate-pulse" />
              ))
            ) : displayAlerts.length === 0 ? (
              <div className="p-12 text-center border border-dashed border-slate-800 rounded-2xl bg-slate-900/30 text-slate-400 text-sm">
                No active incidents matching severity filter.
              </div>
            ) : (
              displayAlerts.map((alert) => {
                const isAcked = acknowledgedIds.includes(alert.id)
                return (
                  <Card
                    key={alert.id}
                    className={`bg-slate-900 border p-5 rounded-2xl transition-all shadow-md ${
                      isAcked
                        ? 'border-slate-800/60 opacity-60'
                        : alert.severity === 'critical'
                        ? 'border-rose-500/40 hover:border-rose-500/60'
                        : 'border-slate-800 hover:border-orange-500/30'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex items-start gap-3">
                        <div
                          className={`p-2.5 rounded-xl mt-0.5 ${
                            alert.severity === 'critical'
                              ? 'bg-rose-500/10 text-rose-400 border border-rose-500/30'
                              : 'bg-orange-500/10 text-orange-400 border border-orange-500/30'
                          }`}
                        >
                          {alert.severity === 'critical' ? (
                            <ShieldAlert className="w-5 h-5" />
                          ) : (
                            <AlertTriangle className="w-5 h-5" />
                          )}
                        </div>

                        <div>
                          <div className="flex items-center gap-2 mb-1">
                            <h4 className="text-sm font-bold text-white">{alert.title}</h4>
                            <span
                              className={`px-2 py-0.5 rounded text-[10px] font-extrabold uppercase ${
                                alert.severity === 'critical'
                                  ? 'bg-rose-500 text-slate-950'
                                  : 'bg-orange-500/20 text-orange-400 border border-orange-500/30'
                              }`}
                            >
                              {alert.severity}
                            </span>
                            {isAcked && (
                              <span className="text-[10px] text-emerald-400 font-bold flex items-center gap-1">
                                <CheckCircle2 className="w-3 h-3" /> Acknowledged
                              </span>
                            )}
                          </div>
                          <p className="text-xs text-slate-400">{alert.message}</p>

                          <div className="flex flex-wrap gap-3 mt-3 text-[11px] text-slate-500">
                            {alert.node_id && <span>Node: <strong className="text-slate-300">{alert.node_id}</strong></span>}
                            {alert.gpu_id && <span>GPU: <strong className="text-slate-300">{alert.gpu_id}</strong></span>}
                            <span>Dispatched via: <strong className="text-slate-300">{alert.channels.join(', ')}</strong></span>
                          </div>
                        </div>
                      </div>

                      {!isAcked && (
                        <button
                          onClick={() => handleAcknowledge(alert.id, alert.title)}
                          className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-xl transition flex-shrink-0"
                        >
                          Ack
                        </button>
                      )}
                    </div>
                  </Card>
                )
              })
            )}
          </div>
        </div>

        {/* Right Col: Notification Channels & Configuration */}
        <div className="space-y-6">
          <Card className="bg-slate-900 border-slate-800 p-6 rounded-2xl space-y-5">
            <div className="flex items-center gap-2 border-b border-slate-800 pb-3">
              <Settings2 className="w-4 h-4 text-orange-400" />
              <h3 className="text-sm font-bold text-white uppercase tracking-wider">Alert Channel Routing</h3>
            </div>

            {/* Slack Integration */}
            <div className="space-y-2 bg-slate-950/60 p-4 rounded-xl border border-slate-800/80">
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-2">
                  <MessageSquare className="w-4 h-4 text-[#ECB22E]" />
                  <span className="text-xs font-bold text-white">Slack Webhook</span>
                </div>
                <input
                  type="checkbox"
                  checked={slackEnabled}
                  onChange={(e) => setSlackEnabled(e.target.checked)}
                  className="accent-orange-500 cursor-pointer"
                />
              </div>
              <input
                type="text"
                value={slackWebhook}
                onChange={(e) => setSlackWebhook(e.target.value)}
                disabled={!slackEnabled}
                className="w-full px-3 py-1.5 bg-slate-900 border border-slate-800 rounded-lg text-xs text-slate-300 font-mono focus:outline-none focus:border-orange-500/50 disabled:opacity-50"
              />
              <p className="text-[10px] text-slate-500">Notifies #infra-gpu-alerts on severity &gt;= High</p>
            </div>

            {/* PagerDuty Integration */}
            <div className="space-y-2 bg-slate-950/60 p-4 rounded-xl border border-slate-800/80">
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-2">
                  <Webhook className="w-4 h-4 text-emerald-400" />
                  <span className="text-xs font-bold text-white">PagerDuty Service Key</span>
                </div>
                <input
                  type="checkbox"
                  checked={pagerdutyEnabled}
                  onChange={(e) => setPagerdutyEnabled(e.target.checked)}
                  className="accent-orange-500 cursor-pointer"
                />
              </div>
              <input
                type="password"
                value={pagerdutyKey}
                onChange={(e) => setPagerdutyKey(e.target.value)}
                disabled={!pagerdutyEnabled}
                className="w-full px-3 py-1.5 bg-slate-900 border border-slate-800 rounded-lg text-xs text-slate-300 font-mono focus:outline-none focus:border-orange-500/50 disabled:opacity-50"
              />
              <p className="text-[10px] text-slate-500">Triggers on-call engineer page on XID Hardware Faults</p>
            </div>

            <button
              onClick={handleSaveChannels}
              className="w-full py-2.5 bg-orange-500 hover:bg-orange-600 text-white font-bold text-xs rounded-xl shadow-lg shadow-orange-500/20 transition"
            >
              Save Notification Channels
            </button>
          </Card>
        </div>
      </div>
    </div>
  )
}
