<template>
  <div v-if="artifacts && artifacts.length > 0" class="artifact-container" data-test-role="artifact-viewer">
    <div v-for="artifact in artifacts" :key="artifact.uuid" class="artifact-item">
      <ArtifactHeader
        :artifact="artifact"
        :is-expanded="isExpanded(artifact.uuid)"
        @toggle-expand="toggleExpanded"
        @copy-content="copyContent"
        @open-in-new-window="openInNewWindow"
      />

      <ArtifactContent
        v-if="isExpanded(artifact.uuid)"
        :artifact="artifact"
        :is-editing="isEditing(artifact.uuid)"
        :editable-content="editableContent[artifact.uuid]"
        @toggle-edit="toggleEdit"
        @save-edit="saveEdit"
        @cancel-edit="cancelEdit"
        @update-editable-content="updateEditableContent"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { type Artifact } from '@/utils/artifacts'
import { copyText } from '@/utils/format'
import { sanitizeHtml } from '@/utils/sanitize'
import ArtifactHeader from './ArtifactHeader.vue'
import ArtifactContent from './ArtifactContent.vue'

interface Props {
  artifacts: Artifact[]
}

defineProps<Props>()

const message = useMessage()
const expandedArtifacts = ref<Set<string>>(new Set())
const editingArtifacts = ref<Set<string>>(new Set())
const editableContent = reactive<Record<string, string>>({})

const isExpanded = (uuid: string) => expandedArtifacts.value.has(uuid)
const isEditing = (uuid: string) => editingArtifacts.value.has(uuid)

const toggleExpanded = (uuid: string) => {
  if (expandedArtifacts.value.has(uuid)) {
    expandedArtifacts.value.delete(uuid)
    return
  }
  expandedArtifacts.value.add(uuid)
}

const copyContent = async (content: string) => {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(content)
    } else {
      copyText({ text: content, origin: true })
    }
    message.success('Content copied to clipboard')
  } catch {
    message.error('Failed to copy content')
  }
}

const openInNewWindow = (content: string) => {
  const newWindow = window.open('', '_blank')
  if (!newWindow) return

  newWindow.document.write(sanitizeHtml(content))
  newWindow.document.close()
}

const toggleEdit = (uuid: string, content: string) => {
  if (editingArtifacts.value.has(uuid)) {
    editingArtifacts.value.delete(uuid)
    return
  }

  editingArtifacts.value.add(uuid)
  editableContent[uuid] = content
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
