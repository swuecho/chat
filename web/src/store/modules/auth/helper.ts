// Hybrid token approach:
// - Access tokens: In-memory (short-lived, secure from XSS)
// - Refresh tokens: httpOnly cookies (long-lived, persistent)

let accessToken: string | null = null

export function getToken(): string | null {
  return accessToken
}

export function setToken(token: string): void {
  accessToken = token
}

export function removeToken(): void {
  accessToken = null
}

// Expiration tracking can still be useful for UI state
const EXPIRE_LOCAL_NAME = 'expiresIn'

export function getExpiresIn(): number | null {
  const stored = window.localStorage.getItem(EXPIRE_LOCAL_NAME)
  return stored ? parseInt(stored, 10) : null
}

export function setExpiresIn(expiresIn: number): void {
  window.localStorage.setItem(EXPIRE_LOCAL_NAME, expiresIn.toString())
}

export function removeExpiresIn(): void {
  window.localStorage.removeItem(EXPIRE_LOCAL_NAME)
}
