<script lang='ts' setup>
import { computed, h, onMounted, ref, watch } from 'vue'
import copy from 'copy-to-clipboard'
import jwt_decode from 'jwt-decode'
import { useRoute } from 'vue-router'
import { NSelect, NSpin, NTabPane, NTabs, useDialog, useMessage } from 'naive-ui'
import { useQuery } from '@tanstack/vue-query'
import Header from '../snapshot/components/Header/index.vue'
import Message from './components/Message/index.vue'
import AnswerHistory from './components/AnswerHistory.vue'
import { useCopyCode } from '@/hooks/useCopyCode'
import { CreateSessionFromSnapshot, fetchChatSnapshot, updateChatBotModel } from '@/api/chat_snapshot'
import { fetchChatModel } from '@/api/chat_model'
import type { ChatModel } from '@/types/chat-models'
import { HoverButton, SvgIcon } from '@/components/common'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'
import { getCurrentDate } from '@/utils/date'
import { useAuthStore, useSessionStore } from '@/store'
import { generateAPIHelper } from '@/service/snapshot'
import { fetchAPIToken } from '@/api/token'

const authStore = useAuthStore()
const sessionStore = useSessionStore()

const route = useRoute()
const dialog = useDialog()
const nui_msg = useMessage()

useCopyCode()

const { isMobile } = useBasicLayout()
// session uuid
const { uuid } = route.params as { uuid: string }

const { data: snapshot_data, isLoading } = useQuery({
  queryKey: ['chatSnapshot', uuid],
  queryFn: async () => await fetchChatSnapshot(uuid),
})

const { data: chatModels, isLoading: modelsLoading } = useQuery<ChatModel[]>({
  queryKey: ['chat_models'],
  queryFn: fetchChatModel,
  enabled: computed(() => Boolean(authStore.getToken)),
})

const currentUserId = computed(() => {
  if (!authStore.getToken)
    return undefined
  try {
    const claims = jwt_decode<{ user_id?: string }>(authStore.getToken)
    return claims.user_id ? Number(claims.user_id) : undefined
  }
  catch {
    return undefined
  }
})
const canEditModel = computed(() =>
  currentUserId.value !== undefined && currentUserId.value === snapshot_data.value?.userId,
)
const modelOptions = computed(() =>
  (chatModels.value ?? [])
    .filter(model => model.isEnable)
    .sort((a, b) => (a.orderNumber || 0) - (b.orderNumber || 0))
    .map(model => ({ label: model.label, value: model.name })),
)
const selectedModel = ref<string>()
const modelUpdating = ref(false)

watch(() => snapshot_data.value?.model, (model) => {
  selectedModel.value = model
}, { immediate: true })

async function handleModelUpdate(model: string) {
  const previousModel = snapshot_data.value?.model
  if (!model || model === previousModel)
    return

  modelUpdating.value = true
  try {
    const updated = await updateChatBotModel(uuid, model)
    snapshot_data.value = updated
    nui_msg.success(t('bot.modelUpdateSuccess'))
  }
  catch (error) {
    selectedModel.value = previousModel
    nui_msg.error(t('bot.modelUpdateFailed'))
  }
  finally {
    modelUpdating.value = false
  }
}

const activeTab = ref('conversation')

const apiToken = ref('')

onMounted(async () => {
  const data = await fetchAPIToken()
  apiToken.value = data.accessToken
})

function format_chat_md(chat: Chat.Message): string {
  return `<sup><kbd><var>${chat.dateTime}</var></kbd></sup>:\n ${chat.text}`
}

