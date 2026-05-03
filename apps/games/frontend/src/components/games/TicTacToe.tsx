import type { GameProps, TicTacToeState } from '../../api/types'

const WINNING_LINES = [
  [0, 1, 2], [3, 4, 5], [6, 7, 8], // rows
  [0, 3, 6], [1, 4, 7], [2, 5, 8], // cols
  [0, 4, 8], [2, 4, 6],             // diagonals
]

function getWinnerLine(board: number[]): number[] | null {
  for (const line of WINNING_LINES) {
    const [a, b, c] = line
    if (board[a] !== 0 && board[a] === board[b] && board[b] === board[c]) {
      return line
    }
  }
  return null
}

export function TicTacToe({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as TicTacToeState
  const winnerLine = getWinnerLine(s.board)
  const isMyTurn = s.turn === playerIdx && !gameOver

  const handleClick = (idx: number) => {
    if (!isMyTurn || s.board[idx] !== 0) return
    onAction({ cell: idx })
  }

  return (
    <div className="game-container">
      <div className="game-info">
        <div
          className={`turn-indicator ${isMyTurn ? 'your-turn' : 'waiting'}`}
        >
          {isMyTurn ? 'Your turn' : "Opponent's turn"}
        </div>
        <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>
          You are <strong style={{ color: playerIdx === 0 ? 'var(--p1)' : 'var(--p2)' }}>
            {playerIdx === 0 ? 'X' : 'O'}
          </strong>
        </div>
      </div>

      <div className="ttt-board">
        {s.board.map((cell, i) => {
          const isWinner = winnerLine?.includes(i) ?? false
          const cls = [
            'ttt-cell',
            cell !== 0 ? 'taken' : '',
            cell === 1 ? 'p1' : cell === 2 ? 'p2' : '',
            isWinner ? 'winner-cell' : '',
          ]
            .filter(Boolean)
            .join(' ')

          return (
            <div
              key={i}
              className={cls}
              onClick={() => handleClick(i)}
            >
              {cell === 1 ? 'X' : cell === 2 ? 'O' : ''}
            </div>
          )
        })}
      </div>

      <div style={{ color: 'var(--text-muted)', fontSize: 13 }}>
        Board cells: 1=X (P1) · 2=O (P2)
      </div>
    </div>
  )
}
