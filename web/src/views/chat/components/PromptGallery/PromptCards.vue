<script lang="ts" setup>
import { NCard, NButton, NSpace } from 'naive-ui'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { SvgIcon } from '@/components/common'

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
      hoverable
      embedded
      class="hover:shadow-lg transition-shadow duration-200 dark:bg-neutral-800"
    >
      <template #header>
        <div class="line-clamp-1 overflow-hidden text-ellipsis font-medium text-gray-900 dark:text-gray-100">
          {{ prompt.key }}
        </div>
      </template>
      <template #header-extra>
        <NButton
          type="primary"
          size="small"
          class="!bg-primary-500 hover:!bg-primary-600 dark:!bg-primary-600 dark:hover:!bg-primary-700"
          @click="emit('usePrompt', prompt.key, prompt.value, prompt?.uuid)"
        >
        <SvgIcon icon="material-symbols:play-arrow" class="w-4 h-4 mr-1" />
        </NButton>
      </template>
      <div class="line-clamp-2 leading-6 overflow-hidden text-ellipsis text-gray-600 dark:text-gray-300">
        {{ prompt.value }}
      </div>
    </NCard>
  </NSpace>
</template>
