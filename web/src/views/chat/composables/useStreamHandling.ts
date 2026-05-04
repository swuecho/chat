import { useAuthStore, useMessageStore } from '@/store'
import { extractStreamingData } from '@/utils/string'
import { extractArtifacts } from '@/utils/artifacts'
import { nowISO } from '@/utils/date'
import { useChat } from '@/views/chat/hooks/useChat'
import { t } from '@/locales'
import { getStreamingUrl } from '@/config/api'

interface ErrorResponse {
  code: number
  message: string
  details?: any
}

interface StreamChunkData {
  choices: Array<{
    delta: {
      content: string
      suggestedQuestions?: string[]
    }
  }>
  id: string
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

  function processStreamChunk(chunk: string, responseIndex: number, sessionUuid: string): void {
    const data = extractStreamingData(chunk)

    if (!data)
      return

    try {
      const parsedData: StreamChunkData = JSON.parse(data)

      const delta = parsedData.choices?.[0]?.delta
      const answerUuid = parsedData.id?.replace('chatcmpl-', '') || parsedData.id

      // Handle both content and suggested questions
      const deltaContent = delta?.content || ''
      const suggestedQuestions = delta?.suggestedQuestions

      // Skip if neither content nor suggested questions are present
      if (!deltaContent && !suggestedQuestions && !parsedData.id) {
        console.warn('Invalid stream chunk structure:', parsedData)
        return
      }

      // Get current message
      const messages = messageStore.getChatSessionDataByUuid(sessionUuid)
      const currentMessage = messages ? (messages[responseIndex] || null) : null

      // Process content if present
      let newText = currentMessage?.text || ''
      let artifacts = currentMessage?.artifacts || []

      if (deltaContent) {
        newText = newText + deltaContent
        artifacts = extractArtifacts(newText)
      }

      // Prepare update object - preserve original timestamp from initial message
      const updateData: any = {
        uuid: answerUuid,
        dateTime: currentMessage?.dateTime || nowISO(),
        text: newText,
        inversion: false,
        error: false,
        loading: false,
        artifacts,
      }

      // Add suggested questions if present
      if (suggestedQuestions && Array.isArray(suggestedQuestions) && suggestedQuestions.length > 0) {
        updateData.suggestedQuestions = suggestedQuestions
        updateData.suggestedQuestionsLoading = false // Clear loading state when questions are received
      }

      updateChat(sessionUuid, responseIndex, updateData)
    }
    catch (error) {
      console.error('Failed to parse stream chunk:', error)
    }
  }

  async function streamChatResponse(
    sessionUuid: string,
    chatUuid: string,
    message: string,
    responseIndex: number,
    onStreamChunk: (chunk: string, responseIndex: number) => void,
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

          for (const line of lines) {
            if (line.trim())
              onStreamChunk(line, responseIndex)
          }
        }

        // Process any remaining data in buffer
        if (buffer.trim())
          onStreamChunk(buffer, responseIndex)
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
    onStreamChunk: (chunk: string, updateIndex: number) => void,
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

          for (const line of lines) {
            if (line.trim())
              onStreamChunk(line, updateIndex)
          }
        }

        // Process any remaining data in buffer
        if (buffer.trim())
          onStreamChunk(buffer, updateIndex)
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
    processStreamChunk,
    streamChatResponse,
    streamRegenerateResponse,
    formatErr,
  }
}
