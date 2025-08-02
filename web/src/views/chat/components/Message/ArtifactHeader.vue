<template>
  <div class="artifact-header">
    <div class="artifact-title">
      <Icon :icon="getArtifactIcon(artifact.type)" class="artifact-icon" />
      <span class="artifact-title-text">{{ artifact.title }}</span>
      <span class="artifact-type">({{ artifact.type }})</span>
    </div>
    <div class="artifact-actions">
      <NButton size="small" @click="$emit('toggle-expand', artifact.uuid)">
        <span class="hidden sm:inline">{{ isExpanded ? 'Collapse' : 'Expand' }}</span>
        <Icon :icon="isExpanded ? 'ri:arrow-up-line' : 'ri:arrow-down-line'" class="sm:hidden" />
      </NButton>
      
      <NButton 
        v-if="artifact.type === 'html'" 
        size="small" 
        @click="$emit('open-in-new-window', artifact.content)"
        title="Open in new window for debugging">
        <Icon icon="ri:external-link-line" />
      </NButton>
      
      <NButton size="small" @click="$emit('copy-content', artifact.content)">
        <Icon icon="ri:file-copy-line" />
      </NButton>
      
      <NButton 
        v-if="isExecutable(artifact)" 
        size="small" 
        type="primary" 
        @click="$emit('run-code', artifact)"
        :disabled="isRunning"
        :loading="isRunning">
        <template #icon>
          <Icon icon="ri:play-line" />
        </template>
        {{ isRunning ? 'Running...' : 'Run Code' }}
      </NButton>
      
      <NButton 
        v-if="isExecutable && hasOutput" 
        size="small" 
        @click="$emit('clear-output', artifact.uuid)">
        <template #icon>
          <Icon icon="ri:delete-bin-line" />
        </template>
        Clear Output
      </NButton>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Icon } from '@iconify/vue'
import { NButton } from 'naive-ui'
import { type Artifact } from '@/typings/chat'

interface Props {
  artifact: Artifact
  isExpanded: boolean
  isRunning: boolean
  hasOutput: boolean
}

defineProps<Props>()

defineEmits<{
  'toggle-expand': [uuid: string]
  'run-code': [artifact: Artifact]
  'clear-output': [uuid: string]
  'copy-content': [content: string]
  'open-in-new-window': [content: string]
}>()

// Artifact type utilities
const isExecutable = (artifact: Artifact) => {
  return artifact.type === 'executable-code' && artifact.isExecutable
}

const getArtifactIcon = (type: string) => {
  const iconMap: Record<string, string> = {
    'code': 'ri:code-line',
    'executable-code': 'ri:code-s-slash-line',
    'html': 'ri:html5-line',
    'svg': 'ri:svg-line',
    'mermaid': 'ri:git-branch-line',
    'json': 'ri:json-line',
    'markdown': 'ri:markdown-line',
  }
  return iconMap[type] || 'ri:file-line'
}
</script>

<style scoped>
.artifact-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.artifact-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.artifact-icon {
  width: 1.25rem;
  height: 1.25rem;
  color: #6b7280;
}

.artifact-title-text {
  font-weight: 500;
  color: #1f2937;
}

.artifact-type {
  font-size: 0.75rem;
  color: #6b7280;
  background: #e5e7eb;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
}

.artifact-actions {
  display: flex;
  gap: 0.5rem;
}

:deep(.dark) .artifact-header {
  background: #374151;
  border-bottom-color: #4b5563;
}

:deep(.dark) .artifact-title-text {
  color: #f3f4f6;
}

:deep(.dark) .artifact-type {
  color: #9ca3af;
  background: #4b5563;
}

:deep(.dark) .artifact-icon {
  color: #9ca3af;
}
</style>