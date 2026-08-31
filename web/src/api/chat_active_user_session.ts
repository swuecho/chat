import { getActiveChatSession, setActiveChatSession } from './generated_client'

export const getUserActiveChatSession = async () => {
  try {
    return await getActiveChatSession()
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

// createOrUpdateUserActiveChatSession
export const createOrUpdateUserActiveChatSession = async (chatSessionUuid: string) => {
  try {
    return await setActiveChatSession({ body: { chatSessionUuid } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
