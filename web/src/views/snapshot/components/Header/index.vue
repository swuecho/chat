<script lang="ts" setup>
import { useRoute } from 'vue-router'
import { nextTick, ref } from 'vue'
import { HoverButton, SvgIcon } from '@/components/common'
import { updateChatSnapshot } from '@/api'

defineProps<Props>()

const route = useRoute()

interface Props {
  title: string
}

const { uuid } = route.params as { uuid: string }

const isEditing = ref<boolean>(false)

const titleRef = ref(null)

function handleHome() {
  window.open('#/snapshot_all', '_blank')
}

async function handleEdit(e: Event) {
  const title_value = (e.target as HTMLInputElement).innerText
  isEditing.value = false
  await updateChatSnapshot(uuid, { title: title_value })
}

async function handleEditTitle() {
  isEditing.value = true
  await nextTick()
  if (titleRef.value)
    // @ts-expect-error focus is ok
    titleRef.value.focus()
}
</script>

<template>
  <header
    class="sticky top-0 left-0 right-0 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 dark:text-white backdrop-blur"
  >
    <div class="relative flex items-center justify-between min-w-0 overflow-hidden h-14">
      <span class="ml-5">
        <HoverButton :tooltip="$t('common.edit')" @click="handleEditTitle">
          <SvgIcon icon="ic:baseline-edit" />
        </HoverButton>
      </span>
      <h1
        ref="titleRef" class="flex-1 overflow-hidden text-ellipsis whitespace-nowrap"
        :class="[isEditing ? 'shadow-green-100' : '']" :contenteditable="isEditing" @blur="handleEdit"
        @dblclick="handleEditTitle"
      >
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
  box-shadow: 0 0 0 1px #18a058;
}
</style>
