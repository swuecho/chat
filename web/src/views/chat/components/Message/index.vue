<script setup lang='ts'>
import { computed, ref } from 'vue'
import { NModal, NInput, NCard, NButton } from 'naive-ui'
import TextComponent from '@/views/components/Message/Text.vue'
import AvatarComponent from '@/views/components/Avatar/MessageAvatar.vue'
import ArtifactViewer from './ArtifactViewer.vue'
import SuggestedQuestions from './SuggestedQuestions.vue'
import { HoverButton, SvgIcon } from '@/components/common'
import { copyText } from '@/utils/format'
import { useUserStore } from '@/store'
import { displayLocaleDate } from '@/utils/date'

interface Props {
  index: number
  dateTime: string
  text?: string
  inversion?: boolean
  error?: boolean
  loading?: boolean
  model?: string
  isPrompt?: boolean
  isPin?: boolean
  pining?: boolean
  artifacts?: Chat.Artifact[]
  suggestedQuestions?: string[]
  suggestedQuestionsLoading?: boolean
  exploreMode?: boolean
}

interface Emit {
  (ev: 'regenerate'): void
  (ev: 'delete'): void
  (ev: 'togglePin'): void
  (ev: 'afterEdit', index: number, text: string): void
  (ev: 'useQuestion', question: string): void
}

const props = defineProps<Props>()

const emit = defineEmits<Emit>()

const textRef = ref()


const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)

const showEditModal = ref(false)
const editedText = ref('')

const code = computed(() => {
  return props?.model?.includes('davinci') ?? false
})

function handleRegenerate() {
  emit('regenerate')
}

function handleCopy() {
  copyText({ text: props.text ?? '' })
}

function handleEdit() {
  editedText.value = props.text || ''
  showEditModal.value = true
}

function handleEditConfirm() {
  emit('afterEdit', props.index, editedText.value)
  showEditModal.value = false
}

function handleDelete() {
  emit('delete')
}

function handleUseQuestion(question: string) {
  emit('useQuestion', question)
}
</script>

<template>
  <div class="chat-message">
    <p class="text-xs text-[#b4bbc4] text-center">{{ displayLocaleDate(dateTime) }}</p>
    <div class="flex w-full mb-6" :class="[{ 'flex-row-reverse': inversion }]">
      <div class="flex items-center justify-center flex-shrink-0 h-8 overflow-hidden rounded-full basis-8"
        :class="[inversion ? 'ml-2' : 'mr-2']">
        <AvatarComponent :inversion="inversion" :model="model" />
      </div>
      <div class="text-sm min-w-0 flex-1" :class="[inversion ? 'items-end' : 'items-start']">
        <p :class="[inversion ? 'text-right' : 'text-left']">
          {{ !inversion ? model : userInfo.name || $t('setting.defaultName') }}
        </p>
        <div class="flex items-end gap-1 mt-2" :class="[inversion ? 'flex-row-reverse' : 'flex-row']">
          <div class="flex flex-col min-w-0">
            <TextComponent ref="textRef" class="message-text" :inversion="inversion" :error="error" :text="text"
              :code="code" :loading="loading" :idex="index" />
            <ArtifactViewer v-if="artifacts && artifacts.length > 0" :artifacts="artifacts" :inversion="inversion"
              data-testid="artifact-viewer" />

          </div>
          <div class="flex flex-col">

            <button v-if="!isPrompt && inversion"
              class="mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
              :disabled="pining" @click="emit('togglePin')">
              <SvgIcon :icon="isPin ? 'ri:unpin-line' : 'ri:pushpin-line'" />
            </button>
            <!-- testid="chat-message-regenerate" not ok, something like testclass -->
            <button v-if="!isPrompt"
              class="chat-message-regenerate mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
              @click="handleRegenerate">
              <SvgIcon icon="ri:restart-line" />
            </button>

          </div>

        </div>
        <div class="flex" :class="[inversion ? 'justify-end' : 'justify-start']">
          <div class="flex items-center">
            <!--
            <AudioPlayer :text="text || ''" :right="inversion" class="mr-2" />
          -->
            <HoverButton :tooltip="$t('common.delete')"
              class="transition text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-300"
              @click="handleDelete">
              <SvgIcon icon="ri:delete-bin-line" />
            </HoverButton>
            <HoverButton :tooltip="$t('common.edit')"
              class="transition text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-300"
              @click="handleEdit">
              <SvgIcon icon="ri:edit-line" />
            </HoverButton>
            <HoverButton :tooltip="$t('chat.copy')"
              class="transition text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-300"
              @click="handleCopy">
              <SvgIcon icon="ri:file-copy-2-line" />
            </HoverButton>

          </div>
        </div>
        <SuggestedQuestions v-if="!inversion && exploreMode && !loading && (suggestedQuestionsLoading || (suggestedQuestions && suggestedQuestions.length > 0))" :questions="suggestedQuestions || []"
          :loading="suggestedQuestionsLoading && (!suggestedQuestions || suggestedQuestions.length === 0)" @useQuestion="handleUseQuestion" />
      </div>
    </div>
  </div>

  <!-- Updated modal for editing -->
  <NModal v-model:show="showEditModal" :mask-closable="false" style="width: 90%; max-width: 800px;">
    <NCard :bordered="false" size="medium" role="dialog" aria-modal="true" :title="$t('common.edit')">

      <NInput v-model:value="editedText" type="textarea" :autosize="{ minRows: 10, maxRows: 20 }" :autofocus="true" />

      <template #footer>
        <div class="flex justify-end space-x-2">
          <NButton type="default" @click="showEditModal = false">
            {{ $t('common.cancel') }}
          </NButton>
          <NButton type="primary" @click="handleEditConfirm">
            {{ $t('common.confirm') }}
          </NButton>
        </div>
      </template>
    </NCard>
  </NModal>
</template>

<style scoped>
.chat-message {
  /* Ensure proper responsive behavior */
  max-width: 100%;
  overflow-x: hidden;
}

/* Mobile responsive improvements */
@media (max-width: 639px) {
  .chat-message {
    /* Better mobile layout */
    word-wrap: break-word;
    overflow-wrap: break-word;
  }

  .message-text {
    /* Ensure text content doesn't break layout */
    max-width: 100%;
    overflow-wrap: break-word;
    word-wrap: break-word;
  }
}
</style>
