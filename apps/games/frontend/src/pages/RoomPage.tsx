import { useEffect, useState, useContext } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import { useGame } from '../hooks/useGame'
import type { Room } from '../api/types'
import { GameOver } from '../components/ui/GameOver'
import { TicTacToe } from '../components/games/TicTacToe'
import { Connect4 } from '../components/games/Connect4'
import { MemoryMatch } from '../components/games/MemoryMatch'
import { Nim } from '../components/games/Nim'
import { Pong } from '../components/games/Pong'
import { LightCycles } from '../components/games/LightCycles'
import { Wordle } from '../components/games/Wordle'
import { AuthContext } from '../App'

export function RoomPage() {
  const { code } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const { user } = useContext(AuthContext)
  const [roomInfo, setRoomInfo] = useState<Room | null>(null)
  const [roomError, setRoomError] = useState('')

  const { playerIdx, gameId, state, gameOver, connected, error: wsError, send } = useGame(code ?? '')

  // Poll room info to show waiting state before WS game starts
  useEffect(() => {
    if (!code) return
    let interval: ReturnType<typeof setInterval>

    const fetchRoom = () => {
      api.getRoom(code)
        .then(r => {
          setRoomInfo(r)
          if (r.players_joined === 2) clearInterval(interval)
        })
        .catch(e => setRoomError(e instanceof Error ? e.message : 'Room not found'))
    }

    fetchRoom()
    interval = setInterval(fetchRoom, 2000)
    return () => clearInterval(interval)
  }, [code])

  if (roomError) {
    return (
      <div className="page" style={{ textAlign: 'center' }}>
        <div className="error-msg" style={{ maxWidth: 400, margin: '40px auto' }}>
          {roomError}
        </div>
        <button className="btn-primary" onClick={() => navigate('/')}>
          Back to Lobby
        </button>
      </div>
    )
  }

  const isWaiting = !state && !wsError

  return (
    <div>
      {/* Nav */}
      <nav className="nav">
        <div className="nav-brand">🎮 Mini Games</div>
        <div className="nav-links">
          <span style={{ color: 'var(--text-muted)', fontSize: 14 }}>
            {user?.username}
          </span>
          <button
            className="btn-ghost"
            onClick={() => navigate('/')}
            style={{ padding: '6px 14px', fontSize: 13 }}
          >
            Leave
          </button>
        </div>
      </nav>

      {/* Room code header */}
      <div style={{ background: 'var(--bg-card)', borderBottom: '1px solid var(--border)', padding: '8px 24px', display: 'flex', alignItems: 'center', gap: 16 }}>
        <span style={{ color: 'var(--text-muted)', fontSize: 13 }}>Room:</span>
        <span style={{ fontWeight: 800, letterSpacing: 4, color: 'var(--accent)', fontSize: 16 }}>{code}</span>
        {roomInfo && (
          <span style={{ color: 'var(--text-muted)', fontSize: 13 }}>
            · {roomInfo.players_joined}/2 players
          </span>
        )}
        {connected && (
          <span style={{
            width: 8, height: 8, borderRadius: '50%',
            background: 'var(--success)',
            display: 'inline-block',
          }} title="Connected" />
        )}
        {playerIdx !== null && (
          <span style={{
            color: playerIdx === 0 ? 'var(--p1)' : 'var(--p2)',
            fontWeight: 600,
            fontSize: 13,
          }}>
            You are Player {playerIdx + 1}
          </span>
        )}
      </div>

      {/* WS error */}
      {wsError && (
        <div className="error-msg" style={{ margin: '16px auto', maxWidth: 600 }}>
          {wsError}
        </div>
      )}

      {/* Waiting for opponent */}
      {isWaiting && (
        <div className="waiting-room">
          <div className="spinner" />
          {roomInfo && roomInfo.players_joined < 2 ? (
            <>
              <h2>Waiting for opponent...</h2>
              <p style={{ color: 'var(--text-muted)' }}>
                Share this room code with a friend:
              </p>
              <div className="room-code-display">{code}</div>
              <p style={{ color: 'var(--text-muted)', fontSize: 14 }}>
                Both players must be connected to start the game.
              </p>
            </>
          ) : (
            <>
              <h2>Connecting to game...</h2>
              <p style={{ color: 'var(--text-muted)' }}>
                Setting up the game board...
              </p>
            </>
          )}
        </div>
      )}

      {/* Game area */}
      {state && playerIdx !== null && (
        <div style={{ padding: '16px 0' }}>
          <GameBoard
            gameId={gameId ?? (roomInfo?.game_id ?? '')}
            state={state}
            playerIdx={playerIdx}
            onAction={send}
            gameOver={gameOver}
          />
        </div>
      )}

      {/* Game over overlay */}
      {gameOver && playerIdx !== null && (
        <GameOver gameOver={gameOver} playerIdx={playerIdx} />
      )}
    </div>
  )
}

interface GameBoardProps {
  gameId: string
  state: object
  playerIdx: number
  onAction: (action: object) => void
  gameOver: { winner: number } | null
}

function GameBoard({ gameId, state, playerIdx, onAction, gameOver }: GameBoardProps) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const props = { state: state as any, playerIdx, onAction, gameOver }

  switch (gameId) {
    case 'tictactoe':
      return <TicTacToe {...props} />
    case 'connect4':
      return <Connect4 {...props} />
    case 'memory':
    case 'memorymatch':
      return <MemoryMatch {...props} />
    case 'nim':
      return <Nim {...props} />
    case 'pong':
      return <Pong {...props} />
    case 'lightcycles':
    case 'light_cycles':
      return <LightCycles {...props} />
    case 'wordle':
      return <Wordle {...props} />
    default:
      return (
        <div className="page" style={{ textAlign: 'center' }}>
          <p style={{ color: 'var(--text-muted)' }}>
            Unknown game: <strong>{gameId}</strong>
          </p>
          <pre style={{
            background: 'var(--bg-input)',
            padding: 16,
            borderRadius: 8,
            textAlign: 'left',
            overflow: 'auto',
            fontSize: 12,
            marginTop: 16,
          }}>
            {JSON.stringify(state, null, 2)}
          </pre>
        </div>
      )
  }
}
