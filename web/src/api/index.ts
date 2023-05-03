import type { AxiosProgressEvent } from 'axios'
import { getChatMessagesBySessionUUID } from './chat_message'
import request from '@/utils/request/axios'
export * from './user'
export * from './chat_model'
export * from './chat_session'
export * from './chat_user_model_privilege'
export * from './chat_content'

export async function fetchChatStream(
  sessionUuid: string,
  chatUuid: string,
  regenerate: boolean,
  prompt: string,
  onDownloadProgress?: (progressEvent: AxiosProgressEvent) => void,
) {
  try {
    const response = await request.post(
      '/chat_stream',
      { regenerate, prompt, sessionUuid, chatUuid },
      { onDownloadProgress },
    )

    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const GetUserData = async (page: number, size: number) => {
  try {
    const response = await request.post('/admin/user_stats', {
      page,
      size,
    })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const UpdateRateLimit = async (email: string, rateLimit: number) => {
  try {
    const response = await request.post('/admin/rate_limit', {
      email,
      rateLimit,
    })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

function format_chat_md(chat: Chat.Chat): string {
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
    const markdown = chatData.map((chat: Chat.Chat) => {
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

export const fetchConversationSnapshot = async (uuid: string): Promise<Chat.Chat[]> => {
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

export const updateUserFullName = async (data: any): Promise<any> => {
  try {
    const response = await request.put('/admin/users', data)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

// CreateSessionFromSnapshot
export const CreateSessionFromSnapshot = async (snapshot_uuid: string) => {
  try {
    const response = await request.post(`/uuid/chat_session_from_snapshot/${snapshot_uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
