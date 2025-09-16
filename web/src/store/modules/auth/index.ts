import { defineStore } from 'pinia'
import { getExpiresIn, getToken, removeExpiresIn, removeToken, setExpiresIn, setToken } from './helper'

export interface AuthState {
  token: string | null  // Access token stored in memory
  expiresIn: number | null
  isRefreshing: boolean // Track if token refresh is in progress
  isInitialized: boolean // Track if auth state has been initialized
  isInitializing: boolean // Track if auth initialization is in progress
}

export const useAuthStore = defineStore('auth-store', {
  state: (): AuthState => ({
    token: getToken(), // Load token normally
    expiresIn: getExpiresIn(),
    isRefreshing: false,
    isInitialized: false,
    isInitializing: false,
  }),

  getters: {
    isValid(): boolean {
      return !!(this.token && this.expiresIn && this.expiresIn > Date.now() / 1000)
    },
    getToken(): string | null {
      return this.token
    },
    getExpiresIn(): number | null {
      return this.expiresIn
    },
    needsRefresh(): boolean {
      // Check if token expires within next 5 minutes
      const fiveMinutesFromNow = Date.now() / 1000 + 300
      return !!(this.expiresIn && this.expiresIn < fiveMinutesFromNow)
    },
    needPermission(): boolean {
      return this.isInitialized && !this.isInitializing && !this.isValid
    }
  },

  actions: {
    async initializeAuth() {
      if (this.isInitialized || this.isInitializing) return

      this.isInitializing = true

      try {
        // Try to refresh token if we have valid expiration
        if (this.expiresIn && this.expiresIn > Date.now() / 1000) {
          console.log('Token expired or about to expire, refreshing...')
          await this.refreshToken()
        } else if (this.token) {
          // Clear expired token
          this.removeToken()
          this.removeExpiresIn()
        }
      } catch (error) {
        // Clear invalid state on error
        this.removeToken()
        this.removeExpiresIn()
      } finally {
        this.isInitializing = false
        this.isInitialized = true
      }
    },
    setToken(token: string) {
      this.token = token
      setToken(token)
    },
    removeToken() {
      this.token = null
      removeToken()
    },
    async refreshToken() {
      if (this.isRefreshing) {
        console.log('Token refresh already in progress, skipping...')
        return // Prevent multiple simultaneous refresh attempts
      }

      console.log('Starting token refresh...')
      this.isRefreshing = true
      try {
        // Call refresh endpoint - refresh token is sent automatically via httpOnly cookie
        const response = await fetch('/api/auth/refresh', {
          method: 'POST',
          credentials: 'include', // Include httpOnly cookies
        })

        console.log('Refresh response status:', response.status)

        if (response.ok) {
          const data = await response.json()
          console.log('Token refresh successful, setting new token')
          this.setToken(data.accessToken)
          this.setExpiresIn(data.expiresIn)
          return data.accessToken
        } else {
          // Refresh failed - user needs to login again
          console.log('Token refresh failed, removing tokens')
          this.removeToken()
          this.removeExpiresIn()
          throw new Error('Token refresh failed')
        }
      } catch (error) {
        console.error('Token refresh error:', error)
        this.removeToken()
        this.removeExpiresIn()
        throw error
      } finally {
        this.isRefreshing = false
      }
    },
    setExpiresIn(expiresIn: number) {
      this.expiresIn = expiresIn
      setExpiresIn(expiresIn)
    },
    removeExpiresIn() {
      this.expiresIn = null
      removeExpiresIn()
    },
  },
})

