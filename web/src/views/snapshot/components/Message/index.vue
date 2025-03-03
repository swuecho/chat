<script setup lang='ts'>
import { computed, ref } from 'vue'
import { NDropdown, NInput, NModal, useMessage } from 'naive-ui'
import { createChatComment } from '@/api/comment'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import TextComponent from '@/views//components/Message/Text.vue'
import AvatarComponent from '@/views/components/Avatar/MessageAvatar.vue'
import { SvgIcon } from '@/components/common'
import { copyText } from '@/utils/format'
import { useIconRender } from '@/hooks/useIconRender'
import { t } from '@/locales'
import { displayLocaleDate } from '@/utils/date'
import { useUserStore } from '@/store'

interface Props {
  sessionUuid: string
  uuid: string
  index: number
  dateTime: string
  model: string
  text: string
  inversion?: boolean
  error?: boolean
  loading?: boolean
  comments?: Array<any>
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

const queryClient = useQueryClient()

const options = [
  {
    label: t('chat.copy'),
    key: 'copyText',
    icon: iconRender({ icon: 'ri:file-copy-2-line' }),
  },
]

const mutation = useMutation({
  mutationFn: () => createChatComment(props.sessionUuid, props.uuid, commentContent.value),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['conversationComments', props.sessionUuid] })
  },
})


async function handleComment() {
  console.log('commenting')
  try {
    isCommenting.value = true


    await mutation.mutateAsync()
    nui_msg.success(t('chat.commentSuccess'))
    showCommentModal.value = false
    commentContent.value = ''
  } catch (error) {
    console.log(error)
    console.log('failed')
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

const formatCommentDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

// fiter comments with uuid using computed
const filterComments = computed(() => {
  if (!props.comments)
    return []
  return props.comments
    .filter((comment: any) => comment.chatMessageUuid === props.uuid)
    .sort((a: any, b: any) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())
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
          -->
            <button class="mb-2 transition text-neutral-300 hover:text-neutral-800 dark:hover:text-neutral-300"
              @click="showCommentModal = true">
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
  <!-- Comments section -->

  <div v-if="filterComments && filterComments.length > 0" class="mt-4" :class="[inversion ? 'pr-12' : 'pl-12']">
    <div v-for="comment in filterComments" :key="comment.uuid"
      class="comment-item mb-3 p-2 bg-gray-100 dark:bg-gray-700 rounded-lg w-1/2" 
      :class="[inversion ? 'ml-auto' : 'mr-auto']">
      <div class="text-xs text-gray-600 dark:text-gray-300">
        <span class="font-medium">{{ comment.authorUsername }}</span>
        <span class="mx-1">•</span>
        <span>{{ formatCommentDate(comment.createdAt) }}</span>
      </div>
      <div class="text-sm mt-1 text-gray-800 dark:text-gray-100">
        {{ comment.content }}
      </div>
    </div>
  </div>
  <NModal v-model:show="showCommentModal" :mask-closable="false">
    <div class="p-6 bg-white dark:bg-[#1a1a1a] rounded-lg w-[90vw] max-w-[500px]">
      <h3 class="mb-4 text-lg font-medium">{{ $t('chat.addComment') }}</h3>
      <NInput v-model:value="commentContent" type="textarea" :placeholder="$t('chat.commentPlaceholder')"
        :autosize="{ minRows: 3, maxRows: 6 }" />
      <div class="flex justify-end gap-2 mt-4">
        <button class="px-4 py-2 text-sm rounded hover:bg-gray-100 dark:hover:bg-gray-700"
          @click="showCommentModal = false">
          {{ $t('common.cancel') }}
        </button>
        <button class="px-4 py-2 text-sm text-white bg-blue-600 rounded hover:bg-blue-700"
          :disabled="!commentContent || isCommenting" @click="handleComment">
          {{ isCommenting ? $t('common.submitting') : $t('common.submit') }}
        </button>
      </div>
    </div>
  </NModal>
</template>
