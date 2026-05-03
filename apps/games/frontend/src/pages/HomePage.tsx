import { useState, useEffect, useContext } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Game } from '../api/types'
import { AuthContext } from '../App'

export function HomePage() {
  const { user, logout } = useContext(AuthContext)
  const navigate = useNavigate()
  const [daily, setDaily] = useState<Game | null>(null)
  const [games, setGames] = useState<Game[]>([])
  const [joinCode, setJoinCode] = useState('')
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    api.getDaily().then(setDaily).catch(() => {})
    api.getGames().then(setGames).catch(() => {})
  }, [])

  const handleCreateRoom = async () => {
    setCreating(true)
    setError('')
    try {
      const { code } = await api.createRoom()
      navigate(`/room/${code}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create room')
    } finally {
      setCreating(false)
    }
  }

  const handleJoinRoom = (e: React.FormEvent) => {
    e.preventDefault()
    const code = joinCode.trim().toUpperCase()
    if (code.length !== 4) {
      setError('Room code must be 4 characters')
      return
    }
    navigate(`/room/${code}`)
  }

  return (
    <div>
      {/* Nav */}
      <nav className="nav">
        <div className="nav-brand">🎮 Mini Games</div>
        <div className="nav-links">
          <span style={{ color: 'var(--text-muted)', fontSize: 14 }}>
            {user?.username}
          </span>
          <Link to="/scores">Leaderboard</Link>
          {user?.role === 'gamemaster' && <Link to="/admin">Admin</Link>}
          <button
            className="btn-ghost"
            onClick={logout}
            style={{ padding: '6px 14px', fontSize: 13 }}
          >
            Logout
          </button>
        </div>
      </nav>

      <div className="page">
        {/* Hero */}
        <div className="home-hero">
          <h1>Mini Games Platform</h1>
          <p style={{ color: 'var(--text-muted)', marginBottom: 20 }}>
            Challenge a friend to a 2-player game
          </p>

          {daily && (
            <div>
              <div className="daily-badge">Today's Game</div>
              <div style={{ fontSize: 48, marginBottom: 8 }}>{daily.emoji}</div>
              <h2 style={{ fontSize: 28, fontWeight: 800, marginBottom: 8 }}>
                {daily.name}
              </h2>
              <p style={{ color: 'var(--text-muted)', maxWidth: 400, margin: '0 auto' }}>
                {daily.description}
              </p>
            </div>
          )}
        </div>

        {error && <div className="error-msg">{error}</div>}

        {/* Room actions */}
        <div className="room-actions">
          <div className="card">
            <h3 style={{ marginBottom: 8 }}>Create Room</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: 14, marginBottom: 16 }}>
              Start a new game and share the code with a friend.
            </p>
            <button
              className="btn-primary"
              onClick={handleCreateRoom}
              disabled={creating}
              style={{ width: '100%', padding: '12px' }}
            >
              {creating ? 'Creating...' : 'Create Room'}
            </button>
          </div>

          <div className="card">
            <h3 style={{ marginBottom: 8 }}>Join Room</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: 14, marginBottom: 16 }}>
              Enter a 4-character room code to join a friend.
            </p>
            <form onSubmit={handleJoinRoom} style={{ display: 'flex', gap: 8 }}>
              <input
                type="text"
                value={joinCode}
                onChange={e => setJoinCode(e.target.value.toUpperCase())}
                placeholder="XXXX"
                maxLength={4}
                style={{
                  textAlign: 'center',
                  letterSpacing: 8,
                  fontWeight: 800,
                  fontSize: 20,
                  textTransform: 'uppercase',
                }}
              />
              <button
                type="submit"
                className="btn-secondary"
                style={{ whiteSpace: 'nowrap' }}
              >
                Join
              </button>
            </form>
          </div>
        </div>

        {/* Games list */}
        <div style={{ marginTop: 40 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
            <h2 style={{ color: 'var(--text-muted)', fontSize: 14, textTransform: 'uppercase', letterSpacing: 1, margin: 0 }}>
              Available Games
            </h2>
            <Link
              to="/games"
              style={{
                fontSize: 13,
                fontWeight: 600,
                color: 'var(--accent)',
                padding: '6px 14px',
                border: '1px solid var(--accent)',
                borderRadius: 8,
                textDecoration: 'none',
              }}
            >
              Browse All Games →
            </Link>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 12 }}>
            {games.map(g => (
              <div
                key={g.id}
                className="card"
                style={{
                  padding: 16,
                  borderLeft: `3px solid ${g.color}`,
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                  <span style={{ fontSize: 20 }}>{g.emoji}</span>
                  <div style={{ fontWeight: 700 }}>{g.name}</div>
                </div>
                <div style={{ color: 'var(--text-muted)', fontSize: 12 }}>{g.description}</div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
