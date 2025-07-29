import request from '@/utils/request/axios'

export const fetchChatModel = async () => {
  try {
    const response = await request.get('/chat_model')
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const updateChatModel = async (id: number, chatModel: any) => {
  try {
    const response = await request.put(`/chat_model/${id}`, chatModel)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const deleteChatModel = async (id: number) => {
  try {
    const response = await request.delete(`/chat_model/${id}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
export const createChatModel = async (chatModel: any) => {
  try {
    const response = await request.post('/chat_model', chatModel)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const fetchDefaultChatModel = async () => {
  try {
    const response = await request.get('/chat_model/default')
    return response.data
  }
  catch (error: any) {
    console.error('Failed to fetch default chat model:', error)
    
    // If default model API fails, try to get all models and use the first enabled one
    if (error.response?.data?.code === 'RES_001' || error.response?.status === 500) {
      console.warn('Default model not found, falling back to first available model')
      try {
        const allModelsResponse = await request.get('/chat_model')
        const enabledModels = allModelsResponse.data?.filter((model: any) => model.isEnable) || []
        
        if (enabledModels.length > 0) {
          // Sort by order number and return first one
          enabledModels.sort((a: any, b: any) => (a.orderNumber || 0) - (b.orderNumber || 0))
          console.log('Using fallback model:', enabledModels[0].name)
          return enabledModels[0]
        }
      } catch (fallbackError) {
        console.error('Failed to fetch fallback model:', fallbackError)
      }
    }
    
    throw error
  }
}
