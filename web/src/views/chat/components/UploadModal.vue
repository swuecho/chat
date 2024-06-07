<template>
        <NModal v-model:show="props.showUploadModal">
                <NCard style="width: 600px" title="Upload" :bordered="false" size="huge" role="dialog"
                        aria-modal="true">
                        <template #header-extra>
                                upload doc or image (txt, png, excel or code file)
                        </template>
                        <NUpload multiple>
                                <NButton id="attach_file_button" data-testid="attach_file_button" type="primary"> Upload
                                </NButton>
                        </NUpload>
                        <template #footer>
                                <NButton @click="$emit('update:showUploadModal', false)">Cancel</NButton>
                        </template>
                </NCard>
        </NModal>
</template>

<script lang="ts" setup>
import { NModal, NCard, NUpload, NButton } from 'naive-ui';
import { ref } from 'vue';

const props = defineProps(['showUploadModal'])
const emit = defineEmits(['update:showUploadModal']);

</script>