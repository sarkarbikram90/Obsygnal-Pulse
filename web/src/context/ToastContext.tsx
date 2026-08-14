import React, { createContext, useContext, useState, useCallback } from 'react'
import { CheckCircle2, AlertCircle, Info, X, Zap } from 'lucide-react'

export type ToastType = 'success' | 'error' | 'info' | 'action'

export interface Toast {
  id: string
  type: ToastType
  title: string
  message?: string
  actionLabel?: string
  onAction?: () => void
  duration?: number
}

interface ToastContextValue {
  toast: {
    success: (title: string, message?: string) => void
    error: (title: string, message?: string) => void
    info: (title: string, message?: string) => void
    action: (title: string, message: string, actionLabel: string, onAction: () => void) => void
    custom: (options: Omit<Toast, 'id'>) => void
  }
}

const ToastContext = createContext<ToastContextValue | null>(null)

export const useToast = () => {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider')
  }
  return context.toast
}

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<Toast[]>([])

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const addToast = useCallback((toastOptions: Omit<Toast, 'id'>) => {
    const id = Math.random().toString(36).substring(2, 9)
    const duration = toastOptions.duration ?? 4000

    const newToast: Toast = { ...toastOptions, id }
    setToasts((prev) => [...prev, newToast])

    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }
  }, [removeToast])

  const toastMethods = {
    success: (title: string, message?: string) => addToast({ type: 'success', title, message }),
    error: (title: string, message?: string) => addToast({ type: 'error', title, message }),
    info: (title: string, message?: string) => addToast({ type: 'info', title, message }),
    action: (title: string, message: string, actionLabel: string, onAction: () => void) =>
      addToast({ type: 'action', title, message, actionLabel, onAction, duration: 6000 }),
    custom: addToast,
  }

  return (
    <ToastContext.Provider value={{ toast: toastMethods }}>
      {children}
      {/* Toast Render Portal */}
      <div className="fixed bottom-6 right-6 z-50 flex flex-col gap-2.5 max-w-md w-full pointer-events-none">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={`pointer-events-auto flex items-start gap-3 p-4 rounded-2xl border shadow-2xl backdrop-blur-xl transition-all duration-300 animate-in slide-in-from-bottom-5 fade-in ${
              t.type === 'success'
                ? 'bg-slate-900/95 border-emerald-500/40 text-slate-100 shadow-emerald-500/10'
                : t.type === 'error'
                ? 'bg-slate-900/95 border-rose-500/40 text-slate-100 shadow-rose-500/10'
                : t.type === 'action'
                ? 'bg-slate-900/95 border-orange-500/40 text-slate-100 shadow-orange-500/10'
                : 'bg-slate-900/95 border-slate-700/60 text-slate-100 shadow-slate-900/40'
            }`}
          >
            <div className="mt-0.5 flex-shrink-0">
              {t.type === 'success' && <CheckCircle2 className="w-5 h-5 text-emerald-400" />}
              {t.type === 'error' && <AlertCircle className="w-5 h-5 text-rose-400" />}
              {t.type === 'action' && <Zap className="w-5 h-5 text-orange-400" />}
              {t.type === 'info' && <Info className="w-5 h-5 text-cyan-400" />}
            </div>

            <div className="flex-1 min-w-0">
              <h4 className="text-xs font-bold tracking-tight text-white">{t.title}</h4>
              {t.message && <p className="text-[11px] text-slate-400 mt-0.5 leading-relaxed">{t.message}</p>}
              {t.actionLabel && t.onAction && (
                <button
                  onClick={() => {
                    t.onAction?.()
                    removeToast(t.id)
                  }}
                  className="mt-2 px-2.5 py-1 bg-orange-500 hover:bg-orange-600 text-white font-bold text-[10px] rounded-lg transition"
                >
                  {t.actionLabel}
                </button>
              )}
            </div>

            <button
              onClick={() => removeToast(t.id)}
              className="text-slate-500 hover:text-slate-300 transition -mr-1 -mt-1 p-1"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
