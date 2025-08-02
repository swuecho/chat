import { defineStore } from 'pinia'
import { router } from '@/router'
import {
  getWorkspaces,
  createWorkspace,
  updateWorkspace,
  deleteWorkspace,
  ensureDefaultWorkspace,
  setDefaultWorkspace,
  updateWorkspaceOrder,
  autoMigrateLegacySessions,
  getAllWorkspaceActiveSessions,
  getChatSessionDefault,
} from '@/api'

import { useSessionStore } from '@/store/modules/session'
import { t } from '@/locales'

export interface WorkspaceState {
  workspaces: Chat.Workspace[]
  activeWorkspaceUuid: string | null
  workspaceActiveSessions: Record<string, string> // workspaceUuid -> sessionUuid
  pendingSessionRestore: { workspaceUuid: string; sessionUuid: string } | null
  isLoading: boolean
}

export const useWorkspaceStore = defineStore('workspace-store', {
  state: (): WorkspaceState => ({
    workspaces: [],
    activeWorkspaceUuid: null,
    workspaceActiveSessions: {},
    pendingSessionRestore: null,
    isLoading: false,
  }),

  getters: {
    getWorkspaceByUuid(state) {
      return (uuid?: string) => {
        if (uuid) {
          return state.workspaces.find(workspace => workspace.uuid === uuid)
        }
        return null
      }
    },

    getDefaultWorkspace(state) {
      return state.workspaces.find(workspace => workspace.isDefault) || null
    },

    activeWorkspace(state) {
      if (state.activeWorkspaceUuid) {
        return state.workspaces.find(workspace => workspace.uuid === state.activeWorkspaceUuid)
      }
      return null
    },

    // Get active session for a specific workspace
    getActiveSessionForWorkspace(state) {
      return (workspaceUuid: string) => {
        return state.workspaceActiveSessions[workspaceUuid] || null
      }
    },
  },

  actions: {
    // Comprehensive initialization method that replaces the old chat store logic
    async initializeApplication() {
      try {
        console.log('🔄 Starting comprehensive application initialization...')

        // Step 1: Handle legacy session migration
        try {
          const migrationResult = await autoMigrateLegacySessions()
          if (migrationResult.hasLegacySessions && migrationResult.migratedSessions > 0) {
            console.log(`🔄 Auto-migrated ${migrationResult.migratedSessions} legacy sessions to default workspace`)

            // Only force refresh if we're not already on a workspace route
            const currentRoute = router.currentRoute.value
            if (currentRoute.name !== 'WorkspaceChat') {
              console.log('🔄 Refreshing page after migration')
              window.location.reload()
              return // Exit early since we're refreshing
            } else {
              console.log('🔄 Skipping refresh - already on workspace route')
            }
          }
        } catch (migrationError) {
          console.warn('⚠️ Legacy session migration failed:', migrationError)
          // Continue with normal sync - don't block the app
        }

        // Step 2: Sync workspaces
        await this.syncWorkspaces()

        // Step 3: Determine workspace context from URL
        const routeBeforeSync = router.currentRoute.value
        const urlWorkspaceUuid = routeBeforeSync.name === 'WorkspaceChat' ? routeBeforeSync.params.workspaceUuid as string : null
        const urlSessionUuid = routeBeforeSync.params.uuid as string
        const isOnDefaultRoute = routeBeforeSync.name === 'DefaultWorkspace'

        // Step 4: Sync workspace active sessions from backend
        await this.syncWorkspaceActiveSessions(urlWorkspaceUuid || undefined, urlSessionUuid || undefined)

        // Step 5: Ensure we have an active workspace
        await this.ensureActiveWorkspace()

        // Step 6: Initialize sessions through session store
        const sessionStore = useSessionStore()
        await sessionStore.syncAllWorkspaceSessions()

        // Step 7: Handle session creation if needed
        await this.ensureUserHasSession()

        // Step 8: Handle navigation
        await this.handleInitialNavigation(urlWorkspaceUuid || undefined, urlSessionUuid || undefined, isOnDefaultRoute)

        console.log('✅ Application initialization completed successfully')
      } catch (error) {
        console.error('❌ Error in initializeApplication:', error)
        throw error
      }
    },

    async syncWorkspaceActiveSessions(urlWorkspaceUuid?: string, urlSessionUuid?: string) {
      try {
        const backendSessions = await getAllWorkspaceActiveSessions()

        // Build workspace active sessions mapping
        this.workspaceActiveSessions = {}
        let globalActiveSession = null

        for (const session of backendSessions) {
          if (session.workspaceUuid) {
            this.workspaceActiveSessions[session.workspaceUuid] = session.chatSessionUuid
            if (!globalActiveSession) {
              globalActiveSession = session
            }
          }
        }

        // Prioritize URL context over backend data
        if (urlWorkspaceUuid && urlSessionUuid) {
          this.activeWorkspaceUuid = urlWorkspaceUuid
          console.log('✅ Used workspace from URL:', { workspaceUuid: urlWorkspaceUuid, sessionUuid: urlSessionUuid })
        } else if (urlWorkspaceUuid) {
          this.activeWorkspaceUuid = urlWorkspaceUuid
          console.log('✅ Used workspace from URL (no session):', urlWorkspaceUuid)
        } else if (globalActiveSession?.workspaceUuid) {
          this.activeWorkspaceUuid = globalActiveSession.workspaceUuid
          console.log('✅ Used workspace from backend:', globalActiveSession.workspaceUuid)
        }
      } catch (error) {
        console.warn('⚠️ Failed to sync workspace active sessions:', error)
      }
    },

    async ensureActiveWorkspace() {
      // If we don't have an active workspace, set to default
      if (!this.activeWorkspaceUuid && this.workspaces.length > 0) {
        const defaultWorkspace = this.workspaces.find(workspace => workspace.isDefault) || this.workspaces[0]
        if (defaultWorkspace) {
          this.activeWorkspaceUuid = defaultWorkspace.uuid
          console.log('✅ Set active workspace to default:', defaultWorkspace.name)
        }
      }
    },

    async ensureUserHasSession() {
      const sessionStore = useSessionStore()

      // Check if user has any sessions
      const allSessions = sessionStore.getAllSessions()
      if (allSessions.length === 0) {
        console.log('🔄 No sessions found for user, creating default session')

        // Ensure we have a default workspace
        const defaultWorkspace = this.workspaces.find(workspace => workspace.isDefault) || null
        if (!defaultWorkspace) {
          console.error('❌ No default workspace found when trying to create default session')
          throw new Error('No default workspace available for session creation')
        }

        // Set active workspace
        this.activeWorkspaceUuid = defaultWorkspace.uuid

        // Create default session
        const new_chat_text = t('chat.new')
        await sessionStore.createSessionInWorkspace(new_chat_text, defaultWorkspace.uuid)
        console.log('✅ Created default session for new user')
      }
    },

    async handleInitialNavigation(urlWorkspaceUuid?: string, urlSessionUuid?: string, isOnDefaultRoute?: boolean) {
      const sessionStore = useSessionStore()

      if (urlSessionUuid && sessionStore.getChatSessionByUuid(urlSessionUuid)) {
        // We have a valid session in URL, set it as active
        const session = sessionStore.getChatSessionByUuid(urlSessionUuid)
        if (session) {
          await sessionStore.setActiveSession(session.workspaceUuid || this.activeWorkspaceUuid, urlSessionUuid)
          console.log('✅ Set active session from URL:', urlSessionUuid)
        }
      } else if (this.activeWorkspaceUuid) {
        // Find a session to activate in the active workspace
        const workspaceSessions = sessionStore.getSessionsByWorkspace(this.activeWorkspaceUuid)
        if (workspaceSessions.length > 0) {
          await sessionStore.setActiveSession(this.activeWorkspaceUuid, workspaceSessions[0].uuid)
          console.log('✅ Set first session as active in workspace')
        }
      }

      // Handle default route navigation
      if (isOnDefaultRoute && this.activeWorkspaceUuid) {
        console.log('✅ Navigating from default route to active workspace')
        await router.push({
          name: 'WorkspaceChat',
          params: { workspaceUuid: this.activeWorkspaceUuid }
        })
      }
    },

    async syncWorkspaces() {
      try {
        this.isLoading = true
        const workspaces = await getWorkspaces()
        this.workspaces = workspaces

        // Ensure we have a default workspace
        const defaultWorkspace = this.workspaces.find(workspace => workspace.isDefault) || null
        if (!defaultWorkspace) {
          await this.ensureDefaultWorkspace()
        }

        // Set active workspace if not already set
        if (!this.activeWorkspaceUuid && this.workspaces.length > 0) {
          const defaultWs = this.workspaces.find(workspace => workspace.isDefault) || this.workspaces[0]
          this.activeWorkspaceUuid = defaultWs.uuid
        }
      } catch (error) {
        console.error('Failed to sync workspaces:', error)
        throw error
      } finally {
        this.isLoading = false
      }
    },

    async ensureDefaultWorkspace() {
      try {
        const defaultWorkspace = await ensureDefaultWorkspace()
        this.workspaces.push(defaultWorkspace)
        this.activeWorkspaceUuid = defaultWorkspace.uuid
        return defaultWorkspace
      } catch (error) {
        console.error('Failed to ensure default workspace:', error)
        throw error
      }
    },

    async setActiveWorkspace(workspaceUuid: string) {
      const workspace = this.workspaces.find(workspace => workspace.uuid === workspaceUuid)
      if (workspace) {
        this.activeWorkspaceUuid = workspaceUuid

        // Restore the previously active session for this workspace
        const activeSessionForWorkspace = this.workspaceActiveSessions[workspaceUuid]

        if (activeSessionForWorkspace) {
          // Emit an event that the chat view can listen to
          this.$patch((state) => {
            state.pendingSessionRestore = {
              workspaceUuid,
              sessionUuid: activeSessionForWorkspace
            }
          })
        }
      }
    },

    // Method to handle session restore (called from chat view)
    restoreActiveSession() {
      const pending = this.pendingSessionRestore
      if (pending) {
        const sessionStore = useSessionStore()
        const session = sessionStore.getChatSessionByUuid(pending.sessionUuid)
        if (session) {
          sessionStore.setActiveSession(pending.workspaceUuid, pending.sessionUuid)
        } else {
          // Session no longer exists, clear the tracking
          delete this.workspaceActiveSessions[pending.workspaceUuid]
        }
        // Clear the pending restore
        this.$patch((state) => {
          state.pendingSessionRestore = null
        })
      }
    },

    async createWorkspace(name: string, description: string = '', color: string = '#6366f1', icon: string = 'folder') {
      try {
        const newWorkspace = await createWorkspace({
          name,
          description,
          color,
          icon,
        })
        this.workspaces.push(newWorkspace)
        return newWorkspace
      } catch (error) {
        console.error('Failed to create workspace:', error)
        throw error
      }
    },

    async updateWorkspace(workspaceUuid: string, updates: any) {
      try {
        const updatedWorkspace = await updateWorkspace(workspaceUuid, updates)
        const index = this.workspaces.findIndex(w => w.uuid === workspaceUuid)
        if (index !== -1) {
          this.workspaces[index] = updatedWorkspace
        }
        return updatedWorkspace
      } catch (error) {
        console.error('Failed to update workspace:', error)
        throw error
      }
    },

    async deleteWorkspace(workspaceUuid: string) {
      try {
        await deleteWorkspace(workspaceUuid)
        this.workspaces = this.workspaces.filter(w => w.uuid !== workspaceUuid)

        // Remove from active sessions tracking
        delete this.workspaceActiveSessions[workspaceUuid]

        // If we deleted the active workspace, switch to default
        if (this.activeWorkspaceUuid === workspaceUuid) {
          const defaultWorkspace = this.workspaces.find(workspace => workspace.isDefault) || null
          if (defaultWorkspace) {
            this.activeWorkspaceUuid = defaultWorkspace.uuid
          } else if (this.workspaces.length > 0) {
            this.activeWorkspaceUuid = this.workspaces[0].uuid
          } else {
            this.activeWorkspaceUuid = null
          }
        }
      } catch (error) {
        console.error('Failed to delete workspace:', error)
        throw error
      }
    },

    async setDefaultWorkspace(workspaceUuid: string) {
      try {
        await setDefaultWorkspace(workspaceUuid)
        // Update local state
        this.workspaces.forEach(workspace => {
          workspace.isDefault = workspace.uuid === workspaceUuid
        })
      } catch (error) {
        console.error('Failed to set default workspace:', error)
        throw error
      }
    },

    async updateWorkspaceOrder(workspaceUuids: string[]) {
      try {
        // API expects individual updates - this needs to be implemented properly
        // For now, just reorder locally
        console.warn('updateWorkspaceOrder API call needs to be implemented')
        // Reorder workspaces locally
        const reorderedWorkspaces: Chat.Workspace[] = []
        workspaceUuids.forEach(uuid => {
          const workspace = this.workspaces.find(w => w.uuid === uuid)
          if (workspace) {
            reorderedWorkspaces.push(workspace)
          }
        })
        this.workspaces = reorderedWorkspaces
      } catch (error) {
        console.error('Failed to update workspace order:', error)
        throw error
      }
    },

    setActiveSessionForWorkspace(workspaceUuid: string, sessionUuid: string) {
      this.workspaceActiveSessions[workspaceUuid] = sessionUuid
    },

    clearActiveSessionForWorkspace(workspaceUuid: string) {
      delete this.workspaceActiveSessions[workspaceUuid]
    },

    navigateToWorkspace(workspaceUuid: string, sessionUuid?: string) {
      const route = sessionUuid
        ? { name: 'WorkspaceChat', params: { workspaceUuid, uuid: sessionUuid } }
        : { name: 'WorkspaceChat', params: { workspaceUuid } }

      router.push(route)
    },
  },
})