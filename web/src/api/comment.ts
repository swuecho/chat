import { createChatComment as createChatCommentRequest, listSessionComments } from '@/api/generated_client'

// createChatComment(messageUUID:string, content:string)
export const createChatComment = async (sessionUUID: string, messageUUID: string, content: string) => {
  return createChatCommentRequest({ path: { sessionUUID, messageUUID }, body: { content } })
}
// return list of comments
// comment (sessionUUID: string, messageUUID: string, content: string, createdAt: string)
export const getConversationComments = async (sessionUUID: string) => {
  return listSessionComments({ path: { sessionUUID } })
}
