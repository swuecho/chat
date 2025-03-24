<script lang='ts' setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
// @ts-ignore
import { v7 as uuidv7 } from 'uuid'
import { NAutoComplete, NButton, NInput, NModal, NSpin, useDialog, useMessage } from 'naive-ui'
import { storeToRefs } from 'pinia'
import html2canvas from 'html2canvas'
import { type OnSelect } from 'naive-ui/es/auto-complete/src/interface'
import { useScroll } from '@/views/chat/hooks/useScroll'
import { useChat } from '@/views/chat/hooks/useChat'
import HeaderMobile from '@/views/chat/components/HeaderMobile/index.vue'
import SessionConfig from '@/views/chat/components/Session/SessionConfig.vue'
import { createChatBot, createChatSnapshot, deleteChatMessage, fetchChatStream, getChatSessionDefault } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { useAppStore, useChatStore, usePromptStore } from '@/store'
import { t } from '@/locales'
import { genTempDownloadLink } from '@/utils/download'
import { nowISO } from '@/utils/date'
import UploadModal from '@/views/chat/components/UploadModal.vue'
import UploaderReadOnly from '@/views/chat/components/UploaderReadOnly.vue'
import ModelSelector from '@/views/chat/components/ModelSelector.vue'
import MessageList from '@/views/chat/components/MessageList.vue'
import PromptGallery from '@/views/chat/components/PromptGallery/index.vue'
import { getDataFromResponseText } from '@/utils/string'
import renderMessage from './RenderMessage.vue'
let controller = new AbortController()

const dialog = useDialog()
const nui_msg = useMessage()

const chatStore = useChatStore()

const { sessionUuid } = defineProps({
  sessionUuid: {
    type: String,
    required: true
  },
});



const { isMobile } = useBasicLayout()
const { addChat, updateChat, updateChatPartial } = useChat()
const { scrollRef, scrollToBottom, scrollToBottomIfAtBottom } = useScroll()
// session uuid
chatStore.syncChatMessages(sessionUuid)

const dataSources = computed(() => chatStore.getChatSessionDataByUuid(sessionUuid))
const chatSession = computed(() => chatStore.getChatSessionByUuid(sessionUuid))

const prompt = ref<string>('')
const loading = ref<boolean>(false)
const showUploadModal = ref<boolean>(false)
const showModal = ref<boolean>(false)
const snapshotLoading = ref<boolean>(false)
const botLoading = ref<boolean>(false)

const appStore = useAppStore()


async function handleAdd() {
  if (dataSources.value.length > 0) {
    const new_chat_text = t('chat.new')
    const default_model_parameters = await getChatSessionDefault(new_chat_text)
    await chatStore.addChatSession(default_model_parameters)
    if (isMobile.value)
      appStore.setSiderCollapsed(true)
  }
}

// 添加PromptStore
const promptStore = usePromptStore()

// 使用storeToRefs，保证store修改后，联想部分能够重新渲染
const { promptList: promptTemplate } = storeToRefs<any>(promptStore)

// 可优化部分
// 搜索选项计算，这里使用value作为索引项，所以当出现重复value时渲染异常(多项同时出现选中效果)
// 理想状态下其实应该是key作为索引项,但官方的renderOption会出现问题，所以就需要value反renderLabel实现
const searchOptions = computed(() => {
  function filterItemsByPrompt(item: { key: string }): boolean {
    const lowerCaseKey = item.key.toLowerCase()
    const lowerCasePrompt = prompt.value.substring(1).toLowerCase()
    return lowerCaseKey.includes(lowerCasePrompt)
  }
  function filterItemsByTitle(item: { title: string }): boolean {
    const lowerCaseKey = item.title.toLowerCase()
    const lowerCasePrompt = prompt.value.substring(1).toLowerCase()
    return lowerCaseKey.includes(lowerCasePrompt)
  }
  if (prompt.value.startsWith('/')) {
    const filterStores = chatStore.history.filter(filterItemsByTitle).map((obj: { uuid: any }) => {
      return {
        label: `UUID|$|${obj.uuid}`,
        value: `UUID|$|${obj.uuid}`,
      }
    })

    const filterPrompts = promptTemplate.value.filter(filterItemsByPrompt).map((obj: { value: any }) => {
      return {
        label: obj.value,
        value: obj.value,
      }
    })
    const all = filterStores.concat(filterPrompts)
    return all
  }
  else {
    return []
  }
})
// value反渲染key
const renderOption = (option: { label: string }) => {
  for (const i of promptTemplate.value) {
    if (i.value === option.label)
      return [i.key]
  }
  for (const chat of chatStore.history) {
    if (`UUID|$|${chat.uuid}` === option.label)
      return [chat.title]
  }
  return []
}

