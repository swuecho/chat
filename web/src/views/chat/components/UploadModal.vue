<template>
        <div>
                <NModal :show="showUploadModal">
                        <!--he { width: ['100vw', '600px'] } syntax is using Naive UI's responsive array notation for styles. Here's what it means:                                                              

 • ['100vw', '600px'] is an array where:                                                                                                                                              
    • The first value 100vw applies to the smallest screen size (mobile)                                                                                                              
    • The second value 600px applies to screens at the sm breakpoint and larger 
     -->
                        <NCard :style="{ width: ['100vw', '600px'] }" :title="$t('chat.uploader_title')"
                                :bordered="false" size="huge" role="dialog" aria-modal="true">
                                <template #header-extra>
                                        <span class="hidden sm:inline">{{ $t('chat.uploader_help_text') }}</span>
                                </template>
                                <Uploader :sessionUuid="sessionUuid" :showUploaderButton="true"></Uploader>
                                <template #footer>
                                        <NButton @click="$emit('update:showUploadModal', false)">{{
                                                $t('chat.uploader_close') }}</NButton>
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
