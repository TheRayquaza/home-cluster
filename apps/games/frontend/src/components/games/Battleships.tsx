import type { GameProps, BattleshipsState } from '../../api/types'

// Cell value constants
// boards: 0=water, 1=ship, 2=hit, 3=miss
// shots:  0=unknown, 1=miss, 2=hit

function cellColor(value: number, isOwn: boolean): string {
  if (value === 2) return '#e64a19' // hit
  if (value === 3) return '#29b6f6' // miss
  if (isOwn && value === 1) return '#546e7a' // own ship
  return '#0d47a1' // water / unknown
}

interface GridProps {
  cells: number[]
  isOwn: boolean
  clickable: boolean
  onCellClick?: (idx: number) => void
}

function BattleGrid({ cells, isOwn, clickable, onCellClick }: GridProps) {
  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(10, 1fr)',
      gap: 2,
      width: 'min(320px, 100%)',
    }}>
      {cells.map((val, idx) => {
        const canClick = clickable && val === 0
        return (
          <div
            key={idx}
            onClick={() => canClick && onCellClick?.(idx)}
            style={{
              aspectRatio: '1',
              background: cellColor(val, isOwn),
              borderRadius: 2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 14,
              fontWeight: 700,
              color: 'white',
              cursor: canClick ? 'crosshair' : 'default',
              transition: 'filter 0.15s',
              filter: canClick ? undefined : undefined,
            }}
            onMouseEnter={e => { if (canClick) (e.currentTarget as HTMLDivElement).style.filter = 'brightness(1.4)' }}
            onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.filter = '' }}
          >
            {val === 2 ? '⊕' : val === 3 ? '·' : ''}
          </div>
        )
      })}
    </div>
  )
}

export function Battleships({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as BattleshipsState
  const oppIdx = playerIdx === 0 ? 1 : 0

  const myBoard: number[] = s.boards?.[playerIdx] ?? Array(100).fill(0)
  const myShots: number[] = s.shots?.[playerIdx] ?? Array(100).fill(0)

  const handleShot = (idx: number) => {
    const x = idx % 10
    const y = Math.floor(idx / 10)
    onAction({ x, y })
  }

  const isMyTurn = s.turn === playerIdx
  const iAmReady = s.ready?.[playerIdx] ?? false
  const oppReady = s.ready?.[oppIdx] ?? false
  const myShipsLeft = s.ships_left?.[playerIdx] ?? 0
  const oppShipsLeft = s.ships_left?.[oppIdx] ?? 0

  if (s.phase === 'placement') {
    return (
      <div className="game-container">
        <h2 style={{ textAlign: 'center', marginBottom: 16 }}>Ship Placement</h2>

        <div style={{ textAlign: 'center', marginBottom: 12, color: 'var(--text-muted)', fontSize: 14 }}>
          Ships: {s.ship_sizes?.join(', ')}
        </div>

        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 16 }}>
          <BattleGrid cells={myBoard} isOwn={true} clickable={false} />
        </div>

        <div style={{ display: 'flex', gap: 12, justifyContent: 'center', flexWrap: 'wrap', marginBottom: 16 }}>
          <button
            className="btn-primary"
            onClick={() => onAction({ action: 'randomize' })}
            disabled={iAmReady}
            style={{ fontSize: 16 }}
          >
            🔀 Randomize
          </button>
          <button
            className="btn-primary"
            onClick={() => onAction({ action: 'ready' })}
            disabled={iAmReady}
            style={{
              fontSize: 16,
              background: iAmReady ? 'var(--success)' : undefined,
            }}
          >
            ✓ Ready
          </button>
        </div>

        <div style={{ textAlign: 'center', color: 'var(--text-muted)', fontSize: 14 }}>
          {iAmReady && !oppReady && 'Waiting for opponent...'}
          {iAmReady && oppReady && 'Both ready — starting!'}
          {!iAmReady && 'Place your ships, then press Ready'}
        </div>
      </div>
    )
  }

  // Combat phase
  const canShoot = isMyTurn && !gameOver

  return (
    <div className="game-container">
      {/* Turn indicator */}
      <div className="game-info">
        {isMyTurn ? (
          <div className="turn-indicator your-turn">Your turn — Fire!</div>
        ) : (
          <div className="turn-indicator waiting">Opponent's turn</div>
        )}
      </div>

      <div style={{
        display: 'flex',
        gap: 24,
        justifyContent: 'center',
        flexWrap: 'wrap',
        marginTop: 8,
      }}>
        {/* Own fleet */}
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontWeight: 700, marginBottom: 8, color: 'var(--text-muted)', fontSize: 13 }}>
            Your fleet — {myShipsLeft} cells left
          </div>
          <BattleGrid cells={myBoard} isOwn={true} clickable={false} />
        </div>

        {/* Enemy waters */}
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontWeight: 700, marginBottom: 8, color: 'var(--text-muted)', fontSize: 13 }}>
            Enemy waters — {oppShipsLeft} cells remaining
          </div>
          <BattleGrid
            cells={myShots}
            isOwn={false}
            clickable={canShoot}
            onCellClick={handleShot}
          />
        </div>
      </div>

      <div style={{ display: 'flex', gap: 24, justifyContent: 'center', marginTop: 12, fontSize: 13, color: 'var(--text-muted)', flexWrap: 'wrap' }}>
        <span><span style={{ color: '#546e7a' }}>■</span> Ship</span>
        <span><span style={{ color: '#e64a19' }}>⊕</span> Hit</span>
        <span><span style={{ color: '#29b6f6' }}>·</span> Miss</span>
        <span><span style={{ color: '#0d47a1' }}>■</span> Water</span>
      </div>
    </div>
  )
}
