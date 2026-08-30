import { describe, expect, it } from 'vitest'
import { consumeAnswerEventStream, readAnswerStreamEvent } from '../sse'

const encoder = new TextEncoder()

function streamFrom(...chunks: string[]): ReadableStream<Uint8Array> {
  return new ReadableStream({
    start(controller) {
      for (const chunk of chunks)
        controller.enqueue(encoder.encode(chunk))
      controller.close()
    },
  })
}

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

describe('consumeAnswerEventStream', () => {
  it('reassembles fragmented CRLF frames and waits for persisted completion', async () => {
    const events: string[] = []
    const stream = streamFrom(
      'event: delta\r\ndata: {"type":"delta","answerId":"answer-1","del',
      'ta":"hello"}\r\n\r\nevent: completed\r\n',
      'data: {"type":"completed","answerId":"answer-1","persisted":true}\r\n\r\n',
    )

    await consumeAnswerEventStream(stream, (event) => {
      events.push(event.type)
    })

    expect(events).toEqual(['delta', 'completed'])
  })

  it('rejects a disconnect before durable completion', async () => {
    const stream = streamFrom('event: delta\ndata: {"type":"delta","delta":"partial"}\n\n')
    await expect(consumeAnswerEventStream(stream, () => {})).rejects.toThrow(
      'The response stream ended before it was saved',
    )
  })

  it('surfaces typed provider failure details', async () => {
    const stream = streamFrom('event: failed\ndata: {"type":"failed","code":"rate_limited","message":"Try later"}\n\n')
    await expect(consumeAnswerEventStream(stream, () => {})).rejects.toThrow('Try later')
  })

  it('treats typed cancellation as an unsuccessful terminal event', async () => {
    const stream = streamFrom('event: canceled\ndata: {"type":"canceled","code":"canceled"}\n\n')
    await expect(consumeAnswerEventStream(stream, () => {})).rejects.toThrow('canceled')
  })

  it('rejects completion before persistence', async () => {
    const stream = streamFrom('event: completed\ndata: {"type":"completed","persisted":false}\n\n')
    await expect(consumeAnswerEventStream(stream, () => {})).rejects.toThrow('The response was not saved')
  })

  it('rejects frames emitted after completion', async () => {
    const stream = streamFrom(
      'event: completed\ndata: {"type":"completed","persisted":true}\n\n',
      'event: delta\ndata: {"type":"delta","delta":"late"}\n\n',
    )
    await expect(consumeAnswerEventStream(stream, () => {})).rejects.toThrow(
      'Received an answer stream event after completion',
    )
  })
})
