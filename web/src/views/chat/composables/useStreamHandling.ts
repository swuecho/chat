import { useMessage } from 'naive-ui'
import { fetchChatStream } from '@/api'
import { getDataFromResponseText } from '@/utils/string'
import { extractArtifacts } from '@/utils/artifacts'
import { nowISO } from '@/utils/date'
import { useChatStore } from '@/store'
import { useChat } from '@/views/chat/hooks/useChat'
import renderMessage from '../components/RenderMessage.vue'
import { t } from '@/locales'

interface StreamProgress {
  event: {
    target: {
      responseText: string
      status: number
    }
  }
}

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

  function handleStreamProgress(progress: StreamProgress, responseIndex: number, sessionUuid: string): void {
    const xhr = progress.event.target
    const { responseText, status } = xhr

    if (status >= 400) {
      handleStreamError(responseText, responseIndex, sessionUuid)
      return
    }

    processStreamChunk(responseText, responseIndex, sessionUuid)
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

  function processStreamChunk(responseText: string, responseIndex: number, sessionUuid: string): void {
    const chunk = getDataFromResponseText(responseText)

    if (!chunk) return

    try {
      const data: StreamChunkData = JSON.parse(chunk)
      
      // Validate data structure
      if (!data.choices?.[0]?.delta?.content || !data.id) {
        console.warn('Invalid stream chunk structure:', data)
        return
      }

      const answer = data.choices[0].delta.content
      const answerUuid = data.id.replace('chatcmpl-', '')
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
    onProgress?: (responseText: string, responseIndex: number) => void
  ): Promise<void> {
    return new Promise((resolve, reject) => {
      fetchChatStream(
        sessionUuid,
        chatUuid,
        false,
        message,
        (progress: any) => {
          try {
            if (onProgress) {
              onProgress(progress, responseIndex)
            } else {
              handleStreamProgress(progress, responseIndex, sessionUuid)
            }
          } catch (error) {
            reject(error)
          }
        },
      ).then(resolve).catch(reject)
    })
  }

  async function streamRegenerateResponse(
    sessionUuid: string,
    chatUuid: string,
    updateIndex: number,
    isRegenerate: boolean,
    onProgress?: (progress: any, updateIndex: number) => void
  ): Promise<void> {
    return new Promise((resolve, reject) => {
      fetchChatStream(
        sessionUuid,
        chatUuid,
        isRegenerate,
        "",
        (progress: any) => {
          try {
            if (onProgress) {
              onProgress(progress, updateIndex)
            } else {
              handleStreamProgress(progress, updateIndex, sessionUuid)
            }
          } catch (error) {
            reject(error)
          }
        },
      ).then(resolve).catch(reject)
    })
  }

  function formatErr(error_json: ErrorResponse): string {
    const message = t(`error.${error_json.code}`) || error_json.message
    return `${error_json.code}: ${message}`
  }

  return {
    handleStreamProgress,
    handleStreamError,
    processStreamChunk,
    streamChatResponse,
    streamRegenerateResponse,
    formatErr
  }
}