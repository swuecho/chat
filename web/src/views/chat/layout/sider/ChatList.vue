<script setup lang='ts'>
import { computed, onMounted } from 'vue'
import { NInput, NPopconfirm, NScrollbar, useMessage } from 'naive-ui'
import Sortable from 'sortablejs'
import { renameChatSession } from '@/api'
import { SvgIcon } from '@/components/common'
import { useAppStore, useChatStore } from '@/store'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'

const { isMobile } = useBasicLayout()
const nui_msg = useMessage()

const appStore = useAppStore()
const chatStore = useChatStore()

const dataSources = computed(() => chatStore.history)

onMounted(async () => {
  await handleSyncChat()
  // Initialize the Sortable object
  const chatMenuElement = document.getElementById('chat-menu')
  const _sortable = new Sortable(chatMenuElement, {
    // Add the draggable class to each item in the list
    handle: '.chat-menu-item',
    // drag handle selector
    ghostClass: 'sortable-ghost',
    animation: 150,
    // Call the onEnd function when an item is dropped
    onEnd: handleDrop,
  })
})

// Define the handleDrop function
function handleDrop(evt: { oldIndex: any; newIndex: any }): void {
  // Get the index of the item that was dragged
  const oldIndex = evt.oldIndex

  // Get the index of the item's new position
  const newIndex = evt.newIndex
  console.log(oldIndex, newIndex)
  // Use splice to move the item to its new position
  dataSources.value.splice(newIndex, 0, dataSources.value.splice(oldIndex, 1)[0])
}

async function handleSyncChat() {
  // if (chatStore.history.length == 1 && chatStore.history[0].title == 'New Chat'
  //   && chatStore.chat[0].data.length <= 0)
  try {
    await chatStore.syncChatSessions()
  }
  catch (error: any) {
    if (error.response?.status === 500)
      nui_msg.error(t('error.syncChatSession'))
    // eslint-disable-next-line no-console
    console.log(error)
  }
}

async function handleSelect({ uuid }: Chat.History) {
  if (isActive(uuid))
    return

  if (chatStore.active)
    chatStore.updateChatSession(chatStore.active, { isEdit: false })

  await chatStore.setActive(uuid)

  if (isMobile.value)
    appStore.setSiderCollapsed(true)
}

function handleEdit({ uuid }: Chat.History, isEdit: boolean, event?: MouseEvent) {
  event?.stopPropagation()
  chatStore.updateChatSession(uuid, { isEdit })
}
function handleSave({ uuid, title }: Chat.History, isEdit: boolean, event?: MouseEvent) {
  event?.stopPropagation()
  chatStore.updateChatSession(uuid, { isEdit })
  // should move to store
  renameChatSession(uuid, title)
}

function handleDelete(index: number, event?: MouseEvent | TouchEvent) {
  event?.stopPropagation()
  chatStore.deleteChatSession(index)
}

function handleEnter({ uuid, title }: Chat.History, isEdit: boolean, event: KeyboardEvent) {
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
  <NScrollbar class="px-4">
    <div>
      <template v-if="!dataSources.length">
        <div class="flex flex-col gap-2 text-sm">
          <div class="flex flex-col items-center mt-4 text-center text-neutral-300">
            <SvgIcon icon="ri:inbox-line" class="mb-2 text-3xl" />
            <span>{{ $t('common.noData') }}</span>
          </div>
        </div>
      </template>
      <template v-else>
        <div id="chat-menu" class="flex flex-col gap-2 text-sm">
          <div v-for="(item, index) of dataSources" :key="index">
            <a
              class="chat-menu-item relative flex items-center gap-3 px-3 py-3 break-all border rounded-md cursor-pointer hover:bg-neutral-100 group dark:border-neutral-800 dark:hover:bg-[#24272e]"
              :class="isActive(item.uuid) && ['border-[#4b9e5f]', 'bg-neutral-100', 'text-[#4b9e5f]', 'dark:bg-[#24272e]', 'dark:border-[#4b9e5f]', 'pr-14']"
              @click="handleSelect(item)"
            >
              <span>
                <SvgIcon icon="ri:message-3-line" />
              </span>
              <div class="relative flex-1 overflow-hidden break-all text-ellipsis whitespace-nowrap">
                <NInput
                  v-if="item.isEdit" v-model:value="item.title" data-testid="edit_session_topic_input" size="tiny"
                  @keypress="handleEnter(item, false, $event)"
                />
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
                  <NPopconfirm
                    placement="bottom" data-testid="confirm_delete_session"
                    @positive-click="handleDelete(index, $event)"
                  >
                    <template #trigger>
                      <button class="p-1">
                        <SvgIcon icon="ri:delete-bin-line" />
                      </button>
                    </template>
                    {{ $t('chat.deleteChatSessionsConfirm') }}
                  </NPopconfirm>
                </template>
              </div>
            </a>
          </div>
        </div>
      </template>
    </div>
  </NScrollbar>
</template>

<style>
.sortable-ghost {
  opacity: 0.5;
  background-color: #fff;
}

.drag-handle {
  cursor: move;
}
</style>