const chatToMarkdown = () => {
  try {
    /*
    uuid: string,
    dateTime: string
    text: string
    inversion?: boolean
    error?: boolean
    loading?: boolean
    isPrompt?: boolean
    */
    const chatData = snapshot_data.value.conversation
    const markdown = chatData.map((chat: Chat.Message) => {
      if (chat.isPrompt)
        return `**system** ${format_chat_md(chat)}}`
      else if (chat.inversion)
        return `**user** ${format_chat_md(chat)}`
      else
        return `**assistant** ${format_chat_md(chat)}`
    }).join('\n\n----\n\n')
    return markdown
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

function handleMarkdown() {
  const dialogBox = dialog.warning({
    title: t('chat.exportMD'),
    content: t('chat.exportMDConfirm'),
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: async () => {
      try {
        dialogBox.loading = true
        const markdown = chatToMarkdown()
        const ts = getCurrentDate()
        const filename = `chat-${ts}.md`
        const blob = new Blob([markdown], { type: 'text/plain;charset=utf-8' })
        const url: string = URL.createObjectURL(blob)
        const link: HTMLAnchorElement = document.createElement('a')
        link.href = url
        link.download = filename
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        dialogBox.loading = false
        nui_msg.success(t('chat.exportSuccess'))
        Promise.resolve()
      }
      catch (error: any) {
        nui_msg.error(t('chat.exportFailed'))
      }
      finally {
        dialogBox.loading = false
      }
    },
  })
}

async function handleChat() {
  if (!authStore.getToken)
    nui_msg.error(t('common.ask_user_register'))
  window.open('/', '_blank')
  const { SessionUuid }: { SessionUuid: string } = await CreateSessionFromSnapshot(uuid)
  const session = sessionStore.getChatSessionByUuid(SessionUuid)
  if (session)
    sessionStore.setActiveSessionWithoutNavigation(session.workspaceUuid, SessionUuid)
}

const footerClass = computed(() => {
  let classes = ['p-4']
  if (isMobile.value)
    classes = ['sticky', 'left-0', 'bottom-0', 'right-0', 'p-2', 'pr-3', 'overflow-hidden']
  return classes
})

function handleShowCode() {
  const postUuid = route.path.split('/')[2]
  const code = generateAPIHelper(postUuid, apiToken.value, window.location.origin)
  const dialogBox = dialog.info({
    title: t('bot.showCode'),
    content: () => h('code', { class: 'whitespace-pre-wrap' }, code),
    positiveText: t('common.copy'),
    onPositiveClick: () => {
      const success = copy(code)
      if (success)
        nui_msg.success(t('common.success'))
      else
        nui_msg.error(t('common.copyFailed'))

      dialogBox.loading = false
    },
  })
}

const scrollRef = ref<HTMLElement | null>(null)
const showScrollToTop = ref(false)

function handleScroll() {
  if (scrollRef.value)
    showScrollToTop.value = scrollRef.value.scrollTop > 100
}

function onScrollToTop() {
  if (scrollRef.value) {
    scrollRef.value.scrollTo({
      top: 0,
      behavior: 'smooth',
    })
    // Force scroll in case smooth scrolling is blocked by browser
    scrollRef.value.scrollTop = 0
  }
}
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <div v-if="isLoading">
      <NSpin size="large" />
    </div>
    <div v-else>
      <Header :title="snapshot_data.title" typ="chatbot" />
      <main class="flex-1 overflow-hidden">
        <div id="scrollRef" ref="scrollRef" class="h-[calc(100vh-6rem)] overflow-y-auto" @scroll="handleScroll">
          <div
            id="image-wrapper" class="w-full max-w-screen-xl m-auto dark:bg-[#101014]"
            :class="[isMobile ? 'p-2' : 'p-4']"
          >
            <div class="flex items-center justify-center mt-4 ">
              <div class="w-4/5 md:w-1/3 mb-3">
                <NSelect
                  v-if="canEditModel"
                  v-model:value="selectedModel"
                  :options="modelOptions"
                  :loading="modelsLoading || modelUpdating"
                  :disabled="modelsLoading || modelUpdating"
                  :placeholder="$t('bot.selectModel')"
                  :fallback-option="false"
                  @update:value="handleModelUpdate"
                />
                <div v-else class="px-3 py-2 border rounded border-gray-200 dark:border-gray-700">
                  {{ snapshot_data.model }}
                </div>
              </div>
            </div>

            <NTabs v-model:value="activeTab" type="line">
              <NTabPane name="conversation" :tab="t('bot.tabs.conversation')">
                <Message
                  v-for="(item, index) of snapshot_data.conversation" :key="index" :date-time="item.dateTime"
                  :model="snapshot_data.model" :text="item.text" :inversion="item.inversion" :error="item.error"
                  :loading="item.loading" :index="index"
                />
                <footer :class="footerClass">
                  <div class="w-full max-w-screen-xl m-auto">
                    <div class="flex items-center justify-between space-x-2">
                      <HoverButton :tooltip="$t('chat_snapshot.showCode')" @click="handleShowCode">
                        <span class="text-xl text-[#4f555e] dark:text-white">
                          <SvgIcon icon="ic:outline-code" />
                        </span>
                      </HoverButton>
                      <HoverButton
                        v-if="!isMobile" :tooltip="$t('chat_snapshot.exportMarkdown')"
                        @click="handleMarkdown"
                      >
                        <span class="text-xl text-[#4f555e] dark:text-white">
                          <SvgIcon icon="mdi:language-markdown" />
                        </span>
                      </HoverButton>
                      <HoverButton :tooltip="$t('chat_snapshot.scrollTop')" @click="onScrollToTop">
                        <span class="text-xl text-[#4f555e] dark:text-white">
                          <SvgIcon icon="material-symbols:vertical-align-top" />
                        </span>
                      </HoverButton>
                    </div>
                  </div>
                </footer>
              </NTabPane>

              <NTabPane name="history" :tab="t('bot.tabs.history')">
                <AnswerHistory :bot-uuid="uuid" />
              </NTabPane>
            </NTabs>
          </div>
        </div>
      </main>
      <div class="floating-button">
        <HoverButton testid="create-chat" :tooltip="$t('chat_snapshot.createChat')" @click="handleChat">
          <span class="text-xl text-[#4f555e] dark:text-white m-auto mx-10">
            <SvgIcon icon="mdi:chat-plus" width="32" height="32" />
          </span>
        </HoverButton>
      </div>
    </div>
  </div>
</template>
