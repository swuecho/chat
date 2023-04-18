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
    class="sticky top-0 left-0 right-0 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 backdrop-blur">
    <div class="relative flex items-center justify-between min-w-0 overflow-hidden h-14">
      <h1 class="flex-1 ml-5 px-4 pr-6 overflow-hidden  text-ellipsis whitespace-nowrap" contenteditable @blur="handleEdit">
        {{ title ?? '' }}
      </h1>
      <div class="flex mr-5 items-center space-x-2">
        <HoverButton @click="handleHome">
          <span class="text-xl text-[#4f555e] dark:text-white">
            <SvgIcon icon="ic:baseline-home" />
          </span>
        </HoverButton>
      </div>
    </div>
  </header>
</template>

<style>
h1[contenteditable] {
  padding: 0.15rem 0.5rem;
  border-radius: 0.15rem;
}

h1[contenteditable]:focus {
  outline: none;
  box-shadow: 0 0 0 2px #01bc77;
}
</style>
