import {
  createChatPrompt as createChatPromptGenerated,
  deleteChatPrompt as deleteChatPromptGenerated,
  updateChatPrompt as updateChatPromptGenerated,
} from './generated_client'

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
    return await createChatPromptGenerated({ body: payload })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const deleteChatPrompt = async (uuid: string) => {
  try {
    return await deleteChatPromptGenerated({ path: { uuid } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const updateChatPrompt = async (chat: Chat.Message) => {
  try {
    return await updateChatPromptGenerated({ path: { uuid: chat.uuid }, body: chat })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
