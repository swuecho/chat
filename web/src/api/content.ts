import {
  deleteChatMessage,
  deleteChatPrompt,
  updateChatMessage,
  updateChatPrompt,
} from './generated_client'
import type { SimpleChatMessage } from './generated_client'

function toSimpleChatMessage(chat: Chat.Message): SimpleChatMessage {
  return {
    uuid: chat.uuid,
    dateTime: chat.dateTime,
    text: chat.text,
    inversion: chat.inversion ?? false,
    error: chat.error ?? false,
    loading: chat.loading ?? false,
    isPin: chat.isPin ?? false,
    isPrompt: chat.isPrompt ?? false,
    artifacts: chat.artifacts,
  }
}

export const deleteChatData = async (chat: Chat.Message) => {
  if (chat?.isPrompt)
    await deleteChatPrompt({ path: { uuid: chat.uuid } })
  else
    await deleteChatMessage({ path: { uuid: chat.uuid } })
}

export const updateChatData = async (chat: Chat.Message) => {
  const body = toSimpleChatMessage(chat)
  if (chat?.isPrompt)
    await updateChatPrompt({ path: { uuid: chat.uuid }, body })
  else
    await updateChatMessage({ path: { uuid: chat.uuid }, body })
}
