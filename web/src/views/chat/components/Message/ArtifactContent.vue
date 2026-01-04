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
      <div v-html="sanitizedSvg" class="svg-content" />
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
          <template v-if="output.type === 'canvas'">
            <canvas :ref="el => registerCanvas(el, output)" class="output-canvas" />
          </template>
          <template v-else-if="output.type === 'matplotlib'">
            <img :ref="el => registerMatplotlib(el, output)" class="output-image" alt="Matplotlib output" />
          </template>
          <template v-else>
            <template v-if="structuredOutputs[output.id]">
              <div class="output-table-wrapper">
                <table class="output-table">
                  <thead>
                    <tr>
                      <th v-for="header in structuredOutputs[output.id]?.headers" :key="header">
                        {{ header }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(row, rowIndex) in structuredOutputs[output.id]?.rows" :key="rowIndex">
                      <td v-for="(cell, cellIndex) in row" :key="cellIndex">
                        <template v-if="isVfsPath(cell)">
                          <button type="button" class="vfs-path-link" @click="handleOpenVfs(cell)">
                            {{ cell }}
                          </button>
                          <button type="button" class="vfs-copy-button" @click="copyPath(cell)" aria-label="Copy path">
                            <Icon icon="ri:file-copy-line" />
                          </button>
                        </template>
                        <span v-else>{{ cell }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </template>
            <template v-else>
              <pre class="output-text">
                <template v-for="(segment, segIndex) in getOutputSegments(output.content)" :key="segIndex">
                  <span v-if="segment.path" class="vfs-path-inline">
                    <button type="button" class="vfs-path-link" @click="handleOpenVfs(segment.path)">
                      {{ segment.text }}
                    </button>
                    <button type="button" class="vfs-copy-button" @click="copyPath(segment.path)"
                      aria-label="Copy path">
                      <Icon icon="ri:file-copy-line" />
                    </button>
                  </span>
                  <span v-else>{{ segment.text }}</span>
                </template>
              </pre>
            </template>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, inject, nextTick } from 'vue'
import { NButton, useMessage } from 'naive-ui'
import { Icon } from '@iconify/vue'
import { type Artifact, type ExecutionResult } from '@/typings/chat'
import MarkdownIt from 'markdown-it'
import { getCodeRunner } from '@/services/codeRunner'
import { sanitizeHtml, sanitizeSvg } from '@/utils/sanitize'
import { copyText } from '@/utils/format'

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
const message = useMessage()

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
  return sanitizeHtml(mdi.render(props.artifact.content))
})

const openVfsAtPath = inject('openVfsAtPath', null)

