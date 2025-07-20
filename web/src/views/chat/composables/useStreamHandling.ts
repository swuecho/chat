import { useMessage } from 'naive-ui'
import { useAuthStore } from '@/store'
import { getDataFromResponseText } from '@/utils/string'
import { extractArtifacts } from '@/utils/artifacts'
import { nowISO } from '@/utils/date'
import { useChatStore } from '@/store'
import { useChat } from '@/views/chat/hooks/useChat'
import renderMessage from '../components/RenderMessage.vue'
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
    }
  }>
  id: string
}

export function useStreamHandling() {
  const nui_msg = useMessage()
  const chatStore = useChatStore()
  const { updateChat } = useChat()
  

  function handleStreamChunk(chunk: string, responseIndex: number, sessionUuid: string): void {
    processStreamChunk(chunk, responseIndex, sessionUuid)
  }

  function handleStreamError(responseText: string, responseIndex: number, sessionUuid: string): void {
    try {
      const errorJson: ErrorResponse = JSON.parse(responseText)
      console.error('Stream error:', errorJson)

      const errorMessage = formatErr(errorJson)
      nui_msg.error(errorMessage, {
        duration: 5000,
        closable: true,
        render: renderMessage
      })

      chatStore.deleteChatByUuid(sessionUuid, responseIndex)
    } catch (parseError) {
      console.error('Failed to parse error response:', parseError)
      nui_msg.error('An unexpected error occurred')
    }
  }

  function processStreamChunk(chunk: string, responseIndex: number, sessionUuid: string): void {
    const data = getDataFromResponseText(chunk)

    if (!data) return

    try {
      const parsedData: StreamChunkData = JSON.parse(data)
      
      // Validate data structure
      if (!parsedData.choices?.[0]?.delta?.content || !parsedData.id) {
        console.warn('Invalid stream chunk structure:', parsedData)
        return
      }

      const answer = parsedData.choices[0].delta.content
      const answerUuid = parsedData.id.replace('chatcmpl-', '')
      const artifacts = extractArtifacts(answer)

      updateChat(sessionUuid, responseIndex, {
        uuid: answerUuid,
        dateTime: nowISO(),
        text: answer,
        inversion: false,
        error: false,
        loading: false,
        artifacts: artifacts,
      })
    } catch (error) {
      console.error('Failed to parse stream chunk:', error)
    }
  }

  async function streamChatResponse(
    sessionUuid: string,
    chatUuid: string,
    message: string,
    responseIndex: number,
    onProgress?: (chunk: string, responseIndex: number) => void
  ): Promise<void> {
    const authStore = useAuthStore()
    const token = authStore.getToken()
    
    try {
      const response = await fetch(getStreamingUrl('/chat_stream'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 'no-cache',
          'Connection': 'keep-alive',
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
        body: JSON.stringify({
          regenerate: false,
          prompt: message,
          sessionUuid,
          chatUuid,
          stream: true,
        }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      if (!response.body) {
        throw new Error('Response body is null')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let accumulated = ''

      try {
        while (true) {
          const { done, value } = await reader.read()
          
          if (done) {
            break
          }

          const chunk = decoder.decode(value, { stream: true })
          accumulated += chunk
          
          if (onProgress) {
            onProgress(accumulated, responseIndex)
          } else {
            handleStreamChunk(accumulated, responseIndex, sessionUuid)
          }
        }
      } finally {
        reader.releaseLock()
      }
    } catch (error) {
      console.error('Stream error:', error)
      handleStreamError(error instanceof Error ? error.message : 'Unknown error', responseIndex, sessionUuid)
      throw error
    }
  }

  async function streamRegenerateResponse(
    sessionUuid: string,
    chatUuid: string,
    updateIndex: number,
    isRegenerate: boolean,
    onProgress?: (chunk: string, updateIndex: number) => void
  ): Promise<void> {
    const authStore = useAuthStore()
    const token = authStore.getToken()
    
    try {
      const response = await fetch(getStreamingUrl('/chat_stream'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 'no-cache',
          'Connection': 'keep-alive',
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
        body: JSON.stringify({
          regenerate: isRegenerate,
          prompt: "",
          sessionUuid,
          chatUuid,
          stream: true,
        }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      if (!response.body) {
        throw new Error('Response body is null')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let accumulated = ''

      try {
        while (true) {
          const { done, value } = await reader.read()
          
          if (done) {
            break
          }

          const chunk = decoder.decode(value, { stream: true })
          accumulated += chunk
          
          if (onProgress) {
            onProgress(accumulated, updateIndex)
          } else {
            handleStreamChunk(accumulated, updateIndex, sessionUuid)
          }
        }
      } finally {
        reader.releaseLock()
      }
    } catch (error) {
      console.error('Stream error:', error)
      handleStreamError(error instanceof Error ? error.message : 'Unknown error', updateIndex, sessionUuid)
      throw error
    }
  }

  function formatErr(error_json: ErrorResponse): string {
    const message = t(`error.${error_json.code}`) || error_json.message
    return `${error_json.code}: ${message}`
  }

  return {
    handleStreamChunk,
    handleStreamError,
    processStreamChunk,
    streamChatResponse,
    streamRegenerateResponse,
    formatErr
  }
}