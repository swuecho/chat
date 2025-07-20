import { ref } from 'vue'
import { useMessage } from 'naive-ui'
// @ts-ignore
import { v7 as uuidv7 } from 'uuid'
import { deleteChatMessage } from '@/api'
import { nowISO } from '@/utils/date'
import { useChat } from '@/views/chat/hooks/useChat'
import { useStreamHandling } from './useStreamHandling'
import { t } from '@/locales'

export function useRegenerate(sessionUuid: string) {
  const nui_msg = useMessage()
  const loading = ref<boolean>(false)
  const { addChat, updateChat, updateChatPartial } = useChat()
  const { streamRegenerateResponse, processStreamChunk } = useStreamHandling()

  let controller = new AbortController()

  function validateRegenerateInput(): boolean {
    return !loading.value
  }

  async function prepareRegenerateContext(
    index: number,
    chat: any,
    dataSources: any[]
  ): Promise<{ updateIndex: number; isRegenerate: boolean }> {
    controller = new AbortController()
    loading.value = true

    let updateIndex = index
    let isRegenerate = true

    if (chat.inversion) {
      const result = await handleUserMessageRegenerate(index, dataSources)
      updateIndex = result.updateIndex
      isRegenerate = result.isRegenerate
    } else {
      updateChat(sessionUuid, index, {
        uuid: chat.uuid,
        dateTime: nowISO(),
        text: '',
        inversion: false,
        error: false,
        loading: true,
      })
    }

    return { updateIndex, isRegenerate }
  }

  async function handleUserMessageRegenerate(
    index: number,
    dataSources: any[]
  ): Promise<{ updateIndex: number; isRegenerate: boolean }> {
    const chatNext = dataSources[index + 1]
    let updateIndex = index + 1
    const isRegenerate = false

    if (chatNext) {
      await deleteChatMessage(chatNext.uuid)
      updateChat(sessionUuid, updateIndex, {
        uuid: chatNext.uuid,
        dateTime: nowISO(),
        text: '',
        inversion: false,
        error: false,
        loading: true,
      })
    } else {
      addChat(sessionUuid, {
        uuid: '',
        dateTime: nowISO(),
        text: '',
        loading: true,
        inversion: false,
        error: false,
      })
    }

    return { updateIndex, isRegenerate }
  }

  function handleRegenerateStreamProgress(progress: any, updateIndex: number): void {
    const xhr = progress.event.target
    const { responseText, status } = xhr

    if (status >= 400) {
      // Handle error - reuse stream error handling
      try {
        const errorJson = JSON.parse(responseText)
        console.error('Stream error:', responseText)
        nui_msg.error(`${errorJson.code} : ${errorJson.message}`)
      } catch (parseError) {
        nui_msg.error('An unexpected error occurred')
      }
      return
    }

    processStreamChunk(responseText, updateIndex, sessionUuid)
  }

  function handleRegenerateError(error: any, chatUuid: string, index: number): void {
    console.error('Regenerate error:', error)

    if (error.message === 'canceled') {
      updateChatPartial(sessionUuid, index, {
        loading: false,
      })
      return
    }

    const errorMessage = error?.message ?? t('common.wrong')

    updateChat(sessionUuid, index, {
      uuid: chatUuid,
      dateTime: nowISO(),
      text: errorMessage,
      inversion: false,
      error: true,
      loading: false,
    })
  }

  async function onRegenerate(index: number, dataSources: any[]): Promise<void> {
    if (!validateRegenerateInput()) return

    const chat = dataSources[index]
    const { updateIndex, isRegenerate } = await prepareRegenerateContext(index, chat, dataSources)

    try {
      await streamRegenerateResponse(
        sessionUuid,
        chat.uuid,
        updateIndex,
        isRegenerate,
        (progress: any, updateIdx: number) => {
          handleRegenerateStreamProgress(progress, updateIdx)
        }
      )
    } catch (error) {
      handleRegenerateError(error, chat.uuid, index)
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    validateRegenerateInput,
    prepareRegenerateContext,
    handleUserMessageRegenerate,
    handleRegenerateStreamProgress,
    handleRegenerateError,
    onRegenerate
  }
}