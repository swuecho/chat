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

export const fetchSnapshotAll = async (): Promise<any> => {
  try {
    const response = await request.get('/uuid/chat_snapshot/all')
    return response.data
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
