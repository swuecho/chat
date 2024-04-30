<script setup lang='ts'>
import { computed, nextTick, ref } from 'vue'
import { NDropdown } from 'naive-ui'
import AudioPlayer from '../AudioPlayer/index.vue'
import TextComponent from '@/views/components/Message/Text.vue'
import AvatarComponent from '@/views/components/Avatar/MessageAvatar.vue'
import { SvgIcon } from '@/components/common'
import { copyText } from '@/utils/format'
import { useIconRender } from '@/hooks/useIconRender'
import { useUserStore } from '@/store'
import { t } from '@/locales'
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

const { iconRender } = useIconRender()

const textRef = ref()

const editable = ref(false)

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)

const options = [
  {
    label: t('common.edit'),
    key: 'editText',
    icon: iconRender({ icon: 'ri:edit-line' }),
  },
  {
    label: t('chat.copy'),
    key: 'copyText',
    icon: iconRender({ icon: 'ri:file-copy-2-line' }),
  },
  {
    label: t('common.delete'),
    key: 'delete',
    icon: iconRender({ icon: 'ri:delete-bin-line' }),
  },
]

function onContentChange(event: FocusEvent, index: number) {
  editable.value = false
  const text = (event.target as HTMLElement).innerText
  emit('afterEdit', index, text)
}

async function handleSelect(key: 'copyText' | 'delete' | 'editText') {
  switch (key) {
    case 'copyText':
      copyText({ text: props.text ?? '' })
      return
    case 'delete':
      emit('delete')
      return
    case 'editText':
      // make the text editable
      editable.value = true
      await nextTick()
      textRef.value.textRef.focus()
      break
  }
}

const code = computed(() => {
  return props?.model?.includes('davinci') ?? false
})

function handleRegenerate() {
  emit('regenerate')
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
          :code="code" :loading="loading" :contenteditable="editable" :idex="index"
          @blur="event => onContentChange(event, index)" />
        <div class="flex flex-col">
          <!-- testid="chat-message-regenerate" not ok, someting like testclass -->
          <button v-if="!inversion"
            class="chat-message-regenerate mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
            @click="handleRegenerate">
            <SvgIcon icon="ri:restart-line" />
          </button>
          <button v-if="!isPrompt && inversion"
            class="mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300" :disabled="pining"
            @click="emit('togglePin')">
            <SvgIcon :icon="isPin ? 'ri:unpin-line' : 'ri:pushpin-line'" />
          </button>
          <NDropdown :placement="!inversion ? 'right' : 'left'" :options="options" @select="handleSelect">
            <button class="transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-200">
              <SvgIcon icon="ri:more-2-fill" />
            </button>
          </NDropdown>
        </div>
      </div>
      <AudioPlayer :text="text || ''" :right="inversion"></AudioPlayer>
    </div>
  </div>
  </div>
</template>
