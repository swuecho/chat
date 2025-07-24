<template>
  <div v-if="artifacts && artifacts.length > 0" class="artifact-container" data-test-role="artifact-viewer">
    <div v-for="artifact in artifacts" :key="artifact.uuid" class="artifact-item">
      <div class="artifact-header">
        <div class="artifact-title">
          <Icon :icon="getArtifactIcon(artifact.type)" class="artifact-icon" />
          <span class="artifact-title-text">{{ artifact.title }}</span>
          <span class="artifact-type">({{ artifact.type }})</span>
        </div>
        <div class="artifact-actions">
          <NButton size="small" @click="toggleExpanded(artifact.uuid)">
            <span class="hidden sm:inline">{{ isExpanded(artifact.uuid) ? 'Collapse' : 'Expand' }}</span>
            <Icon :icon="isExpanded(artifact.uuid) ? 'ri:arrow-up-line' : 'ri:arrow-down-line'" class="sm:hidden" />
          </NButton>
          <NButton v-if="artifact.type === 'html'" size="small" @click="openInNewWindow(artifact.content)" title="Open in new window for debugging">
            <Icon icon="ri:external-link-line" />
          </NButton>
          <NButton size="small" @click="copyContent(artifact.content)">
            <Icon icon="ri:file-copy-line" />
          </NButton>
        </div>
      </div>
      
      <div v-if="isExpanded(artifact.uuid)" class="artifact-content">
        <!-- Code Artifact -->
        <div v-if="artifact.type === 'code' || artifact.type === 'executable-code'" class="code-artifact">
          <!-- Execution controls for executable code -->
          <div v-if="isExecutable(artifact)" class="execution-controls">
            <div class="control-bar">
              <NButton 
                size="small" 
                type="primary" 
                @click="runCode(artifact)" 
                :disabled="isRunning(artifact.uuid)"
                :loading="isRunning(artifact.uuid)">
                <template #icon>
                  <Icon icon="ri:play-line" />
                </template>
                {{ isRunning(artifact.uuid) ? 'Running...' : 'Run Code' }}
              </NButton>
              <NButton 
                size="small" 
                @click="clearOutput(artifact.uuid)" 
                :disabled="!hasOutput(artifact.uuid)">
                <template #icon>
                  <Icon icon="ri:delete-bin-line" />
                </template>
                Clear Output
              </NButton>
              <NButton 
                size="small" 
                @click="toggleEditor(artifact.uuid)" 
                type="tertiary">
                <template #icon>
                  <Icon :icon="isEditing(artifact.uuid) ? 'ri:eye-line' : 'ri:edit-line'" />
                </template>
                {{ isEditing(artifact.uuid) ? 'View' : 'Edit' }}
              </NButton>
            </div>
            
            <div v-if="getLastExecution(artifact.uuid)" class="execution-info">
              <span class="execution-time">
                {{ getLastExecution(artifact.uuid)?.execution_time_ms }}ms
              </span>
              <span class="execution-status" :class="getExecutionStatus(artifact.uuid)">
                {{ getExecutionStatus(artifact.uuid) }}
              </span>
            </div>
          </div>

          <!-- Code display/editor -->
          <div v-if="isEditing(artifact.uuid)" class="code-editor">
            <textarea 
              v-model="editableContent[artifact.uuid]"
              :placeholder="`Enter ${artifact.language || 'JavaScript'} code...`"
              class="code-textarea"
              rows="10"
              @keydown.ctrl.enter="runCode(artifact)"
              @keydown.meta.enter="runCode(artifact)"
            ></textarea>
            <div class="editor-hint">
              Press <kbd>Ctrl/Cmd + Enter</kbd> to run code
            </div>
          </div>
          <div v-else class="code-display">
            <pre><code :class="`language-${artifact.language || 'text'}`" v-html="highlightCode(getCodeContent(artifact), artifact.language)"></code></pre>
          </div>

          <!-- Library management -->
          <div v-if="isExecutable(artifact)" class="library-management">
            <div class="library-header">
              <span class="library-title">
                <Icon icon="ri:package-line" />
                Libraries
              </span>
              <NButton size="tiny" @click="showLibraries(artifact.uuid)" type="tertiary">
                <Icon icon="ri:information-line" />
                Available
              </NButton>
            </div>
            <div v-if="showLibraryList[artifact.uuid]" class="library-list">

              <div v-if="artifact.language === 'python' || artifact.language === 'py'" class="library-info">
                <div class="library-info">
                  Available Python packages: numpy, pandas, matplotlib, scipy, scikit-learn, requests, beautifulsoup4, pillow, sympy, networkx, seaborn, plotly, bokeh, altair
                </div>
                <div class="library-usage">
                  Use <code>import packageName</code> or <code>from packageName import ...</code> in your code
                </div>
              </div>
              <div v-else class="library-info">
                <div class="library-info">
                  Available libraries: lodash, d3, chart.js, moment, axios, rxjs, p5, three, fabric
                </div>
                <div class="library-usage">
                  Use <code>// @import libraryName</code> in your code to auto-load libraries
                </div>

              </div>
            </div>
          </div>

          <!-- Output area -->
          <div v-if="hasOutput(artifact.uuid)" class="execution-output">
            <div class="output-header">
              <span class="output-title">
                <Icon icon="ri:terminal-line" />
                Output
              </span>
              <NButton size="tiny" text @click="clearOutput(artifact.uuid)">
                <Icon icon="ri:close-line" />
              </NButton>
            </div>
            <div class="output-content">
              <div v-for="result in getExecutionOutput(artifact.uuid)" 
                   :key="result.id" 
                   class="output-line"
                   :class="result.type">
                <span class="output-type">{{ result.type }}</span>
                
                <!-- Canvas output -->
                <div v-if="result.type === 'canvas'" class="canvas-output">
                  <canvas 
                    :ref="(el) => setCanvasRef(result.id, el)"
                    class="result-canvas"
                  ></canvas>
                </div>
                

                <!-- Matplotlib plot output -->
                <div v-else-if="result.type === 'matplotlib'" class="matplotlib-output">
                  <img 
                    :ref="(el) => setMatplotlibRef(result.id, el)"
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
        </div>
        
        <!-- HTML Artifact -->
        <div v-else-if="artifact.type === 'html'" class="html-artifact" :class="{ fullscreen: isFullscreen(artifact.uuid) }" :data-test-fullscreen="isFullscreen(artifact.uuid)">
          <div class="html-preview">
            <iframe 
              :srcdoc="processHtmlContent(artifact.content)" 
              sandbox="allow-scripts allow-same-origin allow-forms allow-modals allow-popups"
              class="html-frame"
              loading="lazy"
              :key="htmlRefreshKey[artifact.uuid] || 0"
            ></iframe>
          </div>
          <div class="html-actions">
            <NButton size="tiny" @click="refreshHtmlPreview(artifact.uuid)" title="Refresh preview">
              <Icon icon="ri:refresh-line" />
            </NButton>
            <NButton size="tiny" @click="toggleFullscreen(artifact.uuid)" :title="isFullscreen(artifact.uuid) ? 'Exit fullscreen' : 'Enter fullscreen'">
              <Icon :icon="isFullscreen(artifact.uuid) ? 'ri:fullscreen-exit-line' : 'ri:fullscreen-line'" />
            </NButton>
          </div>
        </div>
        
        <!-- SVG Artifact -->
        <div v-else-if="artifact.type === 'svg'" class="svg-artifact">
          <div class="svg-container">
            <div class="svg-preview" v-html="processSvgContent(artifact.content)"></div>
          </div>
        </div>
        
        <!-- Mermaid Diagram -->
        <div v-else-if="artifact.type === 'mermaid'" class="mermaid-artifact">
          <div class="mermaid-container">
            <div class="mermaid-preview" :id="`mermaid-${artifact.uuid}`">
              {{ artifact.content }}
            </div>
          </div>
        </div>
        
        <!-- JSON Artifact -->
        <div v-else-if="artifact.type === 'json'" class="json-artifact">
          <div class="json-container">
            <div class="json-preview">
              <pre><code class="language-json" v-html="highlightCode(formatJson(artifact.content), 'json')"></code></pre>
            </div>
            <div class="json-actions">
              <NButton size="tiny" @click="validateJson(artifact.content)" title="Validate JSON">
                <Icon icon="ri:check-line" />
              </NButton>
              <NButton size="tiny" @click="formatAndCopyJson(artifact.content)" title="Format and copy">
                <Icon icon="ri:code-box-line" />
              </NButton>
            </div>
          </div>
        </div>
        
        <!-- Default: Plain text -->
        <div v-else class="text-artifact">
          <pre>{{ artifact.content }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, reactive, computed } from 'vue'
