import { inject, ref, type Ref } from 'vue'
import { v7 as uuidv7 } from 'uuid'
import { nowISO } from '@/utils/date'
import { useChat } from '@/views/chat/hooks/useChat'
import { useStreamHandling } from './useStreamHandling'
import { useErrorHandling } from './useErrorHandling'
import { useValidation } from './useValidation'
import { useMessageStore, useSessionStore } from '@/store'
import { extractArtifacts } from '@/utils/artifacts'
import { extractToolCalls } from '@/utils/tooling'
import { getCodeRunner } from '@/services/codeRunner'

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
  sessionUuidRef: Ref<string>,
  scrollToBottom: () => Promise<void>,
  smoothScrollToBottomIfAtBottom: () => Promise<void>,
  vfsContext?: {
    vfsInstance: Ref<any>
    isVFSReady: Ref<boolean>
  }
) {
  const loading = ref<boolean>(false)
  const abortController = ref<AbortController | null>(null)
  const { addChat, updateChat, updateChatPartial } = useChat()
  const { streamChatResponse, processStreamChunk } = useStreamHandling()
  const { handleApiError, showErrorNotification } = useErrorHandling()
  const { validateChatMessage } = useValidation()
  const sessionStore = useSessionStore()
  const messageStore = useMessageStore()
  const vfsInstance = vfsContext?.vfsInstance ?? inject('vfsInstance', ref(null))
  const isVFSReady = vfsContext?.isVFSReady ?? inject('isVFSReady', ref(false))

  const maxToolTurns = 10 
  const toolRunning = ref<boolean>(false)

  const toBase64 = (bytes: Uint8Array) => {
    const chunkSize = 0x8000
    let binary = ''
    for (let i = 0; i < bytes.length; i += chunkSize) {
      const chunk = bytes.subarray(i, i + chunkSize)
      binary += String.fromCharCode(...chunk)
    }
    return btoa(binary)
  }

  function validateConversationInput(message: string): boolean {
    if (loading.value) {
      showErrorNotification('Please wait for the current message to complete')
      return false
    }

    if (!sessionUuidRef.value) {
      showErrorNotification('No active session selected')
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
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid) {
      return
    }

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
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid) {
      return dataSources.length - 1
    }

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
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid) {
      return
    }

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

  function stopStream(): void {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
      loading.value = false
    }
  }

  async function startStream(
    message: string,
    dataSources: any[],
    chatUuid: string
  ): Promise<void> {
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid) {
      loading.value = false
      abortController.value = null
      return
    }

    loading.value = true
    abortController.value = new AbortController()
    let responseIndex = await initializeChatResponse(dataSources)

    try {
      let currentPrompt = message
      let currentChatUuid = chatUuid
      let toolTurns = 0

      while (true) {
        await streamChatResponse(
          sessionUuid,
          currentChatUuid,
          currentPrompt,
          responseIndex,
          async (chunk: string, index: number) => {
            processStreamChunk(chunk, index, sessionUuid)
            await smoothScrollToBottomIfAtBottom()
          },
          abortController.value.signal
        )

        if (toolTurns >= maxToolTurns) {
          break
        }

        const toolPrompt = await handleToolCalls(sessionUuid, responseIndex)
        if (!toolPrompt) {
          break
        }

        toolTurns += 1
        currentPrompt = toolPrompt
        currentChatUuid = uuidv7()
        responseIndex = await initializeChatResponse(dataSources)
      }
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        // Stream was cancelled, no need to show error
        return
      }
      handleStreamingError(error, responseIndex, dataSources)
    } finally {
      loading.value = false
      abortController.value = null
      
      // For sessions in exploreMode, set suggested questions loading state
      const session = sessionStore.getChatSessionByUuid(sessionUuid)
      if (session?.exploreMode && dataSources[responseIndex] && !dataSources[responseIndex].inversion) {
        updateChatPartial(sessionUuid, responseIndex, {
          suggestedQuestionsLoading: true
        })
      }
    }
  }

  const handleToolCalls = async (sessionUuid: string, responseIndex: number) => {
    toolRunning.value = false
    const session = sessionStore.getChatSessionByUuid(sessionUuid)
    if (!session?.codeRunnerEnabled) return null

    const messages = messageStore.getChatSessionDataByUuid(sessionUuid)
    const currentMessage = messages?.[responseIndex]
    if (!currentMessage || currentMessage.inversion) return null

    const { calls, cleanedText } = extractToolCalls(currentMessage.text || '')
    if (!calls.length) return null

    toolRunning.value = true
    updateChat(sessionUuid, responseIndex, {
      ...currentMessage,
      text: cleanedText,
      artifacts: extractArtifacts(cleanedText),
    })

    const runner = getCodeRunner()
    const toolResults = []

    try {
      for (const call of calls) {
        const args = call.arguments || {}

        if (call.name === 'run_code') {
          const language = typeof args.language === 'string' ? args.language : 'python'
          const code = typeof args.code === 'string' ? args.code : ''

          if (!code) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'Missing code to execute.' }],
            })
            continue
          }

          if (!runner.isLanguageSupported(language)) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: `Unsupported language: ${language}` }],
            })
            continue
          }

          try {
            const results = await runner.execute(language, code)
            toolResults.push({
              name: call.name,
              success: !results.some(result => result.type === 'error'),
              results: results.map(result => ({
                type: result.type,
                content: result.content,
              })),
            })
          } catch (error) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: error instanceof Error ? error.message : String(error) }],
            })
          }
          continue
        }

        if (call.name === 'read_vfs') {
          if (!isVFSReady.value || !vfsInstance.value) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'VFS is not ready.' }],
            })
            continue
          }

          const path = typeof args.path === 'string' ? args.path : ''
          const encoding = typeof args.encoding === 'string' ? args.encoding : 'utf8'
          if (!path) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'Missing VFS path.' }],
            })
            continue
          }

          try {
            const data = await vfsInstance.value.readFile(path, encoding === 'binary' ? 'binary' : encoding)
            if (encoding === 'binary' && data instanceof Uint8Array) {
              const base64 = toBase64(data)
              toolResults.push({
                name: call.name,
                success: true,
                results: [{ type: 'vfs', content: base64, encoding: 'base64', path }],
              })
            } else {
              toolResults.push({
                name: call.name,
                success: true,
                results: [{ type: 'vfs', content: String(data), encoding: 'utf8', path }],
              })
            }
          } catch (error) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: error instanceof Error ? error.message : String(error) }],
            })
          }
          continue
        }

        if (call.name === 'write_vfs') {
          if (!isVFSReady.value || !vfsInstance.value) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'VFS is not ready.' }],
            })
            continue
          }

          const path = typeof args.path === 'string' ? args.path : ''
          const encoding = typeof args.encoding === 'string' ? args.encoding : 'utf8'
          const content = typeof args.content === 'string' ? args.content : ''
          if (!path) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'Missing VFS path.' }],
            })
            continue
          }

          try {
            if (encoding === 'base64') {
              const binary = Uint8Array.from(atob(content), c => c.charCodeAt(0))
              await vfsInstance.value.writeFile(path, binary, { binary: true })
            } else {
              await vfsInstance.value.writeFile(path, content, { encoding })
            }

            await runner.syncVFSToWorkers()
            toolResults.push({
              name: call.name,
              success: true,
              results: [{ type: 'vfs', content: 'ok', encoding, path }],
            })
          } catch (error) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: error instanceof Error ? error.message : String(error) }],
            })
          }
          continue
        }

        if (call.name === 'list_vfs') {
          if (!isVFSReady.value || !vfsInstance.value) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'VFS is not ready.' }],
            })
            continue
          }

          const path = typeof args.path === 'string' ? args.path : ''
          if (!path) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'Missing VFS path.' }],
            })
            continue
          }

          try {
            const entries = await vfsInstance.value.readdir(path)
            toolResults.push({
              name: call.name,
              success: true,
              results: [{ type: 'vfs', content: JSON.stringify(entries), path }],
            })
          } catch (error) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: error instanceof Error ? error.message : String(error) }],
            })
          }
          continue
        }

        if (call.name === 'stat_vfs') {
          if (!isVFSReady.value || !vfsInstance.value) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'VFS is not ready.' }],
            })
            continue
          }

          const path = typeof args.path === 'string' ? args.path : ''
          if (!path) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: 'Missing VFS path.' }],
            })
            continue
          }

          try {
            const stat = await vfsInstance.value.stat(path)
            toolResults.push({
              name: call.name,
              success: true,
              results: [{ type: 'vfs', content: JSON.stringify(stat), path }],
            })
          } catch (error) {
            toolResults.push({
              name: call.name,
              success: false,
              results: [{ type: 'error', content: error instanceof Error ? error.message : String(error) }],
            })
          }
          continue
        }

        toolResults.push({
          name: call.name,
          success: false,
          results: [{ type: 'error', content: `Unsupported tool: ${call.name}` }],
        })
      }
    } finally {
      toolRunning.value = false
    }

    const toolBlocks = toolResults.map(result => {
      return `\`\`\`tool_result\n${JSON.stringify(result)}\n\`\`\``
    })

    return `[[TOOL_RESULT]]\n${toolBlocks.join('\n')}`
  }

  return {
    loading,
    toolRunning,
    validateConversationInput,
    addUserMessage,
    initializeChatResponse,
    handleStreamingError,
    startStream,
    stopStream
  }
}
