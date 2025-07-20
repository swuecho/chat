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
        (chunk: string, index: number) => {
          processStreamChunk(chunk, index, sessionUuid)
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