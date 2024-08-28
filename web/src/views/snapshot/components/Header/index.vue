<script lang="ts" setup>
import { useRoute } from 'vue-router'
import { nextTick, ref } from 'vue'
import { HoverButton, SvgIcon } from '@/components/common'
import { updateChatSnapshot } from '@/api'
import { useMutation, useQueryClient } from '@tanstack/vue-query'

const queryClient = useQueryClient()

const props = defineProps<Props>()

const route = useRoute()

interface Props {
  title: string
  typ: string
}

const { uuid } = route.params as { uuid: string }

const isEditing = ref<boolean>(false)

const titleRef = ref(null)

function handleHome() {
  const typ = props.typ
  if (typ === 'snapshot') {
    window.open('#/snapshot_all', '_blank')
  } else if (typ === 'chatbot') {
    window.open('#/bot_all', '_blank')
  }
}

function handleChatHome() {
  window.open('/#/chat/', '_blank')
}

const { mutate } = useMutation({
  mutationFn: async (variables: { uuid: string, title: string }) => await updateChatSnapshot(variables.uuid, { title: variables.title }),
  onSuccess: (data) => {
    queryClient.setQueriesData({ queryKey: ['chatSnapshot', uuid] }, data)
  },
})

const updateTitle = (uuid: string, title: string) => {
  mutate({ uuid: uuid, title: title })
}


async function handleEdit(e: Event) {
  const title_value = (e.target as HTMLInputElement).innerText
  updateTitle(uuid, title_value)
  isEditing.value = false
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
    class="sticky h-16 flex items-center justify-between border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 dark:text-white backdrop-blur  overflow-hidden">
    <div class="flex items-center ml-10">
      <div>
        <HoverButton :tooltip="$t('common.edit')" @click="handleEditTitle">
          <SvgIcon icon="ic:baseline-edit" />
        </HoverButton>
      </div>
      <h1 ref="titleRef" class="flex-1 overflow-hidden text-ellipsis whitespace-nowrap"
        :class="[isEditing ? 'shadow-green-100' : '']" :contenteditable="isEditing" @blur="handleEdit"
        @dblclick="handleEditTitle">
        {{ title ?? '' }}
      </h1>
    </div>
    <div class="flex mr-10 items-center space-x-4">
      <HoverButton @click="handleHome">
        <span class="text-2xl text-[#4f555e] dark:text-white">
          <SvgIcon icon="carbon:table-of-contents" />
        </span>
      </HoverButton>
      <HoverButton @click="handleChatHome">
        <span class="text-2xl text-[#4f555e] dark:text-white">
          <SvgIcon icon="ic:baseline-home" />
        </span>
      </HoverButton>
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
