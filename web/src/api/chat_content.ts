import { deleteChatMessage, updateChatMessage } from './chat_message'
import { deleteChatPrompt, updateChatPrompt } from './chat_prompt'

export const deleteChatData = async (chat: Chat.Chat) => {
  if (chat?.isPrompt)
    await deleteChatPrompt(chat.uuid)
  else
    await deleteChatMessage(chat.uuid)
}

export const updateChatData = async (chat: Chat.Chat) => {
  if (chat?.isPrompt)
    await updateChatPrompt(chat)
  else
    await updateChatMessage(chat)
}
