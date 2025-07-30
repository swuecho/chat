<template>
  <div class="artifact-content">
    <!-- Code Artifact -->
    <div v-if="artifact.type === 'code' || artifact.type === 'executable-code'" class="code-artifact">
      <div v-if="isEditing" class="code-editor">
        <textarea
          :value="editableContent"
          @input="$emit('update-editable-content', artifact.uuid, $event.target.value)"
          class="code-textarea"
          :style="{ height: `${Math.max(200, editableContent.split('\n').length * 20)}px` }"
        />
        <div class="editor-actions">
          <NButton size="small" @click="$emit('save-edit', artifact.uuid)" type="primary">
            Save
          </NButton>
          <NButton size="small" @click="$emit('cancel-edit', artifact.uuid)">
            Cancel
          </NButton>
        </div>
      </div>
      <div v-else class="code-display">
        <pre><code :class="`language-${artifact.language || 'javascript'}`">{{ artifact.content }}</code></pre>
        <div class="code-actions">
          <NButton size="small" @click="$emit('toggle-edit', artifact.uuid, artifact.content)">
            <Icon icon="ri:edit-line" />
            Edit
          </NButton>
        </div>
      </div>
    </div>

    <!-- HTML Artifact -->
    <div v-else-if="artifact.type === 'html'" class="html-artifact">
      <iframe
        :srcdoc="artifact.content"
        class="html-iframe"
        sandbox="allow-scripts"
      />
    </div>

    <!-- SVG Artifact -->
    <div v-else-if="artifact.type === 'svg'" class="svg-artifact">
      <div v-html="artifact.content" class="svg-content" />
    </div>

    <!-- Mermaid Artifact -->
    <div v-else-if="artifact.type === 'mermaid'" class="mermaid-artifact">
      <div class="mermaid-content">{{ artifact.content }}</div>
    </div>

    <!-- JSON Artifact -->
    <div v-else-if="artifact.type === 'json'" class="json-artifact">
      <pre><code class="language-json">{{ formatJson(artifact.content) }}</code></pre>
    </div>

    <!-- Markdown Artifact -->
    <div v-else-if="artifact.type === 'markdown'" class="markdown-artifact">
      <div class="markdown-content" v-html="renderedMarkdown" />
    </div>

    <!-- Execution Output -->
    <div v-if="executionOutputs && executionOutputs.length > 0" class="execution-output">
      <div class="output-header">
        <h4>Execution Output</h4>
      </div>
      <div class="output-content">
        <div v-for="output in executionOutputs" :key="output.id" class="output-item">
          <div class="output-meta">
            <span class="output-type">{{ output.type }}</span>
            <span class="output-time">{{ formatTime(output.timestamp) }}</span>
          </div>
          <pre class="output-text">{{ output.content }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { NButton } from 'naive-ui'
import { Icon } from '@iconify/vue'
import { type Artifact, type ExecutionResult } from '@/typings/chat'
import MarkdownIt from 'markdown-it'

interface Props {
  artifact: Artifact
  isEditing: boolean
  editableContent: string
  executionOutputs: ExecutionResult[]
  showLibraryList: boolean
  canvasRefs: Record<string, HTMLCanvasElement>
  matplotlibRefs: Record<string, HTMLImageElement>
}

const props = defineProps<Props>()

defineEmits<{
  'toggle-edit': [uuid: string, content: string]
  'save-edit': [uuid: string]
  'cancel-edit': [uuid: string]
  'update-editable-content': [uuid: string, content: string]
  'toggle-library-list': [uuid: string]
  'run-code': [artifact: Artifact]
}>()

// Markdown rendering
const mdi = new MarkdownIt()

const renderedMarkdown = computed(() => {
  return mdi.render(props.artifact.content)
})

// Utility functions
const formatJson = (jsonString: string) => {
  try {
    const parsed = JSON.parse(jsonString)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return jsonString
  }
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleTimeString()
}
</script>

<style scoped>
.artifact-content {
  padding: 1rem;
}

/* Code Artifact */
.code-editor {
  margin-bottom: 1rem;
}

.code-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  resize: vertical;
  background: #f9fafb;
}