const sanitizedSvg = computed(() => {
  return sanitizeSvg(props.artifact.content)
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

type StructuredOutput = {
  headers: string[]
  rows: string[][]
}

type OutputSegment = {
  text: string
  path?: string
}

const maxStructuredRows = 50
const maxStructuredColumns = 20
const vfsPathRegex = /\/(?:data|workspace|tmp|uploads)(?:\/[^\s'"`<>]+)+/g

const parseJsonTable = (content: string): StructuredOutput | null => {
  try {
    const parsed = JSON.parse(content)
    if (Array.isArray(parsed) && parsed.length > 0) {
      if (parsed.every(item => item && typeof item === 'object' && !Array.isArray(item))) {
        const headers = Object.keys(parsed[0]).slice(0, maxStructuredColumns)
        const rows = parsed.slice(0, maxStructuredRows).map(row =>
          headers.map(header => String((row as Record<string, unknown>)[header] ?? ''))
        )
        return { headers, rows }
      }

      if (parsed.every(item => Array.isArray(item))) {
        const headers = parsed[0].slice(0, maxStructuredColumns).map((_, index) => `col_${index + 1}`)
        const rows = parsed.slice(0, maxStructuredRows).map(row =>
          (row as unknown[]).slice(0, maxStructuredColumns).map(cell => String(cell ?? ''))
        )
        return { headers, rows }
      }
    }

    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      const entries = Object.entries(parsed).slice(0, maxStructuredRows)
      return {
        headers: ['key', 'value'],
        rows: entries.map(([key, value]) => [String(key), String(value ?? '')])
      }
    }
  } catch {
    return null
  }

  return null
}

const parseCsvLine = (line: string): string[] => {
  const cells: string[] = []
  let current = ''
  let inQuotes = false

  for (let i = 0; i < line.length; i += 1) {
    const char = line[i]
    if (char === '"') {
      if (inQuotes && line[i + 1] === '"') {
        current += '"'
        i += 1
      } else {
        inQuotes = !inQuotes
      }
      continue
    }

    if (char === ',' && !inQuotes) {
      cells.push(current.trim())
      current = ''
      continue
    }

    current += char
  }

  cells.push(current.trim())
  return cells
}

const parseCsvTable = (content: string): StructuredOutput | null => {
  const lines = content.trim().split(/\r?\n/).filter(line => line.trim().length > 0)
  if (lines.length < 2) return null
  if (!lines[0].includes(',')) return null

  const headers = parseCsvLine(lines[0]).slice(0, maxStructuredColumns)
  if (headers.length === 0) return null

  const rows = lines.slice(1, maxStructuredRows + 1).map(line =>
    parseCsvLine(line).slice(0, maxStructuredColumns).map(cell => cell)
  )

  return { headers, rows }
}

const getStructuredOutput = (output: ExecutionResult): StructuredOutput | null => {
  if (!['stdout', 'return', 'log', 'info', 'debug', 'warn'].includes(output.type)) return null
  const content = output.content?.trim()
  if (!content) return null

  const jsonTable = parseJsonTable(content)
  if (jsonTable) return jsonTable

  return parseCsvTable(content)
}

const structuredOutputs = computed<Record<string, StructuredOutput | null>>(() => {
  const outputMap: Record<string, StructuredOutput | null> = {}
  if (!props.executionOutputs) return outputMap
  for (const output of props.executionOutputs) {
    outputMap[output.id] = getStructuredOutput(output)
  }
  return outputMap
})

const isVfsPath = (value: string) => {
  vfsPathRegex.lastIndex = 0
  return vfsPathRegex.test(value)
}

const getOutputSegments = (content: string): OutputSegment[] => {
  const segments: OutputSegment[] = []
  if (!content) return segments

  let lastIndex = 0
  const matches = [...content.matchAll(vfsPathRegex)]

  for (const match of matches) {
    if (!match.index && match.index !== 0) continue
    const start = match.index
    const end = start + match[0].length

    if (start > lastIndex) {
      segments.push({ text: content.slice(lastIndex, start) })
    }

    segments.push({ text: match[0], path: match[0] })
    lastIndex = end
  }

  if (lastIndex < content.length) {
    segments.push({ text: content.slice(lastIndex) })
  }

  return segments
}

const handleOpenVfs = (path: string) => {
  if (openVfsAtPath && typeof openVfsAtPath === 'function') {
    openVfsAtPath(path)
  }
}

const copyPath = async (path: string) => {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(path)
    } else {
      copyText({ text: path, origin: true })
    }
    message.success('Path copied to clipboard')
  } catch (error) {
    message.error('Failed to copy path')
  }
}

const renderCanvas = (output: ExecutionResult, canvasElement: HTMLCanvasElement) => {
  const runner = getCodeRunner()
  runner.renderCanvasToElement(output.content, canvasElement)
}

const renderMatplotlib = (output: ExecutionResult, imgElement: HTMLImageElement) => {
  const runner = getCodeRunner()
  runner.renderMatplotlibToElement(output.content, imgElement)
}

const registerCanvas = (el: HTMLCanvasElement | null, output: ExecutionResult) => {
  if (!el) return
  nextTick(() => renderCanvas(output, el))
}

const registerMatplotlib = (el: HTMLImageElement | null, output: ExecutionResult) => {
  if (!el) return
  nextTick(() => renderMatplotlib(output, el))
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

/* Execution Output */
.output-canvas,
.output-image {
  display: block;
  max-width: 100%;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  background: #fff;
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

.output-table-wrapper {
  overflow-x: auto;
}

.output-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.output-table th,
.output-table td {
  border: 1px solid #e5e7eb;
  padding: 0.5rem;
  text-align: left;
  white-space: nowrap;
}

.output-table th {
  background: #f9fafb;
  font-weight: 600;
}

.vfs-path-link {
  color: #2563eb;
  text-decoration: underline;
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  cursor: pointer;
}

.vfs-path-inline {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.vfs-copy-button {
  background: none;
  border: none;
  padding: 0;
  margin-left: 0.25rem;
  color: #6b7280;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s ease;
}

.vfs-path-inline:hover .vfs-copy-button,
.output-table td:hover .vfs-copy-button {
  opacity: 1;
  pointer-events: auto;
}

.vfs-copy-button:hover {
  color: #111827;
}

:deep(.dark) .vfs-path-link {
  color: #93c5fd;
}

:deep(.dark) .vfs-copy-button {
  color: #9ca3af;
}

:deep(.dark) .vfs-copy-button:hover {
  color: #e5e7eb;
}

:deep(.dark) .output-table th,
:deep(.dark) .output-table td {
  border-color: #374151;
}

:deep(.dark) .output-table th {
  background: #111827;
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
