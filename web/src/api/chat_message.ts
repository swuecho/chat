import {
  deleteChatHistory,
  deleteChatMessage as deleteChatMessageGenerated,
  generateMessageSuggestions,
  getChatHistory,
  updateChatMessage as updateChatMessageGenerated,
} from './generated_client'

export const updateChatMessage = async (chat: Chat.Message) => {
  try {
    return await updateChatMessageGenerated({ path: { uuid: chat.uuid }, body: chat })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const deleteChatMessage = async (uuid: string) => {
  try {
    return await deleteChatMessageGenerated({ path: { uuid } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const getChatMessagesBySessionUUID = async (uuid: string) => {
  try {
    return await getChatHistory({ path: { uuid } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const generateMoreSuggestions = async (messageUuid: string) => {
  try {
    return await generateMessageSuggestions({ path: { uuid: messageUuid } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const clearChatMessagesBySessionUUID = async (sessionUuid: string) => {
  try {
    return await deleteChatHistory({ path: { uuid: sessionUuid } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
