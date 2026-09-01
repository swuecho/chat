import { beforeEach, describe, expect, it, vi } from 'vitest'

import { updateChatSession } from '../chat_session'

const generated = vi.hoisted(() => ({
  createOrUpdateChatSession: vi.fn(),
}))

vi.mock('../generated_client', () => generated)
vi.mock('../chat_model', () => ({ fetchDefaultChatModel: vi.fn() }))
vi.mock('../chat_message', () => ({ clearChatMessagesBySessionUUID: vi.fn() }))

describe('chat session generated client migration', () => {
  beforeEach(() => vi.clearAllMocks())

  it('keeps the UI-to-contract payload mapper at the migration boundary', async () => {
    generated.createOrUpdateChatSession.mockResolvedValue({ uuid: 'session-1' })
    const session = {
      uuid: 'session-1',
      title: 'Topic',
      isEdit: false,
      model: 'model',
      maxLength: 10,
      temperature: 1,
      topP: 1,
      n: 1,
      maxTokens: 2048,
      debug: false,
      summarizeMode: false,
      exploreMode: false,
      artifactEnabled: false,
    }

    await updateChatSession('session-1', session)

    expect(generated.createOrUpdateChatSession).toHaveBeenCalledWith({
      path: { uuid: 'session-1' },
      body: expect.objectContaining({ uuid: 'session-1', topic: 'Topic', model: 'model' }),
    })
  })
})
