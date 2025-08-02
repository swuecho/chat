import { updateChatData } from '@/api'
import { useMessageStore } from '@/store'
import { nowISO } from '@/utils/date'

export function useChat() {
  const messageStore = useMessageStore()

  const getChatByUuidAndIndex = (uuid: string, index: number) => {
    const messages = messageStore.getChatSessionDataByUuid(uuid)
    return messages && messages[index] ? messages[index] : null
  }

  const addChat = (uuid: string, chat: Chat.Message) => {
    messageStore.addMessage(uuid, chat)
  }

  const deleteChat = (uuid: string, index: number) => {
    const messages = messageStore.getChatSessionDataByUuid(uuid)
    if (messages && messages[index]) {
      messageStore.removeMessage(uuid, messages[index].uuid)
    }
  }

  const updateChat = (uuid: string, index: number, chat: Chat.Message) => {
    const messages = messageStore.getChatSessionDataByUuid(uuid)
    if (messages && messages[index]) {
      messageStore.updateMessage(uuid, messages[index].uuid, chat)
    }
  }

  const updateChatPartial = (uuid: string, index: number, chat: Partial<Chat.Message>) => {
    const messages = messageStore.getChatSessionDataByUuid(uuid)
    if (messages && messages[index]) {
      messageStore.updateMessage(uuid, messages[index].uuid, chat)
    }
  }

  const updateChatText = async (uuid: string, index: number, text: string) => {
    const messages = messageStore.getChatSessionDataByUuid(uuid)
    const chat = messages && messages[index] ? messages[index] : null
    if (!chat)
      return
    chat.text = text
    // update time stamp
    chat.dateTime = nowISO()
    messageStore.updateMessage(uuid, chat.uuid, chat)
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
