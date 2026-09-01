<script lang="ts" setup>
import { computed, watch } from 'vue'
import { NButton, NCard, NModal, NSpace, NSpin, useMessage } from 'naive-ui'
import { useQuery } from '@tanstack/vue-query'
import { getAdminSessionMessages } from '@/api/generated_client'
import type { GetAdminSessionMessagesResponse } from '@/api/generated_client'
import { HoverButton, SvgIcon } from '@/components/common'
import TextComponent from '@/views/components/Message/Text.vue'
import AvatarComponent from '@/views/components/Avatar/MessageAvatar.vue'
import { t } from '@/locales'
import { displayLocaleDate } from '@/utils/date'
import { copyText } from '@/utils/format'

interface Props {
  visible: boolean
  sessionId: string
  sessionModel: string
  userEmail: string
}

type ChatMessage = GetAdminSessionMessagesResponse[number]

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const message = useMessage()

const show = computed({
  get: () => props.visible,
  set: (visible: boolean) => emit('update:visible', visible),
})

// Fetch session messages when modal opens
const { data: messages, isLoading, error } = useQuery({
  queryKey: ['sessionMessages', props.sessionId],
  queryFn: async () => {
    if (!props.sessionId)
      return []
    return await getAdminSessionMessages({ path: { sessionUuid: props.sessionId } })
  },
  enabled: computed(() => props.visible && !!props.sessionId),
})

// Format messages for display
const formattedMessages = computed(() => {
  if (!messages.value)
    return []

  return messages.value.map((msg: ChatMessage, index: number) => ({
    uuid: msg.uuid,
    index,
    dateTime: msg.createdAt,
    model: msg.model || props.sessionModel,
    text: msg.content,
    inversion: msg.role === 'user' || (msg.role === 'system' && index === 0),
    error: false,
    loading: false,
    tokenCount: msg.tokenCount,
  }))
})

// Copy message text
function handleCopy(text: string) {
  copyText({ text })
  message.success(t('chat.copySuccess'))
}

// Scroll to top
function scrollToTop() {
  const container = document.querySelector('.session-snapshot-content')
  if (container)
    container.scrollTo({ top: 0, behavior: 'smooth' })
}

// Calculate total tokens
const totalTokens = computed(() => {
  return formattedMessages.value.reduce((sum: number, msg: any) => sum + (msg.tokenCount || 0), 0)
})

// Format date safely
function formatMessageDate(dateString: string) {
  try {
    if (!dateString)
      return ''
    // Handle different date formats from backend
    const date = new Date(dateString)
    if (isNaN(date.getTime())) {
      // If invalid date, return raw string
      return dateString
    }
    return displayLocaleDate(date.toISOString())
  }
  catch (error) {
    console.warn('Invalid date format:', dateString)
    return dateString
  }
}

// Watch for errors
watch(error, (newError) => {
  if (newError)
    message.error(t('common.fetchFailed'))
})
</script>

<template>
  <NModal v-model:show="show" :style="{ width: ['100vw', '90vw', '1200px'] }">
    <NCard
      role="dialog"
      aria-modal="true"
      :title="`${t('admin.sessionSnapshot')} - ${sessionId.slice(0, 8)}...`"
      :bordered="false"
      size="huge"
      class="session-snapshot-modal"
    >
      <template #header-extra>
        <NSpace>
          <span class="text-sm text-gray-500">
            {{ t('admin.model') }}: {{ sessionModel }}
          </span>
          <span v-if="!isLoading" class="text-sm text-gray-500">
            {{ t('admin.totalTokens') }}: {{ totalTokens }}
          </span>
        </NSpace>
      </template>

      <NSpin :show="isLoading">
        <div class="session-snapshot-content" style="max-height: 70vh; overflow-y: auto;">
          <div v-if="formattedMessages.length === 0 && !isLoading" class="text-center py-8 text-gray-500">
            {{ t('common.noData') }}
          </div>

          <div v-else class="space-y-4">
            <div
              v-for="chatMessage in formattedMessages"
              :key="chatMessage.uuid"
              class="flex w-full"
              :class="[chatMessage.inversion ? 'flex-row-reverse' : 'flex-row']"
            >
              <div
                class="flex items-start space-x-2"
                :class="[
                  chatMessage.inversion ? 'flex-row-reverse space-x-reverse ml-4' : 'mr-4',
                ]"
              >
                <!-- Avatar -->
                <div class="flex-shrink-0">
                  <AvatarComponent :image="chatMessage.inversion" />
                </div>

                <!-- Message Content -->
                <div
                  class="max-w-[calc(100%-3rem)] rounded-lg px-4 py-3 relative group"
                  :class="[
                    chatMessage.inversion
                      ? 'text-white'
                      : 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100',
                  ]"
                >
                  <!-- Copy Button -->
                  <div
                    class="absolute -top-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200"
                    :class="[chatMessage.inversion ? '-left-2' : '-right-2']"
                  >
                    <HoverButton size="small" @click="handleCopy(chatMessage.text)">
                      <SvgIcon icon="ri:file-copy-2-line" />
                    </HoverButton>
                  </div>

                  <!-- Message Text -->
                  <div class="message-content">
                    <TextComponent
                      ref="textRef"
                      :inversion="chatMessage.inversion"
                      :error="chatMessage.error"
                      :text="chatMessage.text"
                      :loading="chatMessage.loading"
                      :model="chatMessage.model"
                      :as-raw-text="false"
                    />
                  </div>

                  <!-- Message Info -->
                  <div
                    class="text-xs opacity-70 mt-2 flex items-center gap-2"
                    :class="[chatMessage.inversion ? 'text-blue-100' : 'text-gray-500']"
                  >
                    <span>{{ formatMessageDate(chatMessage.dateTime) }}</span>
                    <span v-if="chatMessage.tokenCount">
                      • {{ chatMessage.tokenCount }} {{ t('admin.tokens') }}
                    </span>
                    <span v-if="!chatMessage.inversion && chatMessage.model">
                      • {{ chatMessage.model }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </NSpin>

      <!-- Footer with actions -->
      <template #footer>
        <div class="flex justify-between items-center">
          <NSpace>
            <NButton size="small" secondary @click="scrollToTop">
              <template #icon>
                <SvgIcon icon="ri:arrow-up-line" />
              </template>
              {{ t('chat.backToTop') }}
            </NButton>
          </NSpace>

          <NSpace>
            <span class="text-sm text-gray-500">
              {{ t('admin.userEmail') }}: {{ userEmail }}
            </span>
          </NSpace>
        </div>
      </template>
    </NCard>
  </NModal>
</template>

<style scoped>
.session-snapshot-modal :deep(.n-card-header) {
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.session-snapshot-content {
  padding: 16px 0;
}

.message-content :deep(.text-component) {
  word-break: break-word;
}

/* Scrollbar styling */
.session-snapshot-content::-webkit-scrollbar {
  width: 6px;
}

.session-snapshot-content::-webkit-scrollbar-track {
  background: var(--scrollbar-color);
  border-radius: 3px;
}

.session-snapshot-content::-webkit-scrollbar-thumb {
  background: var(--scrollbar-hover-color);
  border-radius: 3px;
}

.session-snapshot-content::-webkit-scrollbar-thumb:hover {
  background: var(--primary-color);
}
</style>
