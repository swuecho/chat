<script lang="ts" setup>
import { NCard, NButton, NSpace } from 'naive-ui'
import { useBasicLayout } from '@/hooks/useBasicLayout'

const { isMobile } = useBasicLayout()

defineProps<{
  prompts: any[]
}>()

const emit = defineEmits<{
  (ev: 'usePrompt', key: string, prompt: string, uuid?: string): void
}>()
</script>

<template>
  <NSpace
    :wrap="true"
    :wrap-item="true"
    :size="[16, 16]"
    :item-style="{ width: isMobile ? '100%' : 'calc(50% - 8px)' }"
  >
    <NCard
      v-for="prompt in prompts"
      :key="prompt.key"
      :title="prompt.key"
      hoverable
      embedded
    >
      <template #header-extra>
        <NButton
          type="primary"
          size="small"
          @click="emit('usePrompt', prompt.key, prompt.value, prompt?.uuid)"
        >
          使用
        </NButton>
      </template>
      <div class="line-clamp-2 leading-6 overflow-hidden text-ellipsis">
        {{ prompt.value }}
      </div>
    </NCard>
  </NSpace>
</template>
