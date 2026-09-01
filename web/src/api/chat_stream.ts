import { streamChat } from '@/api/generated_client'
import type { AnswerEvent, ChatRequest } from '@/api/generated/types.gen'

export async function openChatStream(body: ChatRequest, signal?: AbortSignal): Promise<AsyncIterable<AnswerEvent>> {
  const result = await streamChat({ body, signal, sseMaxRetryAttempts: 1 })
  if (!result.stream)
    throw new Error('Chat stream response body is empty')
  return result.stream
}
