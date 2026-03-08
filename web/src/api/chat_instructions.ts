import request from '@/utils/request/axios'

export interface ChatInstructions {
  artifactInstruction: string
}

export const fetchChatInstructions = async (): Promise<ChatInstructions> => {
  const response = await request.get('/chat_instructions')
  return response.data
}