function handleSubmit() {
  onConversationStream()
}

async function onConversationStream() {
  const message = prompt.value

  if (loading.value)
    return

  if (!message || message.trim() === '')
    return

  const chatUuid = uuidv7()

  addChat(
    sessionUuid,
    {
      uuid: chatUuid,
      dateTime: nowISO(),
      text: message,
      inversion: true,
      error: false,
    },
  )
  scrollToBottom()

  loading.value = true
  prompt.value = ''

  // add a blank response
  addChat(
    sessionUuid,
    {
      uuid: '',
      dateTime: nowISO(),
      text: '',
      loading: true,
      inversion: false,
      error: false,
    },
  )
  scrollToBottomIfAtBottom()
  const subscribleStrem = async () => {
    try {
      // Send the request with axios
      const response = fetchChatStream(
        sessionUuid,
        chatUuid,
        false,
        message,
        (progress: any) => {
          const xhr = progress.event.target
          const {
            responseText,
            status,
          } = xhr
          if (status >= 400) {
            const error_json: { code: number; message: string; details: any } = JSON.parse(responseText)
            console.log(responseText)
            nui_msg.error(formatErr(error_json), {
              duration: 5000,
              closable: true,
              render: renderMessage
            })

            chatStore.deleteChatByUuid(sessionUuid, dataSources.value.length - 1)
            // remove last input box
            loading.value = false
          }
          else {
            const chunk = getDataFromResponseText(responseText)
            // Check if the chunk is not empty
            if (chunk) {
              // Parse the JSON data chunk
              try {
                const data = JSON.parse(chunk)
                const answer = data.choices[0].delta.content
                const answer_uuid = data.id.replace('chatcmpl-', '') // use answer id as uuid
                updateChat(
                  sessionUuid,
                  dataSources.value.length - 1,
                  {
                    uuid: answer_uuid,
                    dateTime: nowISO(),
                    text: answer,
                    inversion: false,
                    error: false,
                    loading: false,
                  },
                )
                scrollToBottomIfAtBottom()
              }
              catch (error) {
                // eslint-disable-next-line no-console
                console.log("xxx", error)
              }
            }
          }
        },
      )
      return response
    }
    catch (error: any) {
      // eslint-disable-next-line no-console
      console.log(error)
      const response = error.response
      if (response.status >= 400) {
        nui_msg.error(response.data.message)
        loading.value = false
      }
    }
  }

  await subscribleStrem()
  loading.value = false
}

async function onRegenerate(index: number) {
  if (loading.value)
    return

  controller = new AbortController()

  const chat = dataSources.value[index]

  const chatUuid = chat.uuid
  // from user
  const inversion = chat.inversion

  loading.value = true

  let updateIndex = index;
  let isRegenerate = true;

  if (inversion) {
    // trigger from user message
    const chatNext = dataSources.value[index + 1]
    if (chatNext) {
      updateIndex = index + 1
      isRegenerate = false
      // if there are answer below. then clear
      await deleteChatMessage(chatNext.uuid)
      updateChat(
        sessionUuid,
        updateIndex,
        {
          uuid: chatNext.uuid,
          dateTime: nowISO(),
          text: '',
          inversion: false,
          error: false,
          loading: true,
        },
      )

    } else {
      // add a blank response
      updateIndex = index + 1
      isRegenerate = false
      addChat(
        sessionUuid,
        {
          uuid: '',
          dateTime: nowISO(),
          text: '',
          loading: true,
          inversion: false,
          error: false,
        },
      )
    }
    // if there are answer below. then clear
    // if not, add answer

  } else {
    // clear the old answer for regenerating
    updateChat(
      sessionUuid,
      index,
      {
        uuid: chatUuid,
        dateTime: nowISO(),
        text: '',
        inversion: false,
        error: false,
        loading: true,
      },
    )

  }
  try {
    const subscribleStrem = async () => {
      try {
        // Send the request with axios
        const response = fetchChatStream(
          sessionUuid,
          chatUuid,
          isRegenerate,
          "",
          (progress: any) => {
            const xhr = progress.event.target
            const {
              responseText,
              status
            } = xhr


            if (status >= 400) {
              const error_json: { code: number; message: string; details: any } = JSON.parse(responseText)
              nui_msg.error(formatErr(error_json), {
                duration: 5000,
                closable: true,
                render: renderMessage
              })

              loading.value = false
            }
            else {
              // Extract the JSON data chunk from the responseText
              const chunk = getDataFromResponseText(responseText)

              // Check if the chunk is not empty
              if (chunk) {
                // Parse the JSON data chunk
                const data = JSON.parse(chunk)
                const answer = data.choices[0].delta.content
                const answer_uuid = data.id.replace('chatcmpl-', '') // use answer id as uuid
                updateChat(
                  sessionUuid,
                  updateIndex,
                  {
                    uuid: answer_uuid,
                    dateTime: nowISO(),
                    text: answer,
                    inversion: false,
                    error: false,
                    loading: false,
                  },
                )
              }

            }
          },
        )
        return response
      }
      catch (error) {
        console.error('Error:', error)
        throw error
      }
      finally {
        console.log(loading.value)
        loading.value = false
        console.log(loading.value)
      }
    }

    await subscribleStrem()
  }
  catch (error: any) {
    // TODO: fix  error
    if (error.message === 'canceled') {
      updateChatPartial(
        sessionUuid,
        index,
        {
          loading: false,
        },
      )
      return
    }

    const errorMessage = error?.message ?? t('common.wrong')

    updateChat(
      sessionUuid,
      index,
      {
        uuid: chatUuid,
        dateTime: nowISO(),
        text: errorMessage,
        inversion: false,
        error: true,
        loading: false,
      },
    )
  }
  finally {
    loading.value = false
  }
}

