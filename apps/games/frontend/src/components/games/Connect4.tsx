import type { GameProps, Connect4State } from '../../api/types'

const COLS = 7
const ROWS = 6

export function Connect4({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as Connect4State
  const isMyTurn = s.turn === playerIdx && !gameOver

  const handleColClick = (col: number) => {
    if (!isMyTurn) return
    onAction({ col })
  }

  // Backend: board[col][row], row 0 = bottom.
  // Display: row 0 = top, so flip: backendRow = ROWS-1-displayRow.
  const getCell = (displayRow: number, col: number): number => {
    if (!s.board || !s.board[col]) return 0
    return s.board[col][ROWS - 1 - displayRow] ?? 0
  }

  // Detect winning cells (simple check)
  const winnerCells = new Set<string>()
  if (s.winner > 0) {
    // Check all directions for the winner
    const w = s.winner
    const directions = [[0,1],[1,0],[1,1],[1,-1]]
    for (let r = 0; r < ROWS; r++) {
      for (let c = 0; c < COLS; c++) {
        if (getCell(r, c) !== w) continue
        for (const [dr, dc] of directions) {
          const cells: string[] = []
          let valid = true
          for (let k = 0; k < 4; k++) {
            const nr = r + dr * k
            const nc = c + dc * k
            if (nr < 0 || nr >= ROWS || nc < 0 || nc >= COLS || getCell(nr, nc) !== w) {
              valid = false
              break
            }
            cells.push(`${nr},${nc}`)
          }
          if (valid) cells.forEach(k => winnerCells.add(k))
        }
      }
    }
  }

  return (
    <div className="game-container">
      <div className="game-info">
        <div className={`turn-indicator ${isMyTurn ? 'your-turn' : 'waiting'}`}>
          {isMyTurn ? 'Your turn' : "Opponent's turn"}
        </div>
        <div style={{ color: 'var(--text-muted)', fontSize: 14 }}>
          You are{' '}
          <strong style={{ color: playerIdx === 0 ? '#e53935' : '#fdd835' }}>
            {playerIdx === 0 ? 'Red' : 'Yellow'}
          </strong>
        </div>
      </div>

      <div>
        {/* Column drop buttons */}
        <div style={{ display: 'grid', gridTemplateColumns: `repeat(${COLS}, 60px)`, gap: 4, marginBottom: 4 }}>
          {Array.from({ length: COLS }, (_, c) => (
            <button
              key={c}
              className="c4-col-btn"
              onClick={() => handleColClick(c)}
              disabled={!isMyTurn}
              style={{ opacity: isMyTurn ? 1 : 0.3 }}
            >
              ▼
            </button>
          ))}
        </div>

        {/* Board */}
        <div className="c4-board">
          <div className="c4-grid">
            {Array.from({ length: ROWS }, (_, row) =>
              Array.from({ length: COLS }, (_, col) => {
                const cell = getCell(row, col)
                const key = `${row},${col}`
                const isWin = winnerCells.has(key)
                const cls = [
                  'c4-cell',
                  cell === 1 ? 'p1' : cell === 2 ? 'p2' : '',
                  isWin ? 'winner-cell' : '',
                ]
                  .filter(Boolean)
                  .join(' ')
                return <div key={key} className={cls} />
              })
            )}
          </div>
        </div>
      </div>

      <div style={{ color: 'var(--text-muted)', fontSize: 13 }}>
        Click ▼ to drop a piece in that column
      </div>
    </div>
  )
}
