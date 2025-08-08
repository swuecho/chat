import request from '@/utils/request/axios'

export const updateChatMessage = async (chat: Chat.Message) => {
  try {
    const response = await request.put(`/uuid/chat_messages/${chat.uuid}`, chat)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const deleteChatMessage = async (uuid: string) => {
  try {
    const response = await request.delete(`/uuid/chat_messages/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const getChatMessagesBySessionUUID = async (uuid: string) => {
  try {
    const response = await request.get(`/uuid/chat_messages/chat_sessions/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const generateMoreSuggestions = async (messageUuid: string) => {
  try {
    const response = await request.post(`/uuid/chat_messages/${messageUuid}/generate-suggestions`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
