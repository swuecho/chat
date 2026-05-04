import { type Ref, ref } from 'vue'
import { useStreamHandling } from './useStreamHandling'
import { nowISO } from '@/utils/date'
import { useChat } from '@/views/chat/hooks/useChat'
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
    dataSources: any[],
  ): Promise<{ updateIndex: number; isRegenerate: boolean; targetChatUuid: string }> {
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid)
      return { updateIndex: index, isRegenerate: true, targetChatUuid: chat.uuid }

    loading.value = true

    let updateIndex = index
    let isRegenerate = true
    let targetChatUuid = chat.uuid

    if (chat.inversion) {
      const result = await handleUserMessageRegenerate(index, dataSources)
      updateIndex = result.updateIndex
      isRegenerate = result.isRegenerate
      targetChatUuid = result.targetChatUuid
    }
    else {
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

    return { updateIndex, isRegenerate, targetChatUuid }
  }

  async function handleUserMessageRegenerate(
    index: number,
    dataSources: any[],
  ): Promise<{ updateIndex: number; isRegenerate: boolean; targetChatUuid: string }> {
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid)
      return { updateIndex: index, isRegenerate: false, targetChatUuid: dataSources[index]?.uuid || '' }

    const chatNext = dataSources[index + 1]
    let updateIndex = index + 1
    let isRegenerate = false
    let targetChatUuid = dataSources[index]?.uuid || ''

    if (chatNext) {
      updateChat(sessionUuid, updateIndex, {
        uuid: chatNext.uuid,
        dateTime: nowISO(),
        text: '',
        inversion: false,
        error: false,
        loading: true,
        suggestedQuestionsLoading: true,
      })
      isRegenerate = true
      targetChatUuid = chatNext.uuid
    }
    else {
      updateIndex = dataSources.length
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

    return { updateIndex, isRegenerate, targetChatUuid }
  }

  function handleRegenerateError(error: any, chatUuid: string, index: number): void {
    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid)
      return

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
    }
    loading.value = false
  }

  async function onRegenerate(index: number, dataSources: any[]): Promise<void> {
    if (!validateRegenerateInput())
      return

    const sessionUuid = sessionUuidRef.value
    if (!sessionUuid)
      return

    const chat = dataSources[index]
    abortController.value = new AbortController()
    const { updateIndex, isRegenerate, targetChatUuid } = await prepareRegenerateContext(index, chat, dataSources)

    try {
      await streamRegenerateResponse(
        sessionUuid,
        targetChatUuid,
        updateIndex,
        isRegenerate,
        (chunk: string, updateIdx: number) => {
          processStreamChunk(chunk, updateIdx, sessionUuid)
        },
        abortController.value.signal,
      )
    }
    catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        // Stream was cancelled, no need to show error
        return
      }
      handleRegenerateError(error, targetChatUuid, updateIndex)
    }
    finally {
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
    stopRegenerate,
  }
}
