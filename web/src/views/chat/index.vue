<script setup lang='ts'>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { v4 as uuidv4 } from 'uuid'
import { useRoute } from 'vue-router'
import { NButton, NInput, NModal, useDialog, useMessage } from 'naive-ui'
import html2canvas from 'html2canvas'
import { Message } from './components'
import { useScroll } from './hooks/useScroll'
import { useChat } from './hooks/useChat'
import { useCopyCode } from './hooks/useCopyCode'
import { useUsingContext } from './hooks/useUsingContext'
import HeaderComponent from './components/Header/index.vue'
import SessionConfig from './components/Session/SessionConfig.vue'
import { HoverButton, SvgIcon } from '@/components/common'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { useChatStore } from '@/store'
import { fetchChatStream } from '@/api'
import { t } from '@/locales'

let controller = new AbortController()

const route = useRoute()
const dialog = useDialog()
const nui_msg = useMessage()

const chatStore = useChatStore()

useCopyCode()

const { isMobile } = useBasicLayout()
const { addChat, updateChat, updateChatPartial, updateChatText } = useChat()
const { scrollRef, scrollToBottom } = useScroll()
const { usingContext } = useUsingContext()
// session uuid
const { uuid } = route.params as { uuid: string }
const sessionUuid = uuid
const dataSources = computed(() => chatStore.getChatSessionDataByUuid(sessionUuid))
const conversationList = computed(() => dataSources.value.filter(item => (!item.inversion && !item.error)))

const prompt = ref<string>('')
const loading = ref<boolean>(false)

const showModal = ref<boolean>(false)

function handleSubmit() {
  onConversationStream()
}

async function onConversationStream() {
  const message = prompt.value

  if (loading.value)
    return

  if (!message || message.trim() === '')
    return

  const chatUuid = uuidv4()

  addChat(
    sessionUuid,
    {
      uuid: chatUuid,
      dateTime: new Date().toLocaleString(),
      text: message,
      inversion: true,
      error: false,
      conversationOptions: null,
      requestOptions: { prompt: message, options: null },
    },
  )
  scrollToBottom()

  loading.value = true
  prompt.value = ''

  let options: Chat.ConversationRequest = {}
  const lastContext = conversationList.value[conversationList.value.length - 1]?.conversationOptions

  if (lastContext && usingContext.value)
    options = { ...lastContext }
  options.uuid = sessionUuid.toString()

  // add a blank response
  addChat(
    sessionUuid,
    {
      uuid: '',
      dateTime: new Date().toLocaleString(),
      text: '',
      loading: true,
      inversion: false,
      error: false,
      conversationOptions: null,
      requestOptions: { prompt: message, options: { ...options } },
    },
  )
  scrollToBottom()
  const subscribleStrem = async () => {
    try {
      // Send the request with axios
      const response = fetchChatStream(
        sessionUuid,
        chatUuid,
        false,
        message,
        options,
        (progress: any) => {
          const xhr = progress.event.target
          const {
            responseText,
          } = xhr

          const lastIndex = responseText.lastIndexOf('data: ')

          // Extract the JSON data chunk from the responseText
          const chunk = responseText.slice(lastIndex + 6)

          // Check if the chunk is not empty
          if (chunk) {
            // Parse the JSON data chunk
            const data = JSON.parse(chunk)
            const answer = data.choices[0].delta.content
            const answer_uuid = data.id.replace('chatcmpl-', '') // use answer id as uuid
            updateChat(
              sessionUuid,
              dataSources.value.length - 1,
              {
                uuid: answer_uuid,
                dateTime: new Date().toLocaleString(),
                text: answer,
                inversion: false,
                error: false,
                loading: false,
                conversationOptions: { conversationId: data.conversationId, parentMessageId: data.id },
                requestOptions: { prompt: message, options: { ...options } },
              },
            )
          }
        },
      )
      return response
    }
    catch (error: any) {
      const response = error.response
      if (response.status === 500)
        nui_msg.error(response.data.message)
      throw error
    }
    finally {
      loading.value = false
    }
  }

  await subscribleStrem()
}

// async function _onConversation() {
//   const message = prompt.value

//   if (loading.value)
//     return

//   if (!message || message.trim() === '')
//     return

//   const chatUuid = uuidv4()

//   addChat(
//     sessionUuid,
//     {
//       uuid: chatUuid,
//       dateTime: new Date().toLocaleString(),
//       text: message,
//       inversion: true,
//       error: false,
//       conversationOptions: null,
//       requestOptions: { prompt: message, options: null },
//     },
//   )
//   scrollToBottom()