function formatErr(error_json: { code: number; message: string; details: any }) {
  const message = t(`error.${error_json.code}`) ?? error_json.message
  return `${error_json.code} : ${message}`
}

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

async function handleSnapshot() {
  snapshotLoading.value = true
  try {
    const snapshot = await createChatSnapshot(sessionUuid)
    const snapshot_uuid = snapshot.uuid
    window.open(`#/snapshot/${snapshot_uuid}`, '_blank')
    nui_msg.success(t('chat.snapshotSuccess'))
  } catch (error) {
    nui_msg.error(t('chat.snapshotFailed'))
  } finally {
    snapshotLoading.value = false
  }
}
async function handleCreateBot() {
  botLoading.value = true
  try {
    const snapshot = await createChatBot(sessionUuid)
    const snapshot_uuid = snapshot.uuid
    window.open(`#/snapshot/${snapshot_uuid}`, '_blank')
    nui_msg.success(t('chat.botSuccess'))
  } catch (error) {
    nui_msg.error(t('chat.botFailed'))
  } finally {
    botLoading.value = false
  }
}






function handleClear() {
  if (loading.value)
    return

  dialog.warning({
    title: t('chat.clearChat'),
    content: t('chat.clearChatConfirm'),
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: () => {
      chatStore.clearChatByUuid(sessionUuid)
    },
  })
}

function handleEnter(event: KeyboardEvent) {
  if (!isMobile.value) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      handleSubmit()
    }
  }
  else {
    if (event.key === 'Enter' && event.ctrlKey) {
      event.preventDefault()
      handleSubmit()
    }
  }
}

// function handleStop() {
//   if (loading.value) {
//     controller.abort()
//     loading.value = false
//   }
// }

const handleSelectAutoComplete: OnSelect = function (v: string | number) {
  if (typeof v === 'string' && v.startsWith('UUID|$|')) {
    // set active session to the selected uuid
    chatStore.setActive(v.split('|$|')[1])
  }
}

const placeholder = computed(() => {
  if (isMobile.value)
    return t('chat.placeholderMobile')
  return t('chat.placeholder')
})

const sendButtonDisabled = computed(() => {
  return loading.value || !prompt.value || prompt.value.trim() === ''
})

const footerClass = computed(() => {
  let classes = ['m-2', 'p-2']
  if (isMobile.value)
    classes = ['sticky', 'left-0', 'bottom-0', 'right-0', 'p-2', 'pr-3', 'overflow-hidden']
  return classes
})

onMounted(() => {
  scrollToBottom()
  // init default prompts
  promptStore.getPromptList()
})

onUnmounted(() => {
  if (loading.value)
    controller.abort()
})

