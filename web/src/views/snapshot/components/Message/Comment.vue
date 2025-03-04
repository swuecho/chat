<script setup lang="ts">
import { NTooltip } from 'naive-ui'
import { displayLocaleDate } from '@/utils/date'

interface Props {
  comment: Chat.Comment
  inversion?: boolean
}

defineProps<Props>()
</script>

<template>
  <div class="comment-item mb-3 p-2 bg-gray-50 dark:bg-gray-600 rounded-lg w-1/2" 
       :class="[inversion ? 'ml-auto' : 'mr-auto']">
    <NTooltip>
      <template #trigger>
        <div class="text-xs text-gray-600 dark:text-gray-300 overflow-hidden whitespace-nowrap overflow-ellipsis">
          <span class="font-medium">{{ comment.authorUsername }}</span>
          <span class="mx-1">•</span>
          <span>{{ displayLocaleDate(comment.createdAt) }}</span>
        </div>
      </template>
      {{ comment.authorUsername }} • {{ displayLocaleDate(comment.createdAt) }}
    </NTooltip>
    <div class="text-sm mt-1 text-gray-800 dark:text-gray-100">
      {{ comment.content }}
    </div>
  </div>
</template>

<style scoped>
.comment-item {
  max-width: 80%;
}
</style>
