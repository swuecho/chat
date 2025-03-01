<script setup lang="ts">
import { SvgIcon } from '@/components/common'
import ClaudeIcon from '@/assets/svgs/claude.svg'
import GptIcon from '@/assets/svgs/gpt.svg'
import GeminiIcon from '@/assets/svgs/gemini.svg'
import DeepseekIcon from '@/assets/svgs/deepseek.svg'

defineProps({
  model: String,
})
</script>

<template>
  <div>
    <template v-if="model?.startsWith('claude')">
      <ClaudeIcon />
    </template>
    <template v-else-if="model === 'claude-instant-1'">
      <SvgIcon icon="simple-icons:fastapi" color="brown" />
    </template>
    <template v-else-if="model?.startsWith('gpt-4')">
      <SvgIcon icon="material-symbols:filter-4-rounded" color="green" />
    </template>
    <template v-else-if="model?.includes('davinci')">
      <SvgIcon icon="ph:file-code-bold" color="green" />
    </template>
    <template v-else-if="model?.startsWith('gpt')">
      <GptIcon />
    </template>
    <template v-else-if="model?.includes('gemini')">
      <GeminiIcon />
    </template>
    <template v-else-if="model?.includes('deepseek')">
      <DeepseekIcon />
    </template>
    <template v-else>
      <SvgIcon icon="ri:question-answer-line" width="0.8em" />
    </template>
  </div>
</template>
