import type { App } from 'vue'
import { createI18n } from 'vue-i18n'
import enUS from './en-US.json'
import zhCN from './zh-CN.json'
import zhTW from './zh-TW.json'
import type { Language } from '@/store/modules/app/helper'


const i18n = createI18n({
  locale: navigator.language.split('-')[0],
  fallbackLocale: 'en',
  allowComposition: true,
  messages: {
    'en-US': enUS,
    'zh-CN': zhCN,
    'zh-TW': zhTW,
  },
})

export function t(key: string, values?: Record<string, string>) {
  if (values) {
  return i18n.global.t(key, values)
  } else {
    return i18n.global.t(key)
  }
}

export function setLocale(locale: Language) {
  i18n.global.locale = locale
}

export function setupI18n(app: App) {
  app.use(i18n)
}

export default i18n
