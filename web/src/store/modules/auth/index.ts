import { defineStore } from 'pinia'
import { watch } from 'vue'
import { getExpiresIn, getToken, removeExpiresIn, removeToken, setExpiresIn, setToken } from './helper'
import { createClient } from '@/api/generated/client'
import { refreshAccessToken } from '@/api/generated/sdk.gen'

let activeRefreshPromise: Promise<string | void> | null = null
const refreshClient = createClient({ credentials: 'include', throwOnError: true })

export interface AuthState {
  token: string | null // Access token stored in memory
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
    },
  },

  actions: {
    async initializeAuth() {
      if (this.isInitialized || this.isInitializing)
        return

      this.isInitializing = true

      try {
        const now = Date.now() / 1000
        if (this.expiresIn) {
          const tokenMissing = !this.token
          const expired = this.expiresIn <= now
          if (tokenMissing || expired || this.needsRefresh)
            await this.refreshToken()
        }
        else if (this.token) {
          // Clear expired token
          this.removeToken()
          this.removeExpiresIn()
        }
      }
      catch (error) {
        // Clear invalid state on error
        this.removeToken()
        this.removeExpiresIn()
      }
      finally {
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
      if (this.isRefreshing && activeRefreshPromise)
        return activeRefreshPromise

      this.isRefreshing = true

      const refreshOperation = (async () => {
        try {
          // Use a standalone generated client to avoid the authenticated
          // client's refresh interceptor recursively calling this action.
          const data = await refreshAccessToken({ client: refreshClient })
          this.setToken(data.accessToken)
          this.setExpiresIn(data.expiresIn)
          return data.accessToken
        }
        catch (error) {
          console.error('Token refresh error:', error)
          this.removeToken()
          this.removeExpiresIn()
          throw error
        }
        finally {
          this.isRefreshing = false
          activeRefreshPromise = null
        }
      })()

      activeRefreshPromise = refreshOperation
      return refreshOperation
    },
    setExpiresIn(expiresIn: number) {
      this.expiresIn = expiresIn
      setExpiresIn(expiresIn)
    },
    removeExpiresIn() {
      this.expiresIn = null
      removeExpiresIn()
    },
    async waitForInitialization(timeoutMs = 10000) {
      if (!this.isInitializing)
        return

      await new Promise<void>((resolve) => {
        let stopWatcher: (() => void) | null = null
        const timeoutId = setTimeout(() => {
          if (stopWatcher)
            stopWatcher()
          resolve()
        }, timeoutMs)

        stopWatcher = watch(
          () => this.isInitializing,
          (isInit) => {
            if (!isInit) {
              clearTimeout(timeoutId)
              if (stopWatcher)
                stopWatcher()
              resolve()
            }
          },
          { immediate: false },
        )
      })
    },
  },
})
