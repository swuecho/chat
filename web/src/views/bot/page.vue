<script lang='ts' setup>
import { computed, h, onMounted, ref } from 'vue'
import copy from 'copy-to-clipboard'
import jwt_decode from 'jwt-decode'
import { useRoute } from 'vue-router'
import {
  NAlert,
  NButton,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSpin,
  NTabPane,
  NTabs,
  useDialog,
  useMessage,
} from 'naive-ui'
import { useQuery } from '@tanstack/vue-query'
import Header from '../snapshot/components/Header/index.vue'
import Message from './components/Message/index.vue'
import AnswerHistory from './components/AnswerHistory.vue'
import { useCopyCode } from '@/hooks/useCopyCode'
import { CreateSessionFromSnapshot } from '@/api/chat_snapshot'
import type { ChatModel } from '@/types/chat-models'
import { HoverButton, SvgIcon } from '@/components/common'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'
import { getCurrentDate } from '@/utils/date'
import { useAuthStore, useSessionStore } from '@/store'
import { generateAPIHelper } from '@/service/snapshot'
import { createLongLivedToken, getChatSnapshot, listChatModels, updateChatBotSettings } from '@/api/generated_client'

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
  queryFn: () => getChatSnapshot({ path: { uuid } }),
})

const { data: chatModels, isLoading: modelsLoading } = useQuery<ChatModel[]>({
  queryKey: ['chat_models'],
  queryFn: () => listChatModels(),
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
const canEditBot = computed(() =>
  currentUserId.value !== undefined && currentUserId.value === snapshot_data.value?.userId,
)
const enabledModels = computed(() =>
  (chatModels.value ?? [])
    .filter(model => model.isEnable)
    .sort((a, b) => (a.orderNumber || 0) - (b.orderNumber || 0)),
)
const modelOptions = computed(() =>
  enabledModels.value
    .map(model => ({ label: model.label, value: model.name })),
)
const currentModel = computed(() =>
  (chatModels.value ?? []).find(model => model.name === snapshot_data.value?.model),
)
const modelIsOutdated = computed(() =>
  !modelsLoading.value && Boolean(snapshot_data.value?.model)
  && (!currentModel.value || !currentModel.value.isEnable),
)
const settingsVisible = ref(false)
const settingsSaving = ref(false)
const settingsForm = ref({
  title: '',
  summary: '',
  model: '',
})

function openSettings() {
  const recommendedModel = enabledModels.value.find(model => model.isDefault) ?? enabledModels.value[0]
  settingsForm.value = {
    title: snapshot_data.value?.title ?? '',
    summary: snapshot_data.value?.summary ?? '',
    model: modelIsOutdated.value ? (recommendedModel?.name ?? '') : (snapshot_data.value?.model ?? ''),
  }
  settingsVisible.value = true
}

async function saveSettings() {
  if (!settingsForm.value.title.trim() || !settingsForm.value.model) {
    nui_msg.error(t('bot.settingsRequired'))
    return
  }

  settingsSaving.value = true
  try {
    const updated = await updateChatBotSettings({ path: { uuid }, body: {
      title: settingsForm.value.title.trim(),
      summary: settingsForm.value.summary.trim(),
      model: settingsForm.value.model,
    } })
    snapshot_data.value = updated
    settingsVisible.value = false
    nui_msg.success(t('bot.settingsUpdateSuccess'))
  }
  catch (error) {
    nui_msg.error(t('bot.settingsUpdateFailed'))
  }
  finally {
    settingsSaving.value = false
  }
}

const activeTab = ref('conversation')

const apiToken = ref('')

onMounted(async () => {
  const data = await createLongLivedToken()
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
            <div class="w-4/5 md:w-2/3 mx-auto mt-4 mb-3 space-y-3">
              <NAlert v-if="modelIsOutdated" type="warning" :title="$t('bot.outdatedModelTitle')">
                {{ $t('bot.outdatedModelDescription', { model: snapshot_data.model }) }}
                <template v-if="canEditBot" #action>
                  <NButton size="small" type="warning" @click="openSettings">
                    {{ $t('bot.chooseReplacement') }}
                  </NButton>
                </template>
              </NAlert>
              <div class="flex items-center justify-between gap-3 px-3 py-2 border rounded border-gray-200 dark:border-gray-700">
                <div class="min-w-0">
                  <div class="text-xs text-gray-500 dark:text-gray-400">
                    {{ $t('bot.currentModel') }}
                  </div>
                  <div class="font-medium truncate">
                    {{ currentModel?.label || snapshot_data.model }}
                  </div>
                </div>
                <NButton v-if="canEditBot" secondary @click="openSettings">
                  <template #icon>
                    <SvgIcon icon="material-symbols:settings-outline" />
                  </template>
                  {{ $t('bot.settings') }}
                </NButton>
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
      <NModal
        v-model:show="settingsVisible"
        preset="card"
        class="w-[min(92vw,36rem)]"
        :title="$t('bot.settings')"
        :bordered="false"
      >
        <NForm :model="settingsForm" label-placement="top">
          <NFormItem :label="$t('bot.title')" required>
            <NInput v-model:value="settingsForm.title" maxlength="200" show-count />
          </NFormItem>
          <NFormItem :label="$t('bot.description')">
            <NInput
              v-model:value="settingsForm.summary"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              maxlength="1000"
              show-count
            />
          </NFormItem>
          <NFormItem :label="$t('bot.model')" required>
            <NSelect
              v-model:value="settingsForm.model"
              :options="modelOptions"
              :loading="modelsLoading"
              :disabled="modelsLoading"
              :placeholder="$t('bot.selectModel')"
              :fallback-option="false"
            />
          </NFormItem>
          <NAlert v-if="modelIsOutdated" type="warning" class="mb-4">
            {{ $t('bot.replacementSelected', { model: snapshot_data.model }) }}
          </NAlert>
          <div class="flex justify-end gap-2">
            <NButton :disabled="settingsSaving" @click="settingsVisible = false">
              {{ $t('common.cancel') }}
            </NButton>
            <NButton type="primary" :loading="settingsSaving" @click="saveSettings">
              {{ $t('common.save') }}
            </NButton>
          </div>
        </NForm>
      </NModal>
    </div>
  </div>
</template>
