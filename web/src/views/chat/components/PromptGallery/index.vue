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



// 移动端自适应相关
const { isMobile } = useBasicLayout()

const promptStore = usePromptStore()
const promptList = ref<any>(promptStore.promptList)

interface DataProps {
        renderKey: string
        renderValue: string
        key: string
        value: string
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
        {{ prompt.value }}
      </NCard>
    </NSpace>
  </NSpace>
</template>
