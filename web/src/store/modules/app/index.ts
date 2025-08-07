import { defineStore } from 'pinia'
import type { AppState, Language, Theme } from './helper'
import { getLocalSetting, setLocalSetting } from './helper'
import { store } from '@/store'

const langs: Language[] = ['zh-CN', 'en-US', 'zh-CN']

export const useAppStore = defineStore('app-store', {
  state: (): AppState => getLocalSetting(),
  actions: {
    setSiderCollapsed(collapsed: boolean) {
      this.siderCollapsed = collapsed
      this.recordState()
    },

    setTheme(theme: Theme) {
      this.theme = theme
      this.recordState()
    },

    setLanguage(language: Language) {
      if (this.language !== language) {
        this.language = language
        this.recordState()
      }
    },
    setNextLanguage() {
      const currentLang = this.language
      const nextLang = langs[(langs.indexOf(currentLang) + 1) % langs.length]
      this.language = nextLang
      this.recordState()
    },

    recordState() {
      setLocalSetting(this.$state)
    },
  },
})

export function useAppStoreWithOut() {
  return useAppStore(store)
}
