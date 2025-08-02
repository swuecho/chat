import type { Router } from 'vue-router'
import { useAuthStore } from '@/store/modules/auth'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { useSessionStore } from '@/store/modules/session'
import { store } from '@/store'

// when time expired, remove the username from localstorage
// this will force user re-login after
// rome-ignore lint/suspicious/noExplicitAny: <explanation>
function checkIsTokenExpired(auth_store: any) {
  const accessToken = auth_store.getToken
  if (accessToken) {
    const current_ts = Math.floor(Date.now() / 1000)
    const expiresIn = auth_store.getExpiresIn
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
    const auth_store = useAuthStore(store)
    checkIsTokenExpired(auth_store)

    // Handle workspace context from URL
    if (to.name === 'WorkspaceChat' && to.params.workspaceUuid) {
      const workspaceStore = useWorkspaceStore(store)
      const sessionStore = useSessionStore(store)
      const workspaceUuid = to.params.workspaceUuid as string

      // Set active workspace if it's different from current
      if (workspaceUuid !== workspaceStore.activeWorkspaceUuid) {
        console.log('Setting workspace from URL:', workspaceUuid)
        await workspaceStore.setActiveWorkspace(workspaceUuid)
      }

      // Set active session if provided in URL
      if (to.params.uuid) {
        const sessionUuid = to.params.uuid as string
        if (sessionUuid !== sessionStore.activeSessionUuid) {
          await sessionStore.setActiveSession(workspaceUuid, sessionUuid)
        }
      }
    }

    // Handle default route - let store sync handle navigation to default workspace
    if (to.name === 'DefaultWorkspace') {
      console.log('On default route, letting store handle workspace navigation')
    }

    next()
  })
}
