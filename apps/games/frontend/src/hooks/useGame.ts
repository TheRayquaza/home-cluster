import { useState, useEffect, useRef, useCallback } from 'react'
import type { GameState, GameOverPayload, JoinedPayload } from '../api/types'

interface GameHookResult {
  playerIdx: number | null
  gameId: string | null
  state: GameState | null
  gameOver: GameOverPayload | null
  connected: boolean
  error: string | null
  send: (action: object) => void
}

const MAX_RETRIES = 3
const BASE_DELAY = 1000

export function useGame(roomCode: string): GameHookResult {
  const [playerIdx, setPlayerIdx] = useState<number | null>(null)
  const [gameId, setGameId] = useState<string | null>(null)
  const [state, setState] = useState<GameState | null>(null)
  const [gameOver, setGameOver] = useState<GameOverPayload | null>(null)
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const retriesRef = useRef(0)
  const roomCodeRef = useRef(roomCode)
  const mountedRef = useRef(true)

  useEffect(() => {
    roomCodeRef.current = roomCode
  })

  const connect = useCallback(() => {
    if (!mountedRef.current) return

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const host = window.location.host
    const url = `${proto}://${host}/ws?room=${roomCodeRef.current}`

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      if (!mountedRef.current) return
      retriesRef.current = 0
      setConnected(true)
      setError(null)
      ws.send(JSON.stringify({ type: 'join', payload: { room: roomCodeRef.current } }))
    }

    ws.onmessage = (evt: MessageEvent) => {
      if (!mountedRef.current) return
      try {
        const msg = JSON.parse(evt.data as string) as {
          type: string
          payload: unknown
        }
        switch (msg.type) {
          case 'joined': {
            const p = msg.payload as JoinedPayload
            setPlayerIdx(p.playerIdx)
            setGameId(p.game)
            ws.send(JSON.stringify({ type: 'ready' }))
            break
          }
          case 'state':
            setState(msg.payload as GameState)
            break
          case 'game_over':
            setGameOver(msg.payload as GameOverPayload)
            break
          case 'waiting':
            // opponent not yet connected — stay in waiting UI, not an error
            break
          case 'error':
            setError((msg.payload as { message: string }).message)
            break
        }
      } catch {
        // ignore parse errors
      }
    }

    ws.onclose = () => {
      if (!mountedRef.current) return
      setConnected(false)
      if (retriesRef.current < MAX_RETRIES) {
        const delay = BASE_DELAY * Math.pow(2, retriesRef.current)
        retriesRef.current += 1
        setTimeout(connect, delay)
      } else {
        setError('Connection lost. Please refresh.')
      }
    }

    ws.onerror = () => {
      if (!mountedRef.current) return
      // onclose will handle reconnect
    }
  }, [])

  useEffect(() => {
    mountedRef.current = true
    connect()
    return () => {
      mountedRef.current = false
      wsRef.current?.close()
    }
  }, [connect])

  const send = useCallback((action: object) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'action', payload: action }))
    }
  }, [])

  return { playerIdx, gameId, state, gameOver, connected, error, send }
}
