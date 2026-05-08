import { useState, useEffect, useCallback } from 'react'
import { api } from '../api/client'
import type { User } from '../api/types'

interface AuthState {
  user: User | null
  loading: boolean
}

interface AuthActions {
  logout: () => Promise<void>
}

export function useAuth(): AuthState & AuthActions {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api
      .me()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false))
  }, [])

  const logout = useCallback(async () => {
    const { logout_url } = await api.logout().catch(() => ({ logout_url: '/' }))
    setUser(null)
    window.location.href = logout_url
  }, [])

  return { user, loading, logout }
}
