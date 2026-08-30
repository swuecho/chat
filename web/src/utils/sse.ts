export type AnswerStreamEventType
  = 'started'
  | 'delta'
  | 'reasoning_delta'
  | 'suggested_questions'
  | 'completed'
  | 'failed'
  | 'canceled'

export interface AnswerStreamEvent {
  type: AnswerStreamEventType
  answerId?: string
  delta?: string
  suggestedQuestions?: string[]
  persisted?: boolean
  code?: string
  message?: string
}

export type AnswerStreamEventHandler = (event: AnswerStreamEvent) => void | Promise<void>

const eventTypes = new Set<AnswerStreamEventType>([
  'started', 'delta', 'reasoning_delta', 'suggested_questions',
  'completed', 'failed', 'canceled',
])

export function readAnswerStreamEvent(frame: string): AnswerStreamEvent | null {
  const normalized = frame.replace(/\r\n/g, '\n')
  const eventType = normalized
    .split('\n')
    .find(line => line.startsWith('event:'))
    ?.slice('event:'.length)
    .trim() as AnswerStreamEventType | undefined

  if (!eventType || !eventTypes.has(eventType))
    return null

  const data = normalized
    .split('\n')
    .filter(line => line.startsWith('data:'))
    .map(line => line.slice('data:'.length).trimStart())
    .join('\n')

  try {
    const event = JSON.parse(data) as AnswerStreamEvent
    if (event.type !== eventType)
      throw new Error(`Mismatched ${eventType} stream event`)
    return event
  }
  catch (error) {
    if (error instanceof Error && error.message.startsWith('Mismatched'))
      throw error
    throw new Error(`Invalid ${eventType} stream event`)
  }
}

export async function consumeAnswerEventStream(
  stream: ReadableStream<Uint8Array>,
  onEvent: AnswerStreamEventHandler,
): Promise<void> {
  const reader = stream.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let completed = false

  const processFrame = async (frame: string): Promise<void> => {
    if (!frame.trim())
      return
    if (completed)
      throw new Error('Received an answer stream event after completion')

    const event = readAnswerStreamEvent(frame)
    if (!event)
      throw new Error('Received an untyped answer stream frame')
    if (event.type === 'failed' || event.type === 'canceled')
      throw new Error(event.message || event.code || `Stream ${event.type}`)
    if (event.type === 'completed') {
      if (!event.persisted)
        throw new Error('The response was not saved')
      completed = true
    }
    await onEvent(event)
  }

  const processBuffer = async (flush: boolean): Promise<void> => {
    const normalized = buffer.replace(/\r\n/g, '\n')
    const frames = normalized.split('\n\n')
    buffer = frames.pop() || ''
    for (const frame of frames)
      await processFrame(frame)
    if (flush && buffer.trim()) {
      const finalFrame = buffer
      buffer = ''
      await processFrame(finalFrame)
    }
  }

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done)
        break
      buffer += decoder.decode(value, { stream: true })
      await processBuffer(false)
    }
    buffer += decoder.decode()
    await processBuffer(true)
    if (!completed)
      throw new Error('The response stream ended before it was saved')
  }
  finally {
    reader.releaseLock()
  }
}
