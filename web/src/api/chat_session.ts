import { v7 as uuidv7 } from 'uuid'
import request from '../utils/request/axios'
import { fetchDefaultChatModel } from './chat_model'
import { toUpdateChatSessionPayload } from './chat_session_payload'
import {
  createChatSession as createChatSessionGenerated,
  createOrUpdateChatSession,
  deleteChatSession as deleteChatSessionGenerated,
  listChatSessions,
  updateChatSessionTopic,
} from './generated_client'

export const getChatSessionDefault = async (title: string): Promise<Chat.Session> => {
  const default_model = await fetchDefaultChatModel()
  const uuid = uuidv7()
  return {
    title,
    isEdit: false,
    uuid,
    maxLength: 10,
    temperature: 1,
    model: default_model.name,
    maxTokens: default_model.defaultToken,
    topP: 1,
    n: 1,
    debug: false,
    exploreMode: true,
    artifactEnabled: false,
  }
}

export const getChatSessionsByUser = async () => {
  try {
    return await listChatSessions()
  }
  catch (error) {
    console.error('Error in getChatSessionsByUser:', error)
    throw error
  }
}

export const deleteChatSession = async (uuid: string) => {
  try {
    return await deleteChatSessionGenerated({ path: { uuid } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const createChatSession = async (
  uuid: string,
  name: string,
  model: string | undefined,
  defaultSystemPrompt?: string,
) => {
  try {
    return await createChatSessionGenerated({
      body: {
        uuid,
        topic: name,
        model: model ?? '',
        defaultSystemPrompt: defaultSystemPrompt ?? '',
      },
    })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const renameChatSession = async (uuid: string, name: string) => {
  try {
    return await updateChatSessionTopic({ path: { uuid }, body: { topic: name } })
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
    if (sessionUuid !== session_data.uuid)
      throw new Error('Session UUID does not match the update route')

    return await createOrUpdateChatSession({
      path: { uuid: sessionUuid },
      body: toUpdateChatSessionPayload(session_data),
    })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
