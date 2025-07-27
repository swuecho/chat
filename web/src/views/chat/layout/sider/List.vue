<script setup lang='ts'>
import { computed, onMounted } from 'vue'
import { NInput, NPopconfirm, NScrollbar, useMessage } from 'naive-ui'
import { renameChatSession } from '@/api'
import { SvgIcon } from '@/components/common'
import { useAppStore, useAuthStore, useChatStore } from '@/store'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import ModelAvatar from '@/views/components/Avatar/ModelAvatar.vue'
import { t } from '@/locales'
import { throttle } from 'lodash';

const { isMobile } = useBasicLayout()
const nui_msg = useMessage()

const appStore = useAppStore()
const chatStore = useChatStore()
const authStore = useAuthStore()

const dataSources = computed(() => {
  // If no active workspace, show sessions from all workspaces
  if (!chatStore.activeWorkspace) {
    const allSessions: Chat.Session[] = []
    for (const sessions of Object.values(chatStore.workspaceHistory)) {
      allSessions.push(...sessions)
    }
    return allSessions
  }
  
  // Filter sessions by active workspace - show only sessions belonging to this workspace
  const workspaceSessions = chatStore.getSessionsByWorkspace(chatStore.activeWorkspace)
  return workspaceSessions
})
const isLogined = computed(() => Boolean(authStore.token))

onMounted(async () => {
  console.log('Sider List component mounted, isLogined:', isLogined.value)
  if (isLogined.value) {
    console.log('User is logged in, syncing chat sessions...')
    await handleSyncChat()
  } else {
    console.log('User is not logged in, skipping chat session sync')
  }
})
async function handleSyncChat() {
  const totalSessions = Object.values(chatStore.workspaceHistory).reduce((sum, sessions) => sum + sessions.length, 0)
  console.log('handleSyncChat called, current total sessions:', totalSessions)
  try {
    console.log('Calling chatStore.syncChatSessions()...')
    await chatStore.syncChatSessions()
    const newTotalSessions = Object.values(chatStore.workspaceHistory).reduce((sum, sessions) => sum + sessions.length, 0)
    console.log('Chat sessions synced successfully, new total sessions:', newTotalSessions)
  }
  catch (error: any) {
    console.error('Error syncing chat sessions:', error)
    if (error.response?.status === 500)
      nui_msg.error(t('error.syncChatSession'))
    // eslint-disable-next-line no-console
    console.log(error)
  }
}

async function handleSelect(uuid: string) {
  if (isActive(uuid))
    return

  if (chatStore.active)
    await chatStore.updateChatSessionIfEdited(chatStore.active, { isEdit: false })

  // Use the store's setActive method which now handles workspace-aware routing
  await chatStore.setActive(uuid)

  if (isMobile.value)
    appStore.setSiderCollapsed(true)
}

// throttle handleSelect
// async function handleSelectThrottle() {
//   throttle(async ({ uuid }: Chat.Session) => await handleSelect(uuid), 500)
// } 

// Create a wrapper to debounce the handleSelect function
const throttledHandleSelect = throttle((uuid) => {
  handleSelect(uuid);
}, 500); // 300ms debounce time

function handleEdit({ uuid }: Chat.Session, isEdit: boolean, event?: MouseEvent) {
  event?.stopPropagation()
  chatStore.updateChatSession(uuid, { isEdit })
}
function handleSave({ uuid, title }: Chat.Session, isEdit: boolean, event?: MouseEvent) {
  event?.stopPropagation()
  chatStore.updateChatSession(uuid, { isEdit })
  // should move to store
  renameChatSession(uuid, title)
}

function handleDelete(index: number, event?: MouseEvent | TouchEvent) {
  event?.stopPropagation()
  const session = dataSources.value[index]
  if (session) {
    chatStore.deleteChatSession(session.uuid)
  }
}

function handleEnter({ uuid, title }: Chat.Session, isEdit: boolean, event: KeyboardEvent) {
  event?.stopPropagation()
  if (event.key === 'Enter') {
    chatStore.updateChatSession(uuid, { isEdit })
    renameChatSession(uuid, title)
  }
}

function isActive(uuid: string) {
  return chatStore.active === uuid
}
</script>

<template>
  <NScrollbar class="px-2">
    <div class="flex flex-col gap-1 text-sm">
      <template v-if="!dataSources.length">
        <div class="flex flex-col items-center mt-2 text-center text-neutral-300">
          <SvgIcon icon="ri:inbox-line" class="mb-2 text-3xl" />
          <span>{{ $t('common.noData') }}</span>
        </div>
      </template>
      <template v-else>
        <div v-for="(item, index) of dataSources" :key="index">
          <a class="relative flex items-center gap-2 p-2 break-all border rounded-sm cursor-pointer hover:bg-neutral-100 group dark:border-neutral-800 dark:hover:bg-[#24272e]"
            :class="isActive(item.uuid) && ['border-[#4b9e5f]', 'bg-neutral-100', 'text-[#4b9e5f]', 'dark:bg-[#24272e]', 'dark:border-[#4b9e5f]', 'pr-14']"
            @click="throttledHandleSelect(item.uuid)">
            <span>
              <ModelAvatar :model="item.model" />
            </span>
            <div class="relative flex-1 overflow-hidden break-all text-ellipsis whitespace-nowrap">
              <NInput v-if="item.isEdit" v-model:value="item.title" data-testid="edit_session_topic_input" size="tiny"
                @keypress="handleEnter(item, false, $event)" />
              <span v-else>{{ item.title }}</span>
            </div>
            <div v-if="isActive(item.uuid)" class="absolute z-10 flex visible right-1">
              <template v-if="item.isEdit">
                <button class="p-1" data-testid="save_session_topic" @click="handleSave(item, false, $event)">
                  <SvgIcon icon="ri:save-line" />
                </button>
              </template>
              <template v-else>
                <button class="p-1" data-testid="edit_session_topic">
                  <SvgIcon icon="ri:edit-line" @click="handleEdit(item, true, $event)" />
                </button>
                <div v-if="dataSources.length > 1">
                  <NPopconfirm placement="bottom" data-testid="confirm_delete_session"
                    @positive-click="handleDelete(index, $event)">
                    <template #trigger>
                      <button class="p-1">
                        <SvgIcon icon="ri:delete-bin-line" />
                      </button>
                    </template>
                    {{ $t('chat.deleteChatSessionsConfirm') }}
                  </NPopconfirm>
                </div>
              </template>
            </div>
          </a>
        </div>
      </template>
    </div>
  </NScrollbar>
</template>
