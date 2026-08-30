import { describe, expect, it } from 'vitest'
import { answerEventAsLegacyFrame, readAnswerStreamEvent, readTerminalStreamEvent } from '../sse'

describe('readTerminalStreamEvent', () => {
  it('reads a persisted completion event', () => {
    expect(readTerminalStreamEvent(
      'event: completed\ndata: {"type":"completed","answerId":"answer-1","persisted":true}',
    )).toEqual({
      type: 'completed',
      answerId: 'answer-1',
      persisted: true,
    })
  })

  it('reads CRLF failure events', () => {
    expect(readTerminalStreamEvent(
      'event: failed\r\ndata: {"type":"failed","code":"persistence_failed","persisted":false}\r\n',
    )).toMatchObject({
      type: 'failed',
      code: 'persistence_failed',
    })
  })

  it('ignores token frames', () => {
    expect(readTerminalStreamEvent('data: {"choices":[{"delta":{"content":"hello"}}]}')).toBeNull()
  })

  it('rejects malformed terminal events', () => {
    expect(() => readTerminalStreamEvent('event: completed\ndata: nope')).toThrow(
      'Invalid completed stream event',
    )
  })

  it('reads typed deltas and adapts them for legacy consumers', () => {
    const event = readAnswerStreamEvent('event: delta\ndata: {"type":"delta","answerId":"answer-1","delta":"hello"}')
    expect(event).toMatchObject({ type: 'delta', answerId: 'answer-1', delta: 'hello' })
    expect(answerEventAsLegacyFrame(event!)).toContain('"content":"hello"')
  })

  it('treats cancellation as terminal', () => {
    expect(readTerminalStreamEvent('event: canceled\ndata: {"type":"canceled","code":"canceled"}')).toMatchObject({
      type: 'canceled',
      code: 'canceled',
    })
  })
})
