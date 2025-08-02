import { ref } from 'vue'
import { v7 as uuidv7 } from 'uuid'
import { nowISO } from '@/utils/date'
import { useChat } from '@/views/chat/hooks/useChat'
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

export function useConversationFlow(
  sessionUuid: string,
  scrollToBottom: () => Promise<void>,
  smoothScrollToBottomIfAtBottom: () => Promise<void>
) {
  const loading = ref<boolean>(false)
  const { addChat, updateChat } = useChat()
  const { streamChatResponse, processStreamChunk } = useStreamHandling()
  const { handleApiError, showErrorNotification } = useErrorHandling()
  const { validateChatMessage } = useValidation()

  function validateConversationInput(message: string): boolean {
    if (loading.value) {
      showErrorNotification('Please wait for the current message to complete')
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

  async function addUserMessage(chatUuid: string, message: string): Promise<void> {
    const chatMessage: ChatMessage = {
      uuid: chatUuid,
      dateTime: nowISO(),
      text: message,
      inversion: true,
      error: false,
    }
    
    addChat(sessionUuid, chatMessage)
    await scrollToBottom()
  }

  async function initializeChatResponse(dataSources: any[]): Promise<number> {
    addChat(sessionUuid, {
      uuid: '',
      dateTime: nowISO(),
      text: '',
      loading: true,
      inversion: false,
      error: false,
    })
    await smoothScrollToBottomIfAtBottom()
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

  async function startStream(
    message: string,
    dataSources: any[],
    chatUuid: string
  ): Promise<void> {
    if (!validateConversationInput(message)) return

    loading.value = true
    const responseIndex = await initializeChatResponse(dataSources)

    try {
      await streamChatResponse(
        sessionUuid,
        chatUuid,
        message,
        responseIndex,
        async (chunk: string, index: number) => {
          processStreamChunk(chunk, index, sessionUuid)
          await smoothScrollToBottomIfAtBottom()
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
    startStream
  }
}