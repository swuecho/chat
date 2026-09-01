import { v7 as uuidv7 } from 'uuid'
import { fetchDefaultChatModel } from './chat_model'
import { toUpdateChatSessionPayload } from './chat_session_payload'
import {
  createOrUpdateChatSession,
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
