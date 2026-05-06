import { useState, useMemo } from 'react'
import type { GameProps, BattleshipsState } from '../../api/types'

// boards: 0=water, 1=ship, 2=hit, 3=miss
// shots:  0=unknown, 1=miss, 2=hit

function cellColor(value: number, isOwn: boolean): string {
  if (value === 2) return '#e64a19'
  if (value === 3) return '#29b6f6'
  if (!isOwn && value === 1) return '#29b6f6' // miss on shots grid
  if (isOwn && value === 1) return '#546e7a'  // own ship
  return '#0d47a1'
}

interface GridProps {
  cells: number[]
  isOwn: boolean
  clickable: boolean
  forceClickable?: boolean
  onCellClick?: (idx: number) => void
  onCellHover?: (idx: number | null) => void
  previewCells?: Map<number, boolean>
}

function BattleGrid({ cells, isOwn, clickable, forceClickable, onCellClick, onCellHover, previewCells }: GridProps) {
  return (
    <div style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(10, 1fr)',
      gap: 2,
      // 100vw avoids the flex align-items:center collapse that makes 100% resolve to 0
      width: 'min(320px, calc(100vw - 32px))',
    }}>
      {cells.map((val, idx) => {
        const canClick = forceClickable ? clickable : (clickable && val === 0)
        const previewState = previewCells?.get(idx)
        const bg = previewState === true
          ? 'rgba(76, 175, 80, 0.75)'
          : previewState === false
          ? 'rgba(233, 69, 96, 0.75)'
          : cellColor(val, isOwn)
        return (
          <div
            key={idx}
            onClick={() => canClick && onCellClick?.(idx)}
            onMouseEnter={e => {
              if (onCellHover) {
                onCellHover(idx)
              } else if (canClick) {
                (e.currentTarget as HTMLDivElement).style.filter = 'brightness(1.4)'
              }
            }}
            onMouseLeave={e => {
              if (onCellHover) {
                onCellHover(null)
              } else {
                (e.currentTarget as HTMLDivElement).style.filter = ''
              }
            }}
            style={{
              aspectRatio: '1',
              background: bg,
              borderRadius: 2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 14,
              fontWeight: 700,
              color: 'white',
              cursor: canClick ? 'crosshair' : 'default',
              transition: 'background 0.08s',
            }}
          >
            {val === 2 ? '⊕' : (val === 3 || (!isOwn && val === 1)) ? '·' : ''}
          </div>
        )
      })}
    </div>
  )
}

const SHIP_NAMES: Record<number, string> = { 5: 'Carrier', 4: 'Battleship', 3: 'Cruiser', 2: 'Destroyer' }

