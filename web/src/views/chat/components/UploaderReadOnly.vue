<script setup lang="ts">
import type { UploadCustomRequestOptions, UploadFileInfo } from 'naive-ui'
import { NUpload } from 'naive-ui'
import { useChatFiles } from '../composables/useChatFiles'

const props = defineProps<Props>()

interface Props {
  sessionUuid: string
  showUploaderButton: boolean
}

const sessionUuid = props.sessionUuid
const { deleteFile, downloadFile, fileListData, uploadFile } = useChatFiles(sessionUuid)

// const emit = defineEmits(['update:showUploadModal']);

async function customRequest({ file, onFinish, onError }: UploadCustomRequestOptions) {
  if (!file.file) {
    onError()
    return
  }
  try {
    const response = await uploadFile(file.file)
    file.url = response.url
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
  deleteFile(file)
}

async function handlePreview(file: UploadFileInfo, detail: { event: MouseEvent }) {
  detail.event.preventDefault()
  await handleDownload(file)
}

async function handleDownload(file: UploadFileInfo) {
  await downloadFile(file)
  return false // !!! cancel original download
}
</script>

<template>
  <div v-if="fileListData && fileListData.length">
    <NUpload
      class="w-full max-w-screen-xl m-auto px-4" :custom-request="customRequest"
      :file-list="fileListData" :show-download-button="true" :show-remove-button="false"
      :show-cancel-button="false" @before-upload="beforeUpload"
      @remove="handleRemove" @download="handleDownload" @update:file-list="handleFileListUpdate"
      @preview="handlePreview"
    />
  </div>
</template>
