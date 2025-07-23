import { defineStore } from 'pinia'
import { getExpiresIn, getToken, removeExpiresIn, removeToken, setExpiresIn, setToken } from './helper'
import { store } from '@/store'

export interface AuthState {
  token: string | null  // Access token stored in memory
  expiresIn: number | null
  isRefreshing: boolean // Track if token refresh is in progress
}

export const useAuthStore = defineStore('auth-store', {
  state: (): AuthState => ({
    token: getToken(), // Access token from memory
    expiresIn: getExpiresIn(),
    isRefreshing: false,
  }),

  getters: {
    isValid(): boolean {
      return !!(this.token && this.expiresIn && this.expiresIn > Date.now() / 1000)
    },
    needsRefresh(): boolean {
      // Check if token expires within next 5 minutes
      const fiveMinutesFromNow = Date.now() / 1000 + 300
      return !!(this.expiresIn && this.expiresIn < fiveMinutesFromNow)
    },
  },

  actions: {
    getToken() {
      return this.token
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
        return // Prevent multiple simultaneous refresh attempts
      }
      
      this.isRefreshing = true
      try {
        // Call refresh endpoint - refresh token is sent automatically via httpOnly cookie
        const response = await fetch('/api/auth/refresh', {
          method: 'POST',
          credentials: 'include', // Include httpOnly cookies
        })
        
        if (response.ok) {
          const data = await response.json()
          this.setToken(data.accessToken)
          this.setExpiresIn(data.expiresIn)
          return data.accessToken
        } else {
          // Refresh failed - user needs to login again
          this.removeToken()
          this.removeExpiresIn()
          throw new Error('Token refresh failed')
        }
      } finally {
        this.isRefreshing = false
      }
    },
    getExpiresIn() {
      return this.expiresIn
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

export function useAuthStoreWithout() {
  return useAuthStore(store)
}
