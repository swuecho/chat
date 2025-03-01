<script setup lang='ts'>
import type { CSSProperties } from 'vue'
import { computed, watch, ref } from 'vue'

import { NButton, NLayoutSider, NTooltip } from 'naive-ui'
import List from './List.vue'
import Footer from './Footer.vue'
import { useAppStore, useChatStore } from '@/store'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'
import { SvgIcon } from '@/components/common'
import { getChatSessionDefault } from '@/api'
import { PromptStore } from '@/components/common'

const appStore = useAppStore()
const chatStore = useChatStore()

const { isMobile, isBigScreen } = useBasicLayout()
const show = ref(false)

const collapsed = computed(() => appStore.siderCollapsed)

async function handleAdd() {
  const new_chat_text = t('chat.new')
  //   //
  //   // {
  //   "uuid": "ff511942-43b4-4ebd-9fe8-f2ca405605ac",
  //   "isEdit": false,
  //   "title": "try improve code and",
  //   "maxLength": 10,
  //   "temperature": 1,
  //   "topP": 1,
  //   "maxTokens": 512,
  //   "debug": false
  // }//
  const default_model_parameters = await getChatSessionDefault(new_chat_text)
  await chatStore.addChatSession(default_model_parameters)
  if (isMobile.value)
    appStore.setSiderCollapsed(true)
}

function handleUpdateCollapsed() {
  appStore.setSiderCollapsed(!collapsed.value)
}

const getMobileClass = computed<CSSProperties>(() => {
  if (isMobile.value) {
    return {
      position: 'fixed',
      zIndex: 50,
    }
  }
  return {}
})

const mobileSafeArea = computed(() => {
  if (isMobile.value) {
    return {
      paddingBottom: 'env(safe-area-inset-bottom)',
    }
  }
  return {}
})


watch(
  isMobile,
  (val) => {
    appStore.setSiderCollapsed(val)
  },
  {
    immediate: true,
    flush: 'post',
  },
)



function openBotAll() {
  window.open('/#/bot_all', '_blank')
}

function openAllSnapshot() {
  window.open('/#/snapshot_all', '_blank')
}

</script>

<template>
  <NLayoutSider
    :collapsed="collapsed" :collapsed-width="0"  :width="isBigScreen ? 360 : 260" :show-trigger="isMobile ? false : 'arrow-circle'"
    collapse-mode="transform" position="absolute" bordered :style="getMobileClass"
    @update-collapsed="handleUpdateCollapsed"
  >
    <div class="flex flex-col h-full" :style="mobileSafeArea">
      <main class="flex flex-col flex-1 min-h-0">
        <div class="p-2">
          <NButton dashed block @click="handleAdd">
            <SvgIcon icon="material-symbols:add-circle-outline" /> {{ $t('chat.new') }}
          </NButton>
        </div>
        <div class="flex-1 min-h-0 pb-4 overflow-hidden">
          <List />
        </div>
        <div class="px-2 pb-2">
          <NButtonGroup class="w-full flex">
            <NTooltip placement="bottom">
              <template #trigger>
                <NButton 
                  class="flex-1 !rounded-r-none"
                  @click="openAllSnapshot"
                >
                  <template #icon>
                    <SvgIcon icon="ri:history-line" />
                  </template>
                </NButton>
              </template>
              {{ t('chat_snapshot.title') }}
            </NTooltip>
            
            <NTooltip placement="bottom">
              <template #trigger>
                <NButton 
                  class="flex-1 !rounded-none"
                  @click="openBotAll"
                >
                  <template #icon>
                    <SvgIcon icon="ri:robot-line" />
                  </template>
                </NButton>
              </template>
              {{ t('bot.list') }}
            </NTooltip>
            
            <NTooltip placement="bottom">
              <template #trigger>
                <NButton 
                  class="flex-1 !rounded-l-none"
                  @click="show = true"
                >
                  <template #icon>
                    <SvgIcon icon="ri:lightbulb-line" />
                  </template>
                </NButton>
              </template>
              {{ t('prompt.store') }}
            </NTooltip>
          </NButtonGroup>
        </div>
      </main>
      <Footer />
    </div>
  </NLayoutSider>
  <template v-if="isMobile">
    <div v-show="!collapsed" class="fixed inset-0 z-40 w-full h-full bg-black/40" @click="handleUpdateCollapsed" />
  </template>
  <PromptStore v-model:visible="show" />
</template>
