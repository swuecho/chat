import request from '@/utils/request/axios'

export const createChatSnapshot = async (uuid: string): Promise<any> => {
  try {
    const response = await request.post(`/uuid/chat_snapshot/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}


export const createChatBot = async (uuid: string): Promise<any> => {
  try {
    const response = await request.post(`/uuid/chat_bot/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const fetchChatSnapshot = async (uuid: string): Promise<any> => {
  try {
    const response = await request.get(`/uuid/chat_snapshot/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const fetchSnapshotAll = async (page: number = 1, pageSize: number = 20): Promise<any> => {
  try {
    const response = await request.get(`/uuid/chat_snapshot/all?type=snapshot&page=${page}&page_size=${pageSize}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const fetchSnapshotAllData = async (page: number = 1, pageSize: number = 20): Promise<Snapshot.Snapshot[]> => {
  try {
    const response = await fetchSnapshotAll(page, pageSize)
    // Handle response format: { data: [...], total: n } or just the array
    return Array.isArray(response) ? response : (response.data ?? [])
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const fetchChatbotAll= async (): Promise<any> => {
  try {
    const response = await request.get('/uuid/chat_snapshot/all?type=chatbot')
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const fetchChatbotAllData = async (): Promise<Snapshot.Snapshot[]> => {
  try {
    const response = await fetchChatbotAll()
    // Handle response format: { data: [...], total: n } or just the array
    return Array.isArray(response) ? response : (response.data ?? [])
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const chatSnapshotSearch = async (search: string): Promise<any> => {
  try {
    const response = await request.get(`/uuid/chat_snapshot_search?search=${search}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const updateChatSnapshot = async (uuid: string, data: any): Promise<any> => {
  try {
    const response = await request.put(`/uuid/chat_snapshot/${uuid}`, data)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const fetchSnapshotDelete = async (uuid: string): Promise<any> => {
  try {
    const response = await request.delete(`/uuid/chat_snapshot/${uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
// CreateSessionFromSnapshot
export const CreateSessionFromSnapshot = async (snapshot_uuid: string) => {
  try {
    const response = await request.post(`/uuid/chat_session_from_snapshot/${snapshot_uuid}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
