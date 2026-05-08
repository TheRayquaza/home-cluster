import React from 'react'

export function LoginPage() {
  return (
    <div className="auth-container">
      <div className="auth-box">
        <h1>Welcome back</h1>
        <p>Sign in to play mini-games</p>

        <button
          className="btn-primary"
          onClick={() => { window.location.href = '/api/auth/oidc/login' }}
          style={{ width: '100%', marginTop: 24, padding: '12px', fontSize: 16 }}
        >
          Sign in with Keycloak
        </button>

        <p style={{ marginTop: 16, textAlign: 'center', color: 'var(--text-muted)', fontSize: 14 }}>
          No account? You can register directly on Keycloak.
        </p>
      </div>
    </div>
  )
}
