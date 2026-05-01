const ACCOUNT_ID_KEY = 'keklik_account_id'

export function getAccountId(): string | null {
  return localStorage.getItem(ACCOUNT_ID_KEY)
}

export function setAccountId(id: string): void {
  localStorage.setItem(ACCOUNT_ID_KEY, id)
}

export function clearAccountId(): void {
  localStorage.removeItem(ACCOUNT_ID_KEY)
}

export class ApiError extends Error {
  readonly code: string
  readonly status: number
  readonly payload?: unknown
  readonly conflict?: unknown

  constructor(message: string, code: string, status: number, payload?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
    this.payload = payload
    if (payload && typeof payload === 'object' && 'conflict' in payload) {
      this.conflict = (payload as { conflict?: unknown }).conflict
    }
  }
}

export class NetworkError extends Error {
  constructor() {
    super('Network request failed. Please check your connection and try again.')
    this.name = 'NetworkError'
  }
}

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

async function executeRequest(
  method: string,
  path: string,
  body: unknown,
): Promise<Response> {
  try {
    return await fetch(`${BASE_URL}${path}`, {
      method,
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
  } catch {
    throw new NetworkError()
  }
}

async function tryRefreshSession(): Promise<boolean> {
  try {
    const resp = await fetch(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
    return resp.ok
  } catch {
    return false
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  let response = await executeRequest(method, path, body)

  if (response.status === 401) {
    const refreshed = await tryRefreshSession()
    if (refreshed) {
      response = await executeRequest(method, path, body)
    }
    if (response.status === 401) {
      clearAccountId()
      window.location.replace('/')
      throw new ApiError('Session expired. Please sign in again.', 'unauthenticated', 401)
    }
  }

  if (!response.ok) {
    let code = 'unknown'
    let message = `Request failed with status ${response.status}`
    let payload: unknown
    try {
      const err = (await response.json()) as { code?: string; message?: string }
      payload = err
      if (err.code) code = err.code
      if (err.message) message = err.message
    } catch {
      // non-JSON error body
    }
    throw new ApiError(message, code, response.status, payload)
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
