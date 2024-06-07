<template>
        <div>
                <NModal v-model:show="props.showUploadModal">
                        <NCard style="width: 600px" title="Upload" :bordered="false" size="huge" role="dialog"
                                aria-modal="true">
                                <template #header-extra>
                                        upload doc or image (txt, png, excel or code file)
                                </template>
                                <NUpload multiline action="/api/upload" :headers="headers" :data="data"
                                        :show-download-button="true" @finish="handleFinish" @before-upload="beforeUpload"
                                        @remove="handleRemove" @download="handleDownload">
                                        <NButton id="attach_file_button" data-testid="attach_file_button"
                                                type="primary"> Upload
                                        </NButton>
                                </NUpload>
                                <template #footer>
                                        <NButton @click="$emit('update:showUploadModal', false)">Cancel</NButton>
                                </template>
                        </NCard>
                </NModal>

        </div>
</template>

<script lang="ts" setup>
import { NModal, NCard, NUpload, NButton } from 'naive-ui';
import { ref } from 'vue';
import { useAuthStore } from '@/store'
import { useRoute } from 'vue-router'
import request from '@/utils/request/axios'

const route = useRoute()

const { uuid: sessionUuid } = route.params as { uuid: string }

const props = defineProps(['showUploadModal'])
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
// @ts-ignore
function handleFinish({ file, event }) {
        file.url = JSON.parse(event.currentTarget.response)['url']
        //fileList.value.push(file)
        console.log(file, event)
}

// @ts-ignore
function handleRemove({ file, fileList }) {
        console.log('remove', file)
        // delete file at url
        request.delete(file.url)
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