import type { UploadFileInfo } from 'naive-ui'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { deleteChatFile, downloadChatFile, listChatFiles, uploadChatFile } from '@/api/generated_client'

function fileID(file: UploadFileInfo): number {
  return Number(file.url?.split('/').pop())
}

export function useChatFiles(sessionUuid: string) {
  const queryClient = useQueryClient()
  const queryKey = ['fileList', sessionUuid] as const

  const { data: fileListData } = useQuery({
    queryKey,
    queryFn: async (): Promise<UploadFileInfo[]> => {
      const files = await listChatFiles({ path: { uuid: sessionUuid } })
      return files.map(file => ({
        ...file,
        status: 'finished',
        url: `/api/download/${file.id}`,
        percentage: 100,
      }))
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteChatFile({ path: { id } }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey }),
  })

  async function uploadFile(file: File) {
    const response = await uploadChatFile({
      body: { 'session-uuid': sessionUuid, file },
    })
    await queryClient.invalidateQueries({ queryKey })
    return response
  }

  async function downloadFile(file: UploadFileInfo) {
    const response = await downloadChatFile({ path: { id: fileID(file) } })
    const url = window.URL.createObjectURL(response)
    const link = document.createElement('a')
    link.href = url
    link.download = file.name
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  }

  function deleteFile(file: UploadFileInfo) {
    if (file.url)
      deleteMutation.mutate(fileID(file))
  }

  return { deleteFile, downloadFile, fileListData, uploadFile }
}
