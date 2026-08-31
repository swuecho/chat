import { deleteChatFile, downloadChatFile, listChatFiles, uploadChatFile } from '@/api/generated_client'

// /chat_file/{uuid}/list

const baseURL = '/api'

export async function getChatFilesList(uuid: string) {
  const files = await listChatFiles({ path: { uuid } })
  return files.map(item => ({
    ...item,
    status: 'finished',
    url: `${baseURL}/download/${item.id}`,
    percentage: 100,
  }))
}

export const uploadChatFileForSession = (sessionUUID: string, file: File) =>
  uploadChatFile({ body: { 'session-uuid': sessionUUID, file } })

export const deleteChatFileByID = (id: number) =>
  deleteChatFile({ path: { id } })

export const downloadChatFileByID = (id: number) =>
  downloadChatFile({ path: { id } })
