import { createWorkspace, ensureDefaultWorkspace as ensureDefaultWorkspaceRequest } from './generated_client'

// This adapter is intentionally a workflow rather than a second API surface:
// older servers can fail to create the default workspace, so the UI falls back
// to the equivalent explicit create operation.
export async function ensureDefaultWorkspace(): Promise<Chat.Workspace> {
  try {
    return await ensureDefaultWorkspaceRequest() as Chat.Workspace
  }
  catch (error: any) {
    if (error?.code !== 'RES_001')
      throw error

    return await createWorkspace({
      body: {
        name: 'General',
        description: 'Default workspace',
        color: '#6366f1',
        icon: 'folder',
        isDefault: true,
      },
    }) as Chat.Workspace
  }
}