import { NButton, useMessage } from 'naive-ui'
import { Icon } from '@iconify/vue'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'
import { getCodeRunner, type ExecutionResult } from '@/services/codeRunner'

interface Props {
  artifacts: Chat.Artifact[]
  inversion?: boolean
}

const props = defineProps<Props>()

const message = useMessage()
const expandedArtifacts = ref<Set<string>>(new Set())
const fullscreenArtifacts = ref<Set<string>>(new Set())
const htmlRefreshKey = ref<Record<string, number>>({})

// Code execution state
const runningArtifacts = ref<Set<string>>(new Set())
const editingArtifacts = ref<Set<string>>(new Set())
const editableContent = reactive<Record<string, string>>({})
const executionOutputs = reactive<Record<string, ExecutionResult[]>>({})
const showLibraryList = reactive<Record<string, boolean>>({})
const canvasRefs = reactive<Record<string, HTMLCanvasElement>>({})

const matplotlibRefs = reactive<Record<string, HTMLImageElement>>({})

const codeRunner = getCodeRunner()

// Auto-expand SVG artifacts on mount
const autoExpandSvgArtifacts = () => {
  if (props.artifacts) {
    props.artifacts.forEach(artifact => {
      if (artifact.type === 'svg') {
        expandedArtifacts.value.add(artifact.uuid)
      }
    })
  }
}

