import {
  createChatSessionFromSnapshot,
  listChatSnapshots,
} from '@/api/generated_client'

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

// Keep the legacy export name while using the generated operation underneath.
export const CreateSessionFromSnapshot = async (snapshotUUID: string) => {
  const response = await createChatSessionFromSnapshot({ path: { uuid: snapshotUUID } })
  return { ...response, SessionUuid: response.sessionUuid }
}
