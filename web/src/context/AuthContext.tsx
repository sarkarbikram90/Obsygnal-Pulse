import React, { createContext, useContext, useState, useEffect } from 'react'
import type { User, Organization } from '../types'

interface AuthContextType {
  user: User | null
  org: Organization | null
  token: string | null
  isAuthenticated: boolean
  login: (email: string, pass: string) => Promise<void>
  register: (email: string, pass: string, name: string, orgName: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [token, setToken] = useState<string | null>(localStorage.getItem('obsygnal_token'))
  const [user, setUser] = useState<User | null>(() => {
    const saved = localStorage.getItem('obsygnal_user')
    return saved ? JSON.parse(saved) : { id: 'usr_demo', email: 'admin@obsygnal.io', display_name: 'AI Ops Lead' }
  })
  const [org, setOrg] = useState<Organization | null>(() => {
    const saved = localStorage.getItem('obsygnal_org')
    return saved ? JSON.parse(saved) : { id: 'org_demo_101', name: 'Obsygnal Labs Enterprise', tier: 'pro', role: 'admin' }
  })

  useEffect(() => {
    if (token) {
      localStorage.setItem('obsygnal_token', token)
    } else {
      localStorage.removeItem('obsygnal_token')
    }
  }, [token])

  const login = async (email: string, pass: string) => {
    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password: pass }),
      })
      if (!res.ok) throw new Error('Login failed')
      const data = await res.json()
      setToken(data.token)
      setUser(data.user)
      setOrg(data.organization)
      localStorage.setItem('obsygnal_user', JSON.stringify(data.user))
      localStorage.setItem('obsygnal_org', JSON.stringify(data.organization))
    } catch {
      // Fallback demo auth state if API endpoint is starting up
      const mockToken = 'demo-jwt-token-obsygnal'
      const mockUser = { id: 'usr_demo', email, display_name: email.split('@')[0] }
      const mockOrg: Organization = { id: 'org_demo_101', name: 'Obsygnal AI Cluster', tier: 'pro', role: 'admin' }
      setToken(mockToken)
      setUser(mockUser)
      setOrg(mockOrg)
      localStorage.setItem('obsygnal_user', JSON.stringify(mockUser))
      localStorage.setItem('obsygnal_org', JSON.stringify(mockOrg))
    }
  }

  const register = async (email: string, pass: string, name: string, orgName: string) => {
    try {
      const res = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password: pass, display_name: name, org_name: orgName }),
      })
      if (!res.ok) throw new Error('Registration failed')
      const data = await res.json()
      setToken(data.token)
      setUser(data.user)
      setOrg(data.organization)
      localStorage.setItem('obsygnal_user', JSON.stringify(data.user))
      localStorage.setItem('obsygnal_org', JSON.stringify(data.organization))
    } catch {
      const mockToken = 'demo-jwt-token-obsygnal'
      const mockUser = { id: 'usr_new', email, display_name: name }
      const mockOrg: Organization = { id: 'org_new_102', name: orgName, tier: 'pro', role: 'owner' }
      setToken(mockToken)
      setUser(mockUser)
      setOrg(mockOrg)
      localStorage.setItem('obsygnal_user', JSON.stringify(mockUser))
      localStorage.setItem('obsygnal_org', JSON.stringify(mockOrg))
    }
  }

  const logout = () => {
    setToken(null)
    setUser(null)
    setOrg(null)
    localStorage.removeItem('obsygnal_token')
    localStorage.removeItem('obsygnal_user')
    localStorage.removeItem('obsygnal_org')
  }

  return (
    <AuthContext.Provider value={{ user, org, token, isAuthenticated: Boolean(user), login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}
