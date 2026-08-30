import { useAuthStore, useMessageStore } from '@/store'
import { extractArtifacts } from '@/utils/artifacts'
import { nowISO } from '@/utils/date'
import { type AnswerStreamEvent, readAnswerStreamEvent } from '@/utils/sse'
import { useChat } from '@/views/chat/hooks/useChat'
import { t } from '@/locales'
import { getStreamingUrl } from '@/config/api'

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

  function processAnswerFrame(frame: string, responseIndex: number, onAnswerEvent: (event: AnswerStreamEvent, responseIndex: number) => void): boolean {
    const event = readAnswerStreamEvent(frame)
    if (!event)
      throw new Error('Received an untyped answer stream frame')
    if (event.type === 'failed' || event.type === 'canceled')
      throw new Error(event.message || event.code || `Stream ${event.type}`)
    if (event.type === 'completed') {
      if (!event.persisted)
        throw new Error('The response was not saved')
      onAnswerEvent(event, responseIndex)
      return true
    }
    if (event.type === 'started' || event.type === 'delta' || event.type === 'reasoning_delta' || event.type === 'suggested_questions')
      onAnswerEvent(event, responseIndex)
    return false
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
    const token = authStore.getToken

    try {
      const response = await fetch(getStreamingUrl('/chat_stream'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 'no-cache',
          'Connection': 'keep-alive',
          ...(token && { Authorization: `Bearer ${token}` }),
        },
        body: JSON.stringify({
          regenerate: false,
          prompt: message,
          sessionUuid,
          chatUuid,
          stream: true,
        }),
        signal: abortSignal,
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(handleStreamError(errorText))
      }

      if (!response.body)
        throw new Error('Response body is null')

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let completed = false

      try {
        while (true) {
          const { done, value } = await reader.read()

          if (done)
            break

          const chunk = decoder.decode(value, { stream: true })
          buffer += chunk

          // Process complete SSE messages (handle both \n\n and \r\n\r\n)
          const normalizedBuffer = buffer.replace(/\r\n/g, '\n')
          const lines = normalizedBuffer.split('\n\n')
          // Keep the last potentially incomplete message in buffer
          buffer = lines.pop() || ''

          for (const line of lines)
            completed = processAnswerFrame(line, responseIndex, onAnswerEvent) || completed
        }

        // Process any remaining data in buffer
        if (buffer.trim())
          completed = processAnswerFrame(buffer, responseIndex, onAnswerEvent) || completed

        if (!completed)
          throw new Error('The response stream ended before it was saved')
      }
      finally {
        reader.releaseLock()
      }
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
    const token = authStore.getToken

    try {
      const response = await fetch(getStreamingUrl('/chat_stream'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 'no-cache',
          'Connection': 'keep-alive',
          ...(token && { Authorization: `Bearer ${token}` }),
        },
        body: JSON.stringify({
          regenerate: isRegenerate,
          prompt: '',
          sessionUuid,
          chatUuid,
          stream: true,
        }),
        signal: abortSignal,
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(handleStreamError(errorText))
      }

      if (!response.body)
        throw new Error('Response body is null')

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let completed = false

      try {
        while (true) {
          const { done, value } = await reader.read()

          if (done)
            break

          const chunk = decoder.decode(value, { stream: true })
          buffer += chunk

          // Process complete SSE messages (handle both \n\n and \r\n\r\n)
          const normalizedBuffer = buffer.replace(/\r\n/g, '\n')
          const lines = normalizedBuffer.split('\n\n')
          // Keep the last potentially incomplete message in buffer
          buffer = lines.pop() || ''

          for (const line of lines)
            completed = processAnswerFrame(line, updateIndex, onAnswerEvent) || completed
        }

        // Process any remaining data in buffer
        if (buffer.trim())
          completed = processAnswerFrame(buffer, updateIndex, onAnswerEvent) || completed

        if (!completed)
          throw new Error('The regenerated response stream ended before it was saved')
      }
      finally {
        reader.releaseLock()
      }
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