// Moved to combined onMounted below

const toggleExpanded = (uuid: string) => {
  if (expandedArtifacts.value.has(uuid)) {
    expandedArtifacts.value.delete(uuid)
  } else {
    expandedArtifacts.value.add(uuid)
  }
}

const isExpanded = (uuid: string) => {
  return expandedArtifacts.value.has(uuid)
}

const getArtifactIcon = (type: string) => {
  switch (type) {
    case 'code':
      return 'ri:code-line'
    case 'executable-code':
      return 'ri:play-circle-line'
    case 'html':
      return 'ri:html5-line'
    case 'svg':
      return 'ri:image-line'
    case 'mermaid':
      return 'ri:flow-chart'
    case 'json':
      return 'ri:file-code-line'
    case 'markdown':
      return 'ri:markdown-line'
    default:
      return 'ri:file-line'
  }
}

const formatJson = (jsonString: string) => {
  try {
    const parsed = JSON.parse(jsonString)
    return JSON.stringify(parsed, null, 2)
  } catch (error) {
    return jsonString
  }
}

const validateJson = (jsonString: string) => {
  try {
    JSON.parse(jsonString)
    message.success('Valid JSON')
  } catch (error) {
    message.error(`Invalid JSON: ${error}`)
  }
}

const formatAndCopyJson = async (jsonString: string) => {
  try {
    const formatted = formatJson(jsonString)
    await navigator.clipboard.writeText(formatted)
    message.success('Formatted JSON copied to clipboard')
  } catch (error) {
    message.error('Failed to format and copy JSON')
  }
}

const highlightCode = (code: string, language?: string) => {
  if (!language) {
    return hljs.highlightAuto(code).value
  }
  
  if (hljs.getLanguage(language)) {
    return hljs.highlight(code, { language }).value
  }
  
  return hljs.highlightAuto(code).value
}

const copyContent = async (content: string) => {
  try {
    await navigator.clipboard.writeText(content)
    message.success('Content copied to clipboard')
  } catch (error) {
    message.error('Failed to copy content')
  }
}

const openInNewWindow = (htmlContent: string) => {
  try {
    // Create a blob from the HTML content
    const blob = new Blob([htmlContent], { type: 'text/html' })
    const blobUrl = URL.createObjectURL(blob)
    
    // Open in new window
    const newWindow = window.open(blobUrl, '_blank', 'width=1200,height=800,scrollbars=yes,resizable=yes,toolbar=yes,menubar=yes')
    
    if (newWindow) {
      // Focus on the new window
      newWindow.focus()
      
      // Clean up the blob URL after the window loads
      newWindow.addEventListener('load', () => {
        URL.revokeObjectURL(blobUrl)
      })
      
      message.success('HTML opened in new window')
    } else {
      // Clean up if window failed to open
      URL.revokeObjectURL(blobUrl)
      message.error('Failed to open new window. Please check popup blockers.')
    }
  } catch (error) {
    message.error('Failed to create HTML preview. Please try copying the content instead.')
  }
}

