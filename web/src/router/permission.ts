import type { Router } from 'vue-router'
import { useAuthStore } from '@/store/modules/auth'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { useSessionStore } from '@/store/modules/session'
import { store } from '@/store'

const FIVE_MINUTES_IN_SECONDS = 5 * 60

// Attempt to ensure we have a valid access token before continuing navigation
async function ensureFreshToken(authStore: any) {
  const currentTs = Math.floor(Date.now() / 1000)
  const expiresIn = authStore.getExpiresIn
  const token = authStore.getToken
  //  the user hasn’t logged in
  if (!token && !expiresIn)
    return

  // If we already have a token that is valid for some time, nothing to do
  if (token && expiresIn && expiresIn > currentTs + FIVE_MINUTES_IN_SECONDS)
    return

  try {
    await authStore.refreshToken()
  }
  catch (error) {
    // If refresh fails, make sure state is cleared so UI can prompt user
    authStore.removeToken()
    authStore.removeExpiresIn()
  }
}

export function setupPageGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    const auth_store = useAuthStore(store)
    await ensureFreshToken(auth_store)

    // Handle workspace context from URL
    if (to.name === 'WorkspaceChat' && to.params.workspaceUuid) {
      const workspaceStore = useWorkspaceStore(store)
      const sessionStore = useSessionStore(store)
      const workspaceUuid = to.params.workspaceUuid as string

      // Only set active workspace if it's different and not already loading
      if (workspaceUuid !== workspaceStore.activeWorkspaceUuid && !workspaceStore.isLoading) {
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