//   loading.value = true
//   prompt.value = ''

//   let options: Chat.ConversationRequest = {}
//   const lastContext = conversationList.value[conversationList.value.length - 1]?.conversationOptions

//   if (lastContext && usingContext.value)
//     options = { ...lastContext }
//   options.uuid = sessionUuid.toString()

//   // add a blank response
//   addChat(
//     sessionUuid,
//     {
//       uuid: '',
//       dateTime: new Date().toLocaleString(),
//       text: '',
//       loading: true,
//       inversion: false,
//       error: false,
//       conversationOptions: null,
//       requestOptions: { prompt: message, options: { ...options } },
//     },
//   )
//   scrollToBottom()

//   try {
//     const respData = (await fetchChatAPI<any>(sessionUuid, chatUuid, false, message, options)) as any
//     // scrollToBottom()
//     updateChat(
//       sessionUuid,
//       dataSources.value.length - 1,
//       {
//         uuid: respData.chatUuid,
//         dateTime: new Date().toLocaleString(),
//         text: respData.text,
//         inversion: false,
//         error: false,
//         loading: false,
//         conversationOptions: {},
//         requestOptions: { prompt: message, options: { ...options } },
//       })
//   }
//   catch (err: any) {
//     // What's the fuck? response data is in err
//     console.error(err)
//   }
//   finally {
//     loading.value = false
//   }
// }

