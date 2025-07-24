<script lang='ts' setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { NAutoComplete, NButton, NInput, NModal, NSpin } from 'naive-ui'
// @ts-ignore
import { v7 as uuidv7 } from 'uuid'
import { useScroll } from '@/views/chat/hooks/useScroll'
import HeaderMobile from '@/views/chat/components/HeaderMobile/index.vue'
import SessionConfig from '@/views/chat/components/Session/SessionConfig.vue'
import { HoverButton, SvgIcon } from '@/components/common'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { useChatStore, usePromptStore } from '@/store'
import { t } from '@/locales'
import UploadModal from '@/views/chat/components/UploadModal.vue'
import UploaderReadOnly from '@/views/chat/components/UploaderReadOnly.vue'
import ModelSelector from '@/views/chat/components/ModelSelector.vue'
import MessageList from '@/views/chat/components/MessageList.vue'
import PromptGallery from '@/views/chat/components/PromptGallery/index.vue'
import ArtifactGallery from '@/views/chat/components/ArtifactGallery.vue'
import { useSlashToFocus } from '../hooks/useSlashToFocus'
import JumpToBottom from './JumpToBottom.vue'
import ChatVFSUploader from '@/components/ChatVFSUploader.vue'
import VFSProvider from '@/components/VFSProvider.vue'

// Import extracted composables
import { useConversationFlow } from '../composables/useConversationFlow'
import { useRegenerate } from '../composables/useRegenerate'
import { useSearchAndPrompts } from '../composables/useSearchAndPrompts'
import { useChatActions } from '../composables/useChatActions'

// Create a ref for the input element
const searchInputRef = ref(null);
useSlashToFocus(searchInputRef);

let controller = new AbortController()

const chatStore = useChatStore()
const promptStore = usePromptStore()

const { sessionUuid } = defineProps({
  sessionUuid: {
    type: String,
    required: true
  },
});

const { isMobile } = useBasicLayout()
const { scrollRef, scrollToBottom, scrollToBottomIfAtBottom, smoothScrollToBottomIfAtBottom } = useScroll()

// Initialize composables
const conversationFlow = useConversationFlow(sessionUuid, scrollToBottom, smoothScrollToBottomIfAtBottom)
const regenerate = useRegenerate(sessionUuid)
const searchAndPrompts = useSearchAndPrompts()
const chatActions = useChatActions(sessionUuid)

// Sync chat messages
chatStore.syncChatMessages(sessionUuid)

const dataSources = computed(() => chatStore.getChatSessionDataByUuid(sessionUuid))
const chatSession = computed(() => chatStore.getChatSessionByUuid(sessionUuid))

// Destructure from composables
const { prompt, searchOptions, renderOption, handleSelectAutoComplete, handleUsePrompt } = searchAndPrompts
const {
  snapshotLoading,
  botLoading,
  showUploadModal,
  showModal,
  showArtifactGallery,
  toggleArtifactGallery,
  handleVFSFileUploaded
} = chatActions

// Use loading state from composables
const loading = computed(() => conversationFlow.loading.value || regenerate.loading.value)

async function handleSubmit() {
  const message = prompt.value
  if (conversationFlow.validateConversationInput(message)) {
    prompt.value = '' // Clear the input after validation passes
    const chatUuid = uuidv7()
    await conversationFlow.addUserMessage(chatUuid, message)
    conversationFlow.startStream(message, dataSources.value, chatUuid)
  }
}

async function onRegenerate(index: number) {
  await regenerate.onRegenerate(index, dataSources.value)
}

async function handleAdd() {
  await chatActions.handleAdd(dataSources.value)
}

async function handleSnapshot() {
  await chatActions.handleSnapshot()
}

async function handleCreateBot() {
  await chatActions.handleCreateBot()
}

