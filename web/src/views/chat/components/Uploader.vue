<script setup lang="ts">
import type { UploadCustomRequestOptions, UploadFileInfo } from 'naive-ui'
import { NButton, NUpload } from 'naive-ui'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { deleteChatFileByID, downloadChatFileByID, getChatFilesList, uploadChatFileForSession } from '@/api/chat_file'

const props = defineProps<Props>()

const queryClient = useQueryClient()

interface Props {
  sessionUuid: string
  showUploaderButton: boolean
}

const sessionUuid = props.sessionUuid

// sessionUuid not null.
const { data: fileListData } = useQuery({
  queryKey: ['fileList', sessionUuid],
  queryFn: async () => await getChatFilesList(sessionUuid),
})

const fileDeleteMutation = useMutation({
  mutationFn: async (id: number) => {
    await deleteChatFileByID(id)
  },
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['fileList', sessionUuid] })
  },
})

// const emit = defineEmits(['update:showUploadModal']);

async function customRequest({ file, onFinish, onError }: UploadCustomRequestOptions) {
  if (!file.file) {
    onError()
    return
  }
  try {
    const response = await uploadChatFileForSession(sessionUuid, file.file)
    file.url = response.url
    queryClient.invalidateQueries({ queryKey: ['fileList', sessionUuid] })
    onFinish()
  }
  catch {
    onError()
  }
}

const handleFileListUpdate = (_fileList: UploadFileInfo[]) => {}

function beforeUpload(_data: any) {
  // You can return a Promise to reject the file
  // return Promise.reject(new Error('Invalid file type'))
}
function handleRemove({ file }: { file: UploadFileInfo }) {
  if (file.url)
    fileDeleteMutation.mutate(fileID(file))
}

function fileID(file: UploadFileInfo): number {
  return Number(file.url?.split('/').pop())
}

function handlePreview(file: UploadFileInfo, detail: { event: MouseEvent }) {
  detail.event.preventDefault()
  handleDownload(file)
}

async function handleDownload(file: UploadFileInfo) {
  // get last part of file.url
  const response = await downloadChatFileByID(fileID(file))
  // Create a new Blob object using the response data of the file
  const blob = new Blob([response], { type: 'application/octet-stream' })

  // Create a link element
  const link = document.createElement('a')

  // Set the href property of the link to a URL created from the Blob
  link.href = window.URL.createObjectURL(blob)

  // Set the download attribute of the link to the desired file name
  link.download = file.name

  // Append the link to the body
  document.body.appendChild(link)

  // Programmatically click the link to trigger the download
  link.click()

  // Remove the link from the document
  document.body.removeChild(link)
  return false // !!! cancel original download
}
</script>

<template>
  <div>
    <NUpload
      multiline :custom-request="customRequest" :default-file-list="fileListData"
      :show-download-button="true" @before-upload="beforeUpload"
      @preview="handlePreview" @remove="handleRemove" @download="handleDownload"
      @update:file-list="handleFileListUpdate"
    >
      <NButton
        v-if="showUploaderButton" id="attach_file_button" data-testid="attach_file_button"
        type="primary"
      >
        {{ $t('chat.uploader_button') }}
      </NButton>
    </NUpload>
  </div>
</template>
