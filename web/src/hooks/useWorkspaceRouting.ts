import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useChatStore } from '@/store'

export function useWorkspaceRouting() {
  const router = useRouter()
  const route = useRoute()
  const chatStore = useChatStore()

  // Get current workspace from URL
  const currentWorkspaceFromUrl = computed(() => {
    return route.params.workspaceUuid as string || null
  })

  // Get current session from URL
  const currentSessionFromUrl = computed(() => {
    return route.params.uuid as string || null
  })

  // Check if we're on a workspace-aware route
  const isWorkspaceRoute = computed(() => {
    return route.name === 'WorkspaceChat'
  })

  // Generate workspace-aware URL for a session
  function getSessionUrl(sessionUuid: string, workspaceUuid?: string): string {
    const workspace = workspaceUuid || chatStore.activeWorkspace
    const session = chatStore.getChatSessionByUuid(sessionUuid)
    
    // Use session's workspace if available, otherwise use provided or active workspace
    const targetWorkspace = session?.workspaceUuid || workspace || chatStore.getDefaultWorkspace?.uuid
    
    if (targetWorkspace) {
      return `/#/workspace/${targetWorkspace}/chat/${sessionUuid}`
    }
    // Fallback to default workspace if none found
    const defaultWorkspace = chatStore.getDefaultWorkspace
    return defaultWorkspace ? `/#/workspace/${defaultWorkspace.uuid}/chat/${sessionUuid}` : `/#/`
  }

  // Generate workspace URL (without session)
  function getWorkspaceUrl(workspaceUuid: string): string {
    return `/#/workspace/${workspaceUuid}/chat`
  }

  // Navigate to session with workspace context
  async function navigateToSession(sessionUuid: string, workspaceUuid?: string) {
    const workspace = workspaceUuid || chatStore.activeWorkspace
    const session = chatStore.getChatSessionByUuid(sessionUuid)
    
    // Use session's workspace if available, otherwise use default workspace
    const targetWorkspace = session?.workspaceUuid || workspace || chatStore.getDefaultWorkspace?.uuid
    
    if (targetWorkspace) {
      await router.push({
        name: 'WorkspaceChat',
        params: {
          workspaceUuid: targetWorkspace,
          uuid: sessionUuid
        }
      })
    } else {
      // Fallback to default route if no workspace found
      await router.push({ name: 'DefaultWorkspace' })
    }
  }

  // Navigate to workspace (without specific session)
  async function navigateToWorkspace(workspaceUuid: string) {
    await router.push({
      name: 'WorkspaceChat',
      params: { workspaceUuid }
    })
  }

  // Navigate to first session in workspace, or workspace itself if no sessions
  async function navigateToWorkspaceOrFirstSession(workspaceUuid: string) {
    const workspaceSessions = chatStore.getSessionsByWorkspace(workspaceUuid)
    
    if (workspaceSessions.length > 0) {
      await navigateToSession(workspaceSessions[0].uuid, workspaceUuid)
    } else {
      await navigateToWorkspace(workspaceUuid)
    }
  }

  // Check if current route matches the expected workspace/session
  function isCurrentRoute(sessionUuid?: string, workspaceUuid?: string): boolean {
    const currentSession = currentSessionFromUrl.value
    const currentWorkspace = currentWorkspaceFromUrl.value
    
    if (sessionUuid && sessionUuid !== currentSession) {
      return false
    }
    
    if (workspaceUuid && workspaceUuid !== currentWorkspace) {
      return false
    }
    
    return true
  }

  // Sync URL with current state (useful for redirects after workspace changes)
  async function syncUrlWithState() {
    const activeSession = chatStore.active
    const activeWorkspace = chatStore.activeWorkspace
    
    // If we have an active session and workspace, ensure URL is correct
    if (activeSession && activeWorkspace) {
      const session = chatStore.getChatSessionByUuid(activeSession)
      if (session && session.workspaceUuid === activeWorkspace) {
        // Check if current URL doesn't match expected workspace-aware URL
        if (!isCurrentRoute(activeSession, activeWorkspace)) {
          await navigateToSession(activeSession, activeWorkspace)
        }
      }
    }
  }

  // Handle browser back/forward navigation
  function handleRouteChange() {
    const workspaceFromUrl = currentWorkspaceFromUrl.value
    const sessionFromUrl = currentSessionFromUrl.value
    
    // Update store state to match URL
    if (workspaceFromUrl && workspaceFromUrl !== chatStore.activeWorkspace) {
      chatStore.setActiveWorkspace(workspaceFromUrl)
    }
    
    if (sessionFromUrl && sessionFromUrl !== chatStore.active) {
      chatStore.setActiveLocal(sessionFromUrl)
    }
  }

  return {
    // Computed
    currentWorkspaceFromUrl,
    currentSessionFromUrl,
    isWorkspaceRoute,
    
    // Methods
    getSessionUrl,
    getWorkspaceUrl,
    navigateToSession,
    navigateToWorkspace,
    navigateToWorkspaceOrFirstSession,
    isCurrentRoute,
    syncUrlWithState,
    handleRouteChange
  }
}