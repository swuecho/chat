import request from '@/utils/request/axios'

export const deleteChatPrompt = async (uuid: string) => {
  try {
    const response = await request.delete(`/uuid/chat_prompts/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const updateChatPrompt = async (chat: Chat.Message) => {
  try {
    const response = await request.put(`/uuid/chat_prompts/${chat.uuid}`, chat)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
