import { useEffect, useRef, useCallback } from 'react'
import type { GameProps, PongState } from '../../api/types'

export function Pong({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as PongState
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const keysRef = useRef<Set<string>>(new Set())
  const animRef = useRef<number>(0)
  const lastDirRef = useRef<string>('stop')
  const touchDirRef = useRef<string>('stop')

  // Draw
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !s) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const W = s.width || 800
    const H = s.height || 400
    const paddleH = s.paddle_h || 80
    const paddleW = s.paddle_w || 12

    canvas.width = W
    canvas.height = H

    ctx.fillStyle = '#0a0a14'
    ctx.fillRect(0, 0, W, H)

    ctx.setLineDash([8, 8])
    ctx.strokeStyle = '#2a2a4a'
    ctx.lineWidth = 2
    ctx.beginPath()
    ctx.moveTo(W / 2, 0)
    ctx.lineTo(W / 2, H)
    ctx.stroke()
    ctx.setLineDash([])

    ctx.font = 'bold 48px monospace'
    ctx.textAlign = 'center'
    ctx.fillStyle = 'rgba(224,224,224,0.3)'
    ctx.fillText(String(s.scores?.[0] ?? 0), W / 4, 60)
    ctx.fillText(String(s.scores?.[1] ?? 0), (3 * W) / 4, 60)

    const paddles = s.paddles ?? [H / 2 - paddleH / 2, H / 2 - paddleH / 2]
    ctx.fillStyle = '#e94560'
    ctx.beginPath()
    ctx.roundRect(8, paddles[0], paddleW, paddleH, 4)
    ctx.fill()

    ctx.fillStyle = '#4fc3f7'
    ctx.beginPath()
    ctx.roundRect(W - paddleW - 8, paddles[1], paddleW, paddleH, 4)
    ctx.fill()

    if (s.ball) {
      ctx.fillStyle = 'white'
      ctx.shadowColor = 'white'
      ctx.shadowBlur = 8
      ctx.beginPath()
      ctx.arc(s.ball.x, s.ball.y, 8, 0, Math.PI * 2)
      ctx.fill()
      ctx.shadowBlur = 0
    }

    ctx.font = 'bold 13px sans-serif'
    ctx.fillStyle = playerIdx === 0 ? '#e94560' : '#4fc3f7'
    ctx.textAlign = playerIdx === 0 ? 'left' : 'right'
    ctx.fillText(
      playerIdx === 0 ? '◀ You' : 'You ▶',
      playerIdx === 0 ? 16 : W - 16,
      H - 12
    )
  }, [s, playerIdx])

  // Keyboard + tick loop
  useEffect(() => {
    if (gameOver) return

    const getDir = (): string => {
      const t = touchDirRef.current
      if (t !== 'stop') return t
      if (playerIdx === 0) {
        if (keysRef.current.has('w') || keysRef.current.has('W')) return 'up'
        if (keysRef.current.has('s') || keysRef.current.has('S')) return 'down'
      } else {
        if (keysRef.current.has('ArrowUp')) return 'up'
        if (keysRef.current.has('ArrowDown')) return 'down'
      }
      return 'stop'
    }

    const tick = () => {
      const dir = getDir()
      if (dir !== lastDirRef.current) {
        lastDirRef.current = dir
        onAction({ dir })
      }
      animRef.current = requestAnimationFrame(tick)
    }

    const onKeyDown = (e: KeyboardEvent) => {
      const relevant = ['w', 'W', 's', 'S', 'ArrowUp', 'ArrowDown']
      if (relevant.includes(e.key)) {
        e.preventDefault()
        keysRef.current.add(e.key)
      }
    }
    const onKeyUp = (e: KeyboardEvent) => keysRef.current.delete(e.key)

    window.addEventListener('keydown', onKeyDown)
    window.addEventListener('keyup', onKeyUp)
    animRef.current = requestAnimationFrame(tick)

    return () => {
      window.removeEventListener('keydown', onKeyDown)
      window.removeEventListener('keyup', onKeyUp)
      cancelAnimationFrame(animRef.current)
    }
  }, [playerIdx, onAction, gameOver])

  const press = useCallback((dir: string) => (e: React.PointerEvent) => {
    e.preventDefault()
    touchDirRef.current = dir
  }, [])

  const release = useCallback((e: React.PointerEvent) => {
    e.preventDefault()
    touchDirRef.current = 'stop'
  }, [])

  const btnStyle: React.CSSProperties = {
    fontSize: 32,
    padding: '18px 36px',
    background: playerIdx === 0 ? 'var(--p1)' : 'var(--p2)',
    color: '#fff',
    border: 'none',
    borderRadius: 12,
    cursor: 'pointer',
    userSelect: 'none',
    touchAction: 'none',
    WebkitUserSelect: 'none',
  }

  return (
    <div className="game-container">
      <div className="game-info">
        <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>
          {playerIdx === 0 ? 'W / S · or hold buttons below' : '↑ / ↓ · or hold buttons below'}
        </div>
        <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>First to 7 wins</div>
      </div>
      <canvas
        ref={canvasRef}
        style={{ maxWidth: '100%', height: 'auto', display: 'block', margin: '0 auto' }}
      />
      <div style={{ display: 'flex', gap: 24, justifyContent: 'center', marginTop: 16 }}>
        <button
          style={btnStyle}
          onPointerDown={press('up')}
          onPointerUp={release}
          onPointerLeave={release}
        >▲</button>
        <button
          style={btnStyle}
          onPointerDown={press('down')}
          onPointerUp={release}
          onPointerLeave={release}
        >▼</button>
      </div>
    </div>
  )
}
