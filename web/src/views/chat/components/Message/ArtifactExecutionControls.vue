<template>
  <div v-if="isExecutable" class="execution-controls">
    <div class="control-bar">
      <NButton 
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
        size="small" 
        @click="$emit('clear-output')" 
        :disabled="!hasOutput">
        <template #icon>
          <Icon icon="ri:delete-bin-line" />
        </template>
        Clear Output
      </NButton>
      
      <NButton 
        size="small" 
        @click="$emit('toggle-editor')" 
        type="tertiary">
        <template #icon>
          <Icon :icon="isEditing ? 'ri:eye-line' : 'ri:edit-line'" />
        </template>
        {{ isEditing ? 'View' : 'Edit' }}
      </NButton>
    </div>
    
    <div v-if="lastExecution" class="execution-info">
      <span class="execution-time">
        {{ lastExecution.execution_time_ms }}ms
      </span>
      <span class="execution-status" :class="executionStatus">
        {{ executionStatus }}
      </span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { NButton } from 'naive-ui'
import { Icon } from '@iconify/vue'
/// <reference path="@/typings/chat.d.ts" />
type Artifact = Chat.Artifact
type ExecutionResult = Chat.ExecutionResult

interface Props {
  artifact: Artifact
  isExecutable: boolean
  isRunning: boolean
  hasOutput: boolean
  isEditing: boolean
  lastExecution?: ExecutionResult
}

const props = defineProps<Props>()

defineEmits<{
  'run-code': [artifact: Artifact]
  'clear-output': []
  'toggle-editor': []
}>()

const executionStatus = computed(() => {
  if (!props.lastExecution) return ''
  return props.lastExecution.type === 'error' ? 'error' : 'success'
})
</script>

<style scoped>
.execution-controls {
  margin-bottom: 1rem;
  padding: 0.75rem;
  background: #f8fafc;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
}

.control-bar {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.execution-info {
  display: flex;
  gap: 1rem;
  font-size: 0.75rem;
  color: #64748b;
}

.execution-time {
  font-weight: 500;
}

.execution-status {
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.05em;
}

.execution-status.success {
  color: #059669;
}

.execution-status.error {
  color: #dc2626;
}

:deep(.dark) .execution-controls {
  background: #1e293b;
  border-color: #334155;
}

:deep(.dark) .execution-info {
  color: #94a3b8;
}
</style>