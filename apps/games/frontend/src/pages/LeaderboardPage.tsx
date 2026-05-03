import { useState, useEffect, useContext } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { LeaderboardEntry } from '../api/types'
import { AuthContext } from '../App'

export function LeaderboardPage() {
  const { user } = useContext(AuthContext)
  const [entries, setEntries] = useState<LeaderboardEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.getLeaderboard()
      .then(setEntries)
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <nav className="nav">
        <div className="nav-brand">🎮 Mini Games</div>
        <div className="nav-links">
          {user ? (
            <>
              <Link to="/">Lobby</Link>
              {user.role === 'gamemaster' && <Link to="/admin">Admin</Link>}
            </>
          ) : (
            <>
              <Link to="/login">Login</Link>
              <Link to="/register">Register</Link>
            </>
          )}
        </div>
      </nav>

      <div className="page">
        <h1 style={{ marginBottom: 8 }}>Leaderboard</h1>
        <p style={{ color: 'var(--text-muted)', marginBottom: 24 }}>
          Top players ranked by wins
        </p>

        {error && <div className="error-msg">{error}</div>}

        {loading ? (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <div className="spinner" />
          </div>
        ) : (
          <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
            <table>
              <thead>
                <tr>
                  <th style={{ width: 48 }}>#</th>
                  <th>Player</th>
                  <th>Wins</th>
                  <th>Losses</th>
                  <th>Draws</th>
                  <th>Win Rate</th>
                </tr>
              </thead>
              <tbody>
                {entries.length === 0 ? (
                  <tr>
                    <td colSpan={6} style={{ textAlign: 'center', color: 'var(--text-muted)', padding: 32 }}>
                      No games played yet. Be the first!
                    </td>
                  </tr>
                ) : (
                  entries.map((entry, i) => {
                    const total = entry.wins + entry.losses + entry.draws
                    const winRate = total > 0 ? Math.round((entry.wins / total) * 100) : 0
                    const isMe = user?.username === entry.username
                    return (
                      <tr key={entry.username}>
                        <td>
                          <span style={{
                            fontWeight: 700,
                            color: i === 0 ? '#ffd700' : i === 1 ? '#c0c0c0' : i === 2 ? '#cd7f32' : 'var(--text-muted)'
                          }}>
                            {i + 1}
                          </span>
                        </td>
                        <td>
                          <span style={{
                            fontWeight: isMe ? 700 : 400,
                            color: isMe ? 'var(--accent)' : 'var(--text)',
                          }}>
                            {entry.username}
                            {isMe && ' (you)'}
                          </span>
                        </td>
                        <td style={{ color: 'var(--success)', fontWeight: 600 }}>{entry.wins}</td>
                        <td style={{ color: 'var(--danger)' }}>{entry.losses}</td>
                        <td style={{ color: 'var(--text-muted)' }}>{entry.draws}</td>
                        <td>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <div style={{
                              height: 6,
                              width: 80,
                              background: 'var(--bg-input)',
                              borderRadius: 3,
                              overflow: 'hidden',
                            }}>
                              <div style={{
                                height: '100%',
                                width: `${winRate}%`,
                                background: 'var(--success)',
                                borderRadius: 3,
                              }} />
                            </div>
                            <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>
                              {winRate}%
                            </span>
                          </div>
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
