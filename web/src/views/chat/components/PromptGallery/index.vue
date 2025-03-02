<script lang="ts" setup>
import { ref } from 'vue'
import { NCard, NButton, NSpace } from 'naive-ui'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { usePromptStore } from '@/store/modules'

interface Emit {
  (ev: 'usePrompt', key: string, prompt: string): void
}

const emit = defineEmits<Emit>()
const { isMobile } = useBasicLayout()
const promptStore = usePromptStore()
const promptList = ref(promptStore.promptList)

const handleUsePrompt = (key: string, prompt: string) => {
  emit('usePrompt', key, prompt)
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
            @click="handleUsePrompt(prompt.key, prompt.value)"
          >
            使用
          </NButton>
        </template>
        <div class="prompt-value">
          {{ prompt.value.length > 100 ? prompt.value.substring(0, 100) + '...' : prompt.value }}
        </div>
      </NCard>
    </NSpace>
  </NSpace>
</template>

<style scoped>
.prompt-value {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;  
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.5;
  max-height: 3em;
}
</style>
