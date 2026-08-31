import { getChatInstructions } from '@/api/generated_client'

export interface ChatInstructions {
  artifactInstruction: string
}

export const fetchChatInstructions = async (): Promise<ChatInstructions> => {
  return getChatInstructions()
}
