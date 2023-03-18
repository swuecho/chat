import { updateChatData } from '@/api'
import { useChatStore } from '@/store'

export function useChat() {
  const chatStore = useChatStore()

  const getChatByUuidAndIndex = (uuid: string, index: number) => {
    return chatStore.getChatByUuidAndIndex(uuid, index)
  }

  const addChat = (uuid: string, chat: Chat.Chat) => {
    chatStore.addChatByUuid(uuid, chat)
  }

  const updateChat = (uuid: string, index: number, chat: Chat.Chat) => {
    chatStore.updateChatByUuid(uuid, index, chat)
  }

  const updateChatPartial = (uuid: string, index: number, chat: Partial<Chat.Chat>) => {
    chatStore.updateChatPartialByUuid(uuid, index, chat)
  }

  const updateChatText = async (uuid: string, index: number, text: string) => {
    const chat = chatStore.getChatByUuidAndIndex(uuid, index)
    if (!chat)
      return
    chat.text = text
    // updat time stamp
    chat.dateTime = new Date().toLocaleString()
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
