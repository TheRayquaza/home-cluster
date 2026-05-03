import { useState, useEffect, useContext } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Game, AdminSession } from '../api/types'
import { AuthContext } from '../App'
import { GameCard } from '../components/ui/GameCard'

export function AdminPage() {
  const { user } = useContext(AuthContext)
  const navigate = useNavigate()
  const [games, setGames] = useState<Game[]>([])
  const [daily, setDaily] = useState<Game | null>(null)
  const [sessions, setSessions] = useState<AdminSession[]>([])
  const [selectedGame, setSelectedGame] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  useEffect(() => {
    if (!user || user.role !== 'gamemaster') {
      navigate('/')
      return
    }

    Promise.all([
      api.getGames(),
      api.getDaily(),
      api.getSessions(),
    ])
      .then(([g, d, s]) => {
        setGames(g)
        setDaily(d)
        setSelectedGame(d.id)
        setSessions(s)
      })
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load'))
      .finally(() => setLoading(false))
  }, [user, navigate])

  const handleSetDaily = async () => {
    if (!selectedGame) return
    setSaving(true)
    setError('')
    setSuccess('')
    try {
      await api.setDailyGame(selectedGame)
      const updated = await api.getDaily()
      setDaily(updated)
      setSuccess(`Today's game set to: ${updated.name}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to set daily game')
    } finally {
      setSaving(false)
    }
  }

  const formatDate = (iso: string) => {
    try {
      return new Date(iso).toLocaleString()
    } catch {
      return iso
    }
  }

  const winnerLabel = (idx: number) => {
    if (idx === -1) return 'Draw'
    if (idx === 0) return 'Player 1'
    if (idx === 1) return 'Player 2'
    return 'Ongoing'
  }

  return (
    <div>
      <nav className="nav">
        <div className="nav-brand">🎮 Mini Games</div>
        <div className="nav-links">
          <Link to="/">Lobby</Link>
          <Link to="/scores">Leaderboard</Link>
          <span style={{ color: 'var(--accent)', fontSize: 13, fontWeight: 600 }}>
            ADMIN
          </span>
        </div>
      </nav>

      <div className="page">
        <h1 style={{ marginBottom: 4 }}>Admin Panel</h1>
        <p style={{ color: 'var(--text-muted)', marginBottom: 32 }}>
          Gamemaster controls
        </p>

        {error && <div className="error-msg">{error}</div>}
        {success && (
          <div style={{
            background: 'rgba(76,175,80,0.15)',
            border: '1px solid var(--success)',
            color: 'var(--success)',
            borderRadius: 8,
            padding: '10px 14px',
            marginBottom: 16,
            fontSize: 14,
          }}>
            {success}
          </div>
        )}

        {loading ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <div className="spinner" />
          </div>
        ) : (
          <>
            {/* Daily game control */}
            <div className="admin-section">
              <div className="card">
                <h2 style={{ marginBottom: 4, fontSize: 18 }}>Today's Game</h2>
                {daily && (
                  <p style={{ color: 'var(--text-muted)', fontSize: 14, marginBottom: 20 }}>
                    Currently:{' '}
                    <span style={{ fontSize: 18 }}>{daily.emoji}</span>{' '}
                    <strong style={{ color: 'var(--accent)' }}>{daily.name}</strong>
                  </p>
                )}

                <p style={{ color: 'var(--text-muted)', fontSize: 13, marginBottom: 16 }}>
                  Click a game to select it, then confirm below.
                </p>

                <div style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
                  gap: 14,
                  marginBottom: 20,
                }}>
                  {games.map(g => (
                    <GameCard
                      key={g.id}
                      game={g}
                      isDaily={daily?.id === g.id}
                      compact
                      selected={selectedGame === g.id}
                      onClick={() => setSelectedGame(g.id)}
                      onPlay={() => setSelectedGame(g.id)}
                    />
                  ))}
                </div>

                <button
                  className="btn-primary"
                  onClick={handleSetDaily}
                  disabled={saving || !selectedGame}
                  style={{ padding: '10px 28px' }}
                >
                  {saving ? 'Saving...' : "Set as Today's Game"}
                </button>
              </div>
            </div>

            {/* Sessions table */}
            <div className="admin-section">
              <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: 0.5 }}>
                Recent Sessions ({sessions.length})
              </h2>
              <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
                <table>
                  <thead>
                    <tr>
                      <th>Game</th>
                      <th>Player 1</th>
                      <th>Player 2</th>
                      <th>Result</th>
                      <th>Played</th>
                    </tr>
                  </thead>
                  <tbody>
                    {sessions.length === 0 ? (
                      <tr>
                        <td colSpan={5} style={{ textAlign: 'center', color: 'var(--text-muted)', padding: 32 }}>
                          No sessions yet
                        </td>
                      </tr>
                    ) : (
                      sessions.map(s => (
                        <tr key={s.id}>
                          <td>
                            <span style={{ fontWeight: 600 }}>{s.game_id}</span>
                          </td>
                          <td>{s.player1 || '—'}</td>
                          <td>{s.player2 || '—'}</td>
                          <td>
                            <span style={{
                              color: s.winner_idx === -1
                                ? 'var(--text-muted)'
                                : s.winner_idx === 0
                                  ? 'var(--p1)'
                                  : 'var(--p2)',
                              fontWeight: 600,
                            }}>
                              {winnerLabel(s.winner_idx)}
                            </span>
                          </td>
                          <td style={{ color: 'var(--text-muted)', fontSize: 13 }}>
                            {formatDate(s.played_at)}
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
