<script lang="ts" setup>
import { ref, computed } from 'vue'
import { NCard, NButton, NSpace, NTabs, NTabPane } from 'naive-ui'
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

interface Bot {
  title: string
  uuid: string
  typ: string
}

// Get bot prompts
const botPrompts = computed(() => {
  return (bots.value || [])
    .filter((bot: Bot) => bot.typ === 'chatbot')
    .map((bot: Bot) => ({
      key: bot.title,
      uuid: bot.uuid,
      value: ''
    }))
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

const renderPromptCards = (prompts: any[]) => (
  <NSpace
    :wrap="true"
    :wrap-item="true"
    :size="[16, 16]"
    :item-style="{ width: isMobile ? '100%' : 'calc(50% - 8px)' }"
  >
    {prompts.map(prompt => (
      <NCard
        key={prompt.key}
        title={prompt.key}
        hoverable
        embedded
      >
        {{
          headerExtra: () => (
            <NButton
              type="primary"
              size="small"
              onClick={() => handleUsePrompt(prompt.key, prompt.value, prompt?.uuid)}
            >
              使用
            </NButton>
          ),
          default: () => (
            <div class="line-clamp-2 leading-6 overflow-hidden text-ellipsis">
              {prompt.value}
            </div>
          )
        }}
      </NCard>
    ))}
  </NSpace>
)
</script>

<template>
  <NTabs type="line" animated>
    <NTabPane v-if="botPrompts.length > 0" name="bots" tab="Bots">
      <div class="mt-4">
        {{ renderPromptCards(botPrompts) }}
      </div>
    </NTabPane>
    <NTabPane name="prompts" tab="Prompts">
      <div class="mt-4">
        {{ renderPromptCards(promptStore.promptList) }}
      </div>
    </NTabPane>
  </NTabs>
</template>

