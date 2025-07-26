import { defineStore } from 'pinia'
import { getChatKeys, getLocalState } from './helper'
import { router } from '@/router'
import { store } from '@/store'
import {
  clearSessionChatMessages,
  createChatSession,
  createOrUpdateUserActiveChatSession,
  deleteChatData,
  deleteChatSession,
  updateChatSession as fetchUpdateChatByUuid,
  getChatSessionDefault,
  getChatMessagesBySessionUUID,
  getChatSessionsByUser,
  getUserActiveChatSession,
  renameChatSession,
  getWorkspaces,
  createWorkspace,
  updateWorkspace,
  deleteWorkspace,
  ensureDefaultWorkspace,
  setDefaultWorkspace,
  updateWorkspaceOrder,
  createSessionInWorkspace,
  migrateSessionsToDefaultWorkspace,
  getAllWorkspaceActiveSessions,
  setWorkspaceActiveSession,
  autoMigrateLegacySessions,
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
} from '@/api'

import { t } from '@/locales'

// Session creation lock to prevent race conditions
let isCreatingSession = false

export const useChatStore = defineStore('chat-store', {
  state: (): Chat.ChatState => getLocalState(),

  getters: {
    getChatSessionByCurrentActive(state: Chat.ChatState) {
      const index = state.history.findIndex(
        item => item.uuid === state.active,
      )
      if (index !== -1)
        return state.history[index]
      return null
    },

    getChatSessionByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.history.find(item => item.uuid === uuid)
        return (
          state.history.find(item => item.uuid === state.active)
        )
      }
    },

    getChatSessionDataByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.chat[uuid] ?? []
        if (state.active)
          return state.chat[state.active] ?? []
        return []
      }
    },

    getWorkspaceByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.workspaces.find(workspace => workspace.uuid === uuid)
        if (state.activeWorkspace)
          return state.workspaces.find(workspace => workspace.uuid === state.activeWorkspace)
        return null
      }
    },

    getDefaultWorkspace(state: Chat.ChatState) {
      return state.workspaces.find(workspace => workspace.isDefault) || null
    },

    getSessionsByWorkspace(state: Chat.ChatState) {
      return (workspaceUuid?: string) => {
        if (!workspaceUuid) return []
        return state.history.filter(session => session.workspaceUuid === workspaceUuid)
      }
    },

    // Get active session for a specific workspace
    getActiveSessionForWorkspace(state: Chat.ChatState) {
      return (workspaceUuid: string) => {
        return state.workspaceActiveSessions[workspaceUuid] || null
      }
    },

    // Get active session for current workspace
    getCurrentWorkspaceActiveSession(state: Chat.ChatState) {
      if (!state.activeWorkspace) return null
      return state.workspaceActiveSessions[state.activeWorkspace] || null
    },
  },

  actions: {
    async reloadRoute(uuid?: string) {
      // Use workspace-aware routing if active workspace exists
      if (this.activeWorkspace && uuid) {
        // Find the session to verify it belongs to the active workspace
        const session = this.getChatSessionByUuid(uuid)
        if (session && session.workspaceUuid === this.activeWorkspace) {
          await router.push({ 
            name: 'WorkspaceChat', 
            params: { 
              workspaceUuid: this.activeWorkspace,
              uuid 
            } 
          })
          return
        }
      }
      
      // Fallback to legacy routing
      await router.push({ name: 'Chat', params: { uuid } })
    },

    // Helper method to get workspace-aware URL  
    getSessionUrl(sessionUuid: string): string {
      const session = this.getChatSessionByUuid(sessionUuid)
      if (session && session.workspaceUuid) {
        return `/#/workspace/${session.workspaceUuid}/chat/${sessionUuid}`
      }
      return `/#/chat/${sessionUuid}`
    },

    async syncChatSessions() {
      try {
        // First sync workspaces
        await this.syncWorkspaces()
        
        // Auto-migrate any legacy sessions before doing anything else
        try {
          const migrationResult = await autoMigrateLegacySessions()
          if (migrationResult.hasLegacySessions && migrationResult.migratedSessions > 0) {
            console.log(`🔄 Auto-migrated ${migrationResult.migratedSessions} legacy sessions to default workspace`)
            // Force page refresh to reload with migrated data
            window.location.reload()
            return // Exit early since we're refreshing
          }
        } catch (migrationError) {
          console.warn('⚠️ Legacy session migration failed:', migrationError)
          // Continue with normal sync - don't block the app
        }
        
        // Sync workspace active sessions from backend
        await this.syncWorkspaceActiveSessions()
        
        this.history = await getChatSessionsByUser()
        console.log('📋 Synced sessions from DB:', this.history.length)

        // Check if any sessions need workspace assignment (migration)
        await this.migrateSessionsToDefaultWorkspace()

        if (this.history.length === 0) {
          const new_chat_text = t('chat.new')
          await this.addChatSession(await getChatSessionDefault(new_chat_text))
        }

        let active_session_uuid = this.history[0].uuid

        try {
          const active_session = await getUserActiveChatSession()
          if (active_session) {
            active_session_uuid = active_session.chatSessionUuid
          }
        } catch (activeError) {
          // No active session found, using default
        }

        this.active = active_session_uuid

        if (router.currentRoute.value.params.uuid !== this.active) {
          await this.reloadRoute(this.active)
        }
      } catch (error) {
        console.error('❌ Error in syncChatSessions:', error)
        throw error
      }
    },

    async syncChatMessages(need_uuid: string) {
      if (need_uuid) {
        const messageData = await getChatMessagesBySessionUUID(need_uuid)
        this.chat[need_uuid] = messageData
        await createOrUpdateUserActiveChatSession(need_uuid)
        this.setActiveLocal(need_uuid)
        //await this.reloadRoute(this.active) // !!! this cause cycle
      }
    },

    async addChatSession(history: Chat.Session, chatData: Chat.Message[] = []) {
      await createChatSession(history.uuid, history.title, history.model)
      this.history.unshift(history)
      this.chat[history.uuid] = chatData
      this.active = history.uuid
      this.reloadRoute(this.active)
    },

    async updateChatSession(uuid: string, edit: Partial<Chat.Session>) {
      const index = this.history.findIndex(item => item.uuid === uuid)
      if (index !== -1) {
        this.history[index] = { ...this.history[index], ...edit }
        await fetchUpdateChatByUuid(uuid, this.history[index])
      }
    },

    async updateChatSessionIfEdited(uuid: string, edit: Partial<Chat.Session>) {
      const index = this.history.findIndex(item => item.uuid === uuid)
      if (index !== -1) {
        if (this.history[index].isEdit) {
          this.history[index] = { ...this.history[index], ...edit }
          await fetchUpdateChatByUuid(uuid, this.history[index])
        }
      }
    },

    deleteChatSession(index: number) {
      deleteChatSession(this.history[index].uuid)
      delete this.chat[this.history[index].uuid]
      this.history.splice(index, 1)

      if (this.history.length === 0) {
        this.active = null
        this.reloadRoute()
        return
      }

      if (index > 0 && index <= this.history.length) {
        const uuid = this.history[index - 1].uuid
        this.setActive(uuid)
        return
      }

      if (index === 0) {
        if (this.history.length > 0) {
          const uuid = this.history[0].uuid
          this.setActive(uuid)
        }
      }

      if (index > this.history.length) {
        const uuid = this.history[this.history.length - 1].uuid
        this.setActive(uuid)
      }
    },

    async setActive(uuid: string) {
      this.active = uuid
      
      // Also update the active session for the current workspace
      const session = this.getChatSessionByUuid(uuid)
      if (session && session.workspaceUuid) {
        // Don't await here to keep setActive fast - this is fire-and-forget
        this.setActiveSessionForWorkspace(session.workspaceUuid, uuid)
      }
      
      await createOrUpdateUserActiveChatSession(uuid)
      await this.reloadRoute(uuid)
    },

    async setActiveLocal(uuid: string) {
      this.active = uuid
      
      // Also update the active session for the current workspace
      const session = this.getChatSessionByUuid(uuid)
      if (session && session.workspaceUuid) {
        // Don't await here to keep setActiveLocal fast - this is fire-and-forget
        this.setActiveSessionForWorkspace(session.workspaceUuid, uuid)
      }
    },

    getChatByUuidAndIndex(uuid: string, index: number) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length)
          return this.chat[keys[0]][index]
        return null
      }
      if (keys.includes(uuid))
        return this.chat[uuid][index]
      return null
    },

    async addChatByUuid(uuid: string, chat: Chat.Message) {
      const new_chat_text = t('chat.new')
      const [keys] = getChatKeys(this.chat, false)

      if (!uuid) {
        if (this.history.length === 0) {
          if (isCreatingSession) {
            console.log('🚨 RACE CONDITION BLOCKED: Session creation in progress, skipping')
            return
          }

          console.log('🔒 Creating new session, acquiring lock')
          isCreatingSession = true

          try {
            const default_model_parameters = await getChatSessionDefault(new_chat_text)
            const uuid = default_model_parameters.uuid;
            await createChatSession(uuid, chat.text, default_model_parameters.model)
            this.history.push({ uuid, title: chat.text, isEdit: false })
            this.chat[uuid] = [{ ...chat, isPrompt: true, isPin: false }]
            this.active = uuid
            console.log('✅ Session created successfully:', uuid)
          } catch (error) {
            console.error('❌ Session creation failed:', error)
            throw error
          } finally {
            isCreatingSession = false
          }
        }
        else {
          this.chat[keys[0]].push(chat)
          if (this.history[0].title === new_chat_text) {
            this.history[0].title = chat.text
            renameChatSession(this.history[0].uuid, chat.text.substring(0, 40))
          }
        }
      }

      if (keys.includes(uuid)) {
        if (this.chat[uuid].length === 0)
          this.chat[uuid].push({ ...chat, isPrompt: true, isPin: false })
        else
          this.chat[uuid].push(chat)

        if (this.history[0].title === new_chat_text) {
          this.history[0].title = chat.text
          renameChatSession(this.history[0].uuid, chat.text.substring(0, 40))
        }
      }
    },

    async updateChatByUuid(uuid: string, index: number, chat: Chat.Message) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]][index] = chat
        }
        return
      }

      if (keys.includes(uuid)) {
        this.chat[uuid][index] = chat
      }
    },

    updateChatPartialByUuid(
      uuid: string,
      index: number,
      chat: Partial<Chat.Message>,
    ) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]][index] = { ...this.chat[keys[0]][index], ...chat }
        }
        return
      }

      if (keys.includes(uuid)) {
        this.chat[uuid][index] = {
          ...this.chat[uuid][index],
          ...chat,
        }
      }
    },

    async deleteChatByUuid(uuid: string, index: number) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length) {
          const chatData = this.chat[keys[0]]
          const chat = chatData[index]
          chatData.splice(index, 1)
          if (chat && chat.uuid)
            await deleteChatData(chat)
        }
        return
      }

      if (keys.includes(uuid)) {
        const chatData = this.chat[uuid]
        const chat = chatData[index]
        chatData.splice(index, 1)
        if (chat && chat.uuid)
          await deleteChatData(chat)
      }
    },

    clearChatByUuid(uuid: string) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]] = []
        }
        return
      }
      if (keys.includes(uuid)) {
        const data: Chat.Message[] = []
        for (const chat of this.chat[uuid]) {
          if (chat.isPin || chat.isPrompt)
            data.push(chat)
        }
        this.chat[uuid] = data
        clearSessionChatMessages(uuid)
      }
    },
    // Workspace management actions
    async syncWorkspaces() {
      try {
        this.workspaces = await getWorkspaces()
        console.log('📁 Synced workspaces from DB:', this.workspaces.length)
        
        // Ensure user has a default workspace
        if (this.workspaces.length === 0) {
          try {
            const defaultWorkspace = await ensureDefaultWorkspace()
            this.workspaces.push(defaultWorkspace)
            console.log('✅ Default workspace ensured:', defaultWorkspace.name)
          } catch (ensureError: any) {
            console.warn('⚠️ Failed to ensure default workspace, continuing with empty workspace list')
            // Don't throw here - allow app to continue functioning
            // User can manually create workspaces via UI
          }
        }

        // Set active workspace to default if none selected
        if (!this.activeWorkspace) {
          const defaultWorkspace = this.getDefaultWorkspace
          if (defaultWorkspace) {
            this.activeWorkspace = defaultWorkspace.uuid
          }
        }
      } catch (error) {
        console.error('❌ Error in syncWorkspaces:', error)
        // Set fallback state to prevent app breakage
        this.workspaces = []
        this.activeWorkspace = null
        // Don't throw - allow app to continue with empty workspace state
      }
    },

    async createNewWorkspace(data: CreateWorkspaceRequest) {
      try {
        const workspace = await createWorkspace(data)
        this.workspaces.push(workspace)
        return workspace
      } catch (error) {
        console.error('❌ Error creating workspace:', error)
        throw error
      }
    },

    async updateWorkspaceData(uuid: string, data: UpdateWorkspaceRequest) {
      try {
        const workspace = await updateWorkspace(uuid, data)
        const index = this.workspaces.findIndex(w => w.uuid === uuid)
        if (index !== -1) {
          this.workspaces[index] = workspace
        }
        return workspace
      } catch (error) {
        console.error('❌ Error updating workspace:', error)
        throw error
      }
    },

    async deleteWorkspaceData(uuid: string) {
      try {
        await deleteWorkspace(uuid)
        const index = this.workspaces.findIndex(w => w.uuid === uuid)
        if (index !== -1) {
          this.workspaces.splice(index, 1)
        }
        
        // If deleted workspace was active, switch to default
        if (this.activeWorkspace === uuid) {
          const defaultWorkspace = this.getDefaultWorkspace
          this.activeWorkspace = defaultWorkspace?.uuid || null
        }
      } catch (error) {
        console.error('❌ Error deleting workspace:', error)
        throw error
      }
    },

    async setWorkspaceAsDefault(uuid: string) {
      try {
        const workspace = await setDefaultWorkspace(uuid)
        // Update all workspaces to reflect the new default
        this.workspaces = this.workspaces.map(w => ({
          ...w,
          isDefault: w.uuid === uuid
        }))
        return workspace
      } catch (error) {
        console.error('❌ Error setting default workspace:', error)
        throw error
      }
    },

    async updateWorkspaceOrder(uuid: string, orderPosition: number) {
      try {
        const workspace = await updateWorkspaceOrder(uuid, orderPosition)
        const index = this.workspaces.findIndex(w => w.uuid === uuid)
        if (index !== -1) {
          this.workspaces[index] = workspace
        }
        return workspace
      } catch (error) {
        console.error('❌ Error updating workspace order:', error)
        throw error
      }
    },

    setActiveWorkspace(uuid: string) {
      this.activeWorkspace = uuid
    },

    // Set active session for a specific workspace
    async setActiveSessionForWorkspace(workspaceUuid: string, sessionUuid: string | null) {
      // Update local state immediately for responsive UI
      this.workspaceActiveSessions[workspaceUuid] = sessionUuid
      
      // Persist to backend asynchronously
      if (sessionUuid) {
        try {
          await setWorkspaceActiveSession(workspaceUuid, sessionUuid)
          console.log(`✅ Persisted active session for workspace ${workspaceUuid}: ${sessionUuid}`)
        } catch (error) {
          console.warn(`⚠️ Failed to persist active session for workspace ${workspaceUuid}:`, error)
          // Don't revert local state - user can continue working
        }
      }
    },

    // Set active session for current workspace
    setActiveSessionForCurrentWorkspace(sessionUuid: string | null) {
      if (this.activeWorkspace) {
        this.workspaceActiveSessions[this.activeWorkspace] = sessionUuid
      }
      // Also update global active for backward compatibility
      this.active = sessionUuid
    },

    // Sync workspace active sessions from backend
    async syncWorkspaceActiveSessions() {
      try {
        const backendSessions = await getAllWorkspaceActiveSessions()
        // Convert array to object format
        const workspaceActiveSessions: { [workspaceUuid: string]: string | null } = {}
        for (const session of backendSessions) {
          workspaceActiveSessions[session.workspaceUuid] = session.chatSessionUuid
        }
        this.workspaceActiveSessions = workspaceActiveSessions
        console.log('✅ Synced workspace active sessions from backend:', Object.keys(workspaceActiveSessions).length)
      } catch (error) {
        console.warn('⚠️ Failed to sync workspace active sessions:', error)
        // Continue with empty state - not critical for app functionality
        this.workspaceActiveSessions = {}
      }
    },

    // Switch workspace and update URL accordingly
    async switchToWorkspace(workspaceUuid: string) {
      this.setActiveWorkspace(workspaceUuid)
      
      // Check if this workspace has a previously active session
      const previousActiveSession = this.getActiveSessionForWorkspace(workspaceUuid)
      const workspaceSessions = this.getSessionsByWorkspace(workspaceUuid)
      
      // If there was a previous active session and it still exists, restore it
      if (previousActiveSession && workspaceSessions.some(s => s.uuid === previousActiveSession)) {
        await this.setActive(previousActiveSession)
        return
      }
      
      if (workspaceSessions.length > 0) {
        // Navigate to the first session in the workspace and save it as active
        const firstSession = workspaceSessions[0]
        await this.setActiveSessionForWorkspace(workspaceUuid, firstSession.uuid)
        await this.setActive(firstSession.uuid)
      } else {
        // Create a default session in the empty workspace (similar to syncChatSessions behavior)
        try {
          const new_chat_text = t('chat.new')
          await this.createSessionInActiveWorkspace(new_chat_text)
          console.log('✅ Created default session in new workspace')
        } catch (error) {
          console.error('❌ Failed to create default session in workspace:', error)
          // Fallback to navigation without session
          await router.push({ 
            name: 'WorkspaceChat', 
            params: { workspaceUuid } 
          })
          await this.setActiveSessionForWorkspace(workspaceUuid, null)
          this.active = null
        }
      }
    },

    async migrateSessionsToDefaultWorkspace() {
      try {
        // Find sessions without workspace assignment
        const sessionsWithoutWorkspace = this.history.filter(session => !session.workspaceUuid)
        
        if (sessionsWithoutWorkspace.length === 0) {
          console.log('📁 No sessions need workspace migration')
          return
        }

        const defaultWorkspace = this.getDefaultWorkspace
        if (!defaultWorkspace) {
          console.warn('⚠️  No default workspace found for migration')
          return
        }

        console.log(`📁 Migrating ${sessionsWithoutWorkspace.length} sessions to default workspace`)
        
        // Call backend migration API to persist changes to database
        await migrateSessionsToDefaultWorkspace()
        console.log('✅ Backend migration completed')
        
        // Update frontend state to reflect migration
        for (const session of sessionsWithoutWorkspace) {
          const index = this.history.findIndex(s => s.uuid === session.uuid)
          if (index !== -1) {
            this.history[index] = {
              ...this.history[index],
              workspaceUuid: defaultWorkspace.uuid
            }
          }
        }
        
        console.log('✅ Session workspace migration completed')
      } catch (error) {
        console.error('❌ Error migrating sessions to default workspace:', error)
      }
    },

    async createSessionInActiveWorkspace(topic: string, model?: string) {
      try {
        if (!this.activeWorkspace) {
          throw new Error('No active workspace selected')
        }
        
        const sessionData = await createSessionInWorkspace(this.activeWorkspace, { topic, model })
        
        // Add to history with workspace context
        const session: Chat.Session = {
          uuid: sessionData.uuid,
          title: sessionData.topic,
          isEdit: false,
          model: sessionData.model,
          workspaceUuid: this.activeWorkspace
        }
        
        this.history.unshift(session)
        this.chat[sessionData.uuid] = []
        this.active = sessionData.uuid
        
        // Set as active session for the workspace
        await this.setActiveSessionForWorkspace(this.activeWorkspace, sessionData.uuid)
        
        await this.reloadRoute(sessionData.uuid)
        return sessionData
      } catch (error) {
        console.error('❌ Error creating session in workspace:', error)
        throw error
      }
    },

    clearState() {
      this.history = []
      this.chat = {}
      this.active = null
      this.activeWorkspace = null
      this.workspaceActiveSessions = {}
      this.workspaces = []
    },
  },
})

export function useChatStoreWithout() {
  return useChatStore(store)
}
