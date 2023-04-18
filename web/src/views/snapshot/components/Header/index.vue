<script lang="ts" setup>
import { useRoute } from 'vue-router'
import { HoverButton, SvgIcon } from '@/components/common'
import { updateChatSnapshot } from '@/api'

defineProps<Props>()

const route = useRoute()

interface Props {
  title: string
}

const { uuid } = route.params as { uuid: string }

function handleHome() {
  window.open('#/snapshot_all', '_blank')
}

async function handleEdit(e: Event) {
  const title_value = (e.target as HTMLInputElement).innerText
  await updateChatSnapshot(uuid, { title: title_value })
}
</script>

<template>
  <header
    class="sticky top-0 left-0 right-0 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 backdrop-blur"
  >
    <div class="relative flex items-center justify-between min-w-0 overflow-hidden h-14">
      <h1 class="flex-1 px-4 pr-6 overflow-hidden  text-ellipsis whitespace-nowrap" contenteditable @blur="handleEdit">
        {{ title ?? '' }}
      </h1>
      <div class="flex items-center space-x-2">
        <HoverButton @click="handleHome">
          <span class="text-xl text-[#4f555e] dark:text-white">
            <SvgIcon icon="ic:baseline-home" />
          </span>
        </HoverButton>
      </div>
    </div>
  </header>
</template>
