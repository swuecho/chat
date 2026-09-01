import {
  deleteChatMessage,
  deleteChatPrompt,
  updateChatMessage,
  updateChatPrompt,
} from './generated_client'

export const deleteChatData = async (chat: Chat.Message) => {
  if (chat?.isPrompt)
    await deleteChatPrompt({ path: { uuid: chat.uuid } })
  else
    await deleteChatMessage({ path: { uuid: chat.uuid } })
}

export const updateChatData = async (chat: Chat.Message) => {
  if (chat?.isPrompt)
    await updateChatPrompt({ path: { uuid: chat.uuid }, body: chat })
  else
    await updateChatMessage({ path: { uuid: chat.uuid }, body: chat })
}
