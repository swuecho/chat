import { getChatMessagesBySessionUUID } from './chat_message'

function format_chat_md(chat: Chat.Message): string {
  return `<sup><kbd><var>${chat.dateTime}</var></kbd></sup>:\n ${chat.text}`
}

export const fetchMarkdown = async (uuid: string) => {
  try {
    const chatData = await getChatMessagesBySessionUUID(uuid)
    /*
          uuid: string,
          dateTime: string
          text: string
          inversion?: boolean
          error?: boolean
          loading?: boolean
          isPrompt?: boolean
          */
    const markdown = chatData.map((chat: Chat.Message) => {
      if (chat.isPrompt)
        return `**system** ${format_chat_md(chat)}}`
      else if (chat.inversion)
        return `**user** ${format_chat_md(chat)}`
      else
        return `**assistant** ${format_chat_md(chat)}`
    }).join('\n\n----\n\n')
    return markdown
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const fetchConversationSnapshot = async (uuid: string): Promise<Chat.Message[]> => {
  try {
    const chatData = await getChatMessagesBySessionUUID(uuid)
    /*
          uuid: string,
          dateTime: string
          text: string
          inversion?: boolean
          error?: boolean
          loading?: boolean
          isPrompt?: boolean
          */
    return chatData
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
