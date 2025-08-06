<template>
  <div v-if="artifacts && artifacts.length > 0" class="artifact-container" data-test-role="artifact-viewer">
    <div v-for="artifact in artifacts" :key="artifact.uuid" class="artifact-item">
      <ArtifactHeader 
        :artifact="artifact"
        :is-expanded="isExpanded(artifact.uuid)"
        :is-running="isRunning(artifact.uuid)"
        :has-output="hasOutput(artifact.uuid)"
        @toggle-expand="toggleExpanded"
        @run-code="runCode"
        @clear-output="clearOutput"
        @copy-content="copyContent"
        @open-in-new-window="openInNewWindow"
      />
      
      <ArtifactContent 
        v-if="isExpanded(artifact.uuid)"
        :artifact="artifact"
        :is-editing="isEditing(artifact.uuid)"
        :editable-content="editableContent[artifact.uuid]"
        :execution-outputs="executionOutputs[artifact.uuid]"
        :show-library-list="showLibraryList[artifact.uuid]"
        :canvas-refs="canvasRefs"
        :matplotlib-refs="matplotlibRefs"
        @toggle-edit="toggleEdit"
        @save-edit="saveEdit"
        @cancel-edit="cancelEdit"
        @update-editable-content="updateEditableContent"
        @toggle-library-list="toggleLibraryList"
        @run-code="runCode"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue'
import { useMessage } from 'naive-ui'
import { type Artifact } from '@/utils/artifacts'
import { getCodeRunner, type ExecutionResult } from '@/services/codeRunner'
import { copyText } from '@/utils/format'
import ArtifactHeader from './ArtifactHeader.vue'
import ArtifactContent from './ArtifactContent.vue'

interface Props {
  artifacts: Artifact[]
}

const props = defineProps<Props>()

const message = useMessage()

// State management
const expandedArtifacts = ref<Set<string>>(new Set())
const runningArtifacts = ref<Set<string>>(new Set())
const editingArtifacts = ref<Set<string>>(new Set())
const editableContent = reactive<Record<string, string>>({})
const executionOutputs = reactive<Record<string, ExecutionResult[]>>({})
const showLibraryList = reactive<Record<string, boolean>>({})
const canvasRefs = reactive<Record<string, HTMLCanvasElement>>({})
const matplotlibRefs = reactive<Record<string, HTMLImageElement>>({})

// Computed properties
const isExpanded = (uuid: string) => expandedArtifacts.value.has(uuid)
const isRunning = (uuid: string) => runningArtifacts.value.has(uuid)
const isEditing = (uuid: string) => editingArtifacts.value.has(uuid)
const hasOutput = (uuid: string) => executionOutputs[uuid]?.length > 0

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

// Actions
const toggleExpanded = (uuid: string) => {
  if (expandedArtifacts.value.has(uuid)) {
    expandedArtifacts.value.delete(uuid)
  } else {
    expandedArtifacts.value.add(uuid)
  }
}

const runCode = async (artifact: Artifact) => {
  if (!isExecutable(artifact) || isRunning(artifact.uuid)) return

  runningArtifacts.value.add(artifact.uuid)
  
  try {
    const runner = getCodeRunner(artifact.language || 'javascript')
    const result = await runner.run(artifact.content)
    
    if (!executionOutputs[artifact.uuid]) {
      executionOutputs[artifact.uuid] = []
    }
    executionOutputs[artifact.uuid].push(result)
    
    message.success('Code executed successfully')
  } catch (error) {
    const errorResult: ExecutionResult = {
      id: Date.now().toString(),
      type: 'error',
      content: error instanceof Error ? error.message : 'Unknown error',
      timestamp: new Date().toISOString(),
    }
    
    if (!executionOutputs[artifact.uuid]) {
      executionOutputs[artifact.uuid] = []
    }
    executionOutputs[artifact.uuid].push(errorResult)
    
    message.error('Code execution failed')
  } finally {
    runningArtifacts.value.delete(artifact.uuid)
  }
}

const clearOutput = (uuid: string) => {
  executionOutputs[uuid] = []
}

const copyContent = async (content: string) => {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(content)
    } else {
      copyText({ text: content, origin: true })
    }
    message.success('Content copied to clipboard')
  } catch (error) {
    message.error('Failed to copy content')
  }
}

const openInNewWindow = (content: string) => {
  const newWindow = window.open('', '_blank')
  if (newWindow) {
    newWindow.document.write(content)
    newWindow.document.close()
  }
}

const toggleEdit = (uuid: string, content: string) => {
  if (editingArtifacts.value.has(uuid)) {
    editingArtifacts.value.delete(uuid)
  } else {
    editingArtifacts.value.add(uuid)
    editableContent[uuid] = content
  }
}

const saveEdit = (uuid: string) => {
  editingArtifacts.value.delete(uuid)
  message.success('Changes saved')
}

const cancelEdit = (uuid: string) => {
  editingArtifacts.value.delete(uuid)
  delete editableContent[uuid]
}

const updateEditableContent = (uuid: string, content: string) => {
  editableContent[uuid] = content
}

const toggleLibraryList = (uuid: string) => {
  showLibraryList[uuid] = !showLibraryList[uuid]
}
</script>

<style scoped>
.artifact-container {
  margin-top: 1rem;
}

.artifact-item {
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  overflow: hidden;
  background: white;
}

.artifact-item:hover {
  border-color: #d1d5db;
}

:deep(.dark) .artifact-item {
  background: #1f2937;
  border-color: #374151;
}

:deep(.dark) .artifact-item:hover {
  border-color: #4b5563;
}
</style>