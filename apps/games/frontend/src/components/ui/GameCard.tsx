import type { Game } from '../../api/types'

interface GameCardProps {
  game: Game
  onPlay: () => void
  isDaily?: boolean
  compact?: boolean
  selected?: boolean
  onClick?: () => void
}

export function GameCard({ game, onPlay, isDaily = false, compact = false, selected = false, onClick }: GameCardProps) {
  const cardStyle: React.CSSProperties = {
    background: `linear-gradient(135deg, ${game.color}22, ${game.color}11)`,
    border: selected
      ? `2px solid ${game.color}`
      : `1px solid ${game.color}44`,
    borderRadius: 16,
    padding: compact ? 16 : 24,
    display: 'flex',
    flexDirection: 'column',
    gap: compact ? 8 : 12,
    cursor: onClick ? 'pointer' : 'default',
    transition: 'border-color 0.15s, box-shadow 0.15s',
    boxShadow: selected ? `0 0 0 1px ${game.color}66` : undefined,
  }

  return (
    <div style={cardStyle} onClick={onClick}>
      <div style={{ fontSize: compact ? 36 : 56, textAlign: 'center' }}>{game.emoji}</div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <h3 style={{ margin: 0, fontSize: compact ? 14 : 16 }}>{game.name}</h3>
        {isDaily && (
          <span style={{
            background: 'var(--accent)',
            color: '#fff',
            fontSize: 11,
            padding: '2px 8px',
            borderRadius: 20,
            fontWeight: 700,
            whiteSpace: 'nowrap',
          }}>
            TODAY
          </span>
        )}
      </div>
      {!compact && (
        <p style={{ color: 'var(--text-muted)', fontSize: 14, margin: 0 }}>{game.description}</p>
      )}
      {compact && (
        <p style={{ color: 'var(--text-muted)', fontSize: 12, margin: 0, lineHeight: 1.4 }}>{game.description}</p>
      )}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span style={{
          fontSize: 12,
          color: game.real_time ? '#4fc3f7' : '#aaa',
          background: 'rgba(255,255,255,0.07)',
          padding: '3px 10px',
          borderRadius: 20,
        }}>
          {game.real_time ? '⚡ Real-time' : '♟ Turn-based'}
        </span>
        <button
          className="btn-primary"
          style={{ background: game.color, padding: compact ? '6px 14px' : '8px 20px', fontSize: compact ? 12 : 14 }}
          onClick={e => { e.stopPropagation(); onPlay() }}
        >
          Play
        </button>
      </div>
    </div>
  )
}
