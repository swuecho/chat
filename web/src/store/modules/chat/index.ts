import { defineStore } from 'pinia'
import { getChatKeys, getLocalState } from './helper'
import { router } from '@/router'
import {
  clearSessionChatMessages,
  createChatSession,
  deleteChatData,
  deleteChatSession,
  updateChatSession as fetchUpdateChatByUuid,
  getChatSessionDefault,
  getChatMessagesBySessionUUID,
  getChatSessionsByUser,
  renameChatSession,
  getWorkspaces,
  getSessionsByWorkspace,
  createWorkspace,
  updateWorkspace,
  deleteWorkspace,
  ensureDefaultWorkspace,
  setDefaultWorkspace,
  updateWorkspaceOrder,
  createSessionInWorkspace,
  getAllWorkspaceActiveSessions,
  setWorkspaceActiveSession,
  autoMigrateLegacySessions,
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
  fetchDefaultChatModel,
} from '@/api'

import { t } from '@/locales'

// Session creation lock to prevent race conditions
let isCreatingSession = false

// Navigation lock to prevent race conditions during route changes
let isNavigating = false

// Session switching lock to prevent race conditions during session changes
let isSwitchingSession = false

export const useChatStore = defineStore('chat-store', {
  state: (): Chat.ChatState => getLocalState(),

  getters: {
    getChatSessionByUuid(state: Chat.ChatState) {
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

    getChatSessionDataByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.chat[uuid] ?? []
        return []
      }
    },

    getWorkspaceByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.workspaces.find(workspace => workspace.uuid === uuid)
        return null
      }
    },

    getDefaultWorkspace(state: Chat.ChatState) {
      return state.workspaces.find(workspace => workspace.isDefault) || null
    },

    getSessionsByWorkspace(state: Chat.ChatState) {
      return (workspaceUuid?: string) => {
        if (!workspaceUuid) return []
        return state.workspaceHistory[workspaceUuid] || []
      }
    },

    // Get active session for a specific workspace
    getActiveSessionForWorkspace(state: Chat.ChatState) {
      return (workspaceUuid: string) => {
        // First check if we have a stored active session for this workspace
        if (state.workspaceActiveSessions[workspaceUuid]) {
          return state.workspaceActiveSessions[workspaceUuid]
        }
        // Fallback to current active session if it matches this workspace
        if (state.activeSession.workspaceUuid === workspaceUuid) {
          return state.activeSession.sessionUuid
        }
        return null
      }
    },

    // Get active workspace UUID
    activeWorkspace(state: Chat.ChatState) {
      return state.activeSession.workspaceUuid
    },

    // Get active session UUID (legacy compatibility)
    active(state: Chat.ChatState) {
      return state.activeSession.sessionUuid
    },

    // Get current active chat session
    getChatSessionByCurrentActive(state: Chat.ChatState) {
      if (!state.activeSession.sessionUuid) return null
      // Search across all workspace histories
      for (const sessions of Object.values(state.workspaceHistory)) {
        const session = sessions.find(item => item.uuid === state.activeSession.sessionUuid)
        if (session) return session
      }
      return null
    },
  },

  actions: {
    async reloadRoute(uuid?: string) {
      // Prevent concurrent navigation
      if (isNavigating) {
        console.log('🚫 Navigation already in progress, skipping')
        return
      }

      isNavigating = true
      
      try {
        if (uuid) {
          const session = this.getChatSessionByUuid(uuid)
          if (session && session.workspaceUuid) {
            // Only navigate if we're not already on the correct route
            const currentRoute = router.currentRoute.value
            const isCorrectRoute = currentRoute.name === 'WorkspaceChat' && 
                                  currentRoute.params.workspaceUuid === session.workspaceUuid &&
                                  currentRoute.params.uuid === uuid
            
            if (!isCorrectRoute) {
              console.log('🚀 Navigating to workspace route:', { workspaceUuid: session.workspaceUuid, uuid })
              await router.push({
                name: 'WorkspaceChat',
                params: {
                  workspaceUuid: session.workspaceUuid,
                  uuid
                }
              })
            }
            return
          }
        }
        
        // If no specific session/workspace, navigate to default workspace
        const defaultWorkspace = this.getDefaultWorkspace
        if (defaultWorkspace) {
          console.log('🚀 Navigating to default workspace:', defaultWorkspace.uuid)
          await router.push({
            name: 'WorkspaceChat',
            params: {
              workspaceUuid: defaultWorkspace.uuid,
              uuid: uuid || ''
            }
          })
        } else {
          // Fallback to root if no default workspace
          console.log('🚀 No default workspace, navigating to root')
          await router.push({ name: 'DefaultWorkspace' })
        }
      } finally {
        isNavigating = false
      }
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
        // Auto-migrate any legacy sessions before doing anything else
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

        // First sync workspaces
        await this.syncWorkspaces()

        // Check if we should preserve workspace from URL before syncing backend data
        const routeBeforeSync = router.currentRoute.value
        const urlWorkspaceUuid = routeBeforeSync.name === 'WorkspaceChat' ? routeBeforeSync.params.workspaceUuid as string : null
        const urlSessionUuid = routeBeforeSync.params.uuid as string
        const isOnDefaultRoute = routeBeforeSync.name === 'DefaultWorkspace'
        
        // Sync workspace active sessions from backend, but preserve URL context
        await this.syncWorkspaceActiveSessions(urlWorkspaceUuid || undefined, urlSessionUuid || undefined)

        // Ensure we have an active workspace, preserving URL context
        await this.ensureActiveWorkspace()
        
        // If we had a workspace in URL but lost it, restore it
        if (urlWorkspaceUuid && this.getWorkspaceByUuid(urlWorkspaceUuid) && this.activeSession.workspaceUuid !== urlWorkspaceUuid) {
          this.activeSession.workspaceUuid = urlWorkspaceUuid
          // Also restore session UUID if it was in the URL
          if (urlSessionUuid && this.getChatSessionByUuid(urlSessionUuid)) {
            this.activeSession.sessionUuid = urlSessionUuid
          }
          console.log('✅ Restored workspace and session from URL after sync:', { workspaceUuid: urlWorkspaceUuid, sessionUuid: urlSessionUuid })
        }

        // Only sync sessions for active workspace initially
        this.workspaceHistory = {}
        
        if (this.activeSession.workspaceUuid) {
          console.log('📋 Syncing sessions for active workspace:', this.activeSession.workspaceUuid)
          const workspaceSessions = await getSessionsByWorkspace(this.activeSession.workspaceUuid)
          this.workspaceHistory[this.activeSession.workspaceUuid] = workspaceSessions
          console.log('📋 Synced sessions:', workspaceSessions.length)
        } else {
          // Fallback: if no active workspace, load default workspace sessions
          const defaultWorkspace = this.getDefaultWorkspace
          if (defaultWorkspace) {
            console.log('📋 Syncing sessions for default workspace:', defaultWorkspace.uuid)
            const workspaceSessions = await getSessionsByWorkspace(defaultWorkspace.uuid)
            this.workspaceHistory[defaultWorkspace.uuid] = workspaceSessions
            console.log('📋 Synced sessions:', workspaceSessions.length)
          }
        }

        if (Object.keys(this.workspaceHistory).length === 0 || Object.values(this.workspaceHistory).every(sessions => sessions.length === 0)) {
          console.log('🔄 No sessions found for user, creating default session in default workspace')
          
          // Ensure we have a default workspace before creating a session
          const defaultWorkspace = this.getDefaultWorkspace
          if (!defaultWorkspace) {
            console.error('❌ No default workspace found when trying to create default session')
            throw new Error('No default workspace available for session creation')
          }
          
          console.log('🔄 Creating default session in workspace:', defaultWorkspace.name)
          
          // Create session directly in the default workspace instead of using legacy addChatSession
          try {
            // Set active workspace for session creation
            this.activeSession.workspaceUuid = defaultWorkspace.uuid
            
            const new_chat_text = t('chat.new')
            await this.createSessionInActiveWorkspace(new_chat_text)
            console.log('✅ Created default session in default workspace for new user')
          } catch (error) {
            console.error('❌ Failed to create default session in workspace:', error)
            // Fallback to legacy method if workspace creation fails
            const new_chat_text = t('chat.new')
            const defaultSession = await getChatSessionDefault(new_chat_text)
            // Assign to default workspace
            defaultSession.workspaceUuid = defaultWorkspace.uuid
            await this.addChatSession(defaultSession)
            console.log('✅ Created default session using fallback method')
          }
        }

        // Handle navigation based on current route and active session
        const currentRoute = router.currentRoute.value
        const isOnWorkspaceRoute = currentRoute.name === 'WorkspaceChat'
        
        if (this.activeSession.sessionUuid) {
          const session = this.getChatSessionByUuid(this.activeSession.sessionUuid)
          
          // Only reload route if we're not already on the correct route
          const shouldReload = session && (
            // Different session UUID in URL
            router.currentRoute.value.params.uuid !== this.activeSession.sessionUuid ||
            // Wrong workspace in URL (if session has workspace)
            (session.workspaceUuid && isOnWorkspaceRoute && router.currentRoute.value.params.workspaceUuid !== session.workspaceUuid) ||
            // We're on default route but have an active session
            isOnDefaultRoute
          )
          
          if (shouldReload) {
            console.log('✅ Reloading route to match active session:', this.activeSession.sessionUuid)
            await this.reloadRoute(this.activeSession.sessionUuid)
          } else {
            console.log('✅ Route already matches active session, no reload needed')
          }
        } else if (Object.values(this.workspaceHistory).some(sessions => sessions.length > 0)) {
          // Set first available session as active if no active session
          let firstSession = null
          for (const sessions of Object.values(this.workspaceHistory)) {
            if (sessions.length > 0) {
              firstSession = sessions[0]
              break
            }
          }
          if (firstSession) {
            const workspaceUuid = firstSession.workspaceUuid || this.getDefaultWorkspace?.uuid || null
            await this.setActiveSession(workspaceUuid, firstSession.uuid)
          }
        } else if (isOnDefaultRoute) {
          // If we're on default route but have no sessions, navigate to default workspace
          const defaultWorkspace = this.getDefaultWorkspace
          if (defaultWorkspace) {
            console.log('✅ Navigating from default route to default workspace')
            await router.push({
              name: 'WorkspaceChat',
              params: {
                workspaceUuid: defaultWorkspace.uuid
              }
            })
          }
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
        const session = this.getChatSessionByUuid(need_uuid)
        
        // Only set active session if it's different from current and not already switching
        if (this.activeSession.sessionUuid !== need_uuid && !isSwitchingSession) {
          await this.setActiveSession(session?.workspaceUuid || null, need_uuid)
        } else if (session?.workspaceUuid && this.activeSession.workspaceUuid !== session.workspaceUuid) {
          // Just update the workspace if needed without triggering route reload
          this.setActiveSessionLocal(session.workspaceUuid, need_uuid)
        }
      }
    },

    async addChatSession(history: Chat.Session, chatData: Chat.Message[] = []) {
      await createChatSession(history.uuid, history.title, history.model)
      
      // Refresh workspace sessions to get updated list from backend
      const workspaceUuid = history.workspaceUuid || this.getDefaultWorkspace?.uuid
      if (workspaceUuid) {
        await this.syncWorkspaceSessions(workspaceUuid)
      }
      
      this.chat[history.uuid] = chatData
      await this.setActiveSession(history.workspaceUuid || null, history.uuid)
      this.reloadRoute(history.uuid)
    },

    async updateChatSession(uuid: string, edit: Partial<Chat.Session>) {
      // Find session across all workspace histories
      for (const workspaceUuid in this.workspaceHistory) {
        const sessions = this.workspaceHistory[workspaceUuid]
        const index = sessions.findIndex(item => item.uuid === uuid)
        if (index !== -1) {
          sessions[index] = { ...sessions[index], ...edit }
          await fetchUpdateChatByUuid(uuid, sessions[index])
          return
        }
      }
    },

    async updateChatSessionIfEdited(uuid: string, edit: Partial<Chat.Session>) {
      // Find session across all workspace histories
      for (const workspaceUuid in this.workspaceHistory) {
        const sessions = this.workspaceHistory[workspaceUuid]
        const index = sessions.findIndex(item => item.uuid === uuid)
        if (index !== -1) {
          if (sessions[index].isEdit) {
            sessions[index] = { ...sessions[index], ...edit }
            await fetchUpdateChatByUuid(uuid, sessions[index])
          }
          return
        }
      }
    },

    async deleteChatSession(sessionUuid: string) {
      let deletedSession: Chat.Session | null = null
      let workspaceUuid: string | null = null
      
      // Find the session to delete
      for (const wUuid in this.workspaceHistory) {
        const sessions = this.workspaceHistory[wUuid]
        const session = sessions.find(item => item.uuid === sessionUuid)
        if (session) {
          deletedSession = session
          workspaceUuid = wUuid
          break
        }
      }
      
      if (!deletedSession || !workspaceUuid) return
      
      // Delete from backend
      await deleteChatSession(deletedSession.uuid)
      delete this.chat[deletedSession.uuid]

      // Refresh workspace sessions to get updated list from backend
      await this.syncWorkspaceSessions(workspaceUuid)
      
      // Check if any sessions remain across all loaded workspaces
      const hasAnySessions = Object.values(this.workspaceHistory).some(sessions => sessions.length > 0)
      
      if (!hasAnySessions) {
        this.activeSession = { sessionUuid: null, workspaceUuid: null }
        this.reloadRoute()
        return
      }

      // Find next session in the same workspace, or fall back to any workspace
      let nextSession: Chat.Session | null = null
      const workspaceSessions = this.workspaceHistory[workspaceUuid] || []
      
      if (workspaceSessions.length > 0) {
        // Stay in same workspace - pick first available session
        nextSession = workspaceSessions[0]
      } else {
        // Find first session in any workspace
        for (const sessions of Object.values(this.workspaceHistory)) {
          if (sessions.length > 0) {
            nextSession = sessions[0]
            break
          }
        }
      }

      if (nextSession) {
        await this.setActiveSession(nextSession.workspaceUuid || null, nextSession.uuid)
      }
    },

    async setActiveSession(workspaceUuid: string | null, sessionUuid: string) {
      // Prevent concurrent session switching
      if (isSwitchingSession) {
        console.log('🚫 Session switch already in progress, skipping')
        return
      }

      // Check if we're already on this session to prevent unnecessary switching
      if (this.activeSession.sessionUuid === sessionUuid && this.activeSession.workspaceUuid === workspaceUuid) {
        console.log('✅ Already on target session, skipping switch')
        return
      }

      isSwitchingSession = true
      
      try {
        this.activeSession = { workspaceUuid, sessionUuid }

        // Store active session for this workspace
        if (workspaceUuid) {
          this.workspaceActiveSessions[workspaceUuid] = sessionUuid
          
          try {
            await setWorkspaceActiveSession(workspaceUuid, sessionUuid)
            console.log(`✅ Set active session: workspace=${workspaceUuid}, session=${sessionUuid}`)
          } catch (error) {
            console.warn('⚠️ Failed to persist active session:', error)
          }
        }

        await this.reloadRoute(sessionUuid)
      } finally {
        isSwitchingSession = false
      }
    },

    setActiveSessionLocal(workspaceUuid: string | null, sessionUuid: string) {
      this.activeSession = { workspaceUuid, sessionUuid }
      
      // Store active session for this workspace
      if (workspaceUuid) {
        this.workspaceActiveSessions[workspaceUuid] = sessionUuid
      }
    },

    setActiveWorkspace(workspaceUuid: string) {
      this.activeSession.workspaceUuid = workspaceUuid
    },

    setActiveLocal(sessionUuid: string) {
      this.activeSession.sessionUuid = sessionUuid
    },

    async setActive(sessionUuid: string) {
      // Prevent setting active session if already switching
      if (isSwitchingSession) {
        console.log('🚫 Cannot set active session - switch already in progress')
        return
      }
      
      const session = this.getChatSessionByUuid(sessionUuid)
      if (session) {
        await this.setActiveSession(session.workspaceUuid || null, sessionUuid)
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
        const hasAnySessions = Object.values(this.workspaceHistory).some(sessions => sessions.length > 0)
        if (!hasAnySessions) {
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
            const session = { uuid, title: chat.text, isEdit: false, workspaceUuid: this.activeSession.workspaceUuid || undefined }
            const wUuid = this.activeSession.workspaceUuid || this.getDefaultWorkspace?.uuid || 'default'
            if (!this.workspaceHistory[wUuid]) {
              this.workspaceHistory[wUuid] = []
            }
            this.workspaceHistory[wUuid].push(session)
            this.chat[uuid] = [{ ...chat, isPrompt: true, isPin: false }]
            await this.setActiveSession(this.activeSession.workspaceUuid, uuid)
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
          // Find the first session to update title
          let firstSession = null
          for (const sessions of Object.values(this.workspaceHistory)) {
            if (sessions.length > 0) {
              firstSession = sessions[0]
              break
            }
          }
          if (firstSession && firstSession.title === new_chat_text) {
            firstSession.title = chat.text
            renameChatSession(firstSession.uuid, chat.text.substring(0, 40))
          }
        }
      }

      if (keys.includes(uuid)) {
        if (this.chat[uuid].length === 0)
          this.chat[uuid].push({ ...chat, isPrompt: true, isPin: false })
        else
          this.chat[uuid].push(chat)

        // Find the first session to update title
        let firstSession = null
        for (const sessions of Object.values(this.workspaceHistory)) {
          if (sessions.length > 0) {
            firstSession = sessions[0]
            break
          }
        }
        if (firstSession && firstSession.title === new_chat_text) {
          firstSession.title = chat.text
          renameChatSession(firstSession.uuid, chat.text.substring(0, 40))
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

      } catch (error) {
        console.error('❌ Error in syncWorkspaces:', error)
        // Set fallback state to prevent app breakage
        this.workspaces = []
        this.activeSession.workspaceUuid = null
        // Don't throw - allow app to continue with empty workspace state
      }
    },

    async ensureActiveWorkspace() {
      // Check if we should preserve the current workspace (e.g., from URL)
      const currentRoute = router.currentRoute.value
      const isWorkspaceRoute = currentRoute.name === 'WorkspaceChat' && currentRoute.params.workspaceUuid
      
      // If we're on a workspace route, preserve that workspace
      if (isWorkspaceRoute) {
        const routeWorkspaceUuid = currentRoute.params.workspaceUuid as string
        const workspace = this.getWorkspaceByUuid(routeWorkspaceUuid)
        if (workspace) {
          this.activeSession.workspaceUuid = routeWorkspaceUuid
          console.log('✅ Preserved workspace from URL:', workspace.name)
          return
        }
      }
      
      // If we don't have an active workspace or are on default route, set to default workspace
      if ((!this.activeSession.workspaceUuid || currentRoute.name === 'DefaultWorkspace') && this.workspaces.length > 0) {
        const defaultWorkspace = this.getDefaultWorkspace || this.workspaces[0]
        if (defaultWorkspace) {
          this.activeSession.workspaceUuid = defaultWorkspace.uuid
          console.log('✅ Set active workspace to:', defaultWorkspace.name)
        }
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

        // If deleted workspace was active, clear active session
        if (this.activeSession.workspaceUuid === uuid) {
          this.activeSession = { sessionUuid: null, workspaceUuid: null }
        }
        
        // Remove from workspace active sessions mapping
        if (this.workspaceActiveSessions[uuid]) {
          delete this.workspaceActiveSessions[uuid]
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


    // Sync workspace active sessions from backend
    async syncWorkspaceActiveSessions(urlWorkspaceUuid?: string, urlSessionUuid?: string) {
      try {
        const backendSessions = await getAllWorkspaceActiveSessions()
        
        // Build workspace active sessions mapping
        this.workspaceActiveSessions = {}
        let globalActiveSession = null
        
        for (const session of backendSessions) {
          if (session.workspaceUuid) {
            this.workspaceActiveSessions[session.workspaceUuid] = session.chatSessionUuid
            // Keep track of a session to set as global active (prioritize workspace sessions)
            if (!globalActiveSession) {
              globalActiveSession = session
            }
          }
        }
        
        // Prioritize URL context over backend data
        if (urlWorkspaceUuid && urlSessionUuid) {
          // Use URL workspace and session if available
          this.activeSession = {
            workspaceUuid: urlWorkspaceUuid,
            sessionUuid: urlSessionUuid
          }
          console.log('✅ Used active session from URL over backend:', {
            activeSession: this.activeSession,
            workspaceActiveSessions: this.workspaceActiveSessions
          })
        } else if (urlWorkspaceUuid) {
          // Use URL workspace but try to get session from backend data or workspace mapping
          const sessionFromBackend = this.workspaceActiveSessions[urlWorkspaceUuid]
          this.activeSession = {
            workspaceUuid: urlWorkspaceUuid,
            sessionUuid: sessionFromBackend || null
          }
          console.log('✅ Used workspace from URL with session from backend:', {
            activeSession: this.activeSession,
            workspaceActiveSessions: this.workspaceActiveSessions
          })
        } else if (globalActiveSession) {
          // Fall back to backend active session
          this.activeSession = {
            workspaceUuid: globalActiveSession.workspaceUuid || null,
            sessionUuid: globalActiveSession.chatSessionUuid
          }
          console.log('✅ Used active session from backend:', {
            activeSession: this.activeSession,
            workspaceActiveSessions: this.workspaceActiveSessions
          })
        }
      } catch (error) {
        console.warn('⚠️ Failed to sync workspace active sessions:', error)
        // Continue with empty state - not critical for app functionality
      }
    },

    // Sync sessions for a specific workspace
    async syncWorkspaceSessions(workspaceUuid: string) {
      try {
        console.log('📋 Syncing sessions for workspace:', workspaceUuid)
        const workspaceSessions = await getSessionsByWorkspace(workspaceUuid)
        // Ensure we always have an array, even if API returns null
        const sessions = Array.isArray(workspaceSessions) ? workspaceSessions : []
        this.workspaceHistory[workspaceUuid] = sessions
        console.log('📋 Synced', sessions.length, 'sessions for workspace:', workspaceUuid)
        return sessions
      } catch (error) {
        console.error('❌ Error syncing workspace sessions:', error)
        throw error
      }
    },

    // Switch workspace and update URL accordingly
    async switchToWorkspace(workspaceUuid: string) {
      // Sync sessions for this workspace if not already loaded
      if (!this.workspaceHistory[workspaceUuid]) {
        await this.syncWorkspaceSessions(workspaceUuid)
      }
      
      // Check if this workspace has a previously active session
      const previousActiveSession = this.getActiveSessionForWorkspace(workspaceUuid)
      const workspaceSessions = this.getSessionsByWorkspace(workspaceUuid)

      // If there was a previous active session and it still exists, restore it
      if (previousActiveSession && workspaceSessions.some(s => s.uuid === previousActiveSession)) {
        await this.setActiveSession(workspaceUuid, previousActiveSession)
        return
      }

      if (workspaceSessions.length > 0) {
        // Navigate to the first session in the workspace and save it as active
        const firstSession = workspaceSessions[0]
        await this.setActiveSession(workspaceUuid, firstSession.uuid)
      } else {
        // Create a default session in the empty workspace
        try {
          console.log('🔄 Creating default session for empty workspace:', workspaceUuid)
          
          // Clear any stale session state and set active workspace
          this.activeSession = { sessionUuid: null, workspaceUuid }
          
          const new_chat_text = t('chat.new')
          const sessionData = await this.createSessionInActiveWorkspace(new_chat_text)
          console.log('✅ Created default session in new workspace:', sessionData.uuid)
        } catch (error) {
          console.error('❌ Failed to create default session in workspace:', error)
          console.error('Error details:', error)
          // Fallback to navigation without session
          await router.push({
            name: 'WorkspaceChat',
            params: { workspaceUuid }
          })
          this.activeSession = { sessionUuid: null, workspaceUuid }
        }
      }
    },

    
    // Refresh sessions for current active workspace
    async refreshCurrentWorkspaceSessions() {
      if (this.activeSession.workspaceUuid) {
        await this.syncWorkspaceSessions(this.activeSession.workspaceUuid)
      }
    },
    
    // Get all sessions (lazy loading for compatibility)
    async getAllSessions(): Promise<Chat.Session[]> {
      const allSessions: Chat.Session[] = []
      
      // Load sessions for all workspaces if not already loaded
      for (const workspace of this.workspaces) {
        if (!this.workspaceHistory[workspace.uuid]) {
          await this.syncWorkspaceSessions(workspace.uuid)
        }
        const sessions = this.workspaceHistory[workspace.uuid] || []
        allSessions.push(...sessions)
      }
      
      return allSessions
    },

    async createSessionInActiveWorkspace(topic: string, model?: string) {
      try {
        if (!this.activeSession.workspaceUuid) {
          console.error('❌ No active workspace selected for session creation')
          throw new Error('No active workspace selected')
        }

        // If no model provided, try to get default model from backend
        if (!model) {
          try {
            const defaultModelData = await fetchDefaultChatModel()
            model = defaultModelData.name
            console.log('🔧 Using default model for new session:', model)
          } catch (error) {
            console.warn('⚠️ Failed to fetch default model, backend will handle fallback')
          }
        }

        console.log('🚀 Creating session in workspace:', {
          workspaceUuid: this.activeSession.workspaceUuid,
          topic,
          model
        })

        const sessionData = await createSessionInWorkspace(this.activeSession.workspaceUuid, { topic, model })
        console.log('✅ Session created via API:', sessionData)

        // Refresh workspace sessions to get the latest data from backend
        await this.syncWorkspaceSessions(this.activeSession.workspaceUuid)
        
        this.chat[sessionData.uuid] = []
        console.log('✅ Workspace sessions refreshed')

        // Set as active session for the workspace
        await this.setActiveSession(this.activeSession.workspaceUuid, sessionData.uuid)
        console.log('✅ Session set as active')

        await this.reloadRoute(sessionData.uuid)
        console.log('✅ Route reloaded to new session')
        
        return sessionData
      } catch (error) {
        console.error('❌ Error creating session in workspace:', error)
        throw error
      }
    },

    clearState() {
      this.workspaceHistory = {}
      this.chat = {}
      this.activeSession = { sessionUuid: null, workspaceUuid: null }
      this.workspaceActiveSessions = {}
      this.workspaces = []
    },
  },
})