const handleUsePrompt = (_: string, value: string): void => {
  prompt.value = value
}
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <div>
      <UploadModal :sessionUuid="sessionUuid" :showUploadModal="showUploadModal"
        @update:showUploadModal="showUploadModal = $event" />
    </div>
    <HeaderMobile v-if="isMobile" @add-chat="handleAdd" @snapshot="handleSnapshot" @toggle="showModal = true" />
    <main class="flex-1 overflow-hidden">
      <NModal ref="sessionConfigModal" v-model:show="showModal" :title="$t('chat.sessionConfig')" preset="dialog">
        <SessionConfig id="session-config" ref="sessionConfig" :uuid="sessionUuid" />
      </NModal>
      <div class="flex items-center justify-center mt-2 mb-2">
        <div class="w-4/5 md:w-1/3">
          <ModelSelector :uuid="sessionUuid" :model="chatSession?.model"></ModelSelector>
        </div>
      </div>
      <UploaderReadOnly v-if="!!sessionUuid" :sessionUuid="sessionUuid" :showUploaderButton="false">
      </UploaderReadOnly>
      <div id="scrollRef" ref="scrollRef" class="h-full overflow-hidden overflow-y-auto">
        <div id="image-wrapper" class="w-full max-w-screen-xl mx-auto dark:bg-[#101014] mb-10"
          :class="[isMobile ? 'p-2' : 'p-4']">
          <template v-if="!dataSources.length">
            <div class="flex items-center justify-center m-4 text-center text-neutral-300">
              <SvgIcon icon="ri:bubble-chart-fill" class="mr-2 text-3xl" />
              <span>{{ $t('common.help') }}</span>
            </div>
            <PromptGallery @usePrompt="handleUsePrompt"></PromptGallery>
          </template>
          <template v-else>
            <div>
              <MessageList :session-uuid="sessionUuid" :on-regenerate="onRegenerate" />
            </div>
          </template>
        </div>
      </div>
    </main>
    <footer :class="footerClass">
      <div class="w-full max-w-screen-xl m-auto">
        <div class="flex items-center justify-between space-x-1">
          <HoverButton :tooltip="$t('chat.clearChat')" @click="handleClear">
            <span class="text-xl text-[#4b9e5f] dark:text-white">
              <SvgIcon icon="icon-park-outline:clear" />
            </span>
          </HoverButton>
          <NSpin :show="botLoading">
            <HoverButton v-if="!isMobile" data-testid="snpashot-button" :tooltip="$t('chat.createBot')"
              @click="handleCreateBot">
              <span class="text-xl text-[#4b9e5f] dark:text-white">
                <SvgIcon icon="fluent:bot-add-24-regular" />
              </span>
            </HoverButton>
          </NSpin>

          <NSpin :show="snapshotLoading">
            <HoverButton v-if="!isMobile" data-testid="snpashot-button" :tooltip="$t('chat.chatSnapshot')"
              @click="handleSnapshot">
              <span class="text-xl text-[#4b9e5f] dark:text-white">
                <SvgIcon icon="ic:twotone-ios-share" />
              </span>
            </HoverButton>
          </NSpin>
          <HoverButton v-if="!isMobile" @click="showModal = true" :tooltip="$t('chat.chatSettings')">
            <span class="text-xl text-[#4b9e5f]">
              <SvgIcon icon="teenyicons:adjust-horizontal-solid" />
            </span>
          </HoverButton>
          <NAutoComplete v-model:value="prompt" :options="searchOptions" :render-label="renderOption"
            :on-select="handleSelectAutoComplete">
            <template #default="{ handleInput, handleBlur, handleFocus }">
              <NInput id="message_textarea" v-model:value="prompt" type="textarea" :placeholder="placeholder"
                data-testid="message_textarea" :autosize="{ minRows: 1, maxRows: isMobile ? 4 : 8 }"
                @input="handleInput" @focus="handleFocus" @blur="handleBlur" @keypress="handleEnter" />
            </template>
          </NAutoComplete>
          <button class="!-ml-8 z-10" @click="showUploadModal = true">
            <span class="text-xl text-[#4b9e5f]">
              <SvgIcon icon="clarity:attachment-line" />
            </span>
          </button>
          <NButton id="send_message_button" class="!ml-4" data-testid="send_message_button" type="primary"
            :disabled="sendButtonDisabled" @click="handleSubmit">
            <template #icon>
              <span class="dark:text-black">
                <SvgIcon icon="ri:send-plane-fill" />
              </span>
            </template>
          </NButton>
        </div>
      </div>
    </footer>
  </div>
</template>
