<script lang='ts' setup>
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useDialog, useMessage, NSpin } from 'naive-ui'
import html2canvas from 'html2canvas'
import Message from './components/Message/index.vue'
import { useCopyCode } from '@/hooks/useCopyCode'
import Header from './components/Header/index.vue'
import { CreateSessionFromSnapshot, fetchChatSnapshot } from '@/api/chat_snapshot'
import { HoverButton, SvgIcon } from '@/components/common'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'
import { genTempDownloadLink } from '@/utils/download'
import { getCurrentDate } from '@/utils/date'
import { useAuthStore, useChatStore } from '@/store'
import { useQuery } from '@tanstack/vue-query'
import { getConversationComments } from '@/api/comment'

const authStore = useAuthStore()
const chatStore = useChatStore()

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

const { data: comments } = useQuery({
  queryKey: ['conversationComments', uuid],
  queryFn: async () => await getConversationComments(uuid),
})

function handleExport() {

  const dialogBox = dialog.warning({
    title: t('chat.exportImage'),
    content: t('chat.exportImageConfirm'),
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: async () => {
      try {
        dialogBox.loading = true
        const ele = document.getElementById('image-wrapper')
        const canvas = await html2canvas(ele as HTMLDivElement, {
          useCORS: true,
        })
        const imgUrl = canvas.toDataURL('image/png')
        const tempLink = genTempDownloadLink(imgUrl)
        document.body.appendChild(tempLink)
        tempLink.click()
        document.body.removeChild(tempLink)
        window.URL.revokeObjectURL(imgUrl)
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
    const chatData = snapshot_data.value.conversation;
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
  if (!authStore.getToken())
    nui_msg.error(t('common.ask_user_register'))
  window.open(`/`, '_blank')
  const { SessionUuid }: { SessionUuid: string } = await CreateSessionFromSnapshot(uuid)
  await chatStore.setActiveLocal(SessionUuid)
}

const footerClass = computed(() => {
  let classes = ['p-4']
  if (isMobile.value)
    classes = ['sticky', 'left-0', 'bottom-0', 'right-0', 'p-2', 'pr-3', 'overflow-hidden']
  return classes
})

const scrollRef = ref<HTMLElement | null>(null)

function onScrollToTop() {
  const container = scrollRef.value
  if (!container) return

  console.log('Current scroll position:', container.scrollTop)

  // Try both methods for maximum compatibility
  container.scrollTo({ top: 0, behavior: 'smooth' })
  container.scrollTop = 0

  // Add a small timeout to check if it worked
  setTimeout(() => {
    console.log('New scroll position:', container.scrollTop)
  }, 500)
}
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <div v-if="isLoading">
      <NSpin size="large" />
    </div>
    <div v-else>
      <Header :title="snapshot_data.title" typ="snapshot" />
      <main class="flex-1 overflow-hidden">
        <div ref="scrollRef" class="h-full overflow-y-auto"
          style="height: calc(100vh - 100px); scroll-behavior: smooth;">
          <div id="image-wrapper" class="w-full max-w-screen-xl m-auto dark:bg-[#101014]"
            :class="[isMobile ? 'p-2' : 'p-4']">
            <Message v-for="(item, index) of snapshot_data.conversation" :key="index" :date-time="item.dateTime"
              :model="item?.model || snapshot_data.model" :text="item.text" :inversion="item.inversion"
              :error="item.error" :loading="item.loading" :index="index" :uuid="item.uuid" :session-uuid="uuid"
              :comments="comments" />
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
      <footer :class="footerClass">
        <div class="w-full max-w-screen-xl m-auto">
          <div class="flex items-center justify-between space-x-2">
            <HoverButton v-if="!isMobile" :tooltip="$t('chat_snapshot.exportImage')" @click="handleExport">
              <span class="text-xl text-[#4f555e] dark:text-white">
                <SvgIcon icon="ri:download-2-line" />
              </span>
            </HoverButton>
            <HoverButton v-if="!isMobile" :tooltip="$t('chat_snapshot.exportMarkdown')" @click="handleMarkdown">
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
    </div>
  </div>
</template>
