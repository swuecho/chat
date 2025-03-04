<script setup lang='ts'>
import { computed, ref } from 'vue'
import { NDropdown } from 'naive-ui'
import TextComponent from '@/views//components/Message/Text.vue'
import AvatarComponent from '@/views/components/Avatar/MessageAvatar.vue'
import { SvgIcon } from '@/components/common'
import { copyText } from '@/utils/format'
import { useIconRender } from '@/hooks/useIconRender'
import { t } from '@/locales'
import { displayLocaleDate } from '@/utils/date'
import { useUserStore } from '@/store'

interface Props {
  index: number
  dateTime: string
  model: string
  text: string
  inversion?: boolean
  error?: boolean
  loading?: boolean
}

const props = defineProps<Props>()

const { iconRender } = useIconRender()

const userStore = useUserStore()

const userInfo = computed(() => userStore.userInfo)

const textRef = ref<HTMLElement>()

const options = [
  {
    label: t('chat.copy'),
    key: 'copyText',
    icon: iconRender({ icon: 'ri:file-copy-2-line' }),
  },
]


function handleSelect(key: 'copyText') {
  switch (key) {
    case 'copyText':
      copyText({ text: props.text ?? '' })
  }
}

const code = computed(() => {
  return props?.model?.includes('davinci') ?? false
})
</script>

<template>
  <div class="chat-message">
    <p class="text-xs text-[#b4bbc4] text-center">{{ displayLocaleDate(dateTime) }}</p>
    <div class="flex w-full mb-6 overflow-hidden" :class="[{ 'flex-row-reverse': inversion }]">
      <div class="flex items-center justify-center flex-shrink-0 h-8 overflow-hidden rounded-full basis-8"
        :class="[inversion ? 'ml-2' : 'mr-2']">
        <div class="flex items-center justify-center flex-shrink-0 h-8 overflow-hidden rounded-full basis-8"
          :class="[inversion ? 'ml-2' : 'mr-2']">
          <AvatarComponent :inversion="inversion" :model="model" />
        </div>
      </div>
      <div class="overflow-hidden text-sm " :class="[inversion ? 'items-end' : 'items-start']">
        <p class="text-xs text-[#b4bbc4]" :class="[inversion ? 'text-right' : 'text-left']">
          {{ !inversion ? model : userInfo.name || $t('setting.defaultName') }}
        </p>
        <div class="flex items-end gap-1 mt-2" :class="[inversion ? 'flex-row-reverse' : 'flex-row']">
          <TextComponent ref="textRef" class="message-text" :inversion="inversion" :error="error" :text="text"
            :code="code" :loading="loading" :idex="index" />
          <div class="flex flex-col">
            <!-- 
          <button
            v-if="!inversion"
            class="mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
          >
            <SvgIcon icon="mingcute:voice-fill" />
          </button>
            <button class="mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
              <SvgIcon icon="mdi:comment-outline" />
            </button>
          -->
            <NDropdown :placement="!inversion ? 'right' : 'left'" :options="options" @select="handleSelect">
              <button class="transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-200">
                <SvgIcon icon="ri:more-2-fill" />
              </button>
            </NDropdown>
          </div>
        </div>
      </div>
    </div>
  </div>

</template>