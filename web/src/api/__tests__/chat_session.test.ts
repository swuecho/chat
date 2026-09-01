import { describe, expect, it } from 'vitest'
import { toUpdateChatSessionPayload } from '../../store/modules/session/payload'

describe('toUpdateChatSessionPayload', () => {
  it('maps the UI session to the strict writable API contract', () => {
    const payload = toUpdateChatSessionPayload({
      uuid: '01990a45-8a36-7e51-bf7c-a8df8d6b8e91',
      title: 'Updated topic',
      isEdit: true,
      maxLength: 20,
      temperature: 0,
      model: 'test-model',
      topP: 0,
      n: 1,
      maxTokens: 2048,
      debug: true,
      summarizeMode: true,
      exploreMode: true,
      artifactEnabled: true,
      workspaceUuid: '01990a45-8a36-7e51-bf7c-a8df8d6b8e92',
    })

    expect(payload).toEqual({
      uuid: '01990a45-8a36-7e51-bf7c-a8df8d6b8e91',
      topic: 'Updated topic',
      maxLength: 20,
      temperature: 0,
      model: 'test-model',
      topP: 0,
      n: 1,
      maxTokens: 2048,
      debug: true,
      summarizeMode: true,
      exploreMode: true,
      artifactEnabled: true,
      workspaceUuid: '01990a45-8a36-7e51-bf7c-a8df8d6b8e92',
    })
    expect(payload).not.toHaveProperty('title')
    expect(payload).not.toHaveProperty('isEdit')
  })

  it('preserves valid zero values and supplies defaults only when absent', () => {
    const payload = toUpdateChatSessionPayload({
      uuid: '01990a45-8a36-7e51-bf7c-a8df8d6b8e91',
      title: 'Topic',
      isEdit: false,
      temperature: 0,
      model: 'test-model',
      topP: 0,
      maxTokens: 1024,
    })

    expect(payload.temperature).toBe(0)
    expect(payload.topP).toBe(0)
    expect(payload.n).toBe(1)
    expect(payload.maxLength).toBe(10)
  })
})
