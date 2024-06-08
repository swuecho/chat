<template>
        <div>
                <NModal :show="showUploadModal">
                        <NCard style="width: 600px" title="Upload" :bordered="false" size="huge" role="dialog"
                                aria-modal="true">
                                <template #header-extra>
                                        upload doc, image, audio, video (gemmi flash only)
                                </template>
                                <Uploader :sessionUuid="sessionUuid" :showUploaderButton="true"></Uploader>
                                <template #footer>
                                        <NButton @click="$emit('update:showUploadModal', false)">Cancel</NButton>
                                </template>
                        </NCard>
                </NModal>

        </div>
</template>

<script lang="ts" setup>
import { NModal, NCard, NButton } from 'naive-ui';
import Uploader from './Uploader.vue';

// const props = defineProps({
//   showUploadModal: {
//     type: Boolean,
//     required: true
//   },
//   sessionUuid: {
//     type: String,
//     required: true
//   }
// })

interface Props {
        showUploadModal: boolean,
        sessionUuid: string
}

defineProps<Props>()

</script>