export function Battleships({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as BattleshipsState
  const oppIdx = playerIdx === 0 ? 1 : 0

  const [orientation, setOrientation] = useState<'H' | 'V'>('H')
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)

  const myBoard: number[] = s.boards?.[playerIdx] ?? Array(100).fill(0)
  const myShots: number[] = s.shots?.[playerIdx] ?? Array(100).fill(0)
  const toPlace: number[] = s.to_place?.[playerIdx] ?? []
  const nextSize: number | null = toPlace.length > 0 ? toPlace[0] : null

  const isMyTurn = s.turn === playerIdx
  const iAmReady = s.ready?.[playerIdx] ?? false
  const oppReady = s.ready?.[oppIdx] ?? false
  const myShipsLeft = s.ships_left?.[playerIdx] ?? 0
  const oppShipsLeft = s.ships_left?.[oppIdx] ?? 0

  const canPlace = s.phase === 'placement' && !iAmReady && nextSize !== null

  // Compute preview for ship placement hover
  const previewCells = useMemo((): Map<number, boolean> => {
    if (hoverIdx === null || !canPlace) return new Map()
    const x0 = hoverIdx % 10
    const y0 = Math.floor(hoverIdx / 10)
    const cells: number[] = []
    let valid = true

    if (orientation === 'H') {
      if (x0 + nextSize! > 10) valid = false
      for (let i = 0; i < nextSize!; i++) {
        const idx = y0 * 10 + x0 + i
        cells.push(idx)
        if (myBoard[idx] !== 0) valid = false
      }
    } else {
      if (y0 + nextSize! > 10) valid = false
      for (let i = 0; i < nextSize!; i++) {
        const idx = (y0 + i) * 10 + x0
        cells.push(idx)
        if (myBoard[idx] !== 0) valid = false
      }
    }

    const result = new Map<number, boolean>()
    for (const idx of cells) result.set(idx, valid)
    return result
  }, [hoverIdx, canPlace, orientation, nextSize, myBoard])

  const handlePlacementClick = (idx: number) => {
    if (!canPlace) return
    const x = idx % 10
    const y = Math.floor(idx / 10)
    onAction({ action: 'place', x, y, horizontal: orientation === 'H' })
  }

  // ─── Placement phase ──────────────────────────────────────────────────────
  if (s.phase === 'placement') {
    const shipSizes: number[] = s.ship_sizes ?? [5, 4, 3, 3, 2]
    const placedCount = shipSizes.length - toPlace.length

    return (
      <div className="game-container">
        <h2 style={{ textAlign: 'center' }}>Ship Placement</h2>

        {/* Ship queue */}
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', justifyContent: 'center' }}>
          {shipSizes.map((size, i) => {
            const placed = i < placedCount
            const current = i === placedCount && !iAmReady
            return (
              <div
                key={i}
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: 4,
                  opacity: placed ? 0.35 : 1,
                }}
              >
                <div style={{
                  fontSize: 11,
                  color: current ? 'var(--success)' : 'var(--text-muted)',
                  fontWeight: current ? 700 : 400,
                }}>
                  {placed ? '✓' : current ? '→' : ''} {SHIP_NAMES[size] ?? `${size}-cell`}
                </div>
                <div style={{ display: 'flex', gap: 2 }}>
                  {Array.from({ length: size }).map((_, j) => (
                    <div key={j} style={{
                      width: 14,
                      height: 14,
                      background: placed ? '#555' : current ? 'var(--success)' : '#546e7a',
                      borderRadius: 2,
                    }} />
                  ))}
                </div>
              </div>
            )
          })}
        </div>

        {/* Grid */}
        <BattleGrid
          cells={myBoard}
          isOwn={true}
          clickable={canPlace}
          forceClickable={canPlace}
          onCellClick={handlePlacementClick}
          onCellHover={canPlace ? setHoverIdx : undefined}
          previewCells={canPlace ? previewCells : undefined}
        />

        {/* Orientation + actions */}
        {!iAmReady && (
          <div style={{ display: 'flex', gap: 10, justifyContent: 'center', flexWrap: 'wrap' }}>
            {nextSize !== null && (
              <button
                className="btn-secondary"
                onClick={() => setOrientation(o => o === 'H' ? 'V' : 'H')}
                style={{ fontSize: 14, minWidth: 120 }}
              >
                {orientation === 'H' ? '↔ Horizontal' : '↕ Vertical'}
              </button>
            )}
            <button
              className="btn-ghost"
              onClick={() => onAction({ action: 'clear' })}
              style={{ fontSize: 14 }}
            >
              🗑 Clear
            </button>
            <button
              className="btn-primary"
              onClick={() => onAction({ action: 'randomize' })}
              style={{ fontSize: 14 }}
            >
              🔀 Randomize
            </button>
            <button
              className="btn-primary"
              onClick={() => onAction({ action: 'ready' })}
              disabled={toPlace.length > 0}
              style={{
                fontSize: 14,
                background: toPlace.length === 0 ? 'var(--success)' : undefined,
              }}
            >
              ✓ Ready
            </button>
          </div>
        )}

        <div style={{ textAlign: 'center', color: 'var(--text-muted)', fontSize: 14 }}>
          {iAmReady && !oppReady && 'Waiting for opponent...'}
          {iAmReady && oppReady && 'Both ready — starting!'}
          {!iAmReady && nextSize !== null && `Click grid to place ${SHIP_NAMES[nextSize] ?? `${nextSize}-cell ship`} (${nextSize} cells)`}
          {!iAmReady && nextSize === null && 'All ships placed — press Ready!'}
        </div>
      </div>
    )
  }

  // ─── Combat phase ─────────────────────────────────────────────────────────
  const canShoot = isMyTurn && !gameOver

  return (
    <div className="game-container">
      <div className="game-info">
        {isMyTurn ? (
          <div className="turn-indicator your-turn">Your turn — Fire!</div>
        ) : (
          <div className="turn-indicator waiting">Opponent's turn</div>
        )}
      </div>

      <div style={{ display: 'flex', gap: 24, justifyContent: 'center', flexWrap: 'wrap' }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontWeight: 700, marginBottom: 8, color: 'var(--text-muted)', fontSize: 13 }}>
            Your fleet — {myShipsLeft} cells left
          </div>
          <BattleGrid cells={myBoard} isOwn={true} clickable={false} />
        </div>

        <div style={{ textAlign: 'center' }}>
          <div style={{ fontWeight: 700, marginBottom: 8, color: 'var(--text-muted)', fontSize: 13 }}>
            Enemy waters — {oppShipsLeft} cells remaining
          </div>
          <BattleGrid
            cells={myShots}
            isOwn={false}
            clickable={canShoot}
            onCellClick={idx => {
              const x = idx % 10
              const y = Math.floor(idx / 10)
              onAction({ x, y })
            }}
          />
        </div>
      </div>

      <div style={{ display: 'flex', gap: 24, justifyContent: 'center', fontSize: 13, color: 'var(--text-muted)', flexWrap: 'wrap' }}>
        <span><span style={{ color: '#546e7a' }}>■</span> Ship</span>
        <span><span style={{ color: '#e64a19' }}>⊕</span> Hit</span>
        <span><span style={{ color: '#29b6f6' }}>·</span> Miss</span>
        <span><span style={{ color: '#0d47a1' }}>■</span> Water</span>
      </div>
    </div>
  )
}
