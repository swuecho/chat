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

export function readTerminalStreamEvent(frame: string): AnswerStreamEvent | null {
  const event = readAnswerStreamEvent(frame)
  if (!event || !['completed', 'failed', 'canceled'].includes(event.type))
    return null
  return event
}

export function answerEventAsLegacyFrame(event: AnswerStreamEvent): string {
  return `data: ${JSON.stringify({
    id: event.answerId,
    choices: [{ delta: { content: event.delta || '', suggestedQuestions: event.suggestedQuestions } }],
  })}`
}
