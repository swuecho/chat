<script lang="ts" setup>
import { ref, computed } from 'vue'
import { NCard, NButton, NSpace } from 'naive-ui'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { usePromptStore } from '@/store/modules'
import { fetchChatSnapshot, fetchSnapshotAll } from '@/api'
import { useQuery } from '@tanstack/vue-query'

interface Emit {
  (ev: 'usePrompt', key: string, prompt: string): void
}

const emit = defineEmits<Emit>()
const { isMobile } = useBasicLayout()
const promptStore = usePromptStore()

// Fetch bots data
const { data: bots } = useQuery({
  queryKey: ['bots'],
  queryFn: async () => await fetchSnapshotAll(),
})


// Combine bots and prompts
interface Bot {
  title: string
  uuid: string
  typ: string
}

const promptList = computed(() => {
  const botPrompts = (bots.value || []).filter((bot: Bot) => bot.typ='chatbot').
  map((bot: Bot) => ({
    key: bot.title,
    uuid: bot.uuid,
    value: ''
  }))
  
  return [...botPrompts, ...promptStore.promptList]
})

const handleUsePrompt = (key: string, prompt: string, uuid?: string) => {
        if (uuid) {
          fetchChatSnapshot(uuid).then((data) => {
            emit('usePrompt', key, data.conversation[0].text)
          })
        } else {
          emit('usePrompt', key, prompt)
        }
}

</script>

<template>
  <NSpace vertical>
    <NSpace
      :wrap="true"
      :wrap-item="true"
      :size="[16, 16]"
      :item-style="{ width: isMobile ? '100%' : 'calc(50% - 8px)' }"
    >
      <NCard
        v-for="prompt in promptList"
        :key="prompt.key"
        :title="prompt.key"
        hoverable
        embedded
      >
        <template #header-extra>
          <NButton
            type="primary"
            size="small"
            @click="handleUsePrompt(prompt.key, prompt.value, prompt?.uuid)"
          >
            使用
          </NButton>
        </template>
        <div class="line-clamp-2 leading-6 overflow-hidden text-ellipsis">
          {{  prompt.value }}
        </div>
      </NCard>
    </NSpace>
  </NSpace>
</template>

