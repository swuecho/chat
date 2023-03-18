import type { Router } from 'vue-router'
import { useAuthStoreWithout } from '@/store/modules/auth'
// when time expired, remove the username from localstorage
// this will force user re-login after
// rome-ignore lint/suspicious/noExplicitAny: <explanation>
function checkIsTokenExpired(auth_store: any) {
  const accessToken = auth_store.getToken()
  if (accessToken) {
    const current_ts = Math.floor(Date.now() / 1000)
    const expiresIn = auth_store.getExpiresIn()
    if (expiresIn) {
      const expired = expiresIn < current_ts
      if (expired) {
        auth_store.removeToken()
        return expired
      }
    }
    else {
      return true
    }
  }
}

export function setupPageGuard(router: Router) {
  router.beforeEach(async (from, to, next) => {
    const auth_store = useAuthStoreWithout()
    checkIsTokenExpired(auth_store)
    next()
  })
}
