<template>
  <div v-if="artifacts && artifacts.length > 0" class="artifact-container">
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
        <div v-if="artifact.type === 'code'" class="code-artifact">
          <pre><code :class="`language-${artifact.language || 'text'}`" v-html="highlightCode(artifact.content, artifact.language)"></code></pre>
        </div>
        
        <!-- HTML Artifact -->
        <div v-else-if="artifact.type === 'html'" class="html-artifact" :class="{ fullscreen: isFullscreen(artifact.uuid) }">
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
import { ref, onMounted } from 'vue'
import { NButton, useMessage } from 'naive-ui'
import { Icon } from '@iconify/vue'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

interface Props {
  artifacts: Chat.Artifact[]
  inversion?: boolean
}

const props = defineProps<Props>()

const message = useMessage()
const expandedArtifacts = ref<Set<string>>(new Set())
const fullscreenArtifacts = ref<Set<string>>(new Set())
const htmlRefreshKey = ref<Record<string, number>>({})

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

// Call auto-expand when component mounts
onMounted(() => {
  autoExpandSvgArtifacts()
})

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
</script>

<style scoped>
.artifact-container {
  margin-top: 12px;
  width: 100%;
  max-width: 100%;
  /* Ensure artifact viewer doesn't interfere with message layout */
  contain: layout style;
  /* Prevent z-index issues that could hide buttons */
  position: relative;
  z-index: 1;
}

.artifact-item {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  margin-bottom: 8px;
  overflow: hidden;
  background: var(--artifact-bg);
  width: 100%;
  max-width: 100%;
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
  flex: 1;
  min-width: 0;
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
  white-space: pre;
  -webkit-overflow-scrolling: touch;
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
  max-height: 300px;
  color: var(--text-color);
  overflow: hidden;
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
</style>