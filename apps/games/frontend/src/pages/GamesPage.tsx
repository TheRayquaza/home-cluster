import { useState, useEffect, useContext } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Game } from '../api/types'
import { AuthContext } from '../App'
import { GameCard } from '../components/ui/GameCard'

const MEMORY_SIZES = [
  { pairs: 4,  label: '4×2',  desc: 'Easy (8 cards)' },
  { pairs: 6,  label: '4×3',  desc: 'Normal (12 cards)' },
  { pairs: 8,  label: '4×4',  desc: 'Hard (16 cards)' },
  { pairs: 12, label: '6×4',  desc: 'Expert (24 cards)' },
]

export function GamesPage() {
  const { user, logout } = useContext(AuthContext)
  const navigate = useNavigate()
  const [games, setGames] = useState<Game[]>([])
  const [daily, setDaily] = useState<Game | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [creating, setCreating] = useState<string | null>(null)
  const [memoryPicker, setMemoryPicker] = useState(false)
  const [memoryPairs, setMemoryPairs] = useState(8)

  useEffect(() => {
    Promise.all([api.getGames(), api.getDaily()])
      .then(([g, d]) => { setGames(g); setDaily(d) })
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load games'))
      .finally(() => setLoading(false))
  }, [])

  const handlePlay = async (gameId: string, config?: Record<string, unknown>) => {
    if (creating) return
    setCreating(gameId)
    setError('')
    try {
      const { code } = await api.createRoomForGame(gameId, config)
      navigate(`/room/${code}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create room')
      setCreating(null)
    }
  }

  const handleMemoryPlay = () => {
    setMemoryPicker(true)
  }

  const confirmMemory = () => {
    setMemoryPicker(false)
    handlePlay('memory', { pairs: memoryPairs })
  }

  return (
    <div>
      <nav className="nav">
        <div className="nav-brand">🎮 Mini Games</div>
        <div className="nav-links">
          {user ? (
            <>
              <span style={{ color: 'var(--text-muted)', fontSize: 14 }}>{user.username}</span>
              <Link to="/">Lobby</Link>
              <Link to="/scores">Leaderboard</Link>
              {user.role === 'gamemaster' && <Link to="/admin">Admin</Link>}
              <button className="btn-ghost" onClick={logout} style={{ padding: '6px 14px', fontSize: 13 }}>Logout</button>
            </>
          ) : (
            <>
              <Link to="/scores">Leaderboard</Link>
              <Link to="/login">Login</Link>
            </>
          )}
        </div>
      </nav>

      <div className="page">
        <div style={{ marginBottom: 32 }}>
          <h1 style={{ marginBottom: 8 }}>Game Gallery</h1>
          <p style={{ color: 'var(--text-muted)' }}>Browse all games and start playing with a friend.</p>
        </div>

        {error && <div className="error-msg">{error}</div>}

        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}>
            <div className="spinner" />
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 20 }}>
            {games.map(game => (
              <GameCard
                key={game.id}
                game={game}
                isDaily={daily?.id === game.id}
                onPlay={game.id === 'memory' ? handleMemoryPlay : () => handlePlay(game.id)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Memory grid size modal */}
      {memoryPicker && (
        <div
          style={{
            position: 'fixed', inset: 0,
            background: 'rgba(0,0,0,0.7)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            zIndex: 1000,
          }}
          onClick={() => setMemoryPicker(false)}
        >
          <div
            style={{
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 16,
              padding: 32,
              maxWidth: 400,
              width: '90%',
            }}
            onClick={e => e.stopPropagation()}
          >
            <h2 style={{ marginBottom: 8 }}>🃏 Memory Match</h2>
            <p style={{ color: 'var(--text-muted)', marginBottom: 24, fontSize: 14 }}>
              Choose the grid size for this game.
            </p>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 24 }}>
              {MEMORY_SIZES.map(s => (
                <div
                  key={s.pairs}
                  onClick={() => setMemoryPairs(s.pairs)}
                  style={{
                    padding: '16px 12px',
                    borderRadius: 12,
                    border: memoryPairs === s.pairs
                      ? '2px solid #9c27b0'
                      : '1px solid var(--border)',
                    background: memoryPairs === s.pairs
                      ? 'rgba(156,39,176,0.15)'
                      : 'var(--bg-input)',
                    cursor: 'pointer',
                    textAlign: 'center',
                    transition: 'all 0.15s',
                  }}
                >
                  <div style={{ fontSize: 22, fontWeight: 800, marginBottom: 4 }}>{s.label}</div>
                  <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>{s.desc}</div>
                </div>
              ))}
            </div>

            <div style={{ display: 'flex', gap: 12 }}>
              <button className="btn-ghost" onClick={() => setMemoryPicker(false)} style={{ flex: 1 }}>
                Cancel
              </button>
              <button
                className="btn-primary"
                onClick={confirmMemory}
                disabled={!!creating}
                style={{ flex: 2, background: '#9c27b0' }}
              >
                {creating ? 'Creating...' : 'Create Room'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
