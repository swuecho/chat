export interface TerminalStreamEvent {
  type: 'completed' | 'failed'
  answerId?: string
  persisted?: boolean
  code?: string
  message?: string
}

export function readTerminalStreamEvent(frame: string): TerminalStreamEvent | null {
  const normalized = frame.replace(/\r\n/g, '\n')
  const eventType = normalized
    .split('\n')
    .find(line => line.startsWith('event:'))
    ?.slice('event:'.length)
    .trim()

  if (eventType !== 'completed' && eventType !== 'failed')
    return null

  const data = normalized
    .split('\n')
    .filter(line => line.startsWith('data:'))
    .map(line => line.slice('data:'.length).trimStart())
    .join('\n')

  try {
    return JSON.parse(data) as TerminalStreamEvent
  }
  catch {
    throw new Error(`Invalid ${eventType} stream event`)
  }
}
