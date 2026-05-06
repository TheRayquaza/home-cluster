export interface User {
  id: string
  username: string
  role: string
}

export interface Game {
  id: string
  name: string
  description: string
  emoji: string
  color: string
  real_time: boolean
}

export type DailyGame = Game

export interface LeaderboardEntry {
  username: string
  wins: number
  losses: number
  draws: number
}

export interface Room {
  game_id: string
  players_joined: 0 | 1 | 2
}

export interface AdminSession {
  id: string
  game_id: string
  player1: string
  player2: string
  winner_idx: number
  played_at: string
}

// WebSocket message types
export interface WsMessage {
  type: string
  payload?: unknown
}

export interface JoinedPayload {
  playerIdx: number
  game: string
}

export interface GameOverPayload {
  winner: number
}

export interface ErrorPayload {
  message: string
}

// Game state shapes
export interface TicTacToeState {
  board: number[]
  turn: number
  winner: number
}

export interface Connect4State {
  board: number[][]
  turn: number
  winner: number
}

export interface MemoryMatchState {
  cards: number[]
  revealed: boolean[]
  matched: boolean[]
  scores: number[]
  turn: number
  pending: number[]
  waiting: boolean
  pairs: number
  cols: number
  rows: number
}

export interface NimState {
  sticks: number
  turn: number
}

export interface PongState {
  ball: { x: number; y: number; vx: number; vy: number }
  paddles: number[]
  scores: number[]
  width: number
  height: number
  paddle_h: number
  paddle_w: number
}

export interface LightCyclesState {
  width: number
  height: number
  players: Array<{ x: number; y: number; dir: string }>
  trails: number[][][]
  alive: boolean[]
  phase?: string
  countdown?: number
}

export interface WordleState {
  guesses: string[][]
  hints: number[][][]
  solved: boolean[]
  attempts: number[]
  game_over: boolean
  word_length?: number
  revealed_word?: string
}

export interface BattleshipsState {
  phase: 'placement' | 'combat'
  boards: number[][]
  shots: number[][]
  ready: boolean[]
  ships_left: number[]
  turn: number
  winner: number
  ship_sizes: number[]
  to_place: (number[] | null)[]
}

export type GameState =
  | TicTacToeState
  | Connect4State
  | MemoryMatchState
  | NimState
  | PongState
  | LightCyclesState
  | WordleState
  | BattleshipsState

export interface GameProps {
  state: GameState
  playerIdx: number
  onAction: (action: object) => void
  gameOver: GameOverPayload | null
}
