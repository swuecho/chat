import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  createSessionInWorkspace,
  createWorkspace,
  deleteWorkspace,
  getSessionsByWorkspace,
  getWorkspace,
  getWorkspaces,
  updateWorkspace,
} from '../chat_workspace'

const generated = vi.hoisted(() => ({
  listWorkspaces: vi.fn(),
  getWorkspace: vi.fn(),
  createWorkspace: vi.fn(),
  updateWorkspace: vi.fn(),
  deleteWorkspace: vi.fn(),
  updateWorkspaceOrder: vi.fn(),
  setDefaultWorkspace: vi.fn(),
  ensureDefaultWorkspace: vi.fn(),
  createWorkspaceSession: vi.fn(),
  listWorkspaceSessions: vi.fn(),
  autoMigrateLegacySessions: vi.fn(),
}))

vi.mock('../generated_client', () => generated)

describe('chat workspace generated client migration', () => {
  beforeEach(() => vi.clearAllMocks())

  it('uses generated operations with typed path parameters', async () => {
    generated.listWorkspaces.mockResolvedValue([{ uuid: 'ws-1' }])
    generated.getWorkspace.mockResolvedValue({ uuid: 'ws-1' })
    generated.createWorkspace.mockResolvedValue({ uuid: 'ws-1' })
    generated.updateWorkspace.mockResolvedValue({ uuid: 'ws-1' })
    generated.deleteWorkspace.mockResolvedValue({})
    generated.createWorkspaceSession.mockResolvedValue({ uuid: 'session-1' })
    generated.listWorkspaceSessions.mockResolvedValue([{ uuid: 'session-1' }])

    await expect(getWorkspaces()).resolves.toEqual([{ uuid: 'ws-1' }])
    await getWorkspace('ws-1')
    await createWorkspace({ name: 'General', description: 'Default', color: '#6366f1', icon: 'folder', isDefault: true })
    await updateWorkspace('ws-1', { name: 'Renamed' })
    await deleteWorkspace('ws-1')
    await createSessionInWorkspace('ws-1', { topic: 'Hello', model: 'model' })
    await getSessionsByWorkspace('ws-1')

    expect(generated.getWorkspace).toHaveBeenCalledWith({ path: { uuid: 'ws-1' } })
    expect(generated.createWorkspace).toHaveBeenCalledWith({
      body: { name: 'General', description: 'Default', color: '#6366f1', icon: 'folder', isDefault: true },
    })
    expect(generated.updateWorkspace).toHaveBeenCalledWith({ path: { uuid: 'ws-1' }, body: { name: 'Renamed' } })
    expect(generated.deleteWorkspace).toHaveBeenCalledWith({ path: { uuid: 'ws-1' } })
    expect(generated.createWorkspaceSession).toHaveBeenCalledWith({
      path: { uuid: 'ws-1' },
      body: { topic: 'Hello', model: 'model', defaultSystemPrompt: '' },
    })
    expect(generated.listWorkspaceSessions).toHaveBeenCalledWith({ path: { uuid: 'ws-1' } })
  })

  it('defaults optional session fields at the contract boundary', async () => {
    generated.createWorkspaceSession.mockResolvedValue({ uuid: 'session-1' })

    await createSessionInWorkspace('ws-1', { topic: 'Hello' })

    expect(generated.createWorkspaceSession).toHaveBeenCalledWith({
      path: { uuid: 'ws-1' },
      body: { topic: 'Hello', model: '', defaultSystemPrompt: '' },
    })
  })
})
