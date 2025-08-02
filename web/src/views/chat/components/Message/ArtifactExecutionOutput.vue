<template>
  <div v-if="hasOutput" class="execution-output">
    <div class="output-header">
      <span class="output-title">
        <Icon icon="ri:terminal-line" />
        Output
      </span>
      <NButton size="tiny" text @click="$emit('clear-output')">
        <Icon icon="ri:close-line" />
      </NButton>
    </div>
    
    <div class="output-content">
      <div v-for="result in executionResults" 
           :key="result.id" 
           class="output-line"
           :class="result.type">
        <span class="output-type">{{ result.type }}</span>
        
        <!-- Canvas output -->
        <div v-if="result.type === 'canvas'" class="canvas-output">
          <canvas 
            :ref="(el) => $emit('set-canvas-ref', result.id, el as HTMLCanvasElement)"
            class="result-canvas"
          ></canvas>
        </div>
        
        <!-- Matplotlib plot output -->
        <div v-else-if="result.type === 'matplotlib'" class="matplotlib-output">
          <img 
            :ref="(el) => $emit('set-matplotlib-ref', result.id, el as HTMLImageElement)"
            class="result-plot"
            alt="Matplotlib plot"
          />
        </div>
        
        <!-- Regular text output -->
        <span v-else class="output-text">{{ result.content }}</span>
        
        <span class="output-time">{{ formatTime(result.timestamp) }}</span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { NButton } from 'naive-ui'
import { Icon } from '@iconify/vue'
/// <reference path="@/typings/chat.d.ts" />
interface ExecutionResult {
  id: string
  type: 'log' | 'error' | 'return' | 'stdout' | 'warn' | 'info' | 'debug' | 'canvas' | 'matplotlib'
  content: string
  timestamp: string
  execution_time_ms?: number
}

interface Props {
  hasOutput: boolean
  executionResults: ExecutionResult[]
}

defineProps<Props>()

defineEmits<{
  'clear-output': []
  'set-canvas-ref': [id: string, el: HTMLCanvasElement | null]
  'set-matplotlib-ref': [id: string, el: HTMLImageElement | null]
}>()

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleTimeString()
}
</script>

<style scoped>
.execution-output {
  margin-top: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  overflow: hidden;
}

.output-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0.75rem;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.output-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #475569;
}

.output-content {
  max-height: 400px;
  overflow-y: auto;
  padding: 0.5rem;
}

.output-line {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.75rem;
  padding: 0.5rem;
  background: #f9fafb;
  border-radius: 0.375rem;
  border-left: 3px solid #6b7280;
}

.output-line:last-child {
  margin-bottom: 0;
}

.output-line.log {
  border-left-color: #3b82f6;
}

.output-line.error {
  border-left-color: #ef4444;
  background: #fef2f2;
}

.output-line.return {
  border-left-color: #10b981;
}

.output-type {
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #6b7280;
}

.output-line.log .output-type {
  color: #3b82f6;
}

.output-line.error .output-type {
  color: #ef4444;
}

.output-line.return .output-type {
  color: #10b981;
}

.output-text {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875rem;
  white-space: pre-wrap;
  word-break: break-all;
  color: #374151;
}

.output-time {
  font-size: 0.75rem;
  color: #6b7280;
  align-self: flex-end;
}

.canvas-output,
.matplotlib-output {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
  background: white;
  border-radius: 0.25rem;
}

.result-canvas,
.result-plot {
  max-width: 100%;
  height: auto;
  border-radius: 0.25rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
}

:deep(.dark) .execution-output {
  border-color: #374151;
}

:deep(.dark) .output-header {
  background: #374151;
  border-bottom-color: #4b5563;
}

:deep(.dark) .output-title {
  color: #d1d5db;
}

:deep(.dark) .output-line {
  background: #1f2937;
}

:deep(.dark) .output-line.error {
  background: #4b1414;
}

:deep(.dark) .output-text {
  color: #f3f4f6;
}

:deep(.dark) .output-time {
  color: #9ca3af;
}

:deep(.dark) .canvas-output,
:deep(.dark) .matplotlib-output {
  background: #111827;
}
</style>