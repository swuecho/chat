import { ref } from 'vue'
// @ts-ignore
import { v7 as uuidv7 } from 'uuid'
import { nowISO } from '@/utils/date'
import { useChat } from '@/views/chat/hooks/useChat'
import { useScroll } from '@/views/chat/hooks/useScroll'
import { useStreamHandling } from './useStreamHandling'
import { useErrorHandling } from './useErrorHandling'
import { useValidation } from './useValidation'

interface ChatMessage {
  uuid: string
  dateTime: string
  text: string
  inversion: boolean
  error: boolean
  loading?: boolean
  artifacts?: any[]
}

export function useConversationFlow(sessionUuid: string) {
  const loading = ref<boolean>(false)
  const { addChat, updateChat } = useChat()
  const { scrollToBottom, scrollToBottomIfAtBottom } = useScroll()
  const { streamChatResponse, processStreamChunk } = useStreamHandling()
  const { handleApiError, showErrorNotification } = useErrorHandling()
  const { validateChatMessage, validateSessionUuid } = useValidation()

  function validateConversationInput(message: string): boolean {
    if (loading.value) {
      showErrorNotification('Please wait for the current message to complete')
      return false
    }

    // Validate session UUID
    const sessionValidation = validateSessionUuid(sessionUuid)
    if (!sessionValidation.isValid) {
      showErrorNotification('Invalid session')
      return false
    }

    // Validate message content
    const messageValidation = validateChatMessage(message)
    if (!messageValidation.isValid) {
      showErrorNotification(messageValidation.errors[0])
      return false
    }

    return true
  }

  function addUserMessage(chatUuid: string, message: string): void {
    const chatMessage: ChatMessage = {
      uuid: chatUuid,
      dateTime: nowISO(),
      text: message,
      inversion: true,
      error: false,
    }
    
    addChat(sessionUuid, chatMessage)
    scrollToBottom()
    loading.value = true
  }

  function initializeChatResponse(dataSources: any[]): number {
    addChat(sessionUuid, {
      uuid: '',
      dateTime: nowISO(),
      text: '',
      loading: true,
      inversion: false,
      error: false,
    })
    scrollToBottomIfAtBottom()
    return dataSources.length - 1
  }

  function handleStreamingError(error: any, responseIndex: number, dataSources: any[]): void {
    handleApiError(error, 'conversation-stream')

    const lastMessage = dataSources[responseIndex]
    if (lastMessage) {
      const errorMessage: ChatMessage = {
        uuid: lastMessage.uuid || uuidv7(),
        dateTime: nowISO(),
        text: 'Failed to get response. Please try again.',
        inversion: false,
        error: true,
        loading: false,
      }
      
      updateChat(sessionUuid, responseIndex, errorMessage)
    }
  }

  async function onConversationStream(
    message: string,
    dataSources: any[]
  ): Promise<void> {
    if (!validateConversationInput(message)) return

    const chatUuid = uuidv7()
    addUserMessage(chatUuid, message)
    const responseIndex = initializeChatResponse(dataSources)

    try {
      await streamChatResponse(
        sessionUuid,
        chatUuid,
        message,
        responseIndex,
        (progress: any, index: number) => {
          const xhr = progress.event.target
          const { responseText, status } = xhr

          if (status >= 400) {
            // Handle error
            try {
              const errorJson = JSON.parse(responseText)
              console.error('Stream error:', responseText)
              showErrorNotification(`${errorJson.code} : ${errorJson.message}`)
            } catch (parseError) {
              showErrorNotification('An unexpected error occurred')
            }
            return
          }

          processStreamChunk(responseText, index, sessionUuid)
          scrollToBottomIfAtBottom()
        }
      )
    } catch (error) {
      handleStreamingError(error, responseIndex, dataSources)
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    validateConversationInput,
    addUserMessage,
    initializeChatResponse,
    handleStreamingError,
    onConversationStream
  }
}