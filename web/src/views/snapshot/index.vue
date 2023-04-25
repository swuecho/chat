<script lang='ts' setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useDialog, useMessage } from 'naive-ui'
import html2canvas from 'html2canvas'
import Message from './components/Message/index.vue'
import { useCopyCode } from './hooks/useCopyCode'
import Header from './components/Header/index.vue'
import { CreateSessionFromSnapshot, fetchChatSnapshot } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'
import { genTempDownloadLink } from '@/utils/download'
import { getCurrentDate } from '@/utils/date'
import { useAuthStore } from '@/store'

const authStore = useAuthStore()

const route = useRoute()
const dialog = useDialog()
const nui_msg = useMessage()

useCopyCode()

const { isMobile } = useBasicLayout()
// session uuid
const { uuid } = route.params as { uuid: string }

const dataSources = ref<Chat.Chat[]>([])
const title = ref<string>('')
const model = ref<string>('')

onMounted(async () => {
  const snapshot = await fetchChatSnapshot(uuid)
  dataSources.value.push(...snapshot.Conversation)
  title.value = snapshot.Title
  model.value = snapshot.Model
})

const loading = ref<boolean>(false)

function handleExport() {
  if (loading.value)
    return

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
function format_chat_md(chat: Chat.Chat): string {
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
    const chatData = dataSources.value
    const markdown = chatData.map((chat: Chat.Chat) => {
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
  if (loading.value)
    return

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

  const { SessionUuid }: { SessionUuid: string } = await CreateSessionFromSnapshot(uuid)
  // open link at static/#/chat/{SessionUuid}
  window.open(`static/#/chat/${SessionUuid}`, '_blank')
}

const footerClass = computed(() => {
  let classes = ['p-4']
  if (isMobile.value)
    classes = ['sticky', 'left-0', 'bottom-0', 'right-0', 'p-2', 'pr-3', 'overflow-hidden']
  return classes
})

function onScrollToTop() {
  const scrollRef = document.querySelector('#scrollRef')
  if (scrollRef)
    nextTick(() => scrollRef.scrollTop = 0)
}
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <Header :title="title" />
    <main class="flex-1 overflow-hidden">
      <div id="scrollRef" ref="scrollRef" class="h-full overflow-hidden overflow-y-auto">
        <div
          id="image-wrapper" class="w-full max-w-screen-xl m-auto dark:bg-[#101014]"
          :class="[isMobile ? 'p-2' : 'p-4']"
        >
          <Message
            v-for="(item, index) of dataSources" :key="index" class="chat-message" :date-time="item.dateTime"
            :model="model" :text="item.text" :inversion="item.inversion" :error="item.error" :loading="item.loading"
            :index="index"
          />
        </div>
        <div class="flex justify-center items-center">
          <HoverButton :tooltip="$t('chat_snapshot.continueChat')" @click="handleChat">
            <span class="text-xl text-[#4f555e] dark:text-white m-auto mx-10">
              <SvgIcon icon="mdi:chat-plus" width="40" height="40" />
            </span>
          </HoverButton>
        </div>
      </div>
    </main>
    <div class="floating-button">
      <HoverButton :tooltip="$t('chat_snapshot.continueChat')" @click="handleChat">
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
</template>

<style>
/* CSS for the button */
.floating-button {
  position: fixed;
  bottom: 10vh;
  right: 10vmin;
  z-index: 99;
  padding: 0.5em;
  border-radius: 50%;
  cursor: pointer;
  background-color: #4ff09a;
  box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.2), 0 6px 20px 0 rgba(0, 0, 0, 0.19);
}
</style>
