import { useMessage } from 'naive-ui'
import { useAuthStore } from '@/store'
import { extractStreamingData } from '@/utils/string'
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
    const data = extractStreamingData(chunk)

    if (!data) return

    try {
      const parsedData: StreamChunkData = JSON.parse(data)
      
      // Validate data structure
      if (!parsedData.choices?.[0]?.delta?.content || !parsedData.id) {
        console.warn('Invalid stream chunk structure:', parsedData)
        return
      }

      const deltaContent = parsedData.choices[0].delta.content
      const answerUuid = parsedData.id.replace('chatcmpl-', '')
      
      // Get current message to append delta content
      const currentMessage = chatStore.getChatByUuidAndIndex(sessionUuid, responseIndex)
      const currentText = currentMessage?.text || ''
      const newText = currentText + deltaContent
      
      const artifacts = extractArtifacts(newText)

      updateChat(sessionUuid, responseIndex, {
        uuid: answerUuid,
        dateTime: nowISO(),
        text: newText,
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
    onStreamChunk: (chunk: string, responseIndex: number) => void
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
        const errorText = await response.text()
        handleStreamError(errorText, responseIndex, sessionUuid)
        return
      }

      if (!response.body) {
        throw new Error('Response body is null')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      try {
        while (true) {
          const { done, value } = await reader.read()
          
          if (done) {
            break
          }

          const chunk = decoder.decode(value, { stream: true })
          console.log('chunk', chunk)
          buffer += chunk
          
          // Process complete SSE messages
          const lines = buffer.split('\n\n')
          // Keep the last potentially incomplete message in buffer
          buffer = lines.pop() || ''
          
          for (const line of lines) {
            if (line.trim()) {
              onStreamChunk(line, responseIndex)
            }
          }

        }
        
        // Process any remaining data in buffer
        if (buffer.trim()) {
          onStreamChunk(buffer, responseIndex)
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
    onStreamChunk: (chunk: string, updateIndex: number) => void
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
        const errorText = await response.text()
        handleStreamError(errorText, updateIndex, sessionUuid)
        return
      }

      if (!response.body) {
        throw new Error('Response body is null')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      try {
        while (true) {
          const { done, value } = await reader.read()
          
          if (done) {
            break
          }

          const chunk = decoder.decode(value, { stream: true })
          buffer += chunk
          
          // Process complete SSE messages
          const lines = buffer.split('\n\n')
          // Keep the last potentially incomplete message in buffer
          buffer = lines.pop() || ''
          
          for (const line of lines) {
            if (line.trim()) {
              onStreamChunk(line, updateIndex)
            }
          }
         
        }
        
        // Process any remaining data in buffer
        if (buffer.trim()) {
          onStreamChunk(buffer, updateIndex)
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
    handleStreamError,
    processStreamChunk,
    streamChatResponse,
    streamRegenerateResponse,
    formatErr
  }
}