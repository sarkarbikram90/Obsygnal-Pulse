import { useState } from 'react'
import { Outlet, Link, useLocation } from 'react-router-dom'
import { Activity, LayoutDashboard, Lightbulb, AlertTriangle, HardDrive, Cpu, Bell, Key, CreditCard, UserCheck } from 'lucide-react'
import { cn } from '../lib/utils'
import { useAuth } from '../context/AuthContext'
import { AuthModal } from './AuthModal'
import { APIKeyManager } from './APIKeyManager'

const navItems = [
  { name: 'Overview', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Recommendations', href: '/recommendations', icon: Lightbulb },
  { name: 'Inefficiencies', href: '/inefficiencies', icon: AlertTriangle },
  { name: 'Alerts', href: '/alerts', icon: Bell },
  { name: 'Nodes', href: '/nodes', icon: HardDrive },
  { name: 'Jobs', href: '/jobs', icon: Cpu },
  { name: 'Billing & Metering', href: '/billing', icon: CreditCard },
]

export function Layout() {
  const location = useLocation()
  const { user, org, isAuthenticated } = useAuth()
  const [authOpen, setAuthOpen] = useState(false)
  const [apiKeyOpen, setApiKeyOpen] = useState(false)

  return (
    <div className="flex h-screen bg-slate-950 text-slate-100 overflow-hidden font-sans selection:bg-orange-500/30">
      {/* Sidebar */}
      <aside className="w-64 flex flex-col border-r border-slate-800/80 bg-slate-900/60 backdrop-blur-xl">
        <div className="h-16 flex items-center px-6 border-b border-slate-800/80">
          <Activity className="w-6 h-6 text-orange-500 mr-3 animate-pulse" />
          <span className="font-extrabold text-lg tracking-tight bg-gradient-to-r from-orange-400 to-rose-500 bg-clip-text text-transparent">
            Obsygnal Pulse
          </span>
        </div>

        {/* Tenant Organization Switcher */}
        <div className="px-4 py-3 border-b border-slate-800/60 bg-slate-950/40">
          <div className="text-[10px] uppercase font-semibold text-slate-400 tracking-wider">Active Workspace</div>
          <div className="flex items-center justify-between mt-1">
            <span className="text-xs font-bold text-slate-200 truncate">{org?.name || 'Obsygnal Labs'}</span>
            <span className="px-2 py-0.5 text-[10px] font-extrabold uppercase bg-orange-500/10 text-orange-400 border border-orange-500/30 rounded-full">
              {org?.tier || 'Pro'}
            </span>
          </div>
        </div>
        
        <nav className="flex-1 py-4 px-3 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname.startsWith(item.href)
            
            return (
              <Link
                key={item.name}
                to={item.href}
                className={cn(
                  "group flex items-center px-3 py-2.5 text-sm font-medium rounded-xl transition-all duration-200",
                  isActive 
                    ? "bg-gradient-to-r from-orange-500/15 to-rose-500/10 text-orange-400 border border-orange-500/20 shadow-sm" 
                    : "text-slate-400 hover:bg-slate-800/50 hover:text-slate-100"
                )}
              >
                <Icon className={cn(
                  "mr-3 flex-shrink-0 h-5 w-5 transition-colors",
                  isActive ? "text-orange-400" : "text-slate-500 group-hover:text-slate-300"
                )} />
                {item.name}
              </Link>
            )
          })}
        </nav>

        {/* Sidebar Footer Buttons */}
        <div className="p-3 border-t border-slate-800/80 space-y-1">
          <button
            onClick={() => setApiKeyOpen(true)}
            className="w-full flex items-center px-3 py-2 text-xs font-semibold text-slate-300 hover:bg-slate-800/60 hover:text-white rounded-lg transition"
          >
            <Key className="mr-2.5 h-4 w-4 text-orange-400" />
            Agent API Keys
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
        {/* Top Header */}
        <header className="h-16 flex items-center justify-between px-8 border-b border-slate-800/80 bg-slate-900/40 backdrop-blur-md z-10 sticky top-0">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 px-3 py-1.5 bg-rose-500/10 border border-rose-500/30 rounded-xl text-xs">
              <span className="w-2 h-2 rounded-full bg-rose-500 animate-ping"></span>
              <span className="text-slate-300 font-medium">Monthly Wasted Compute:</span>
              <span className="font-extrabold text-rose-400">$6,820.00 / mo</span>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            {isAuthenticated ? (
              <div className="flex items-center gap-3">
                <div className="text-right hidden sm:block">
                  <div className="text-xs font-bold text-slate-200">{user?.display_name}</div>
                  <div className="text-[10px] text-slate-400">{user?.email}</div>
                </div>
                <div className="h-9 w-9 rounded-full bg-gradient-to-tr from-orange-500 to-rose-500 flex items-center justify-center text-xs font-bold text-white shadow-lg shadow-orange-500/20 ring-2 ring-slate-900">
                  {user?.display_name.substring(0, 2).toUpperCase() || 'AI'}
                </div>
              </div>
            ) : (
              <button
                onClick={() => setAuthOpen(true)}
                className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-xs font-bold rounded-xl transition shadow-lg shadow-orange-500/20"
              >
                <UserCheck className="w-4 h-4" />
                Sign In / Register
              </button>
            )}
          </div>
        </header>

        {/* Scrollable Page Content */}
        <div className="flex-1 overflow-auto bg-slate-950 p-8">
          <div className="max-w-7xl mx-auto">
            <Outlet />
          </div>
        </div>
      </main>

      {/* Modals */}
      <AuthModal isOpen={authOpen} onClose={() => setAuthOpen(false)} />
      <APIKeyManager isOpen={apiKeyOpen} onClose={() => setApiKeyOpen(false)} />
    </div>
  )
}
