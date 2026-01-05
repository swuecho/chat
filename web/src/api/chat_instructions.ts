import request from '@/utils/request/axios'

export interface ChatInstructions {
  artifactInstruction: string
  toolInstruction: string
}

export const fetchChatInstructions = async (): Promise<ChatInstructions> => {
  const response = await request.get('/chat_instructions')
  return response.data
}
