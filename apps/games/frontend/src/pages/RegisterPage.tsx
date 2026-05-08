import { useEffect } from 'react'

export function RegisterPage() {
  useEffect(() => {
    window.location.href = '/api/auth/oidc/login'
  }, [])

  return null
}
