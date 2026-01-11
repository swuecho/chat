import request from '@/utils/request/axios'

export interface CreateChatPromptPayload {
  uuid: string
  chatSessionUuid: string
  role: string
  content: string
  tokenCount: number
  userId: number
  createdBy: number
  updatedBy: number
}

export const createChatPrompt = async (payload: CreateChatPromptPayload) => {
  try {
    const response = await request.post('/chat_prompts', payload)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

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
