import request from '@/utils/request/axios'

export interface CreateWorkspaceRequest {
  name: string
  description?: string
  color?: string
  icon?: string
  isDefault?: boolean
}

export interface UpdateWorkspaceRequest {
  name: string
  description?: string
  color?: string
  icon?: string
}

export interface CreateSessionInWorkspaceRequest {
  topic: string
  model?: string
}

// Get all workspaces for the current user
export const getWorkspaces = async (): Promise<Chat.Workspace[]> => {
  try {
    const response = await request.get('/workspaces')
    return response.data
  }
  catch (error) {
    console.error('Error fetching workspaces:', error)
    throw error
  }
}

// Get a specific workspace by UUID
export const getWorkspace = async (uuid: string): Promise<Chat.Workspace> => {
  try {
    const response = await request.get(`/workspaces/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(`Error fetching workspace ${uuid}:`, error)
    throw error
  }
}

// Create a new workspace
export const createWorkspace = async (data: CreateWorkspaceRequest): Promise<Chat.Workspace> => {
  try {
    const response = await request.post('/workspaces', data)
    return response.data
  }
  catch (error) {
    console.error('Error creating workspace:', error)
    throw error
  }
}

// Update an existing workspace
export const updateWorkspace = async (uuid: string, data: UpdateWorkspaceRequest): Promise<Chat.Workspace> => {
  try {
    const response = await request.put(`/workspaces/${uuid}`, data)
    return response.data
  }
  catch (error) {
    console.error(`Error updating workspace ${uuid}:`, error)
    throw error
  }
}

// Delete a workspace
export const deleteWorkspace = async (uuid: string): Promise<void> => {
  try {
    await request.delete(`/workspaces/${uuid}`)
  }
  catch (error) {
    console.error(`Error deleting workspace ${uuid}:`, error)
    throw error
  }
}

// Update workspace order
export const updateWorkspaceOrder = async (uuid: string, orderPosition: number): Promise<Chat.Workspace> => {
  try {
    const response = await request.put(`/workspaces/${uuid}/reorder`, { orderPosition })
    return response.data
  }
  catch (error) {
    console.error(`Error updating workspace order ${uuid}:`, error)
    throw error
  }
}

// Set workspace as default
export const setDefaultWorkspace = async (uuid: string): Promise<Chat.Workspace> => {
  try {
    const response = await request.put(`/workspaces/${uuid}/set-default`)
    return response.data
  }
  catch (error) {
    console.error(`Error setting default workspace ${uuid}:`, error)
    throw error
  }
}

// Ensure user has a default workspace
export const ensureDefaultWorkspace = async (): Promise<Chat.Workspace> => {
  try {
    const response = await request.post('/workspaces/default')
    return response.data
  }
  catch (error) {
    console.error('Error ensuring default workspace:', error)
    throw error
  }
}

// Create a session in a specific workspace
export const createSessionInWorkspace = async (workspaceUuid: string, data: CreateSessionInWorkspaceRequest) => {
  try {
    const response = await request.post(`/workspaces/${workspaceUuid}/sessions`, data)
    return response.data
  }
  catch (error) {
    console.error(`Error creating session in workspace ${workspaceUuid}:`, error)
    throw error
  }
}

// Get all sessions in a workspace
export const getSessionsByWorkspace = async (workspaceUuid: string) => {
  try {
    const response = await request.get(`/workspaces/${workspaceUuid}/sessions`)
    return response.data
  }
  catch (error) {
    console.error(`Error fetching sessions for workspace ${workspaceUuid}:`, error)
    throw error
  }
}