<template>
  <div v-if="artifacts && artifacts.length > 0" class="artifact-container">
    <div v-for="artifact in artifacts" :key="artifact.uuid" class="artifact-item">
      <div class="artifact-header">
        <div class="artifact-title">
          <Icon :icon="getArtifactIcon(artifact.type)" class="artifact-icon" />
          <span>{{ artifact.title }}</span>
          <span class="artifact-type">({{ artifact.type }})</span>
        </div>
        <div class="artifact-actions">
          <NButton size="small" @click="toggleExpanded(artifact.uuid)">
            {{ isExpanded(artifact.uuid) ? 'Collapse' : 'Expand' }}
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
        <div v-else-if="artifact.type === 'html'" class="html-artifact">
          <div class="html-preview">
            <iframe 
              :srcdoc="artifact.content" 
              sandbox="allow-scripts allow-same-origin"
              class="html-frame"
            ></iframe>
          </div>
        </div>
        
        <!-- SVG Artifact -->
        <div v-else-if="artifact.type === 'svg'" class="svg-artifact">
          <div class="svg-container" v-html="artifact.content"></div>
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
import { ref } from 'vue'
import { NButton, useMessage } from 'naive-ui'
import { Icon } from '@iconify/vue'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

interface Props {
  artifacts: Chat.Artifact[]
  inversion?: boolean
}

defineProps<Props>()

const message = useMessage()
const expandedArtifacts = ref<Set<string>>(new Set())

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
</script>

<style scoped>
.artifact-container {
  margin-top: 12px;
}

.artifact-item {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  margin-bottom: 8px;
  overflow: hidden;
  background: var(--artifact-bg);
}

.artifact-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
}

.artifact-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: var(--text-color);
}

.artifact-icon {
  font-size: 16px;
  color: var(--primary-color);
}

.artifact-type {
  font-size: 12px;
  color: var(--text-color-secondary);
  background: var(--tag-bg);
  padding: 2px 6px;
  border-radius: 4px;
}

.artifact-actions {
  display: flex;
  gap: 8px;
}

.artifact-content {
  padding: 16px;
}

.code-artifact pre {
  margin: 0;
  padding: 16px;
  background: var(--code-bg);
  border-radius: 6px;
  overflow-x: auto;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.code-artifact code {
  font-size: 13px;
  line-height: 1.5;
}

.html-artifact {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
}

.html-frame {
  width: 100%;
  height: 300px;
  border: none;
  background: white;
}

.svg-artifact {
  display: flex;
  justify-content: center;
  padding: 16px;
  background: var(--artifact-content-bg);
  border-radius: 6px;
}

.svg-container {
  max-width: 100%;
  max-height: 400px;
}

.text-artifact pre {
  margin: 0;
  padding: 16px;
  background: var(--code-bg);
  border-radius: 6px;
  overflow-x: auto;
  white-space: pre-wrap;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
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