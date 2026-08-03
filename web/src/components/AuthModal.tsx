import React, { useState } from 'react'
import { useAuth } from '../context/AuthContext'

interface AuthModalProps {
  isOpen: boolean
  onClose: () => void
}

export const AuthModal: React.FC<AuthModalProps> = ({ isOpen, onClose }) => {
  const { login, register } = useAuth()
  const [isRegister, setIsRegister] = useState(false)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [orgName, setOrgName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!isOpen) return null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      if (isRegister) {
        await register(email, password, name, orgName)
      } else {
        await login(email, password)
      }
      onClose()
    } catch {
      setError('Authentication failed. Please check your credentials.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="w-full max-w-md bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden text-slate-100">
        <div className="p-6 border-b border-slate-800 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-2xl">🔥</span>
            <h3 className="text-xl font-bold bg-gradient-to-r from-orange-400 to-rose-500 bg-clip-text text-transparent">
              {isRegister ? 'Create Obsygnal Account' : 'Sign in to Obsygnal-Pulse'}
            </h3>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition">✕</button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 text-sm bg-rose-500/10 border border-rose-500/30 text-rose-400 rounded-lg">
              {error}
            </div>
          )}

          {isRegister && (
            <>
              <div>
                <label className="block text-xs uppercase tracking-wider font-semibold text-slate-400 mb-1">
                  Full Name
                </label>
                <input
                  type="text"
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Bikram Sarkar"
                  className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-orange-500"
                />
              </div>
              <div>
                <label className="block text-xs uppercase tracking-wider font-semibold text-slate-400 mb-1">
                  Organization / Company Name
                </label>
                <input
                  type="text"
                  required
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  placeholder="Acme AI Labs"
                  className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-orange-500"
                />
              </div>
            </>
          )}

          <div>
            <label className="block text-xs uppercase tracking-wider font-semibold text-slate-400 mb-1">
              Work Email
            </label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="admin@company.com"
              className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-orange-500"
            />
          </div>

          <div>
            <label className="block text-xs uppercase tracking-wider font-semibold text-slate-400 mb-1">
              Password
            </label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-lg text-sm text-slate-100 focus:outline-none focus:border-orange-500"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 px-4 bg-gradient-to-r from-orange-500 to-rose-600 hover:from-orange-600 hover:to-rose-700 text-white font-medium rounded-xl shadow-lg shadow-orange-500/20 transition disabled:opacity-50"
          >
            {loading ? 'Processing...' : isRegister ? 'Start Free SaaS Trial' : 'Sign In'}
          </button>

          <div className="text-center pt-2">
            <button
              type="button"
              onClick={() => setIsRegister(!isRegister)}
              className="text-xs text-slate-400 hover:text-orange-400 transition underline"
            >
              {isRegister ? 'Already have an account? Sign in' : "Don't have a workspace? Create one"}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
