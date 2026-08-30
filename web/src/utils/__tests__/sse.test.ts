import { describe, expect, it } from 'vitest'
import { readAnswerStreamEvent } from '../sse'

describe('readAnswerStreamEvent', () => {
  it('reads a persisted completion event', () => {
    expect(readAnswerStreamEvent(
      'event: completed\ndata: {"type":"completed","answerId":"answer-1","persisted":true}',
    )).toEqual({
      type: 'completed',
      answerId: 'answer-1',
      persisted: true,
    })
  })

  it('reads CRLF failure events', () => {
    expect(readAnswerStreamEvent(
      'event: failed\r\ndata: {"type":"failed","code":"persistence_failed","persisted":false}\r\n',
    )).toMatchObject({
      type: 'failed',
      code: 'persistence_failed',
    })
  })

  it('rejects untyped frames', () => {
    expect(readAnswerStreamEvent('data: {"choices":[{"delta":{"content":"hello"}}]}')).toBeNull()
  })

  it('rejects malformed terminal events', () => {
    expect(() => readAnswerStreamEvent('event: completed\ndata: nope')).toThrow(
      'Invalid completed stream event',
    )
  })

  it('reads typed deltas', () => {
    const event = readAnswerStreamEvent('event: delta\ndata: {"type":"delta","answerId":"answer-1","delta":"hello"}')
    expect(event).toMatchObject({ type: 'delta', answerId: 'answer-1', delta: 'hello' })
  })

  it('treats cancellation as terminal', () => {
    expect(readAnswerStreamEvent('event: canceled\ndata: {"type":"canceled","code":"canceled"}')).toMatchObject({
      type: 'canceled',
      code: 'canceled',
    })
  })
})
