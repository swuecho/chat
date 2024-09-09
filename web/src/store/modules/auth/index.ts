import { defineStore } from 'pinia'
import { getExpiresIn, getToken, removeExpiresIn, removeToken, setExpiresIn, setToken } from './helper'
import { store } from '@/store'

export interface AuthState {
  token: string | undefined
  expiresIn: number | undefined
}

export const useAuthStore = defineStore('auth-store', {
  state: (): AuthState => ({
    token: getToken(),
    expiresIn: getExpiresIn(),
  }),

  getters: {
    isValid(): boolean {
      return !!(this.token && this.expiresIn && this.expiresIn > Date.now() / 1000)
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
      this.token = undefined
      removeToken()
    },
    getExpiresIn() {
      return this.expiresIn
    },
    setExpiresIn(expiresIn: number) {
      this.expiresIn = expiresIn
      setExpiresIn(expiresIn)
    },
    removeExpiresIn() {
      this.expiresIn = undefined
      removeExpiresIn()
    },

  },
})

export function useAuthStoreWithout() {
  return useAuthStore(store)
}
