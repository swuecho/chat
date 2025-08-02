import { defineStore } from 'pinia'
import { router } from '@/router'
import {
  createChatSession,
  deleteChatSession,
  renameChatSession,
  updateChatSession,
  getSessionsByWorkspace,
  createSessionInWorkspace,
} from '@/api'
import { useWorkspaceStore } from '../workspace'

export interface SessionState {
  workspaceHistory: Record<string, Chat.Session[]> // workspaceUuid -> sessions
  activeSessionUuid: string | null
  isLoading: boolean
  isCreatingSession: boolean
  isSwitchingSession: boolean
}

export const useSessionStore = defineStore('session-store', {
  state: (): SessionState => ({
    workspaceHistory: {},
    activeSessionUuid: null,
    isLoading: false,
    isCreatingSession: false,
    isSwitchingSession: false,
  }),

  getters: {
    getChatSessionByUuid(state) {
      return (uuid?: string) => {
        if (uuid) {
          // Search across all workspace histories
          for (const sessions of Object.values(state.workspaceHistory)) {
            const session = sessions.find(item => item.uuid === uuid)
            if (session) return session
          }
        }
        return null
      }
    },

    getSessionsByWorkspace(state) {
      return (workspaceUuid?: string) => {
        if (!workspaceUuid) return []
        return state.workspaceHistory[workspaceUuid] || []
      }
    },

    activeSession(state) {
      if (state.activeSessionUuid) {
        // Search across all workspace histories
        for (const sessions of Object.values(state.workspaceHistory)) {
          const session = sessions.find(item => item.uuid === state.activeSessionUuid)
          if (session) return session
        }
      }
      return null
    },

    // Get session URL for navigation
    getSessionUrl() {
      return (sessionUuid: string): string => {
        // Search across all workspace histories
        for (const sessions of Object.values(this.workspaceHistory)) {
          const session = sessions.find(item => item.uuid === sessionUuid)
          if (session && session.workspaceUuid) {
            return `/#/workspace/${session.workspaceUuid}/chat/${sessionUuid}`
          }
        }
        return `/#/chat/${sessionUuid}`
      }
    },
  },

  actions: {
    async syncWorkspaceSessions(workspaceUuid: string) {
      try {
        this.isLoading = true
        const sessions = await getSessionsByWorkspace(workspaceUuid)
        this.workspaceHistory[workspaceUuid] = sessions
        return sessions
      } catch (error) {
        console.error(`Failed to sync sessions for workspace ${workspaceUuid}:`, error)
        throw error
      } finally {
        this.isLoading = false
      }
    },

    async syncAllWorkspaceSessions() {
      try {
        this.isLoading = true
        const workspaceStore = useWorkspaceStore()

        // Sync sessions for all workspaces
        for (const workspace of workspaceStore.workspaces) {
          const sessions = await getSessionsByWorkspace(workspace.uuid)
          this.workspaceHistory[workspace.uuid] = sessions
        }
      } catch (error) {
        console.error('Failed to sync all workspace sessions:', error)
        throw error
      } finally {
        this.isLoading = false
      }
    },

    async createSessionInWorkspace(title: string, workspaceUuid?: string, model?: string) {
      if (this.isCreatingSession) {
        return null
      }

      this.isCreatingSession = true

      try {
        const workspaceStore = useWorkspaceStore()
        const targetWorkspaceUuid = workspaceUuid || workspaceStore.activeWorkspaceUuid

        if (!targetWorkspaceUuid) {
          throw new Error('No workspace available for session creation')
        }

        const newSession = await createSessionInWorkspace(targetWorkspaceUuid, {
          topic: title,
          model,
        })

        // Add to workspace history
        if (!this.workspaceHistory[targetWorkspaceUuid]) {
          this.workspaceHistory[targetWorkspaceUuid] = []
        }
        this.workspaceHistory[targetWorkspaceUuid].unshift(newSession)

        // Set as active session
        await this.setActiveSession(targetWorkspaceUuid, newSession.uuid)

        return newSession
      } catch (error) {
        console.error('Failed to create session in workspace:', error)
        throw error
      } finally {
        this.isCreatingSession = false
      }
    },

    async createLegacySession(session: Chat.Session) {
      try {
        await createChatSession(session.uuid, session.title, session.model)

        // Refresh workspace sessions to get updated list from backend
        const workspaceUuid = session.workspaceUuid
        if (workspaceUuid) {
          await this.syncWorkspaceSessions(workspaceUuid)
        }

        await this.setActiveSession(workspaceUuid || null, session.uuid)
        return session
      } catch (error) {
        console.error('Failed to create legacy session:', error)
        throw error
      }
    },

    async updateSession(uuid: string, updates: Partial<Chat.Session>) {
      try {
        console.log('updateSession called with uuid:', uuid, 'updates:', updates)
        console.log('Current workspaceHistory:', this.workspaceHistory)

        // Find session across all workspace histories
        for (const workspaceUuid in this.workspaceHistory) {
          const sessions = this.workspaceHistory[workspaceUuid]
          const index = sessions.findIndex(item => item.uuid === uuid)
          if (index !== -1) {
            console.log('Found session in workspace:', workspaceUuid, 'at index:', index)
            // Update local state
            sessions[index] = { ...sessions[index], ...updates }

            // Update backend - use the appropriate API method
            if (updates.title !== undefined) {
              // If only title is changing, use the rename endpoint
              await renameChatSession(uuid, sessions[index].title)
            } else {
              // For other updates (like model), use the full update endpoint
              await updateChatSession(uuid, sessions[index])
            }

            return sessions[index]
          }
        }

        // If session not found locally, try to update it on the backend anyway
        // This handles cases where the session exists on the server but not in local state
        console.log('Session not found locally, attempting backend update')
        try {
          const session = this.getChatSessionByUuid(uuid)
          if (session) {
            console.log('Found session via getter, updating')
            const updatedSession = { ...session, ...updates }
            await updateChatSession(uuid, updatedSession)
            return updatedSession
          }
        } catch (backendError) {
          console.error('Backend update also failed:', backendError)
        }

        throw new Error(`Session ${uuid} not found`)
      } catch (error) {
        console.error('Failed to update session:', error)
        throw error
      }
    },

    async deleteSession(sessionUuid: string) {
      try {
        // Find session and its workspace
        let workspaceUuid: string | null = null
        for (const [wUuid, sessions] of Object.entries(this.workspaceHistory)) {
          const index = sessions.findIndex(item => item.uuid === sessionUuid)
          if (index !== -1) {
            workspaceUuid = wUuid
            break
          }
        }

        if (workspaceUuid) {
          // Remove from workspace history
          this.workspaceHistory[workspaceUuid] = this.workspaceHistory[workspaceUuid].filter(
            session => session.uuid !== sessionUuid
          )
        }

        // Delete from backend
        await deleteChatSession(sessionUuid)

        // Clear active session if it was deleted
        if (this.activeSessionUuid === sessionUuid) {
          await this.setNextActiveSession(workspaceUuid)
        }

        // Clear from workspace active sessions
        if (workspaceUuid) {
          const workspaceStore = useWorkspaceStore()
          workspaceStore.clearActiveSessionForWorkspace(workspaceUuid)
        }
      } catch (error) {
        console.error('Failed to delete session:', error)
        throw error
      }
    },

    async setActiveSession(workspaceUuid: string | null, sessionUuid: string) {
      if (this.isSwitchingSession) {
        return
      }

      this.isSwitchingSession = true

      try {
        this.activeSessionUuid = sessionUuid

        // Update workspace active session tracking
        if (workspaceUuid) {
          const workspaceStore = useWorkspaceStore()
          workspaceStore.setActiveSessionForWorkspace(workspaceUuid, sessionUuid)
        }

        // Navigate to the session
        await this.navigateToSession(sessionUuid)
      } catch (error) {
        console.error('Failed to set active session:', error)
        throw error
      } finally {
        this.isSwitchingSession = false
      }
    },

    async setNextActiveSession(workspaceUuid: string | null) {
      if (workspaceUuid && this.workspaceHistory[workspaceUuid]?.length > 0) {
        // Set first available session in the same workspace
        const nextSession = this.workspaceHistory[workspaceUuid][0]
        await this.setActiveSession(workspaceUuid, nextSession.uuid)
      } else {
        // Find any available session
        for (const [wUuid, sessions] of Object.entries(this.workspaceHistory)) {
          if (sessions.length > 0) {
            await this.setActiveSession(wUuid, sessions[0].uuid)
            return
          }
        }
        // No sessions available
        this.activeSessionUuid = null
      }
    },

    async navigateToSession(sessionUuid: string) {
      const session = this.getChatSessionByUuid(sessionUuid)
      if (session && session.workspaceUuid) {
        const workspaceStore = useWorkspaceStore()
        await workspaceStore.navigateToWorkspace(session.workspaceUuid, sessionUuid)
      } else {
        // Fallback to legacy route
        await router.push({ name: 'Chat', params: { uuid: sessionUuid } })
      }
    },

    // Helper method to clear all sessions for a workspace
    clearWorkspaceSessions(workspaceUuid: string) {
      this.workspaceHistory[workspaceUuid] = []

      // Clear active session if it was in this workspace
      const activeSession = this.activeSession
      if (activeSession && activeSession.workspaceUuid === workspaceUuid) {
        this.activeSessionUuid = null
      }
    },

    // Helper method to get all sessions across all workspaces
    getAllSessions() {
      const allSessions: Chat.Session[] = []
      for (const sessions of Object.values(this.workspaceHistory)) {
        allSessions.push(...sessions)
      }
      return allSessions
    },

    // Legacy compatibility method - maps to createSessionInWorkspace
    async addSession(session: Chat.Session) {
      return await this.createSessionInWorkspace(session.title, session.workspaceUuid, session.model)
    },
  },
})