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
  type Workspace,
} from '@/api'

export interface WorkspaceState {
  workspaces: Workspace[]
  activeWorkspaceUuid: string | null
  workspaceActiveSessions: Record<string, string> // workspaceUuid -> sessionUuid
  isLoading: boolean
}

export const useWorkspaceStore = defineStore('workspace-store', {
  state: (): WorkspaceState => ({
    workspaces: [],
    activeWorkspaceUuid: null,
    workspaceActiveSessions: {},
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
        return state.getWorkspaceByUuid(state.activeWorkspaceUuid)
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
    async syncWorkspaces() {
      try {
        this.isLoading = true
        const workspaces = await getWorkspaces()
        this.workspaces = workspaces

        // Ensure we have a default workspace
        const defaultWorkspace = this.getDefaultWorkspace
        if (!defaultWorkspace) {
          await this.ensureDefaultWorkspace()
        }

        // Set active workspace if not already set
        if (!this.activeWorkspaceUuid && this.workspaces.length > 0) {
          this.activeWorkspaceUuid = this.getDefaultWorkspace?.uuid || this.workspaces[0].uuid
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
      const workspace = this.getWorkspaceByUuid(workspaceUuid)
      if (workspace) {
        this.activeWorkspaceUuid = workspaceUuid
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

    async updateWorkspace(workspaceUuid: string, updates: Partial<Workspace>) {
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
          const defaultWorkspace = this.getDefaultWorkspace
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
        await updateWorkspaceOrder(workspaceUuids)
        // Reorder workspaces locally
        const reorderedWorkspaces: Workspace[] = []
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