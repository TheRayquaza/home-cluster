import { useNavigate } from 'react-router-dom'
import type { GameOverPayload } from '../../api/types'

interface Props {
  gameOver: GameOverPayload
  playerIdx: number
}

export function GameOver({ gameOver, playerIdx }: Props) {
  const navigate = useNavigate()
  const { winner } = gameOver

  let title: string
  let subtitle: string
  let emoji: string

  if (winner === -1) {
    title = "It's a Draw!"
    subtitle = 'Well played by both!'
    emoji = '🤝'
  } else if (winner === playerIdx) {
    title = 'You Win!'
    subtitle = 'Congratulations!'
    emoji = '🏆'
  } else {
    title = 'You Lose'
    subtitle = 'Better luck next time!'
    emoji = '😔'
  }

  return (
    <div className="game-over-overlay">
      <div className="game-over-box">
        <div style={{ fontSize: 64, marginBottom: 8 }}>{emoji}</div>
        <h1>{title}</h1>
        <p>{subtitle}</p>
        <div className="game-over-actions">
          <button
            className="btn-primary"
            onClick={() => navigate('/')}
          >
            Back to Lobby
          </button>
          <button
            className="btn-ghost"
            onClick={() => navigate('/scores')}
          >
            Leaderboard
          </button>
        </div>
      </div>
    </div>
  )
}
