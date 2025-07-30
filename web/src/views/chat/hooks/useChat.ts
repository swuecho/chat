import { updateChatData } from '@/api'
import { useMessageStore } from '@/store'
import { nowISO } from '@/utils/date'

export function useChat() {
  const messageStore = useMessageStore()

  const getChatByUuidAndIndex = (uuid: string, index: number) => {
    return messageStore.getChatByUuidAndIndex(uuid, index)
  }

  const addChat = (uuid: string, chat: Chat.Message) => {
    messageStore.addChatByUuid(uuid, chat)
  }

  const deleteChat = (uuid: string, index: number) => {
    messageStore.deleteChatByUuid(uuid, index)
  }

  const updateChat = (uuid: string, index: number, chat: Chat.Message) => {
    messageStore.updateChatByUuid(uuid, index, chat)
  }

  const updateChatPartial = (uuid: string, index: number, chat: Partial<Chat.Message>) => {
    messageStore.updateChatPartialByUuid(uuid, index, chat)
  }

  const updateChatText = async (uuid: string, index: number, text: string) => {
    const chat = messageStore.getChatByUuidAndIndex(uuid, index)
    if (!chat)
      return
    chat.text = text
    // update time stamp
    chat.dateTime = nowISO()
    messageStore.updateChatByUuid(uuid, index, chat)
    // sync text to server
    await updateChatData(chat)
  }

  return {
    addChat,
    deleteChat,
    updateChat,
    updateChatText,
    updateChatPartial,
    getChatByUuidAndIndex,
  }
}
