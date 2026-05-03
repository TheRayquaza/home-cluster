import { useEffect, useRef, useCallback } from 'react'
import type { GameProps, LightCyclesState } from '../../api/types'

const CELL_SIZE = 15
const PLAYER_COLORS = ['#e94560', '#4fc3f7']
const TRAIL_COLORS = ['rgba(233,69,96,0.5)', 'rgba(79,195,247,0.5)']

export function LightCycles({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as LightCyclesState
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const lastDirRef = useRef<string>('')

  // Draw
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !s) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const W = (s.width || 40) * CELL_SIZE
    const H = (s.height || 30) * CELL_SIZE
    canvas.width = W
    canvas.height = H

    ctx.fillStyle = '#0a0a14'
    ctx.fillRect(0, 0, W, H)

    ctx.strokeStyle = 'rgba(255,255,255,0.03)'
    ctx.lineWidth = 0.5
    for (let x = 0; x <= s.width; x++) {
      ctx.beginPath(); ctx.moveTo(x * CELL_SIZE, 0); ctx.lineTo(x * CELL_SIZE, H); ctx.stroke()
    }
    for (let y = 0; y <= s.height; y++) {
      ctx.beginPath(); ctx.moveTo(0, y * CELL_SIZE); ctx.lineTo(W, y * CELL_SIZE); ctx.stroke()
    }

    s.trails?.forEach((trail, pIdx) => {
      ctx.fillStyle = TRAIL_COLORS[pIdx]
      trail.forEach(([tx, ty]) => {
        ctx.fillRect(tx * CELL_SIZE + 1, ty * CELL_SIZE + 1, CELL_SIZE - 2, CELL_SIZE - 2)
      })
    })

    s.players?.forEach((player, pIdx) => {
      if (!s.alive?.[pIdx]) {
        ctx.strokeStyle = 'rgba(255,255,255,0.4)'
        ctx.lineWidth = 2
        const x = player.x * CELL_SIZE, y = player.y * CELL_SIZE
        ctx.beginPath()
        ctx.moveTo(x + 2, y + 2); ctx.lineTo(x + CELL_SIZE - 2, y + CELL_SIZE - 2)
        ctx.moveTo(x + CELL_SIZE - 2, y + 2); ctx.lineTo(x + 2, y + CELL_SIZE - 2)
        ctx.stroke()
        return
      }
      ctx.fillStyle = PLAYER_COLORS[pIdx]
      ctx.shadowColor = PLAYER_COLORS[pIdx]
      ctx.shadowBlur = 8
      ctx.fillRect(player.x * CELL_SIZE + 1, player.y * CELL_SIZE + 1, CELL_SIZE - 2, CELL_SIZE - 2)
      ctx.shadowBlur = 0
      if (pIdx === playerIdx) {
        ctx.strokeStyle = 'white'; ctx.lineWidth = 1.5
        ctx.strokeRect(player.x * CELL_SIZE + 1, player.y * CELL_SIZE + 1, CELL_SIZE - 2, CELL_SIZE - 2)
      }
    })
  }, [s, playerIdx])

  // Keyboard
  useEffect(() => {
    if (gameOver) return
    const onKeyDown = (e: KeyboardEvent) => {
      let dir: string | null = null
      if (playerIdx === 0) {
        if (e.key === 'w' || e.key === 'W') dir = 'up'
        else if (e.key === 's' || e.key === 'S') dir = 'down'
        else if (e.key === 'a' || e.key === 'A') dir = 'left'
        else if (e.key === 'd' || e.key === 'D') dir = 'right'
      } else {
        if (e.key === 'ArrowUp') dir = 'up'
        else if (e.key === 'ArrowDown') dir = 'down'
        else if (e.key === 'ArrowLeft') dir = 'left'
        else if (e.key === 'ArrowRight') dir = 'right'
      }
      if (dir && dir !== lastDirRef.current) {
        e.preventDefault()
        lastDirRef.current = dir
        onAction({ dir })
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [playerIdx, onAction, gameOver])

  const sendDir = useCallback((dir: string) => (e: React.PointerEvent) => {
    e.preventDefault()
    if (dir !== lastDirRef.current) {
      lastDirRef.current = dir
      onAction({ dir })
    }
  }, [onAction])

  const alive = s?.alive ?? [true, true]
  const color = playerIdx === 0 ? 'var(--p1)' : 'var(--p2)'

  const dBtn = (dir: string, label: string): React.ReactNode => (
    <button
      onPointerDown={sendDir(dir)}
      style={{
        fontSize: 24, padding: '14px 20px',
        background: color, color: '#fff', border: 'none',
        borderRadius: 10, cursor: 'pointer', touchAction: 'none',
        userSelect: 'none', WebkitUserSelect: 'none',
        minWidth: 56, minHeight: 56,
      }}
    >{label}</button>
  )

  return (
    <div className="game-container">
      <div className="game-info">
        <div style={{ color, fontWeight: 700 }}>
          {alive[playerIdx] ? 'Alive' : 'Eliminated'}
        </div>
        <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>
          {playerIdx === 0 ? 'WASD · or D-pad below' : 'Arrow keys · or D-pad below'}
        </div>
      </div>

      <canvas
        ref={canvasRef}
        style={{ maxWidth: '100%', height: 'auto', display: 'block', margin: '0 auto' }}
      />

      {/* D-pad */}
      <div style={{ display: 'grid', gridTemplateColumns: '56px 56px 56px', gap: 6, margin: '16px auto 0', width: 'fit-content' }}>
        <div />
        {dBtn('up', '▲')}
        <div />
        {dBtn('left', '◀')}
        {dBtn('down', '▼')}
        {dBtn('right', '▶')}
      </div>

      <div style={{ display: 'flex', gap: 24, fontSize: 13, color: 'var(--text-muted)', marginTop: 8 }}>
        <span style={{ color: 'var(--p1)' }}>● P1 Red</span>
        <span style={{ color: 'var(--p2)' }}>● P2 Blue</span>
      </div>
    </div>
  )
}
