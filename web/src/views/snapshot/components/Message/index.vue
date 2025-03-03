<script setup lang='ts'>
import { computed, ref } from 'vue'
import { NDropdown, NInput, NModal } from 'naive-ui'
import { createChatComment } from '@/api/comment'
import TextComponent from '@/views//components/Message/Text.vue'
import AvatarComponent from '@/views/components/Avatar/MessageAvatar.vue'
import { SvgIcon } from '@/components/common'
import { copyText } from '@/utils/format'
import { useIconRender } from '@/hooks/useIconRender'
import { t } from '@/locales'
import { displayLocaleDate } from '@/utils/date'
import { useUserStore } from '@/store'

interface Props {
  uuid: string
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

const showCommentModal = ref(false)
const commentContent = ref('')
const isCommenting = ref(false)
const nui_msg = useMessage()

const options = [
  {
    label: t('chat.copy'),
    key: 'copyText',
    icon: iconRender({ icon: 'ri:file-copy-2-line' }),
  },
]

async function handleComment() {
  try {
    isCommenting.value = true
    await createChatComment(snapshot_data.value.uuid, props.uuid, commentContent.value)
    nui_msg.success(t('chat.commentSuccess'))
    showCommentModal.value = false
    commentContent.value = ''
  } catch (error) {
    nui_msg.error(t('chat.commentFailed'))
  } finally {
    isCommenting.value = false
  }
}

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
    <div
      class="flex items-center justify-center flex-shrink-0 h-8 overflow-hidden rounded-full basis-8"
      :class="[inversion ? 'ml-2' : 'mr-2']"
    >
      <div
        class="flex items-center justify-center flex-shrink-0 h-8 overflow-hidden rounded-full basis-8"
        :class="[inversion ? 'ml-2' : 'mr-2']"
      >
        <AvatarComponent :inversion="inversion" :model="model" />
      </div>
    </div>
    <div class="overflow-hidden text-sm " :class="[inversion ? 'items-end' : 'items-start']">
      <p class="text-xs text-[#b4bbc4]" :class="[inversion ? 'text-right' : 'text-left']">
        {{ !inversion ? model : userInfo.name || $t('setting.defaultName') }}
      </p>
      <div class="flex items-end gap-1 mt-2" :class="[inversion ? 'flex-row-reverse' : 'flex-row']">
        <TextComponent
          ref="textRef" class="message-text" :inversion="inversion" :error="error" :text="text" :code="code"
          :loading="loading" :idex="index"
        />
        <div class="flex flex-col">
          <!-- 
          <button
            v-if="!inversion"
            class="mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
          >
            <SvgIcon icon="mingcute:voice-fill" />
          </button>
          -->
          <button
            class="mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
            @click="showCommentModal = true"
          >
            <SvgIcon icon="mdi:comment-outline" />
          </button>
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

  <NModal v-model:show="showCommentModal" :mask-closable="false">
    <div class="p-6 bg-white dark:bg-[#1a1a1a] rounded-lg w-[90vw] max-w-[500px]">
      <h3 class="mb-4 text-lg font-medium">{{ $t('chat.addComment') }}</h3>
      <NInput
        v-model:value="commentContent"
        type="textarea"
        :placeholder="$t('chat.commentPlaceholder')"
        :autosize="{ minRows: 3, maxRows: 6 }"
      />
      <div class="flex justify-end gap-2 mt-4">
        <button
          class="px-4 py-2 text-sm rounded hover:bg-gray-100 dark:hover:bg-gray-700"
          @click="showCommentModal = false"
        >
          {{ $t('common.cancel') }}
        </button>
        <button
          class="px-4 py-2 text-sm text-white bg-blue-600 rounded hover:bg-blue-700"
          :disabled="!commentContent || isCommenting"
          @click="handleComment"
        >
          {{ isCommenting ? $t('common.submitting') : $t('common.submit') }}
        </button>
      </div>
    </div>
  </NModal>
</template>
  chat: {
    addComment: 'Add Comment',
    commentPlaceholder: 'Enter your comment...',
    commentSuccess: 'Comment added successfully',
    commentFailed: 'Failed to add comment',
  },
