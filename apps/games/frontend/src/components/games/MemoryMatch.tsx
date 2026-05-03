import { useState, useEffect } from 'react'
import type { GameProps, MemoryMatchState } from '../../api/types'

const EMOJIS = ['🍎','🍊','🍋','🍇','🍓','🫐','🍒','🥝','🌸','🦋','⭐','🎯','🎸','🚀','🦊','🐬','💎','🌈']

export function MemoryMatch({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as MemoryMatchState
  const isMyTurn = s.turn === playerIdx && !gameOver
  const [noMatchMsg, setNoMatchMsg] = useState(false)

  useEffect(() => {
    if (s.waiting) {
      setNoMatchMsg(true)
      const t = setTimeout(() => setNoMatchMsg(false), 900)
      return () => clearTimeout(t)
    }
  }, [s.waiting])

  const handleCardClick = (idx: number) => {
    if (!isMyTurn) return
    if (s.matched?.[idx]) return
    if (s.revealed?.[idx]) return
    // KEY FIX: when waiting=true (unmatched pair showing), allow click to
    // trigger resolveWaiting on the server, even if pending.length === 2
    if (!s.waiting && (s.pending?.length ?? 0) >= 2) return
    onAction({ card: idx })
  }

  const cols = s.cols ?? 4

  return (
    <div className="game-container">
      <div className="game-info">
        <div className={`turn-indicator ${isMyTurn ? 'your-turn' : 'waiting'}`}>
          {isMyTurn
            ? (s.waiting ? 'Pick a new card to continue' : 'Your turn')
            : "Opponent's turn"}
        </div>
        {noMatchMsg && (
          <div style={{ color: 'var(--danger)', fontWeight: 700, fontSize: 14 }}>
            No match!
          </div>
        )}
      </div>

      <div className="memory-scores">
        <div style={{ color: 'var(--p1)' }}>P1: {s.scores?.[0] ?? 0} pairs</div>
        <div style={{ color: 'var(--p2)' }}>P2: {s.scores?.[1] ?? 0} pairs</div>
        <div style={{ color: 'var(--text-muted)', fontSize: 13 }}>
          {(s.cards?.length ?? 16) / 2} pairs · {(s.cols ?? 4)}×{(s.rows ?? 4)} grid
        </div>
      </div>

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: `repeat(${cols}, 1fr)`,
          gap: 8,
          maxWidth: cols * 72,
          margin: '0 auto',
        }}
      >
        {(s.cards ?? []).map((cardVal, i) => {
          const isRevealed = s.revealed?.[i] ?? false
          const isMatched = s.matched?.[i] ?? false
          const isPendingNoMatch = (s.pending?.includes(i) ?? false) && s.waiting

          return (
            <div
              key={i}
              onClick={() => handleCardClick(i)}
              style={{
                width: 60,
                height: 60,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: 26,
                borderRadius: 10,
                cursor: isMyTurn && !isMatched && !isRevealed ? 'pointer' : 'default',
                background: isMatched
                  ? 'rgba(76,175,80,0.2)'
                  : isPendingNoMatch
                    ? 'rgba(233,69,96,0.15)'
                    : isRevealed
                      ? 'var(--bg-card)'
                      : 'var(--bg-input)',
                border: isMatched
                  ? '2px solid var(--success)'
                  : isPendingNoMatch
                    ? '2px solid var(--danger)'
                    : isRevealed
                      ? '2px solid var(--border)'
                      : '2px solid transparent',
                transform: isRevealed || isMatched ? 'scale(1)' : 'scale(0.95)',
                transition: 'all 0.15s ease',
                userSelect: 'none',
                WebkitUserSelect: 'none',
              }}
            >
              {isRevealed || isMatched ? EMOJIS[cardVal % EMOJIS.length] : '?'}
            </div>
          )
        })}
      </div>

      <div style={{ color: 'var(--text-muted)', fontSize: 13, marginTop: 8 }}>
        Find all matching pairs · Match = extra turn
      </div>
    </div>
  )
}
