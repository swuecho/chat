import request from '@/utils/request/axios'

// createChatComment(messageUUID:string, content:string)
export const createChatComment = async (sessionUUID: string , messageUUID: string, content: string) => {
  try {
    const response = await request.post(`/uuid/conversation/${sessionUUID}/chat_messages/${messageUUID}/comments`, {
      content
    })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
// return list of comments
// comment (sessionUUID: string, messageUUID: string, content: string, createdAt: string)
export const getConversationComments = async (sessionUUID: string) => {
  try {
    const response = await request.get(`/uuid/conversations/${sessionUUID}/comments`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}