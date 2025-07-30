<script lang="ts" setup>
import { computed, nextTick } from 'vue'
import { HoverButton, SvgIcon } from '@/components/common'
import { useAppStore, useSessionStore } from '@/store'

interface Emit {
  (ev: 'snapshot'): void
  (ev: 'toggle'): void
  (ev: 'addChat'): void
}

const emit = defineEmits<Emit>()

const appStore = useAppStore()
const sessionStore = useSessionStore()

const collapsed = computed(() => appStore.siderCollapsed)
const currentChatSession = computed(() => sessionStore.activeSession)

function handleUpdateCollapsed() {
  appStore.setSiderCollapsed(!collapsed.value)
}

function onScrollToTop() {
  const scrollRef = document.querySelector('#scrollRef')
  if (scrollRef)
    nextTick(() => scrollRef.scrollTop = 0)
}

function toggle() {
  emit('toggle')
}

function handleSnapshot() {
  emit('snapshot')
}

function handleAdd() {
  emit('addChat')
}
</script>

<template>
  <header
    class="sticky top-0 left-0 right-0 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 backdrop-blur">
    <div class="relative flex items-center justify-between min-w-0 overflow-hidden h-14">
      <div class="flex items-center">
        <button class="flex items-center justify-center w-11 h-11" @click="handleUpdateCollapsed">
          <SvgIcon v-if="collapsed" class="text-2xl" icon="ri:align-justify" />
          <SvgIcon v-else class="text-2xl" icon="ri:align-right" />
        </button>
      </div>
      <h1 class="flex-1 px-4 pr-6 overflow-hidden cursor-pointer select-none text-ellipsis whitespace-nowrap"
        @dblclick="onScrollToTop">
        {{ currentChatSession?.title ?? '' }}
      </h1>
      <div class="flex items-center space-x-2">
        <HoverButton :tooltip="$t('chat.chatSnapshot')" @click="handleSnapshot">
          <span class="text-xl text-[#4b9e5f] dark:text-white">
            <SvgIcon icon="ic:twotone-ios-share" />
          </span>
        </HoverButton>
        <HoverButton :tooltip="$t('chat.adjustParameters')" @click="toggle">
          <span class="text-xl text-[#4b9e5f]">
            <SvgIcon icon="teenyicons:adjust-horizontal-solid" />
          </span>
        </HoverButton>
        <HoverButton :tooltip="$t('chat.new')" @click="handleAdd">
          <span class="text-xl text-[#4b9e5f]">
            <SvgIcon icon="material-symbols:add-circle-outline" />
          </span>
        </HoverButton>
   
      </div>
    </div>
  </header>
</template>