.code-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.editor-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.code-display {
  position: relative;
}

.code-display pre {
  margin: 0;
  padding: 0.75rem;
  background: #f8fafc;
  border-radius: 0.375rem;
  overflow-x: auto;
}

.code-actions {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  opacity: 0;
  transition: opacity 0.2s;
}

.code-display:hover .code-actions {
  opacity: 1;
}

/* HTML Artifact */
.html-iframe {
  width: 100%;
  height: 400px;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background: white;
}

/* SVG Artifact */
.svg-content {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.375rem;
}

.svg-content svg {
  max-width: 100%;
  height: auto;
}

/* Mermaid Artifact */
.mermaid-content {
  text-align: center;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.375rem;
}

/* JSON Artifact */
.json-artifact pre {
  margin: 0;
  padding: 0.75rem;
  background: #f8fafc;
  border-radius: 0.375rem;
  overflow-x: auto;
}

/* Markdown Artifact */
.markdown-content {
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.375rem;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3) {
  margin-top: 0;
  margin-bottom: 0.5rem;
  color: #1f2937;
}

.markdown-content :deep(p) {
  margin-bottom: 0.75rem;
}

.markdown-content :deep(code) {
  background: #e5e7eb;
  padding: 0.125rem 0.25rem;
  border-radius: 0.25rem;
  font-size: 0.875em;
}

.markdown-content :deep(pre) {
  background: #f8fafc;
  padding: 0.75rem;
  border-radius: 0.375rem;
  overflow-x: auto;
  margin-bottom: 0.75rem;
}

/* Execution Output */
.execution-output {
  margin-top: 1rem;
  border-top: 1px solid #e5e7eb;
  padding-top: 1rem;
}

.output-header {
  display: flex;
  align-items: center;
  margin-bottom: 0.5rem;
}

.output-header h4 {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.output-content {
  max-height: 300px;
  overflow-y: auto;
}

.output-item {
  margin-bottom: 0.75rem;
  padding: 0.5rem;
  background: #f9fafb;
  border-radius: 0.375rem;
  border-left: 3px solid #6b7280;
}

.output-item:last-child {
  margin-bottom: 0;
}

.output-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.25rem;
}

.output-type {
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.output-type[data-type="log"] {
  color: #3b82f6;
}

.output-type[data-type="error"] {
  color: #ef4444;
}

.output-type[data-type="return"] {
  color: #10b981;
}

.output-time {
  font-size: 0.75rem;
  color: #6b7280;
}

.output-text {
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875rem;
  white-space: pre-wrap;
  word-break: break-all;
}

/* Dark mode */
:deep(.dark) .code-textarea {
  background: #1f2937;
  border-color: #374151;
  color: #f3f4f6;
}

:deep(.dark) .code-display pre {
  background: #1f2937;
}

:deep(.dark) .html-iframe {
  background: #1f2937;
  border-color: #374151;
}

:deep(.dark) .svg-content,
:deep(.dark) .mermaid-content,
:deep(.dark) .json-artifact pre,
:deep(.dark) .markdown-content {
  background: #1f2937;
}

:deep(.dark) .markdown-content h1,
:deep(.dark) .markdown-content h2,
:deep(.dark) .markdown-content h3 {
  color: #f3f4f6;
}

:deep(.dark) .markdown-content code {
  background: #374151;
}

:deep(.dark) .markdown-content pre {
  background: #1f2937;
}

:deep(.dark) .execution-output {
  border-top-color: #374151;
}

:deep(.dark) .output-header h4 {
  color: #d1d5db;
}

:deep(.dark) .output-item {
  background: #1f2937;
  border-left-color: #6b7280;
}

:deep(.dark) .output-time {
  color: #9ca3af;
}
</style>