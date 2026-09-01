import {
  getDefaultChatModel,
  listChatModels,
} from '@/api/generated_client'

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
