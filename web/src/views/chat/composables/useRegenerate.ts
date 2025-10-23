import { ref, type Ref } from 'vue'
import { deleteChatMessage } from '@/api'
import { nowISO } from '@/utils/date'
import { useChat } from '@/views/chat/hooks/useChat'
import { useStreamHandling } from './useStreamHandling'
import { t } from '@/locales'

export function useRegenerate(sessionUuidRef: Ref<string>) {
  const loading = ref<boolean>(false)
  const abortController = ref<AbortController | null>(null)
  const { addChat, updateChat, updateChatPartial } = useChat()
  const { streamRegenerateResponse, processStreamChunk } = useStreamHandling()


  function validateRegenerateInput(): boolean {
    return !loading.value
  }

  async function prepareRegenerateContext(
    index: number,
    chat: any,
    dataSources: any[]
  ): Promise<{ updateIndex: number; isRegenerate: boolean }> {
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid) {
      return { updateIndex: index, isRegenerate: true }
    }

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
        suggestedQuestionsLoading: true,
      })
    }

    return { updateIndex, isRegenerate }
  }

  async function handleUserMessageRegenerate(
    index: number,
    dataSources: any[]
  ): Promise<{ updateIndex: number; isRegenerate: boolean }> {
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid) {
      return { updateIndex: index, isRegenerate: false }
    }

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
        suggestedQuestionsLoading: true,
      })
    } else {
      addChat(sessionUuid, {
        uuid: '',
        dateTime: nowISO(),
        text: '',
        loading: true,
        inversion: false,
        error: false,
        suggestedQuestionsLoading: true,
      })
    }

    return { updateIndex, isRegenerate }
  }

  function handleRegenerateError(error: any, chatUuid: string, index: number): void {
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid) {
      return
    }

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

  function stopRegenerate(): void {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
      loading.value = false
    }
  }

  async function onRegenerate(index: number, dataSources: any[]): Promise<void> {
    if (!validateRegenerateInput()) return

    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid) {
      return
    }

    const chat = dataSources[index]
    abortController.value = new AbortController()
    const { updateIndex, isRegenerate } = await prepareRegenerateContext(index, chat, dataSources)

    try {
      await streamRegenerateResponse(
        sessionUuid,
        chat.uuid,
        updateIndex,
        isRegenerate,
        (chunk: string, updateIdx: number) => {
          processStreamChunk(chunk, updateIdx, sessionUuid)
        },
        abortController.value.signal
      )
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        // Stream was cancelled, no need to show error
        return
      }
      handleRegenerateError(error, chat.uuid, index)
    } finally {
      loading.value = false
      abortController.value = null
    }
  }

  return {
    loading,
    validateRegenerateInput,
    prepareRegenerateContext,
    handleUserMessageRegenerate,
    handleRegenerateError,
    onRegenerate,
    stopRegenerate
  }
}
