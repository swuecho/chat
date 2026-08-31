import { beforeEach, describe, expect, it, vi } from 'vitest'

import { deleteChatSession, getChatSessionsByUser, renameChatSession, updateChatSession } from '../chat_session'

const generated = vi.hoisted(() => ({
  createChatSession: vi.fn(),
  createOrUpdateChatSession: vi.fn(),
  deleteChatSession: vi.fn(),
  listChatSessions: vi.fn(),
  updateChatSessionTopic: vi.fn(),
}))

vi.mock('../generated_client', () => generated)
vi.mock('../chat_model', () => ({ fetchDefaultChatModel: vi.fn() }))
vi.mock('../chat_message', () => ({ clearChatMessagesBySessionUUID: vi.fn() }))

describe('chat session generated client migration', () => {
  beforeEach(() => vi.clearAllMocks())

  it('uses generated operations with typed path parameters', async () => {
    generated.listChatSessions.mockResolvedValue([{ uuid: 'session-1' }])
    generated.deleteChatSession.mockResolvedValue({})
    generated.updateChatSessionTopic.mockResolvedValue({ uuid: 'session-1' })

    await expect(getChatSessionsByUser()).resolves.toEqual([{ uuid: 'session-1' }])
    await deleteChatSession('session-1')
    await renameChatSession('session-1', 'Renamed')

    expect(generated.deleteChatSession).toHaveBeenCalledWith({ path: { uuid: 'session-1' } })
    expect(generated.updateChatSessionTopic).toHaveBeenCalledWith({
      path: { uuid: 'session-1' },
      body: { topic: 'Renamed' },
    })
  })

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
