<template>
        <div v-if="fileListData && fileListData.length">
                <NUpload class="w-full max-w-screen-xl m-auto px-4" :action="actionURL" :headers="headers" :data="data"
                        :file-list="fileListData" :show-download-button="true" :show-remove-button="false"
                        :show-cancel-button="false" @finish="handleFinish" @before-upload="beforeUpload"
                        @remove="handleRemove" @download="handleDownload" @update:file-list="handleFileListUpdate"
                        @preview="handlePreview">
                </NUpload>
        </div>
</template>

<script setup lang="ts">
import { NUpload, UploadFileInfo } from 'naive-ui';
import { computed, ref } from 'vue';
import { useAuthStore } from '@/store'
import request from '@/utils/request/axios'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getChatFilesList } from '@/api/chat_file'

const queryClient = useQueryClient()

const baseURL = "/api"

const actionURL = baseURL + '/upload'

interface Props {
        sessionUuid: string
        showUploaderButton: boolean
}

const props = defineProps<Props>()

const sessionUuid = props.sessionUuid

const fileListQueryKey = computed(() => ['fileList', sessionUuid]);

// sessionUuid not null.
const { data: fileListData } = useQuery({
        queryKey: fileListQueryKey,
        queryFn: async () => await getChatFilesList(sessionUuid)
})

const fileDeleteMutation = useMutation({
        mutationFn: async (url: string) => {
                await request.delete(url)
        },
        onSuccess: () => {
                queryClient.invalidateQueries({ queryKey: ['fileList', sessionUuid] })
        },
})


// const emit = defineEmits(['update:showUploadModal']);

// login modal will appear when there is no token
const authStore = useAuthStore()

const token = authStore.getToken

const headers = ref({
        'Authorization': 'Bearer ' + token
})

const data = ref({
        'session-uuid': sessionUuid
})

const handleFileListUpdate = (fileList: UploadFileInfo[]) => {
        console.log(fileList)
}

function beforeUpload(data: any) {
        console.log(data.file)
        // You can return a Promise to reject the file
        // return Promise.reject(new Error('Invalid file type'))
}
/**
 * Handles the completion of a file upload.
 *
 * @param {object} options - An object containing the file and the event.
 * @param {File} options.file - The uploaded file.
 * @param {Event} options.event - The upload event.
 * @returns {void}
 */
function handleFinish({ file, event }: { file: UploadFileInfo, event?: ProgressEvent }): UploadFileInfo | undefined {
        console.log(file, event)
        if (!event) {
                return
        }
        // Type assertion for ProgressEvent target
        const target = event.target as XMLHttpRequest
        if (target?.response) {
                const response = JSON.parse(target.response)
                file.url = response.url
        }
        //fileList.value.push(file)
        console.log(file, event)
        queryClient.invalidateQueries({ queryKey: ['fileList', sessionUuid] })
        return file

}

function fileUrl(file: UploadFileInfo): string {
        const file_id = file.url?.split('/').pop();
        const url = `/download/${file_id}`
        return url
}

function handleRemove({ file }: { file: UploadFileInfo }) {
        console.log('remove', file)
        if (file.url) {
                const url = fileUrl(file)
                fileDeleteMutation.mutate(url)
        }
        console.log(file.url)
}

async function handlePreview(file: UploadFileInfo, detail: { event: MouseEvent }) {
        detail.event.preventDefault()
        await handleDownload(file)
}


async function handleDownload(file: UploadFileInfo) {
        console.log('download', file)
        const url = fileUrl(file)
        let response = await request.get(url, {
                responseType: 'blob', // Important: set the response type to blob
        })
        // Create a new Blob object using the response data of the file
        const blob = new Blob([response.data], { type: 'application/octet-stream' });

        // Create a link element
        const link = document.createElement('a');

        // Set the href property of the link to a URL created from the Blob
        link.href = window.URL.createObjectURL(blob);

        // Set the download attribute of the link to the desired file name
        link.download = file.name;

        // Append the link to the body
        document.body.appendChild(link);

        // Programmatically click the link to trigger the download
        link.click();

        // Remove the link from the document
        document.body.removeChild(link);
        return false //!!! cancel original download
}

</script>