const processHtmlContent = (content: string) => {
  let processedContent = content.trim()
  
  // Add basic HTML structure if missing
  if (!processedContent.includes('<html') && !processedContent.includes('<!DOCTYPE')) {
    processedContent = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>HTML Artifact</title>
  <style>
    body { margin: 0; padding: 16px; font-family: system-ui, -apple-system, sans-serif; }
    * { box-sizing: border-box; }
    
    /* Custom scrollbar styling for iframe content */
    * {
      scrollbar-width: thin;
      scrollbar-color: rgba(155, 155, 155, 0.5) transparent;
    }

    *::-webkit-scrollbar {
      width: 4px;
      height: 4px;
    }

    *::-webkit-scrollbar-track {
      background: transparent;
      border-radius: 3px;
    }

    *::-webkit-scrollbar-thumb {
      background: rgba(155, 155, 155, 0.5);
      border-radius: 3px;
      transition: background 0.2s ease;
    }

    *::-webkit-scrollbar-thumb:hover {
      background: rgba(155, 155, 155, 0.8);
    }

    *::-webkit-scrollbar-thumb:active {
      background: rgba(155, 155, 155, 1);
    }

    *::-webkit-scrollbar-corner {
      background: transparent;
    }
  </style>
</head>
<body>
${processedContent}
</body>
</html>`
  }
  
  return processedContent
}

const refreshHtmlPreview = (uuid: string) => {
  htmlRefreshKey.value[uuid] = Date.now()
  message.success('HTML preview refreshed')
}

const toggleFullscreen = (uuid: string) => {
  if (fullscreenArtifacts.value.has(uuid)) {
    fullscreenArtifacts.value.delete(uuid)
  } else {
    fullscreenArtifacts.value.add(uuid)
  }
}

const isFullscreen = (uuid: string) => {
  return fullscreenArtifacts.value.has(uuid)
}

const processSvgContent = (content: string) => {
  let processedContent = content.trim()
  
  // Ensure the SVG is properly formatted
  if (processedContent.includes('<svg')) {
    // Replace currentColor with theme-appropriate colors
    processedContent = processedContent.replace(/fill="currentColor"/g, 'fill="#4a5568"')
    processedContent = processedContent.replace(/stroke="currentColor"/g, 'stroke="#4a5568"')
    
    // Ensure the SVG has viewBox for proper scaling
    if (!processedContent.includes('viewBox=')) {
      const widthMatch = processedContent.match(/width="(\d+)"/)
      const heightMatch = processedContent.match(/height="(\d+)"/)
      if (widthMatch && heightMatch) {
        const width = widthMatch[1]
        const height = heightMatch[1]
        processedContent = processedContent.replace('<svg', `<svg viewBox="0 0 ${width} ${height}"`)
      }
    }
  }
  
  return processedContent
}

// Code execution methods
const isExecutable = (artifact: Chat.Artifact) => {
  return artifact.type === 'executable-code' || 
         (artifact.type === 'code' && artifact.language && codeRunner.isLanguageSupported(artifact.language))
}

const isRunning = (uuid: string) => {
  return runningArtifacts.value.has(uuid)
}

const isEditing = (uuid: string) => {
  return editingArtifacts.value.has(uuid)
}

const hasOutput = (uuid: string) => {
  return executionOutputs[uuid] && executionOutputs[uuid].length > 0
}

const getCodeContent = (artifact: Chat.Artifact) => {
  return editableContent[artifact.uuid] || artifact.content
}

const getExecutionOutput = (uuid: string) => {
  return executionOutputs[uuid] || []
}

const getLastExecution = (uuid: string) => {
  const output = executionOutputs[uuid]
  if (!output || output.length === 0) return null
  return output[output.length - 1]
}

const getExecutionStatus = (uuid: string) => {
  const output = executionOutputs[uuid]
  if (!output || output.length === 0) return 'ready'
  
  const hasError = output.some(result => result.type === 'error')
  return hasError ? 'error' : 'success'
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleTimeString()
}

const toggleEditor = (uuid: string) => {
  if (editingArtifacts.value.has(uuid)) {
    editingArtifacts.value.delete(uuid)
  } else {
    editingArtifacts.value.add(uuid)
    // Initialize editable content if not exists
    if (!editableContent[uuid]) {
      const artifact = props.artifacts.find(a => a.uuid === uuid)
      if (artifact) {
        editableContent[uuid] = artifact.content
      }
    }
  }
}

const clearOutput = (uuid: string) => {
  executionOutputs[uuid] = []
}

const showLibraries = (uuid: string) => {
  showLibraryList[uuid] = !showLibraryList[uuid]
}

const setCanvasRef = (resultId: string, el: HTMLCanvasElement | null) => {
  if (el) {
    canvasRefs[resultId] = el
    // Find the result and render canvas
    const result = Object.values(executionOutputs).flat().find(r => r.id === resultId)
    if (result && result.type === 'canvas') {
      codeRunner.renderCanvasToElement(result.content, el)
    }
  }
}


const setMatplotlibRef = (resultId: string, el: HTMLImageElement | null) => {
  if (el) {
    matplotlibRefs[resultId] = el
    // Find the result and render matplotlib plot
    const result = Object.values(executionOutputs).flat().find(r => r.id === resultId)
    if (result && result.type === 'matplotlib') {
      codeRunner.renderMatplotlibToElement(result.content, el)
    }
  }
}


const runCode = async (artifact: Chat.Artifact) => {
  if (!artifact.language) {
    message.error('No language specified for code execution')
    return
  }

  if (!codeRunner.isLanguageSupported(artifact.language)) {
    message.error(`Language ${artifact.language} is not supported for execution`)
    return
  }

  const uuid = artifact.uuid
  runningArtifacts.value.add(uuid)
  
  try {
    const code = getCodeContent(artifact)

    const results = await codeRunner.execute(artifact.language, code, uuid)
    
    executionOutputs[uuid] = results
    
    // Handle canvas and matplotlib output rendering

    setTimeout(() => {
      results.forEach(result => {
        if (result.type === 'canvas' && canvasRefs[result.id]) {
          codeRunner.renderCanvasToElement(result.content, canvasRefs[result.id])

        } else if (result.type === 'matplotlib' && matplotlibRefs[result.id]) {
          codeRunner.renderMatplotlibToElement(result.content, matplotlibRefs[result.id])

        }
      })
    }, 100) // Allow DOM to update first
    
    // Show success message
    const hasError = results.some(result => result.type === 'error')
    const hasCanvas = results.some(result => result.type === 'canvas')

    const hasMatplotlib = results.some(result => result.type === 'matplotlib')
    
    if (hasError) {
      message.error('Code execution completed with errors')
    } else if (hasCanvas || hasMatplotlib) {

      message.success('Code executed successfully with graphics output')
    } else {
      message.success('Code executed successfully')
    }
  } catch (error) {
    executionOutputs[uuid] = [{
      id: Date.now().toString(),
      type: 'error',
      content: error instanceof Error ? error.message : String(error),
      timestamp: new Date().toISOString()
    }]
    message.error('Code execution failed')
  } finally {
    runningArtifacts.value.delete(uuid)
  }
}

// Initialize editable content for artifacts
onMounted(() => {
  autoExpandSvgArtifacts()
  
  // Initialize editable content for all artifacts
  if (props.artifacts) {
    props.artifacts.forEach(artifact => {
      if (isExecutable(artifact)) {
        editableContent[artifact.uuid] = artifact.content
        // Auto-expand executable artifacts
        expandedArtifacts.value.add(artifact.uuid)
      }
    })
  }
})
</script>

<style scoped>
.artifact-container {
  margin-top: 12px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  /* Ensure artifact viewer doesn't interfere with message layout */
  contain: layout style;
  /* Use isolation to prevent interference with parent elements and tests */
  isolation: isolate;
  /* Prevent horizontal overflow */
  overflow-x: hidden;
}

.artifact-item {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  margin-bottom: 8px;
  overflow: hidden;
  background: var(--artifact-bg);
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
}

.artifact-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
}

@media (min-width: 640px) {
  .artifact-header {
    padding: 12px 16px;
    flex-wrap: nowrap;
  }
}

.artifact-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: var(--text-color);
  flex: 1 1 auto;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
}

@media (min-width: 640px) {
  .artifact-title {
    gap: 8px;
  }
}

.artifact-title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
}

@media (min-width: 640px) {
  .artifact-title-text {
    font-size: 16px;
  }
}

.artifact-icon {
  font-size: 14px;
  color: var(--primary-color);
  flex-shrink: 0;
}

@media (min-width: 640px) {
  .artifact-icon {
    font-size: 16px;
  }
}

.artifact-type {
  font-size: 10px;
  color: var(--text-color-secondary);
  background: var(--tag-bg);
  padding: 2px 4px;
  border-radius: 4px;
  flex-shrink: 0;
  white-space: nowrap;
}

@media (min-width: 640px) {
  .artifact-type {
    font-size: 12px;
    padding: 2px 6px;
  }
}

.artifact-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

@media (min-width: 640px) {
  .artifact-actions {
    gap: 8px;
  }
}

.artifact-content {
  padding: 12px;
  overflow: hidden;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
}

@media (min-width: 640px) {
  .artifact-content {
    padding: 16px;
  }
}

.code-artifact pre {
  margin: 0;
  padding: 12px;
  background: var(--code-bg);
  border-radius: 6px;
  overflow-x: auto;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  max-width: 100%;
  width: 100%;
  min-width: 0;
  white-space: pre;
  -webkit-overflow-scrolling: touch;
  box-sizing: border-box;
}

@media (min-width: 640px) {
  .code-artifact pre {
    padding: 16px;
  }
}

.code-artifact code {
  font-size: 12px;
  line-height: 1.4;
  display: block;
  white-space: pre;
  overflow-wrap: normal;
  word-break: normal;
}

@media (min-width: 640px) {
  .code-artifact code {
    font-size: 13px;
    line-height: 1.5;
  }
}

.html-artifact {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
  max-width: 100%;
  width: 100%;
  min-width: 0;
  box-sizing: border-box;
}

.html-preview {
  position: relative;
  width: 100%;
  height: 200px;
  background: white;
  border-radius: 6px;
  overflow: hidden;
}

@media (min-width: 640px) {
  .html-preview {
    height: 300px;
  }
}

@media (min-width: 1024px) {
  .html-preview {
    height: 400px;
  }
}

.html-frame {
  width: 100%;
  height: 100%;
  border: none;
  background: white;
  display: block;
  transition: height 0.3s ease;
  max-width: 100%;
}

.html-actions {
  display: flex;
  gap: 4px;
  padding: 6px 8px;
  background: var(--artifact-header-bg);
  border-top: 1px solid var(--border-color);
  justify-content: flex-end;
  flex-wrap: wrap;
}

@media (min-width: 640px) {
  .html-actions {
    padding: 8px 12px;
    flex-wrap: nowrap;
  }
}

/* Fullscreen mode */
.html-artifact.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: white;
  border-radius: 0;
  margin: 0;
  max-width: 100vw;
  max-height: 100vh;
  /* Ensure fullscreen doesn't interfere with E2E tests */
  contain: strict;
}

.html-artifact.fullscreen .html-preview {
  height: calc(100vh - 40px);
  border-radius: 0;
}

@media (min-width: 640px) {
  .html-artifact.fullscreen .html-preview {
    height: calc(100vh - 50px);
  }
}

.html-artifact.fullscreen .html-actions {
  position: absolute;
  top: 0;
  right: 0;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  z-index: 10000;
  border: 1px solid var(--border-color);
  border-radius: 0 0 0 6px;
  padding: 4px 6px;
}

@media (min-width: 640px) {
  .html-artifact.fullscreen .html-actions {
    padding: 8px 12px;
  }
}

.svg-artifact {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px;
  background: var(--artifact-content-bg);
  border-radius: 6px;
  min-height: 80px;
  overflow: hidden;
}

@media (min-width: 640px) {
  .svg-artifact {
    padding: 16px;
  }
}

.svg-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  max-width: 100%;
  width: 100%;
  max-height: 300px;
  color: var(--text-color);
  overflow: hidden;
  box-sizing: border-box;
}

@media (min-width: 640px) {
  .svg-container {
    max-height: 400px;
  }
}

.svg-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
  min-height: 80px;
  width: 100%;
  overflow: hidden;
}

@media (min-width: 640px) {
  .svg-preview {
    padding: 16px;
  }
}

.svg-preview svg {
  max-width: 100%;
  max-height: 250px;
  width: auto;
  height: auto;
  display: block;
}

@media (min-width: 640px) {
  .svg-preview svg {
    max-height: 300px;
  }
}

/* Dark mode adjustments */
[data-theme='dark'] .svg-preview {
  background: #2d2d2d;
  border-color: #3c3c3c;
}

.mermaid-artifact {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
  max-width: 100%;
  width: 100%;
  min-width: 0;
  box-sizing: border-box;
}

.mermaid-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px;
  background: var(--artifact-content-bg);
  min-height: 120px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

@media (min-width: 640px) {
  .mermaid-container {
    padding: 16px;
  }
}

.mermaid-preview {
  max-width: 100%;
  overflow: auto;
  -webkit-overflow-scrolling: touch;
}

.json-artifact {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
  max-width: 100%;
}

.json-container {
  background: var(--artifact-content-bg);
  overflow: hidden;
}

.json-preview {
  max-height: 300px;
  overflow: auto;
  -webkit-overflow-scrolling: touch;
}

@media (min-width: 640px) {
  .json-preview {
    max-height: 400px;
  }
}

.json-preview pre {
  margin: 0;
  padding: 12px;
  background: var(--code-bg);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  overflow-x: auto;
  white-space: pre;
}

@media (min-width: 640px) {
  .json-preview pre {
    padding: 16px;
    font-size: 13px;
    line-height: 1.5;
  }
}

.json-actions {
  display: flex;
  gap: 4px;
  padding: 6px 8px;
  background: var(--artifact-header-bg);
  border-top: 1px solid var(--border-color);
  justify-content: flex-end;
  flex-wrap: wrap;
}

@media (min-width: 640px) {
  .json-actions {
    padding: 8px 12px;
    flex-wrap: nowrap;
  }
}

.text-artifact pre {
  margin: 0;
  padding: 12px;
  background: var(--code-bg);
  border-radius: 6px;
  overflow-x: auto;
  white-space: pre-wrap;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  max-width: 100%;
  -webkit-overflow-scrolling: touch;
  word-break: break-word;
  overflow-wrap: break-word;
}

@media (min-width: 640px) {
  .text-artifact pre {
    padding: 16px;
    font-size: 14px;
    line-height: 1.5;
  }
}

/* Mobile responsive fixes */
@media (max-width: 639px) {
  .artifact-container {
    /* Ensure artifacts don't break mobile layout */
    overflow-x: hidden;
    word-wrap: break-word;
    overflow-wrap: break-word;
  }
  
  .artifact-header {
    /* Force wrap on small screens for better button accessibility */
    flex-wrap: wrap;
    min-height: 44px; /* Touch-friendly minimum height */
  }
  
  .artifact-actions {
    /* Stack actions vertically on very small screens */
    flex-wrap: wrap;
    justify-content: flex-end;
  }
  
  .code-artifact pre {
    /* Better code display on mobile */
    font-size: 11px;
    line-height: 1.3;
    word-break: break-all;
    overflow-wrap: break-word;
  }
}

/* Theme variables */
:root {
  --artifact-bg: #fafafa;
  --artifact-header-bg: #f0f0f0;
  --artifact-content-bg: #ffffff;
  --code-bg: #f6f8fa;
  --tag-bg: #e1e7ff;
  --border-color: #e1e4e8;
  --text-color: #24292e;
  --text-color-secondary: #6a737d;
  --primary-color: #0366d6;
}

[data-theme='dark'] {
  --artifact-bg: #1e1e1e;
  --artifact-header-bg: #2d2d2d;
  --artifact-content-bg: #252526;
  --code-bg: #1e1e1e;
  --tag-bg: #3c3c3c;
  --border-color: #3c3c3c;
  --text-color: #cccccc;
  --text-color-secondary: #8c8c8c;
  --primary-color: #58a6ff;
}

/* Code execution styles */
.execution-controls {
  margin-bottom: 12px;
  padding: 12px;
  background: var(--artifact-header-bg);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.control-bar {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.execution-info {
  margin-top: 8px;
  display: flex;
  gap: 12px;
  align-items: center;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.execution-time {
  color: var(--primary-color);
  font-weight: 500;
}

.execution-status {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  text-transform: uppercase;
}

.execution-status.success {
  background: #22c55e;
  color: white;
}

.execution-status.error {
  background: #ef4444;
  color: white;
}

.execution-status.ready {
  background: var(--tag-bg);
  color: var(--text-color-secondary);
}

.code-editor {
  margin-bottom: 12px;
}

.code-textarea {
  width: 100%;
  min-height: 200px;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--code-bg);
  color: var(--text-color);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: vertical;
  outline: none;
  box-sizing: border-box;
}

.code-textarea:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(88, 166, 255, 0.2);
}

.editor-hint {
  margin-top: 6px;
  font-size: 11px;
  color: var(--text-color-secondary);
  text-align: right;
}

.editor-hint kbd {
  background: var(--tag-bg);
  color: var(--text-color);
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 10px;
  font-family: inherit;
}

.code-display {
  /* No additional styles needed, uses existing code-artifact styles */
}

.execution-output {
  margin-top: 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
  background: var(--artifact-content-bg);
}

.output-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color);
}

.output-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

.output-content {
  max-height: 300px;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.output-line {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 6px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 11px;
  line-height: 1.4;
}

.output-line:last-child {
  border-bottom: none;
}

.output-line.log {
  background: rgba(59, 130, 246, 0.05);
}

.output-line.error {
  background: rgba(239, 68, 68, 0.05);
  color: #ef4444;
}

.output-line.warn {
  background: rgba(245, 158, 11, 0.05);
  color: #f59e0b;
}

.output-line.info {
  background: rgba(34, 197, 94, 0.05);
  color: #22c55e;
}

.output-line.return {
  background: rgba(168, 85, 247, 0.05);
  color: #a855f7;
}

.output-line.debug {
  background: rgba(107, 114, 128, 0.05);
  color: #6b7280;
}

.output-type {
  flex-shrink: 0;
  font-weight: 500;
  text-transform: uppercase;
  font-size: 9px;
  padding: 2px 4px;
  border-radius: 3px;
  background: var(--tag-bg);
  color: var(--text-color-secondary);
  min-width: 40px;
  text-align: center;
}

.output-text {
  flex: 1;
  word-break: break-word;
  white-space: pre-wrap;
  color: var(--text-color);
}

.output-time {
  flex-shrink: 0;
  font-size: 9px;
  color: var(--text-color-secondary);
  opacity: 0.7;
}

/* Library management styles */
.library-management {
  margin-bottom: 12px;
  padding: 8px 12px;
  background: var(--artifact-header-bg);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.library-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color);
}

.library-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

.library-list {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--border-color);
}

.library-info {
  font-size: 11px;
  color: var(--text-color-secondary);
  margin-bottom: 6px;
  line-height: 1.4;
}

.library-usage {
  font-size: 10px;
  color: var(--text-color-secondary);
  font-style: italic;
}

.library-usage code {
  background: var(--tag-bg);
  padding: 1px 4px;
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

/* Canvas output styles */
.canvas-output {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
  margin: 4px 0;
}


/* Matplotlib plot output styles */
.matplotlib-output {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e9ecef;
  margin: 4px 0;
}

.result-plot {
  max-width: 100%;
  max-height: 500px;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}


.result-canvas {
  max-width: 100%;
  max-height: 400px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

[data-theme='dark'] .canvas-output {
  background: #2d2d2d;
  border-color: #3c3c3c;
}

[data-theme='dark'] .result-canvas {
  background: #ffffff;
  border-color: #3c3c3c;
}

/* Enhanced output line styles for canvas */
.output-line.canvas {
  background: rgba(99, 102, 241, 0.05);
  flex-direction: column;
  align-items: stretch;
}

.output-line.canvas .output-type {
  align-self: flex-start;
  background: #6366f1;
  color: white;
}

/* Mobile responsive adjustments */
@media (max-width: 639px) {
  .execution-controls {
    padding: 8px;
  }
  
  .control-bar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .control-bar > * {
    width: 100%;
  }
  
  .execution-info {
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
  }
  
  .code-textarea {
    min-height: 150px;
    font-size: 12px;
  }
  
  .output-line {
    flex-direction: column;
    gap: 4px;
    padding: 8px;
  }
  
  .output-type {
    align-self: flex-start;
  }
  
  .output-time {
    align-self: flex-end;
  }
  
  .library-management {
    padding: 6px 8px;
  }
  
  .library-header {
    font-size: 11px;
  }
  
  .library-info, .library-usage {
    font-size: 10px;
  }
  
  .result-canvas {
    max-height: 250px;
  }
  
  .canvas-output {
    padding: 8px;
  }
}

/* Custom scrollbar styling for all scrollable elements */
.output-content,
.json-preview,
.mermaid-container,
.code-artifact pre {
  scrollbar-width: thin;
  scrollbar-color: rgba(155, 155, 155, 0.5) transparent;
}

.output-content::-webkit-scrollbar,
.json-preview::-webkit-scrollbar,
.mermaid-container::-webkit-scrollbar,
.code-artifact pre::-webkit-scrollbar {
  width: 8px;
}

.output-content::-webkit-scrollbar-track,
.json-preview::-webkit-scrollbar-track,
.mermaid-container::-webkit-scrollbar-track,
.code-artifact pre::-webkit-scrollbar-track {
  background: transparent;
  border-radius: 4px;
}

.output-content::-webkit-scrollbar-thumb,
.json-preview::-webkit-scrollbar-thumb,
.mermaid-container::-webkit-scrollbar-thumb,
.code-artifact pre::-webkit-scrollbar-thumb {
  background: rgba(155, 155, 155, 0.5);
  border-radius: 4px;
  transition: background 0.2s ease;
}

.output-content::-webkit-scrollbar-thumb:hover,
.json-preview::-webkit-scrollbar-thumb:hover,
.mermaid-container::-webkit-scrollbar-thumb:hover,
.code-artifact pre::-webkit-scrollbar-thumb:hover {
  background: rgba(155, 155, 155, 0.8);
}

.output-content::-webkit-scrollbar-thumb:active,
.json-preview::-webkit-scrollbar-thumb:active,
.mermaid-container::-webkit-scrollbar-thumb:active,
.code-artifact pre::-webkit-scrollbar-thumb:active {
  background: rgba(155, 155, 155, 1);
}

/* Dark mode scrollbar for artifact viewer */
[data-theme='dark'] .output-content::-webkit-scrollbar-thumb,
[data-theme='dark'] .json-preview::-webkit-scrollbar-thumb,
[data-theme='dark'] .mermaid-container::-webkit-scrollbar-thumb,
[data-theme='dark'] .code-artifact pre::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
}

[data-theme='dark'] .output-content::-webkit-scrollbar-thumb:hover,
[data-theme='dark'] .json-preview::-webkit-scrollbar-thumb:hover,
[data-theme='dark'] .mermaid-container::-webkit-scrollbar-thumb:hover,
[data-theme='dark'] .code-artifact pre::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}

[data-theme='dark'] .output-content::-webkit-scrollbar-thumb:active,
[data-theme='dark'] .json-preview::-webkit-scrollbar-thumb:active,
[data-theme='dark'] .mermaid-container::-webkit-scrollbar-thumb:active,
[data-theme='dark'] .code-artifact pre::-webkit-scrollbar-thumb:active {
  background: rgba(255, 255, 255, 0.7);
}

@media (max-width: 768px) {
  /* Thinner scrollbar on mobile */
  .output-content::-webkit-scrollbar,
  .json-preview::-webkit-scrollbar,
  .mermaid-container::-webkit-scrollbar,
  .code-artifact pre::-webkit-scrollbar {
    width: 4px;
  }
}
</style>