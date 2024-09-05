<script setup lang='ts'>
import { computed, ref } from 'vue'
import { NModal, NInput, NCard, NButton } from 'naive-ui'
import AudioPlayer from '../AudioPlayer/index.vue'
import TextComponent from '@/views/components/Message/Text.vue'
import AvatarComponent from '@/views/components/Avatar/MessageAvatar.vue'
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
}

interface Emit {
  (ev: 'regenerate'): void
  (ev: 'delete'): void
  (ev: 'togglePin'): void
  (ev: 'afterEdit', index: number, text: string): void
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
</script>

<template>
  <div class="chat-message">
    <p class="text-xs text-[#b4bbc4] text-center">{{ displayLocaleDate(dateTime) }}</p>
    <div class="flex w-full mb-6 overflow-hidden" :class="[{ 'flex-row-reverse': inversion }]">
      <div class="flex items-center justify-center flex-shrink-0 h-8 overflow-hidden rounded-full basis-8"
        :class="[inversion ? 'ml-2' : 'mr-2']">
        <AvatarComponent :inversion="inversion" :model="model" />
      </div>
      <div class="overflow-hidden text-sm " :class="[inversion ? 'items-end' : 'items-start']">
        <p :class="[inversion ? 'text-right' : 'text-left']">
          {{ !inversion ? model : userInfo.name || $t('setting.defaultName') }}
        </p>
        <div class="flex items-end gap-1 mt-2" :class="[inversion ? 'flex-row-reverse' : 'flex-row']">
          <TextComponent ref="textRef" class="message-text" :inversion="inversion" :error="error" :text="text"
            :code="code" :loading="loading" :idex="index" />
          <div class="flex flex-col">
            <!-- testid="chat-message-regenerate" not ok, someting like testclass -->
            <button v-if="!inversion"
              class="chat-message-regenerate mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
              @click="handleRegenerate">
              <SvgIcon icon="ri:restart-line" />
            </button>
            <button v-if="!isPrompt && inversion"
              class="mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
              :disabled="pining" @click="emit('togglePin')">
              <SvgIcon :icon="isPin ? 'ri:unpin-line' : 'ri:pushpin-line'" />
            </button>


          </div>
        </div>
        <div class="flex" :class="[inversion ? 'justify-end' : 'justify-start']">
          <div class="flex items-center">
            <!--
            <AudioPlayer :text="text || ''" :right="inversion" class="mr-2" />
          -->
            <HoverButton :tooltip="$t('common.edit')"
              class="transition text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-300" @click="handleEdit">
              <SvgIcon icon="ri:edit-line" />
            </HoverButton>
            <HoverButton :tooltip="$t('chat.copy')"
              class="transition text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-300" @click="handleCopy">
              <SvgIcon icon="ri:file-copy-2-line" />
            </HoverButton>
            <HoverButton :tooltip="$t('common.delete')"
              class="transition text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-300" @click="handleDelete">
              <SvgIcon icon="ri:delete-bin-line" />
            </HoverButton>
          </div>
        </div>

      </div>
    </div>
  </div>

  <!-- Updated modal for editing -->
  <NModal v-model:show="showEditModal" :mask-closable="false"  style="width: 90%; max-width: 800px;">
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
