import {
  createChatBot as createChatBotRequest,
  createChatSessionFromSnapshot,
  createChatSnapshot as createChatSnapshotRequest,
  deleteChatSnapshot,
  getChatSnapshot,
  listChatSnapshots,
  searchChatSnapshots,
  updateChatBotModel as updateChatBotModelRequest,
  updateChatBotSettings as updateChatBotSettingsRequest,
  updateChatSnapshot as updateChatSnapshotRequest,
} from '@/api/generated_client'
import type { UpdateChatSnapshotData } from '@/api/generated/types.gen'

export const createChatSnapshot = (uuid: string) =>
  createChatSnapshotRequest({ path: { uuid } })

export const createChatBot = (uuid: string) =>
  createChatBotRequest({ path: { uuid } })

export const updateChatBotModel = (uuid: string, model: string) =>
  updateChatBotModelRequest({ path: { uuid }, body: { model } })

export interface UpdateChatBotSettingsRequest {
  title: string
  summary: string
  model: string
}

export const updateChatBotSettings = (uuid: string, settings: UpdateChatBotSettingsRequest) =>
  updateChatBotSettingsRequest({ path: { uuid }, body: settings })

export const fetchChatSnapshot = (uuid: string) =>
  getChatSnapshot({ path: { uuid } })

export const fetchSnapshotAll = (page = 1, pageSize = 20) =>
  listChatSnapshots({
    query: { type: 'snapshot', limit: pageSize, offset: (page - 1) * pageSize },
  })

export const fetchSnapshotAllData = async (page = 1, pageSize = 20): Promise<Snapshot.Snapshot[]> => {
  const response = await fetchSnapshotAll(page, pageSize)
  return response.items as Snapshot.Snapshot[]
}

export const fetchChatbotAll = async (): Promise<Snapshot.Snapshot[]> => {
  const response = await listChatSnapshots({ query: { type: 'chatbot' } })
  return response.items as Snapshot.Snapshot[]
}

export const fetchChatbotAllData = fetchChatbotAll

export const chatSnapshotSearch = (search: string) =>
  searchChatSnapshots({ query: { search } })

export const updateChatSnapshot = (uuid: string, data: UpdateChatSnapshotData['body']) =>
  updateChatSnapshotRequest({ path: { uuid }, body: data })

export const fetchSnapshotDelete = (uuid: string) =>
  deleteChatSnapshot({ path: { uuid } })

// Keep the legacy export name while using the generated operation underneath.
export const CreateSessionFromSnapshot = async (snapshotUUID: string) => {
  const response = await createChatSessionFromSnapshot({ path: { uuid: snapshotUUID } })
  return { ...response, SessionUuid: response.sessionUuid }
}
