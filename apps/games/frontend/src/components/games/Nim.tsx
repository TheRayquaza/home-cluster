import type { GameProps, NimState } from '../../api/types'

export function Nim({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as NimState
  const isMyTurn = s.turn === playerIdx && !gameOver

  const handleTake = (n: number) => {
    if (!isMyTurn || n > s.sticks) return
    onAction({ take: n })
  }

  return (
    <div className="game-container">
      <div className="game-info">
        <div className={`turn-indicator ${isMyTurn ? 'your-turn' : 'waiting'}`}>
          {isMyTurn ? 'Your turn' : "Opponent's turn"}
        </div>
        <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>
          Sticks remaining: <strong style={{ color: 'var(--accent)' }}>{s.sticks}</strong>
        </div>
      </div>

      <div className="nim-sticks">
        {Array.from({ length: s.sticks }, (_, i) => (
          <div key={i} className="nim-stick" />
        ))}
        {s.sticks === 0 && (
          <div style={{ color: 'var(--text-muted)' }}>No sticks left!</div>
        )}
      </div>

      <div className="nim-actions">
        {[1, 2, 3].map(n => (
          <button
            key={n}
            className="btn-primary"
            disabled={!isMyTurn || n > s.sticks}
            onClick={() => handleTake(n)}
          >
            Take {n}
          </button>
        ))}
      </div>

      <div className="nim-hint">
        Take 1, 2, or 3 sticks. The player who takes the LAST stick LOSES.
      </div>
    </div>
  )
}
