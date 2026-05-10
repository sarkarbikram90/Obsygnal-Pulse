import { Outlet, Link, useLocation } from 'react-router-dom'
import { Activity, LayoutDashboard, Lightbulb, AlertTriangle, Settings, HardDrive, Cpu, Search, Bell } from 'lucide-react'
import { cn } from '../lib/utils'

const navItems = [
  { name: 'Overview', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Recommendations', href: '/recommendations', icon: Lightbulb },
  { name: 'Inefficiencies', href: '/inefficiencies', icon: AlertTriangle },
  { name: 'Alerts', href: '/alerts', icon: Bell },
  { name: 'Nodes', href: '/nodes', icon: HardDrive },
  { name: 'Jobs', href: '/jobs', icon: Cpu },
]

export function Layout() {
  const location = useLocation()

  return (
    <div className="flex h-screen bg-neutral-950 text-neutral-100 overflow-hidden font-sans selection:bg-indigo-500/30">
      {/* Sidebar */}
      <aside className="w-64 flex flex-col border-r border-neutral-800 bg-neutral-900/50 backdrop-blur-xl">
        <div className="h-16 flex items-center px-6 border-b border-neutral-800">
          <Activity className="w-6 h-6 text-indigo-500 mr-3" />
          <span className="font-semibold text-lg tracking-tight">Obsygnal Pulse</span>
        </div>
        
        <nav className="flex-1 py-6 px-4 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname.startsWith(item.href)
            
            return (
              <Link
                key={item.name}
                to={item.href}
                className={cn(
                  "group flex items-center px-3 py-2.5 text-sm font-medium rounded-lg transition-all duration-200",
                  isActive 
                    ? "bg-indigo-500/10 text-indigo-400" 
                    : "text-neutral-400 hover:bg-neutral-800/50 hover:text-neutral-100"
                )}
              >
                <Icon className={cn(
                  "mr-3 flex-shrink-0 h-5 w-5 transition-colors",
                  isActive ? "text-indigo-400" : "text-neutral-500 group-hover:text-neutral-300"
                )} />
                {item.name}
              </Link>
            )
          })}
        </nav>

        <div className="p-4 border-t border-neutral-800">
          <a href="/settings" className="group flex items-center px-3 py-2.5 text-sm font-medium rounded-lg text-neutral-400 hover:bg-neutral-800/50 hover:text-neutral-100 transition-all">
            <Settings className="mr-3 h-5 w-5 text-neutral-500 group-hover:text-neutral-300" />
            Settings
          </a>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
        {/* Top Header */}
        <header className="h-16 flex items-center justify-between px-8 border-b border-neutral-800 bg-neutral-900/30 backdrop-blur-md z-10 sticky top-0">
          <div className="flex items-center flex-1">
            <div className="relative w-96">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-neutral-500" />
              <input 
                type="text" 
                placeholder="Search nodes, jobs, GPUs..." 
                className="w-full bg-neutral-900 border border-neutral-800 text-sm rounded-full pl-10 pr-4 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500/50 transition-all placeholder:text-neutral-600"
              />
            </div>
          </div>
          <div className="flex items-center space-x-4">
            <button className="relative p-2 text-neutral-400 hover:text-neutral-100 transition-colors rounded-full hover:bg-neutral-800">
              <Bell className="h-5 w-5" />
              <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-rose-500 border-2 border-neutral-900"></span>
            </button>
            <div className="h-8 w-8 rounded-full bg-gradient-to-tr from-indigo-500 to-purple-500 flex items-center justify-center text-sm font-medium text-white shadow-lg shadow-indigo-500/20 ring-2 ring-neutral-900">
              JD
            </div>
          </div>
        </header>

        {/* Scrollable Page Content */}
        <div className="flex-1 overflow-auto bg-neutral-950 p-8">
          <div className="max-w-7xl mx-auto">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  )
}
