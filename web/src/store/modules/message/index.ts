import { defineStore } from 'pinia'
import {
  getChatMessagesBySessionUUID,
  clearSessionChatMessages,
  generateMoreSuggestions,
} from '@/api'
import { deleteChatData } from '@/api'
import { createChatPrompt } from '@/api/chat_prompt'
import { DEFAULT_SYSTEM_PROMPT } from '@/constants/chat'
import { nowISO } from '@/utils/date'
import { v7 as uuidv7 } from 'uuid'
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
        
        // Initialize batching structure for messages with suggested questions
        const processedMessageData = messageData.map((message: Chat.Message) => {
          if (message.suggestedQuestions && message.suggestedQuestions.length > 0) {
            // If batches don't exist, create the first batch from existing questions
            if (!message.suggestedQuestionsBatches || message.suggestedQuestionsBatches.length === 0) {
              // Split suggestions into batches of 3 (assuming original suggestions come in groups of 3)
              const batches: string[][] = []
              for (let i = 0; i < message.suggestedQuestions.length; i += 3) {
                batches.push(message.suggestedQuestions.slice(i, i + 3))
              }
              
              return {
                ...message,
                suggestedQuestionsBatches: batches,
                currentSuggestedQuestionsBatch: batches.length - 1, // Show the last batch (most recent)
                suggestedQuestions: batches[batches.length - 1] || message.suggestedQuestions, // Show last batch
              }
            }
          }
          return message
        })
        
        if (processedMessageData.length === 0) {
          try {
            const prompt = await createChatPrompt({
              uuid: uuidv7(),
              chatSessionUuid: sessionUuid,
              role: 'system',
              content: DEFAULT_SYSTEM_PROMPT,
              tokenCount: 0,
              userId: 0,
              createdBy: 0,
              updatedBy: 0,
            })

            processedMessageData.unshift({
              uuid: prompt.uuid,
              dateTime: prompt.updatedAt || nowISO(),
              text: prompt.content || DEFAULT_SYSTEM_PROMPT,
              inversion: true,
              error: false,
              loading: false,
              isPrompt: true,
            })
          } catch (error) {
            console.error(`Failed to create default system prompt for session ${sessionUuid}:`, error)
          }
        }

        this.chat[sessionUuid] = processedMessageData

        // Update active session if needed
        const sessionStore = useSessionStore()
        if (sessionStore.activeSessionUuid !== sessionUuid) {
          const session = sessionStore.getChatSessionByUuid(sessionUuid)
          if (session?.workspaceUuid) {
            await sessionStore.setActiveSession(session.workspaceUuid, sessionUuid)
          }
        }

        return processedMessageData
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

    async removeMessage(sessionUuid: string, messageUuid: string) {
      try {
        const message = this.chat[sessionUuid]?.find(msg => msg.uuid === messageUuid)
        if (!message) {
          return
        }
        // Call the API to delete the message from the server
        await deleteChatData(message)
        // Remove the message from local state after successful API call
        if (this.chat[sessionUuid]) {
          this.chat[sessionUuid] = this.chat[sessionUuid].filter(
            msg => msg.uuid !== messageUuid
          )
        }
      } catch (error) {
        console.error(`Failed to delete message ${messageUuid}:`, error)
        throw error
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

    // Generate more suggested questions for a message
    async generateMoreSuggestedQuestions(sessionUuid: string, messageUuid: string) {
      try {
        // Set generating state for the message
        this.updateMessage(sessionUuid, messageUuid, { suggestedQuestionsGenerating: true })

        const response = await generateMoreSuggestions(messageUuid)
        const { newSuggestions, allSuggestions } = response

        // Get existing message
        const messages = this.chat[sessionUuid] || []
        const messageIndex = messages.findIndex(msg => msg.uuid === messageUuid)
        
        if (messageIndex !== -1) {
          const message = messages[messageIndex]
          
          // Initialize batches if they don't exist
          let suggestedQuestionsBatches = message.suggestedQuestionsBatches || []
          
          // If this is the first time, create the first batch from existing questions
          if (suggestedQuestionsBatches.length === 0 && message.suggestedQuestions) {
            suggestedQuestionsBatches.push(message.suggestedQuestions)
          }
          
          // Add the new suggestions as a new batch
          suggestedQuestionsBatches.push(newSuggestions)
          
          // Update the message with new data - show the new batch, not all suggestions
          this.updateMessage(sessionUuid, messageUuid, {
            suggestedQuestions: newSuggestions, // Show only the new batch
            suggestedQuestionsBatches,
            currentSuggestedQuestionsBatch: suggestedQuestionsBatches.length - 1, // Set to the new batch
            suggestedQuestionsGenerating: false,
          })
        }

        return response
      } catch (error) {
        // Clear generating state on error
        this.updateMessage(sessionUuid, messageUuid, { suggestedQuestionsGenerating: false })
        console.error('Failed to generate more suggestions:', error)
        throw error
      }
    },

    // Navigate to previous suggestions batch
    previousSuggestedQuestionsBatch(sessionUuid: string, messageUuid: string) {
      const messages = this.chat[sessionUuid] || []
      const messageIndex = messages.findIndex(msg => msg.uuid === messageUuid)
      
      if (messageIndex !== -1) {
        const message = messages[messageIndex]
        const batches = message.suggestedQuestionsBatches || []
        const currentBatch = message.currentSuggestedQuestionsBatch || 0
        
        if (currentBatch > 0 && batches.length > 0) {
          const newBatchIndex = currentBatch - 1
          this.updateMessage(sessionUuid, messageUuid, {
            suggestedQuestions: batches[newBatchIndex],
            currentSuggestedQuestionsBatch: newBatchIndex,
          })
        }
      }
    },

    // Navigate to next suggestions batch
    nextSuggestedQuestionsBatch(sessionUuid: string, messageUuid: string) {
      const messages = this.chat[sessionUuid] || []
      const messageIndex = messages.findIndex(msg => msg.uuid === messageUuid)
      
      if (messageIndex !== -1) {
        const message = messages[messageIndex]
        const batches = message.suggestedQuestionsBatches || []
        const currentBatch = message.currentSuggestedQuestionsBatch || 0
        
        if (currentBatch < batches.length - 1) {
          const newBatchIndex = currentBatch + 1
          this.updateMessage(sessionUuid, messageUuid, {
            suggestedQuestions: batches[newBatchIndex],
            currentSuggestedQuestionsBatch: newBatchIndex,
          })
        }
      }
    },
  },
})
