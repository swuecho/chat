// getUserActiveChatSession
import request from '@/utils/request/axios'

export const getUserActiveChatSession = async () => {
  try {
    const response = await request.get('/uuid/user_active_chat_session')
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

// createOrUpdateUserActiveChatSession
export const createOrUpdateUserActiveChatSession = async (chatSessionUuid: string) => {
  try {
    const response = await request.put('/uuid/user_active_chat_session', {
      chatSessionUuid,
    })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
