<template>
        <div>
                <NModal v-model:show="props.showUploadModal">
                        <NCard style="width: 600px" title="Upload" :bordered="false" size="huge" role="dialog"
                                aria-modal="true">
                                <template #header-extra>
                                        upload doc or image (txt, png, excel or code file)
                                </template>
                                <NUpload multiline action="/api/upload" :headers="{
                                        'naive-info': 'hello!'
                                }" :data="{
        'naive-data': 'cool! naive!'
}" @finish="handleFinish" @before-upload="beforeUpload">
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


const props = defineProps(['showUploadModal'])
const emit = defineEmits(['update:showUploadModal']);

function beforeUpload(data) {
        console.log(data.file)
        // You can return a Promise to reject the file
        // return Promise.reject(new Error('Invalid file type'))
}
function handleFinish({ file, event }) {
        console.log(file, event)
}

</script>