import { useState, useEffect, useCallback, useRef } from 'react'
import type { GameProps, WordleState } from '../../api/types'

const KEYBOARD_ROWS = [
  ['Q','W','E','R','T','Y','U','I','O','P'],
  ['A','S','D','F','G','H','J','K','L'],
  ['ENTER','Z','X','C','V','B','N','M','⌫'],
]

export function Wordle({ state, playerIdx, onAction, gameOver }: GameProps) {
  const s = state as WordleState
  const wordLen = s.word_length ?? 5
  const [input, setInput] = useState('')
  const [shake, setShake] = useState(false)
  const _inputRef = useRef<HTMLInputElement>(null)

  const myAttempts = s.attempts?.[playerIdx] ?? 0
  const mySolved = s.solved?.[playerIdx] ?? false
  // guesses[playerIdx] = array of guess strings
  const myGuesses: string[] = (s.guesses?.[playerIdx] ?? []) as unknown as string[]
  // hints[playerIdx][guessIdx][charIdx]
  const myHints: number[][] = s.hints?.[playerIdx] ?? []
  const oppIdx = playerIdx === 0 ? 1 : 0
  const oppGuesses: string[] = (s.guesses?.[oppIdx] ?? []) as unknown as string[]
  const oppHints: number[][] = s.hints?.[oppIdx] ?? []
  const canGuess = !mySolved && !s.game_over && !gameOver && myAttempts < 6

  // Build keyboard letter states from my guesses
  const letterStates: Record<string, number> = {}
  myGuesses.forEach((guess, gi) => {
    const hints = myHints[gi] ?? []
    for (let li = 0; li < guess.length; li++) {
      const letter = guess[li]
      const h = hints[li] ?? 0
      const existing = letterStates[letter] ?? -1
      if (h > existing) letterStates[letter] = h
    }
  })

  const submitGuess = useCallback(() => {
    if (input.length !== wordLen) {
      setShake(true)
      setTimeout(() => setShake(false), 400)
      return
    }
    onAction({ guess: input.toUpperCase() })
    setInput('')
  }, [input, wordLen, onAction])

  const handleKey = useCallback((key: string) => {
    if (!canGuess) return
    if (key === 'ENTER') {
      submitGuess()
    } else if (key === '⌫' || key === 'Backspace') {
      setInput(p => p.slice(0, -1))
    } else if (/^[A-Za-z]$/.test(key) && input.length < wordLen) {
      setInput(p => p + key.toUpperCase())
    }
  }, [canGuess, input, wordLen, submitGuess])

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      handleKey(e.key === 'Backspace' ? 'Backspace' : e.key)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [handleKey])

  const renderGrid = (
    guesses: string[],
    hints: number[][],
    attempts: number,
    isMe: boolean,
    isSolved: boolean
  ) => {
    return (
      <div className="wordle-grid" style={{ gridTemplateColumns: `repeat(${wordLen}, 1fr)` }}>
        {Array.from({ length: 6 }, (_, row) => {
          const guess: string = guesses[row] ?? ''
          const rowHints: number[] = hints[row] ?? []
          const isCurrentRow = isMe && row === attempts && !isSolved && !s.game_over

          return Array.from({ length: wordLen }, (_, col) => {
            const letter = isCurrentRow ? (input[col] ?? '') : (guess[col] ?? '')
            const hint: number | undefined = (!isCurrentRow && guess.length > 0) ? rowHints[col] : undefined

            const bgColor = hint === undefined
              ? 'var(--bg-input)'
              : hint === 2 ? '#538d4e'
              : hint === 1 ? '#b59f3b'
              : '#3a3a3c'

            const borderColor = isCurrentRow && letter
              ? 'var(--text-muted)'
              : hint !== undefined
                ? 'transparent'
                : 'var(--border)'

            const textColor = hint !== undefined ? 'white' : 'var(--text)'
            // For opponent: show cell colors but hide letters
            const displayLetter = isMe ? letter : ''

            return (
              <div
                key={`${row}-${col}`}
                className="wordle-cell"
                style={{
                  border: `2px solid ${borderColor}`,
                  color: textColor,
                  background: bgColor,
                  animation: shake && isCurrentRow ? 'shake 0.4s' : undefined,
                  fontWeight: 800,
                }}
              >
                {displayLetter}
              </div>
            )
          })
        })}
      </div>
    )
  }

  return (
    <div className="game-container" style={{ paddingBottom: 200 }}>
      <div className="game-info">
        {canGuess ? (
          <div className="turn-indicator your-turn">Your turn to guess</div>
        ) : mySolved ? (
          <div className="turn-indicator" style={{
            background: 'rgba(76,175,80,0.2)',
            borderColor: 'var(--success)',
            color: 'var(--success)',
            border: '1px solid var(--success)',
          }}>
            Solved in {myAttempts}!
          </div>
        ) : (
          <div className="turn-indicator waiting">Waiting...</div>
        )}
      </div>

      <div className="wordle-container">
        {/* My grid */}
        <div className="wordle-player">
          <h3>You (P{playerIdx + 1})</h3>
          <div style={{ animation: shake ? 'shake 0.4s' : undefined }}>
            {renderGrid(myGuesses, myHints, myAttempts, true, mySolved)}
          </div>
          {canGuess && (
            <div className="wordle-input" style={{ marginTop: 8 }}>
              <input
                ref={_inputRef}
                value={input}
                onChange={e => {
                  const v = e.target.value.toUpperCase().replace(/[^A-Z]/g, '')
                  if (v.length <= wordLen) setInput(v)
                }}
                onKeyDown={e => {
                  if (e.key === 'Enter') submitGuess()
                }}
                placeholder="GUESS"
                maxLength={wordLen}
                style={{ letterSpacing: 6, fontWeight: 800, fontSize: 18, textAlign: 'center' }}
                autoFocus
              />
              <button
                className="btn-primary"
                onClick={submitGuess}
                disabled={input.length !== wordLen}
              >
                Enter
              </button>
            </div>
          )}
        </div>

        {/* Opponent grid */}
        <div className="wordle-player">
          <h3>Opponent (P{oppIdx + 1})</h3>
          {renderGrid(oppGuesses, oppHints, s.attempts?.[oppIdx] ?? 0, false, s.solved?.[oppIdx] ?? false)}
          {(s.solved?.[oppIdx]) && (
            <div style={{ color: 'var(--success)', fontWeight: 700, marginTop: 8 }}>
              Solved in {s.attempts?.[oppIdx]}!
            </div>
          )}
        </div>
      </div>

      {/* Revealed word banner */}
      {(gameOver || s.game_over) && s.revealed_word && (
        <div style={{ textAlign: 'center', marginTop: 16, padding: '12px 24px', background: 'var(--bg-card)', borderRadius: 12, border: '1px solid var(--border)' }}>
          <div style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 4 }}>The word was</div>
          <div style={{ fontSize: 28, fontWeight: 800, letterSpacing: 4 }}>{s.revealed_word}</div>
        </div>
      )}

      {/* On-screen keyboard — fixed footer */}
      <div className="wordle-keyboard" style={{
        position: 'fixed',
        bottom: 0,
        left: 0,
        right: 0,
        background: 'var(--bg-card)',
        padding: 8,
        borderTop: '1px solid var(--border)',
        zIndex: 100,
      }}>
        {KEYBOARD_ROWS.map((row, ri) => (
          <div key={ri} className="wordle-key-row">
            {row.map(key => {
              const letterState = letterStates[key]
              return (
                <button
                  key={key}
                  onClick={() => handleKey(key)}
                  disabled={!canGuess}
                  style={{
                    minWidth: key.length > 1 ? 56 : 32,
                    padding: '10px 6px',
                    background: letterState === 2 ? '#538d4e'
                      : letterState === 1 ? '#b59f3b'
                      : letterState === 0 ? '#3a3a3c'
                      : '#818384',
                    color: 'white',
                    borderRadius: 4,
                    fontSize: 12,
                    fontWeight: 700,
                    opacity: canGuess ? 1 : 0.5,
                    cursor: canGuess ? 'pointer' : 'not-allowed',
                    border: 'none',
                    transition: 'background 0.2s',
                  }}
                >
                  {key}
                </button>
              )
            })}
          </div>
        ))}
      </div>

      <style>{`
        @keyframes shake {
          0%, 100% { transform: translateX(0); }
          20% { transform: translateX(-8px); }
          40% { transform: translateX(8px); }
          60% { transform: translateX(-5px); }
          80% { transform: translateX(5px); }
        }
      `}</style>
    </div>
  )
}
