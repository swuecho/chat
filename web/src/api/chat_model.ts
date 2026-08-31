import {
  createChatModel as createChatModelRequest,
  deleteChatModel as deleteChatModelRequest,
  getDefaultChatModel,
  getTitleChatModel,
  listChatModels,
  setTitleChatModel,
  updateChatModel as updateChatModelRequest,
} from '@/api/generated_client'
import type { CreateChatModelData } from '@/api/generated/types.gen'

type ChatModelInput = CreateChatModelData['body']

export const fetchChatModel = () => listChatModels()

export const updateChatModel = (id: number, chatModel: ChatModelInput) =>
  updateChatModelRequest({ path: { id }, body: chatModel })

export const deleteChatModel = (id: number) =>
  deleteChatModelRequest({ path: { id } })

export const createChatModel = (chatModel: ChatModelInput) =>
  createChatModelRequest({ body: chatModel })

export const fetchDefaultChatModel = async () => {
  try {
    return await getDefaultChatModel()
  }
  catch (error) {
    console.warn('Default model not found, falling back to first available model')
    try {
      const models = await listChatModels()
      const enabledModels = models.filter(model => model.isEnable)
        .sort((a, b) => (a.orderNumber || 0) - (b.orderNumber || 0))
      if (enabledModels.length > 0)
        return enabledModels[0]
    }
    catch (fallbackError) {
      console.error('Failed to fetch fallback model:', fallbackError)
    }
    throw error
  }
}

export const fetchTitleChatModel = () => getTitleChatModel()

export const updateTitleChatModel = (modelId: number) =>
  setTitleChatModel({ body: { modelId } })
