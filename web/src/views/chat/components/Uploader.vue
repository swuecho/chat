<template>
        <div>
                <NUpload multiline action="/api/upload" :headers="headers" :data="data"
                        :default-file-list="fileListData" :show-download-button="true" @finish="handleFinish"
                        @before-upload="beforeUpload" @remove="handleRemove" @download="handleDownload">
                        <NButton v-if="showUploaderButton" id="attach_file_button" data-testid="attach_file_button"
                                type="primary"> Upload
                        </NButton>
                </NUpload>
        </div>
</template>

<script setup lang="ts">
import { NUpload, NButton, UploadFileInfo } from 'naive-ui';
import { ref } from 'vue';
import { useAuthStore } from '@/store'
import request from '@/utils/request/axios'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { getChatFilesList } from '@/api/chat_file'

const queryClient = useQueryClient()

interface Props {
        sessionUuid: string
        showUploaderButton: boolean
}

const props = defineProps<Props>()

const sessionUuid = props.sessionUuid

// sessionUuid not null.
const { data: fileListData } = useQuery({
        queryKey: ['fileList', sessionUuid],
        queryFn: async () => await getChatFilesList(sessionUuid)
})

const fileDeleteMutation = useMutation({
        mutationFn: async (url: string) => {
                request.delete(url)
        },
        onSuccess: () => {
                queryClient.invalidateQueries({ queryKey: ['fileList', sessionUuid] })
        },
})




// const emit = defineEmits(['update:showUploadModal']);

// login modal will appear when there is no token
const authStore = useAuthStore()

const token = authStore.getToken()

const headers = ref({
        'Authorization': 'Bearer ' + token
})

const data = ref({
        'session-uuid': sessionUuid
})


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
        if (!event) {
                return
        }
        // @ts-ignore
        file.url = JSON.parse(event.currentTarget.response)['url']
        //fileList.value.push(file)
        console.log(file, event)
        queryClient.invalidateQueries({ queryKey: ['fileList', sessionUuid] })
        return file

}

// @ts-ignore
function handleRemove({ file }: { file: UploadFileInfo }) {
        console.log('remove', file)
        if (file.url) {
                fileDeleteMutation.mutate(file.url)
        }
        console.log(file.url)
}

// @ts-ignore
async function handleDownload(file) {
        console.log('download', file)
        let response = await request.get(file.url, {
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