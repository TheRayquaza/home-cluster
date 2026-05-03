interface Props {
  idx: number
  username?: string
  isCurrent?: boolean
}

export function PlayerBadge({ idx, username, isCurrent = false }: Props) {
  const label = username ?? `Player ${idx + 1}`
  const symbol = idx === 0 ? '●' : '○'
  return (
    <span
      className={`player-badge p${idx + 1}`}
      style={isCurrent ? { boxShadow: '0 0 0 2px currentColor' } : undefined}
    >
      {symbol} {label}
      {isCurrent && ' (You)'}
    </span>
  )
}
