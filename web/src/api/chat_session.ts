import { v7 as uuidv7 } from 'uuid'
import { fetchDefaultChatModel } from './chat_model'
import request from '@/utils/request/axios'

export const getChatSessionDefault = async (title: string): Promise<Chat.Session> => {
  const default_model = await fetchDefaultChatModel()
  const uuid = uuidv7()
  return {
    title,
    isEdit: false,
    uuid,
    maxLength: 4,
    temperature: 1,
    model: default_model.name,
    maxTokens: default_model.defaultToken,
    topP: 1,
    n: 1,
    debug: false,
  }
}

export const getChatSessionsByUser = async () => {
  try {
    const response = await request.get('/chat_sessions/users')
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const deleteChatSession = async (uuid: string) => {
  try {
    const response = await request.delete(`/uuid/chat_sessions/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const createChatSession = async (uuid: string, name: string, model: string | undefined) => {
  try {
    const response = await request.post('/uuid/chat_sessions', {
      uuid,
      topic: name,
      model
    })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const renameChatSession = async (uuid: string, name: string) => {
  try {
    const response = await request.put(`/uuid/chat_sessions/topic/${uuid}`, { topic: name })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const clearSessionChatMessages = async (sessionUuid: string) => {
  try {
    const response = await request.delete(`/uuid/chat_messages/chat_sessions/${sessionUuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const updateChatSession = async (sessionUuid: string, session_data: Chat.Session) => {
  try {
    const response = await request.put(`/uuid/chat_sessions/${sessionUuid}`, session_data)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
