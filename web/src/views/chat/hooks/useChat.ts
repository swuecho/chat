import { updateChatData } from '@/api'
import { useChatStore } from '@/store'
import { nowISO } from '@/utils/date'

export function useChat() {
  const chatStore = useChatStore()

  const getChatByUuidAndIndex = (uuid: string, index: number) => {
    return chatStore.getChatByUuidAndIndex(uuid, index)
  }

  const addChat = (uuid: string, chat: Chat.Message) => {
    chatStore.addChatByUuid(uuid, chat)
  }

  const updateChat = (uuid: string, index: number, chat: Chat.Message) => {
    chatStore.updateChatByUuid(uuid, index, chat)
  }

  const updateChatPartial = (uuid: string, index: number, chat: Partial<Chat.Message>) => {
    chatStore.updateChatPartialByUuid(uuid, index, chat)
  }

  const updateChatText = async (uuid: string, index: number, text: string) => {
    const chat = chatStore.getChatByUuidAndIndex(uuid, index)
    if (!chat)
      return
    chat.text = text
    // update time stamp
    chat.dateTime = nowISO()
    chatStore.updateChatByUuid(uuid, index, chat)
    // sync text to server
    await updateChatData(chat)
  }

  return {
    addChat,
    updateChat,
    updateChatText,
    updateChatPartial,
    getChatByUuidAndIndex,
  }
}
