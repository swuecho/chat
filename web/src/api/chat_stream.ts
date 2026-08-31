import { streamChat } from '@/api/generated_client'
import type { ChatRequest } from '@/api/generated/types.gen'

export async function openChatStream(body: ChatRequest, signal?: AbortSignal): Promise<ReadableStream<Uint8Array>> {
  const stream = await streamChat({ body, signal, parseAs: 'stream' })
  if (!stream)
    throw new Error('Chat stream response body is empty')
  return stream as unknown as ReadableStream<Uint8Array>
}