function handleClear() {
  chatActions.handleClear(loading)
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

const placeholder = computed(() => {
  if (isMobile.value)
    return t('chat.placeholderMobile')
  return t('chat.placeholder')
})

const sendButtonDisabled = computed(() => {
  return loading.value || !prompt.value || (typeof prompt.value === 'string' ? prompt.value.trim() === '' : true)
})

const footerClass = computed(() => {
  let classes = ['m-2', 'p-2']
  if (isMobile.value)
    classes = ['p-2', 'pr-3', 'overflow-hidden']
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

// VFS event handlers with stream response functionality
const handleCodeExampleAddedWithStream = async (codeInfo: any) => {
  await chatActions.handleCodeExampleAdded(codeInfo, (chatUuid: string, message: string) => {
    return conversationFlow.startStream(message, dataSources.value, chatUuid)
  })
}

// VFS Upload Modal state and handler
const showVFSUploadModal = ref(false)

function handleUpload() {
  showVFSUploadModal.value = true
}
</script>

<template>
  <VFSProvider>
    <div class="flex flex-col w-full h-full">
      <div>
        <UploadModal :sessionUuid="sessionUuid" :showUploadModal="showUploadModal"
          @update:showUploadModal="showUploadModal = $event" />
        <ChatVFSUploader :session-uuid="sessionUuid" :showUploadModal="showVFSUploadModal"
          @update:showUploadModal="showVFSUploadModal = $event" @file-uploaded="handleVFSFileUploaded"
          @code-example-added="handleCodeExampleAddedWithStream" />
      </div>
      <HeaderMobile v-if="isMobile" @add-chat="handleAdd" @snapshot="handleSnapshot" @toggle="showModal = true" />
      <main class="flex-1 overflow-hidden flex flex-col">
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
        <div id="scrollRef" ref="scrollRef" class="flex-1 overflow-hidden overflow-y-auto">
          <div v-if="!showArtifactGallery" id="image-wrapper"
            class="w-full max-w-screen-xl mx-auto dark:bg-[#101014] mb-10" :class="[isMobile ? 'p-2' : 'p-4']">
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
          <div v-else class="h-full">
            <ArtifactGallery />
          </div>
          <JumpToBottom v-if="dataSources.length > 1 && !showArtifactGallery" targetSelector="#scrollRef"
            :scrollThresholdShow="200" />

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

            <HoverButton tooltip="Upload files to VFS for code runners" @click="handleUpload">
              <span class="text-xl text-[#4b9e5f] dark:text-white">
                <SvgIcon icon="mdi:folder-open" />
              </span>
            </HoverButton>

            <HoverButton v-if="!isMobile" @click="toggleArtifactGallery"
              :tooltip="showArtifactGallery ? 'Hide Gallery' : 'Show Gallery'">
              <span class="text-xl text-[#4b9e5f] dark:text-white">
                <SvgIcon icon="ri:gallery-line" />
              </span>
            </HoverButton>

            <HoverButton v-if="!isMobile" @click="showModal = true" :tooltip="$t('chat.chatSettings')">
              <span class="text-xl text-[#4b9e5f]">
                <SvgIcon icon="teenyicons:adjust-horizontal-solid" />
              </span>
            </HoverButton>
            <NAutoComplete v-model:value="prompt" :options="searchOptions" :render-label="renderOption"
              :on-select="handleSelectAutoComplete">
              <template #default="{ handleInput, handleBlur, handleFocus }">
                <NInput ref="searchInputRef" id="message_textarea" :value="prompt" type="textarea"
                  :placeholder="placeholder" data-testid="message_textarea"
                  :autosize="{ minRows: 1, maxRows: isMobile ? 4 : 8 }" @input="handleInput" @focus="handleFocus"
                  @blur="handleBlur" @keypress="handleEnter" />
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
  </VFSProvider>
</template>

<style scoped>
/* Custom scrollbar styling */
#scrollRef {
  scrollbar-width: thin;
  scrollbar-color: rgba(155, 155, 155, 0.5) transparent;
}

#scrollRef::-webkit-scrollbar {
  width: 8px;
}

#scrollRef::-webkit-scrollbar-track {
  background: transparent;
  border-radius: 4px;
}

#scrollRef::-webkit-scrollbar-thumb {
  background: rgba(155, 155, 155, 0.5);
  border-radius: 4px;
  transition: background 0.2s ease;
}

#scrollRef::-webkit-scrollbar-thumb:hover {
  background: rgba(155, 155, 155, 0.8);
}

#scrollRef::-webkit-scrollbar-thumb:active {
  background: rgba(155, 155, 155, 1);
}

/* Dark mode scrollbar */
.dark #scrollRef::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
}

.dark #scrollRef::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}

.dark #scrollRef::-webkit-scrollbar-thumb:active {
  background: rgba(255, 255, 255, 0.7);
}

@media (max-width: 768px) {

  /* Thinner scrollbar on mobile */
  #scrollRef::-webkit-scrollbar {
    width: 4px;
  }
}
</style>