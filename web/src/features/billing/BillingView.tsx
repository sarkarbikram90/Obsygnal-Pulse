import React, { useEffect, useState } from 'react'
import { api } from '../../api/client'
import type { UsageSummary } from '../../types'

export const BillingView: React.FC = () => {
  const [usage, setUsage] = useState<UsageSummary | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getUsageSummary()
      .then(setUsage)
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return <div className="p-8 text-center text-slate-400">Loading subscription details...</div>
  }

  const percentUsed = usage ? Math.min(100, (usage.gpu_hours_used / usage.gpu_hours_limit) * 100) : 0

  return (
    <div className="space-y-8 p-6 text-slate-100">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold bg-gradient-to-r from-orange-400 to-rose-400 bg-clip-text text-transparent">
          Subscription & GPU Metering
        </h1>
        <p className="text-sm text-slate-400">Manage usage tiers, billing limits, and monitored GPU compute metrics</p>
      </div>

      {/* Usage Meter Card */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="md:col-span-2 p-6 bg-slate-900 border border-slate-800 rounded-2xl space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <span className="text-xs uppercase tracking-wider font-semibold text-slate-400">Monthly Usage Meter</span>
              <h3 className="text-xl font-bold text-white mt-0.5">{usage?.gpu_hours_used.toLocaleString()} GPU Hours</h3>
            </div>
            <span className="px-3 py-1 bg-orange-500/10 text-orange-400 border border-orange-500/30 text-xs font-semibold rounded-full uppercase">
              {usage?.tier} Plan
            </span>
          </div>

          {/* Progress Bar */}
          <div className="space-y-1.5">
            <div className="flex justify-between text-xs text-slate-400">
              <span>Used: {usage?.gpu_hours_used} GPU-hrs</span>
              <span>Quota: {usage?.gpu_hours_limit.toLocaleString()} GPU-hrs</span>
            </div>
            <div className="w-full h-3 bg-slate-950 rounded-full overflow-hidden border border-slate-800">
              <div
                className="h-full bg-gradient-to-r from-orange-500 to-rose-500 rounded-full transition-all duration-500"
                style={{ width: `${percentUsed}%` }}
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-4 pt-2 text-xs border-t border-slate-800/80">
            <div>
              <div className="text-slate-400">Billing Cycle</div>
              <div className="font-semibold text-slate-200 mt-0.5">{usage?.billing_cycle}</div>
            </div>
            <div>
              <div className="text-slate-400">Monitored GPUs</div>
              <div className="font-semibold text-slate-200 mt-0.5">{usage?.monitored_gpus} GPUs</div>
            </div>
            <div>
              <div className="text-slate-400">Net Savings Realized</div>
              <div className="font-semibold text-emerald-400 mt-0.5">+${usage?.estimated_savings_usd.toLocaleString()}</div>
            </div>
          </div>
        </div>

        {/* ROI summary card */}
        <div className="p-6 bg-gradient-to-br from-slate-900 to-slate-950 border border-slate-800 rounded-2xl flex flex-col justify-between">
          <div>
            <span className="text-xs uppercase tracking-wider font-semibold text-orange-400">SaaS Return on Investment</span>
            <div className="text-3xl font-extrabold text-white mt-2">4.4x ROI</div>
            <p className="text-xs text-slate-400 mt-2 leading-relaxed">
              Obsygnal-Pulse prevented <span className="text-emerald-400 font-medium">${usage?.estimated_savings_usd.toLocaleString()}</span> in idle GPU waste this month.
            </p>
          </div>
          <button
            onClick={() => alert('Plan upgrade request submitted to account executive.')}
            className="w-full py-3 px-4 bg-orange-500 hover:bg-orange-600 text-white font-medium rounded-xl text-xs transition shadow-lg shadow-orange-500/20"
          >
            Upgrade to Enterprise Tier
          </button>
        </div>
      </div>

      {/* Pricing Tiers Table */}
      <div className="space-y-4">
        <h3 className="text-lg font-bold text-slate-200">Available SaaS Tiers</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* Free */}
          <div className="p-6 bg-slate-900 border border-slate-800 rounded-2xl space-y-4">
            <div>
              <h4 className="text-lg font-bold text-slate-200">Developer Free</h4>
              <div className="text-2xl font-extrabold text-white mt-1">$0 <span className="text-xs text-slate-400 font-normal">/ mo</span></div>
            </div>
            <ul className="space-y-2 text-xs text-slate-300">
              <li>✓ Up to 50 GPU Hours / mo</li>
              <li>✓ Single Cluster Agent</li>
              <li>✓ Basic Inefficiency Alerts</li>
              <li>✓ 7-Day Metric Retention</li>
            </ul>
          </div>

          {/* Pro */}
          <div className="p-6 bg-slate-900 border-2 border-orange-500 rounded-2xl space-y-4 relative">
            <span className="absolute -top-3 right-4 px-2.5 py-0.5 bg-orange-500 text-slate-950 font-bold text-[10px] uppercase rounded-full">Active Plan</span>
            <div>
              <h4 className="text-lg font-bold text-orange-400">Growth Pro</h4>
              <div className="text-2xl font-extrabold text-white mt-1">$499 <span className="text-xs text-slate-400 font-normal">/ mo</span></div>
            </div>
            <ul className="space-y-2 text-xs text-slate-300">
              <li>✓ Up to 5,000 GPU Hours / mo</li>
              <li>✓ Multi-Cluster K8s Agents</li>
              <li>✓ Automated Recommendations</li>
              <li>✓ 90-Day OLAP ClickHouse Data</li>
              <li>✓ Slack & Webhook Integrations</li>
            </ul>
          </div>

          {/* Enterprise */}
          <div className="p-6 bg-slate-900 border border-slate-800 rounded-2xl space-y-4">
            <div>
              <h4 className="text-lg font-bold text-rose-400">Enterprise</h4>
              <div className="text-2xl font-extrabold text-white mt-1">Custom <span className="text-xs text-slate-400 font-normal">annual</span></div>
            </div>
            <ul className="space-y-2 text-xs text-slate-300">
              <li>✓ Unlimited GPU Hours</li>
              <li>✓ 1-Click K8s Auto-Remediation</li>
              <li>✓ SOC2 Compliance Audit Logs</li>
              <li>✓ Dedicated SLA & Support</li>
              <li>✓ Custom AI Model Profiling</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
