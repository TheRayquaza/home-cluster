import { useState, useEffect, useCallback } from 'react'
import { api } from '../api/client'
import type { User } from '../api/types'

interface AuthState {
  user: User | null
  loading: boolean
  error: string | null
}

interface AuthActions {
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  register: (username: string, password: string) => Promise<void>
}

export function useAuth(): AuthState & AuthActions {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api
      .me()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false))
  }, [])

  const login = useCallback(async (username: string, password: string) => {
    setError(null)
    try {
      const u = await api.login(username, password)
      setUser(u)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Login failed'
      setError(msg)
      throw e
    }
  }, [])

  const logout = useCallback(async () => {
    await api.logout().catch(() => {})
    setUser(null)
  }, [])

  const register = useCallback(
    async (username: string, password: string) => {
      setError(null)
      try {
        await api.register(username, password)
        await login(username, password)
      } catch (e) {
        const msg = e instanceof Error ? e.message : 'Registration failed'
        setError(msg)
        throw e
      }
    },
    [login]
  )

  return { user, loading, error, login, logout, register }
}
