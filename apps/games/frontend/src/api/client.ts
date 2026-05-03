import type {
  User,
  Game,
  LeaderboardEntry,
  Room,
  AdminSession,
} from './types'

const BASE = '/api'

async function request<T>(
  method: string,
  path: string,
  body?: unknown
): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method,
    credentials: 'include',
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText)
    throw new Error(text || `HTTP ${res.status}`)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  // Auth
  register(username: string, password: string): Promise<User> {
    return request('POST', '/auth/register', { username, password })
  },
  login(username: string, password: string): Promise<User> {
    return request('POST', '/auth/login', { username, password })
  },
  logout(): Promise<void> {
    return request('POST', '/auth/logout')
  },
  me(): Promise<User> {
    return request('GET', '/me')
  },

  // Games
  getDaily(): Promise<Game> {
    return request('GET', '/daily')
  },
  getGames(): Promise<Game[]> {
    return request('GET', '/games')
  },
  getLeaderboard(): Promise<LeaderboardEntry[]> {
    return request('GET', '/leaderboard')
  },

  // Rooms
  createRoom(): Promise<{ code: string }> {
    return request('POST', '/rooms')
  },
  createRoomForGame(game_id: string, config?: Record<string, unknown>): Promise<{ code: string }> {
    return request('POST', '/rooms', { game_id, config })
  },
  getRoom(code: string): Promise<Room> {
    return request('GET', `/rooms/${code}`)
  },

  // Admin
  setDailyGame(game_id: string): Promise<void> {
    return request('PUT', '/admin/daily', { game_id })
  },
  getSessions(): Promise<AdminSession[]> {
    return request('GET', '/admin/sessions')
  },
}