async function onRegenerate(index: number) {
  if (loading.value)
    return

  controller = new AbortController()

  const { requestOptions } = dataSources.value[index]

  const message = requestOptions?.prompt ?? ''

  let options: Chat.ConversationRequest = {}

  if (requestOptions.options)
    options = { ...requestOptions.options }

  loading.value = true
  const chatUuid = dataSources.value[index].uuid
  updateChat(
    sessionUuid,
    index,
    {
      uuid: chatUuid,
      dateTime: new Date().toLocaleString(),
      text: '',
      inversion: false,
      error: false,
      loading: true,
      conversationOptions: null,
      requestOptions: { prompt: message, ...options },
    },
  )

  try {
    const subscribleStrem = async () => {
      try {
        // Send the request with axios
        const response = fetchChatStream(
          sessionUuid,
          chatUuid,
          true,
          message,
          options,
          (progress: any) => {
            const xhr = progress.event.target
            const {
              responseText,
            } = xhr

            const lastIndex = responseText.lastIndexOf('data: ')

            // Extract the JSON data chunk from the responseText
            const chunk = responseText.slice(lastIndex + 6)

            // Check if the chunk is not empty
            if (chunk) {
              // Parse the JSON data chunk
              const data = JSON.parse(chunk)
              const answer = data.choices[0].delta.content
              const answer_uuid = data.id.replace('chatcmpl-', '') // use answer id as uuid
              updateChat(
                sessionUuid,
                index,
                {
                  uuid: answer_uuid,
                  dateTime: new Date().toLocaleString(),
                  text: answer,
                  inversion: false,
                  error: false,
                  loading: false,
                  conversationOptions: { conversationId: data.conversationId, parentMessageId: data.id },
                  requestOptions: { prompt: message, options: { ...options } },
                },
              )
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
        loading.value = false
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
        dateTime: new Date().toLocaleString(),
        text: errorMessage,
        inversion: false,
        error: true,
        loading: false,
        conversationOptions: null,
        requestOptions: { prompt: message, ...options },
      },
    )
  }
  finally {
    loading.value = false
  }
}

function handleExport() {
  if (loading.value)
    return

  const d = dialog.warning({
    title: t('chat.exportImage'),
    content: t('chat.exportImageConfirm'),
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: async () => {
      try {
        d.loading = true
        const ele = document.getElementById('image-wrapper')
        const canvas = await html2canvas(ele as HTMLDivElement, {
          useCORS: true,
        })
        const imgUrl = canvas.toDataURL('image/png')
        const tempLink = document.createElement('a')
        tempLink.style.display = 'none'
        tempLink.href = imgUrl
        tempLink.setAttribute('download', 'chat-shot.png')
        if (typeof tempLink.download === 'undefined')
          tempLink.setAttribute('target', '_blank')

        document.body.appendChild(tempLink)
        tempLink.click()
        document.body.removeChild(tempLink)
        window.URL.revokeObjectURL(imgUrl)
        d.loading = false
        nui_msg.success(t('chat.exportSuccess'))
        Promise.resolve()
      }
      catch (error: any) {
        nui_msg.error(t('chat.exportFailed'))
      }
      finally {
        d.loading = false
      }
    },
  })
}
// The user wants to delete the message with the given index.
// If the message is already being deleted, we ignore the request.
// If the user confirms that they want to delete the message, we call
// the deleteChatByUuid function from the chat store.
function handleDelete(index: number) {
  if (loading.value)
    return

  dialog.warning({
    title: t('chat.deleteMessage'),
    content: t('chat.deleteMessageConfirm'),
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: async () => {
      chatStore.deleteChatByUuid(sessionUuid, index)
    },
  })
}

function handleAfterEdit(index: number, text: string) {
  if (loading.value)
    return

  updateChatText(
    sessionUuid,
    index,
    text,
  )
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

function handleStop() {
  if (loading.value) {
    controller.abort()
    loading.value = false
  }
}

const placeholder = computed(() => {
  if (isMobile.value)
    return t('chat.placeholderMobile')
  return t('chat.placeholder')
})

const buttonDisabled = computed(() => {
  return loading.value || !prompt.value || prompt.value.trim() === ''
})

const footerClass = computed(() => {
  let classes = ['p-4']
  if (isMobile.value)
    classes = ['sticky', 'left-0', 'bottom-0', 'right-0', 'p-2', 'pr-3', 'overflow-hidden']
  return classes
})

onMounted(() => {
  scrollToBottom()
})

onUnmounted(() => {
  if (loading.value)
    controller.abort()
})
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <HeaderComponent
      v-if="isMobile" :using-context="usingContext" @export="handleExport"
      @toggle-using-context="showModal = true"
    />
    <main class="flex-1 overflow-hidden">
      <NModal ref="sessionConfigModal" v-model:show="showModal">
        <SessionConfig
          ref="sessionConfig" :uuid="sessionUuid"
        />
      </NModal>
      <div id="scrollRef" ref="scrollRef" class="h-full overflow-hidden overflow-y-auto">
        <div
          id="image-wrapper" class="w-full max-w-screen-xl m-auto dark:bg-[#101014]"
          :class="[isMobile ? 'p-2' : 'p-4']"
        >
          <template v-if="!dataSources.length">
            <div class="flex items-center justify-center mt-4 text-center text-neutral-300">
              <SvgIcon icon="ri:bubble-chart-fill" class="mr-2 text-3xl" />
              <span>{{ $t('common.help') }}</span>
            </div>
          </template>
          <template v-else>
            <div>
              <Message
                v-for="(item, index) of dataSources" :key="index" :date-time="item.dateTime" :text="item.text"
                :inversion="item.inversion" :error="item.error" :loading="item.loading" :index="index"
                @regenerate="onRegenerate(index)" @delete="handleDelete(index)" @after-edit="handleAfterEdit"
              />
              <div class="sticky bottom-0 left-0 flex justify-center">
                <NButton v-if="loading" type="warning" @click="handleStop">
                  <template #icon>
                    <SvgIcon icon="ri:stop-circle-line" />
                  </template>
                  {{ $t('chat.stopAnswer') }}
                </NButton>
              </div>
            </div>
          </template>
        </div>
      </div>
    </main>
    <footer :class="footerClass">
      <div class="w-full max-w-screen-xl m-auto">
        <div class="flex items-center justify-between space-x-2">
          <HoverButton @click="handleClear">
            <span class="text-xl text-[#4f555e] dark:text-white">
              <SvgIcon icon="ri:delete-bin-line" />
            </span>
          </HoverButton>
          <HoverButton v-if="!isMobile" @click="handleExport">
            <span class="text-xl text-[#4f555e] dark:text-white">
              <SvgIcon icon="ri:download-2-line" />
            </span>
          </HoverButton>
          <HoverButton v-if="!isMobile" @click="showModal = true">
            <span class="text-xl" :class="{ 'text-[#4b9e5f]': usingContext, 'text-[#a8071a]': !usingContext }">
              <SvgIcon icon="ri:chat-history-line" />
            </span>
          </HoverButton>

          <NInput
            id="message_textarea" v-model:value="prompt" data-testid="message_textarea" type="textarea"
            :autosize="{ minRows: 1, maxRows: isMobile ? 4 : 8 }" :placeholder="placeholder" @keypress="handleEnter"
          />
          <NButton
            id="send_message_button" data-testid="send_message_button" type="primary" :disabled="buttonDisabled"
            @click="handleSubmit"
          >
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
