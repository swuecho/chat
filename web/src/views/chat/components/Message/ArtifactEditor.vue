<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { NButton, NInput, NSelect, useMessage } from 'naive-ui'
import { Icon } from '@iconify/vue'

const props = defineProps<{ modelValue: string; language: string; title?: string }>()
const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:language': [value: string]
  'update:title': [value: string]
}>()

const message = useMessage()
const content = ref(props.modelValue)
watch(() => props.modelValue, value => content.value = value)
watch(content, value => emit('update:modelValue', value))

const languageOptions = [
  'text', 'javascript', 'typescript', 'python', 'go', 'html', 'css', 'svg', 'mermaid', 'json', 'markdown',
].map(value => ({ label: value, value }))
const lineCount = computed(() => content.value ? content.value.split('\n').length : 0)

async function copyContent() {
  try {
    await navigator.clipboard.writeText(content.value)
    message.success('Content copied')
  }
  catch {
    message.error('Unable to copy content')
  }
}

function downloadContent() {
  const extensions: Record<string, string> = { javascript: 'js', typescript: 'ts', python: 'py', markdown: 'md', text: 'txt' }
  const extension = extensions[props.language] || props.language || 'txt'
  const safeTitle = (props.title || 'artifact').replace(/[^a-z0-9_-]+/gi, '-').replace(/^-|-$/g, '') || 'artifact'
  const url = URL.createObjectURL(new Blob([content.value], { type: 'text/plain;charset=utf-8' }))
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `${safeTitle}.${extension}`
  anchor.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div class="artifact-editor">
    <div class="artifact-fields">
      <NInput :value="title" :placeholder="$t('artifact.titlePlaceholder')" @update:value="$emit('update:title', $event)" />
      <NSelect :value="language" :options="languageOptions" filterable tag @update:value="$emit('update:language', $event)" />
    </div>
    <NInput
      v-model:value="content" type="textarea" :placeholder="$t('artifact.contentPlaceholder')"
      :autosize="{ minRows: 14, maxRows: 28 }" class="artifact-content-input" spellcheck="false"
    />
    <div class="artifact-footer">
      <span>{{ $t('artifact.editorStats', { lines: lineCount, characters: content.length }) }}</span>
      <div class="artifact-actions">
        <NButton size="small" @click="copyContent">
          <template #icon>
            <Icon icon="ri:file-copy-line" />
          </template>{{ $t('artifact.copy') }}
        </NButton>
        <NButton size="small" @click="downloadContent">
          <template #icon>
            <Icon icon="ri:download-line" />
          </template>{{ $t('artifact.download') }}
        </NButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.artifact-editor { display: grid; gap: 12px; }
.artifact-fields { display: grid; grid-template-columns: minmax(0, 1fr) 220px; gap: 12px; }
.artifact-content-input :deep(textarea) { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
.artifact-footer { display: flex; align-items: center; justify-content: space-between; gap: 12px; color: var(--n-text-color-3); font-size: 12px; }
.artifact-actions { display: flex; gap: 8px; }
@media (max-width: 640px) {
  .artifact-fields { grid-template-columns: 1fr; }
  .artifact-footer { align-items: flex-start; flex-direction: column; }
}
</style>
