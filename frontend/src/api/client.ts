const SESSION_KEY = 'keklik_session'

export interface Session {
  token: string
  accountId: string
}

export function getSession(): Session | null {
  const raw = sessionStorage.getItem(SESSION_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as Session
  } catch {
    return null
  }
}

export function saveSession(session: Session): void {
  sessionStorage.setItem(SESSION_KEY, JSON.stringify(session))
}

export function clearSession(): void {
  sessionStorage.removeItem(SESSION_KEY)
}

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export class NetworkError extends Error {
  constructor() {
    super('Network request failed. Please check your connection and try again.')
    this.name = 'NetworkError'
  }
}

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const session = getSession()

  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (session) {
    headers['Authorization'] = `Bearer ${session.token}`
  }

  let response: Response
  try {
    response = await fetch(`${BASE_URL}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  } catch {
    throw new NetworkError()
  }

  if (response.status === 401) {
    clearSession()
    window.location.replace('/signin')
    throw new ApiError('Session expired. Please sign in again.', 'unauthenticated', 401)
  }

  if (!response.ok) {
    let code = 'unknown'
    let message = `Request failed with status ${response.status}`
    try {
      const err = (await response.json()) as { code?: string; message?: string }
      if (err.code) code = err.code
      if (err.message) message = err.message
    } catch {
      // non-JSON error body
    }
    throw new ApiError(message, code, response.status)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, body),
  delete: <T>(path: string, body?: unknown) => request<T>('DELETE', path, body),
}
