import { defineStore } from 'pinia'
import {
  getChatMessagesBySessionUUID,
  clearSessionChatMessages,
} from '@/api'
import { useSessionStore } from '../session'

export interface MessageState {
  chat: Record<string, Chat.Message[]> // sessionUuid -> messages
  isLoading: Record<string, boolean> // sessionUuid -> isLoading
}

export const useMessageStore = defineStore('message-store', {
  state: (): MessageState => ({
    chat: {},
    isLoading: {},
  }),

  getters: {
    getChatSessionDataByUuid(state) {
      return (uuid?: string) => {
        if (uuid) {
          return state.chat[uuid] || []
        }
        return []
      }
    },

    getIsLoadingBySession(state) {
      return (sessionUuid: string) => {
        return state.isLoading[sessionUuid] || false
      }
    },

    // Get last message for a session
    getLastMessageForSession(state) {
      return (sessionUuid: string) => {
        const messages = state.chat[sessionUuid] || []
        return messages[messages.length - 1] || null
      }
    },

    // Get all messages for active session
    activeSessionMessages(state) {
      const sessionStore = useSessionStore()
      if (sessionStore.activeSessionUuid) {
        return state.chat[sessionStore.activeSessionUuid] || []
      }
      return []
    },
  },

  actions: {
    async syncChatMessages(sessionUuid: string) {
      if (!sessionUuid) {
        return
      }

      this.isLoading[sessionUuid] = true

      try {
        const messageData = await getChatMessagesBySessionUUID(sessionUuid)
        this.chat[sessionUuid] = messageData

        // Update active session if needed
        const sessionStore = useSessionStore()
        if (sessionStore.activeSessionUuid !== sessionUuid) {
          const session = sessionStore.getChatSessionByUuid(sessionUuid)
          if (session?.workspaceUuid) {
            await sessionStore.setActiveSession(session.workspaceUuid, sessionUuid)
          }
        }

        return messageData
      } catch (error) {
        console.error(`Failed to sync messages for session ${sessionUuid}:`, error)
        throw error
      } finally {
        this.isLoading[sessionUuid] = false
      }
    },

    addMessage(sessionUuid: string, message: Chat.Message) {
      if (!this.chat[sessionUuid]) {
        this.chat[sessionUuid] = []
      }
      this.chat[sessionUuid].push(message)
    },

    addMessages(sessionUuid: string, messages: Chat.Message[]) {
      if (!this.chat[sessionUuid]) {
        this.chat[sessionUuid] = []
      }
      this.chat[sessionUuid].push(...messages)
    },

    updateMessage(sessionUuid: string, messageUuid: string, updates: Partial<Chat.Message>) {
      const messages = this.chat[sessionUuid]
      if (messages) {
        const index = messages.findIndex(msg => msg.uuid === messageUuid)
        if (index !== -1) {
          messages[index] = { ...messages[index], ...updates }
        }
      }
    },

    removeMessage(sessionUuid: string, messageUuid: string) {
      if (this.chat[sessionUuid]) {
        this.chat[sessionUuid] = this.chat[sessionUuid].filter(
          msg => msg.uuid !== messageUuid
        )
      }
    },

    clearSessionMessages(sessionUuid: string) {
      try {
        clearSessionChatMessages(sessionUuid)
        // Keep the first message (system prompt) and clear the rest
        const messages = this.chat[sessionUuid] || []
        if (messages.length > 0) {
          this.chat[sessionUuid] = [messages[0]] // Keep only the first message
        } else {
          this.chat[sessionUuid] = []
        }
      } catch (error) {
        console.error(`Failed to clear messages for session ${sessionUuid}:`, error)
        throw error
      }
    },

    updateLastMessage(sessionUuid: string, updates: Partial<Chat.Message>) {
      const messages = this.chat[sessionUuid]
      if (messages && messages.length > 0) {
        const lastIndex = messages.length - 1
        messages[lastIndex] = { ...messages[lastIndex], ...updates }
      }
    },

    // Helper method to set loading state
    setLoading(sessionUuid: string, isLoading: boolean) {
      this.isLoading[sessionUuid] = isLoading
    },

    // Helper method to get message count for a session
    getMessageCount(sessionUuid: string) {
      return this.chat[sessionUuid]?.length || 0
    },

    // Helper method to clear all messages
    clearAllMessages() {
      this.chat = {}
      this.isLoading = {}
    },

    // Helper method to remove session data
    removeSessionData(sessionUuid: string) {
      delete this.chat[sessionUuid]
      delete this.isLoading[sessionUuid]
    },

    // Helper method to check if session has messages
    hasMessages(sessionUuid: string) {
      return this.chat[sessionUuid]?.length > 0
    },

    // Helper method to get messages by type
    getMessagesByType(sessionUuid: string, type: 'user' | 'assistant') {
      const messages = this.chat[sessionUuid] || []
      return messages.filter(msg => {
        if (type === 'user') {
          return msg.inversion
        }
        return !msg.inversion
      })
    },

    // Helper method to get pinned messages
    getPinnedMessages(sessionUuid: string) {
      const messages = this.chat[sessionUuid] || []
      return messages.filter(msg => msg.isPin)
    },

    // Helper method to get messages with artifacts
    getMessagesWithArtifacts(sessionUuid: string) {
      const messages = this.chat[sessionUuid] || []
      return messages.filter(msg => msg.artifacts && msg.artifacts.length > 0)
    },

    // Helper method to get messages by date range
    getMessagesByDateRange(sessionUuid: string, startDate: string, endDate: string) {
      const messages = this.chat[sessionUuid] || []
      return messages.filter(msg => {
        const messageDate = new Date(msg.dateTime)
        return messageDate >= new Date(startDate) && messageDate <= new Date(endDate)
      })
    },

    // Helper method to search messages
    searchMessages(sessionUuid: string, query: string) {
      const messages = this.chat[sessionUuid] || []
      const lowercaseQuery = query.toLowerCase()
      return messages.filter(msg =>
        msg.text.toLowerCase().includes(lowercaseQuery)
      )
    },

    // Helper method to get loading messages for a session
    getLoadingMessages(sessionUuid: string) {
      const messages = this.chat[sessionUuid] || []
      return messages.filter(msg => msg.loading)
    },

    // Helper method to get error messages for a session
    getErrorMessages(sessionUuid: string) {
      const messages = this.chat[sessionUuid] || []
      return messages.filter(msg => msg.error)
    },

    // Helper method to get prompt messages for a session
    getPromptMessages(sessionUuid: string) {
      const messages = this.chat[sessionUuid] || []
      return messages.filter(msg => msg.isPrompt)
    },
  },
})