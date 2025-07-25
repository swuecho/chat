import type { Router } from 'vue-router'
import { useAuthStoreWithout } from '@/store/modules/auth'
import { useChatStoreWithout } from '@/store/modules/chat'
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
  router.beforeEach(async (to, from, next) => {
    const auth_store = useAuthStoreWithout()
    checkIsTokenExpired(auth_store)
    
    // Handle workspace context from URL
    if (to.name === 'WorkspaceChat' && to.params.workspaceUuid) {
      const chat_store = useChatStoreWithout()
      const workspaceUuid = to.params.workspaceUuid as string
      
      // Set active workspace if it's different from current
      if (workspaceUuid !== chat_store.activeWorkspace) {
        console.log('Setting workspace from URL:', workspaceUuid)
        chat_store.setActiveWorkspace(workspaceUuid)
      }
      
      // Set active session if provided in URL
      if (to.params.uuid) {
        const sessionUuid = to.params.uuid as string
        if (sessionUuid !== chat_store.active) {
          chat_store.setActiveLocal(sessionUuid)
        }
      }
    }
    
    next()
  })
}
