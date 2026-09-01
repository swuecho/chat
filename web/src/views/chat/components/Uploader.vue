<script setup lang="ts">
import type { UploadCustomRequestOptions, UploadFileInfo } from 'naive-ui'
import { NButton, NUpload } from 'naive-ui'
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

function handlePreview(file: UploadFileInfo, detail: { event: MouseEvent }) {
  detail.event.preventDefault()
  handleDownload(file)
}

async function handleDownload(file: UploadFileInfo) {
  await downloadFile(file)
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
