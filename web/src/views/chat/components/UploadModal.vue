<script lang="ts" setup>
import { NButton, NCard, NModal } from 'naive-ui'
import Uploader from './Uploader.vue'

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
  showUploadModal: boolean
  sessionUuid: string
}

defineProps<Props>()
defineEmits<{
  'update:showUploadModal': [value: boolean]
}>()
</script>

<template>
  <div>
    <NModal :show="showUploadModal">
      <NCard
        :style="{ width: 'min(100vw, 600px)' }" :title="$t('chat.uploader_title')"
        :bordered="false" size="huge" role="dialog" aria-modal="true"
      >
        <template #header-extra>
          <span class="hidden sm:inline">{{ $t('chat.uploader_help_text') }}</span>
        </template>
        <Uploader :session-uuid="sessionUuid" :show-uploader-button="true" />
        <template #footer>
          <NButton @click="$emit('update:showUploadModal', false)">
            {{
              $t('chat.uploader_close') }}
          </NButton>
        </template>
      </NCard>
    </NModal>
  </div>
</template>
