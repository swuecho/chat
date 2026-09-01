import { useAuthStore, useMessageStore } from '@/store'
import { extractArtifacts } from '@/utils/artifacts'
import { nowISO } from '@/utils/date'
import { type AnswerStreamEvent, consumeAnswerEvents } from '@/utils/sse'
import { useChat } from '@/views/chat/hooks/useChat'
import { t } from '@/locales'
import { openChatStream } from '@/api/chat_stream'

interface ErrorResponse {
  code: number
  message: string
  details?: any
}

export function useStreamHandling() {
  const messageStore = useMessageStore()
  const { updateChat } = useChat()

  function handleStreamError(responseText: string): string {
    try {
      const errorJson: ErrorResponse = JSON.parse(responseText)
      console.error('Stream error:', errorJson)
      return formatErr(errorJson)
    }
    catch (parseError) {
      console.error('Failed to parse error response:', parseError)
      const trimmedText = responseText.trim()
      return trimmedText || 'An unexpected error occurred'
    }
  }

  function processAnswerEvent(event: AnswerStreamEvent, responseIndex: number, sessionUuid: string): void {
    const messages = messageStore.getChatSessionDataByUuid(sessionUuid)
    const currentMessage = messages ? (messages[responseIndex] || null) : null
    let newText = currentMessage?.text || ''
    let artifacts = currentMessage?.artifacts || []

    if ((event.type === 'delta' || event.type === 'reasoning_delta') && event.delta) {
      newText += event.delta
      artifacts = extractArtifacts(newText)
    }

    const updateData: any = {
      uuid: event.answerId || currentMessage?.uuid || '',
      dateTime: currentMessage?.dateTime || nowISO(),
      text: newText,
      inversion: false,
      error: false,
      loading: false,
      artifacts,
    }
    if (event.type === 'suggested_questions' && event.suggestedQuestions?.length) {
      updateData.suggestedQuestions = event.suggestedQuestions
      updateData.suggestedQuestionsLoading = false
    }
    updateChat(sessionUuid, responseIndex, updateData)
  }

  async function streamChatResponse(
    sessionUuid: string,
    chatUuid: string,
    message: string,
    responseIndex: number,
    onAnswerEvent: (event: AnswerStreamEvent, responseIndex: number) => void,
    abortSignal?: AbortSignal,
  ): Promise<void> {
    const authStore = useAuthStore()
    await authStore.initializeAuth()
    if (!authStore.isValid || authStore.needsRefresh) {
      try {
        await authStore.refreshToken()
      }
      catch (error) {
        authStore.removeToken()
        authStore.removeExpiresIn()
        throw new Error(t('error.NotAuthorized') || 'Please log in first')
      }
    }
    try {
      const stream = await openChatStream({ regenerate: false, prompt: message, sessionUuid, chatUuid, stream: true }, abortSignal)
      await consumeAnswerEvents(stream, event => onAnswerEvent(event, responseIndex))
    }
    catch (error) {
      if (error instanceof Error && error.name === 'AbortError')
        return
      throw error
    }
  }

  async function streamRegenerateResponse(
    sessionUuid: string,
    chatUuid: string,
    updateIndex: number,
    isRegenerate: boolean,
    onAnswerEvent: (event: AnswerStreamEvent, updateIndex: number) => void,
    abortSignal?: AbortSignal,
  ): Promise<void> {
    const authStore = useAuthStore()
    await authStore.initializeAuth()
    if (!authStore.isValid || authStore.needsRefresh) {
      try {
        await authStore.refreshToken()
      }
      catch (error) {
        authStore.removeToken()
        authStore.removeExpiresIn()
        throw new Error(t('error.NotAuthorized') || 'Please log in first')
      }
    }
    try {
      const stream = await openChatStream({ regenerate: isRegenerate, prompt: '', sessionUuid, chatUuid, stream: true }, abortSignal)
      await consumeAnswerEvents(stream, event => onAnswerEvent(event, updateIndex))
    }
    catch (error) {
      if (error instanceof Error && error.name === 'AbortError')
        return
      throw error
    }
  }

  function formatErr(error_json: ErrorResponse): string {
    const message = t(`error.${error_json.code}`) || error_json.message
    return `${error_json.code}: ${message}`
  }

  return {
    handleStreamError,
    processAnswerEvent,
    streamChatResponse,
    streamRegenerateResponse,
    formatErr,
  }
}
