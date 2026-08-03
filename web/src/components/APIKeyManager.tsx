import React, { useState, useEffect } from 'react'
import { api } from '../api/client'
import type { APIKey } from '../types'

interface APIKeyManagerProps {
  isOpen: boolean
  onClose: () => void
}

export const APIKeyManager: React.FC<APIKeyManagerProps> = ({ isOpen, onClose }) => {
  const [keys, setKeys] = useState<APIKey[]>([])
  const [keyName, setKeyName] = useState('')
  const [description, setDescription] = useState('')
  const [newlyCreatedKey, setNewlyCreatedKey] = useState<APIKey | null>(null)
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (isOpen) {
      api.getAPIKeys().then(setKeys).catch(() => {})
    }
  }, [isOpen])

  if (!isOpen) return null

  const handleCreateKey = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!keyName) return
    setLoading(true)
    try {
      const created = await api.createAPIKey(keyName, description)
      setNewlyCreatedKey(created)
      setKeys((prev) => [created, ...prev])
      setKeyName('')
      setDescription('')
    } catch {
      alert('Failed to generate API Key')
    } finally {
      setLoading(false)
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="w-full max-w-2xl bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden text-slate-100">
        <div className="p-6 border-b border-slate-800 flex items-center justify-between">
          <div>
            <h3 className="text-lg font-bold text-white">Cluster Agent API Keys</h3>
            <p className="text-xs text-slate-400">Machine-to-machine keys for obsygnal-agent metric collectors</p>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition">✕</button>
        </div>

        <div className="p-6 space-y-6 max-h-[80vh] overflow-y-auto">
          {newlyCreatedKey?.secret_key && (
            <div className="p-4 bg-emerald-500/10 border border-emerald-500/30 rounded-xl space-y-2">
              <div className="flex items-center justify-between text-xs font-semibold text-emerald-400">
                <span>🔑 New API Secret Generated — Copy now (will not be shown again!)</span>
              </div>
              <div className="flex items-center gap-2">
                <code className="flex-1 bg-slate-950 p-2.5 rounded-lg text-xs font-mono text-emerald-300 break-all select-all border border-emerald-900/50">
                  {newlyCreatedKey.secret_key}
                </code>
                <button
                  onClick={() => copyToClipboard(newlyCreatedKey.secret_key!)}
                  className="px-3 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-medium transition"
                >
                  {copied ? 'Copied!' : 'Copy'}
                </button>
              </div>
            </div>
          )}

          {/* Form */}
          <form onSubmit={handleCreateKey} className="p-4 bg-slate-950/50 border border-slate-800 rounded-xl space-y-3">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-300">Generate New Collector Key</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <input
                type="text"
                required
                placeholder="Key Name (e.g. Prod-H100-Cluster)"
                value={keyName}
                onChange={(e) => setKeyName(e.target.value)}
                className="px-3 py-2 bg-slate-900 border border-slate-800 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-orange-500"
              />
              <input
                type="text"
                placeholder="Description (optional)"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="px-3 py-2 bg-slate-900 border border-slate-800 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-orange-500"
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white font-medium rounded-lg text-xs transition"
            >
              {loading ? 'Generating...' : '+ Generate API Key'}
            </button>
          </form>

          {/* Key List */}
          <div className="space-y-2">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-400">Active API Keys</h4>
            {keys.map((k) => (
              <div key={k.id} className="flex items-center justify-between p-3 bg-slate-950 border border-slate-800/60 rounded-xl">
                <div>
                  <div className="text-xs font-bold text-slate-200">{k.name}</div>
                  <div className="text-[11px] font-mono text-slate-500">{k.prefix}</div>
                </div>
                <div className="text-right text-[11px] text-slate-400">
                  <div>Created {new Date(k.created_at).toLocaleDateString()}</div>
                  <div className="text-emerald-400">Last used: {k.last_used || 'Just now'}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
