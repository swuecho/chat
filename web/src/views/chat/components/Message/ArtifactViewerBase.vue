<script lang="ts" setup>
import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import ArtifactHeader from './ArtifactHeader.vue'
import ArtifactContent from './ArtifactContent.vue'
import { type Artifact } from '@/utils/artifacts'
import { copyText } from '@/utils/format'

interface Props {
  artifacts: Artifact[]
}

defineProps<Props>()

const message = useMessage()
const expandedArtifacts = ref<Set<string>>(new Set())
const isExpanded = (uuid: string) => expandedArtifacts.value.has(uuid)

const toggleExpanded = (uuid: string) => {
  if (expandedArtifacts.value.has(uuid)) {
    expandedArtifacts.value.delete(uuid)
    return
  }
  expandedArtifacts.value.add(uuid)
}

const copyContent = async (content: string) => {
  try {
    if (navigator.clipboard?.writeText)
      await navigator.clipboard.writeText(content)
    else
      copyText({ text: content, origin: true })

    message.success('Content copied to clipboard')
  }
  catch {
    message.error('Failed to copy content')
  }
}

const downloadContent = (artifact: Artifact) => {
  const extensions: Record<string, string> = { javascript: 'js', typescript: 'ts', python: 'py', markdown: 'md', text: 'txt' }
  const extension = extensions[artifact.language || ''] || artifact.language || artifact.type || 'txt'
  const filename = (artifact.title || 'artifact').replace(/[^a-z0-9_-]+/gi, '-').replace(/^-|-$/g, '') || 'artifact'
  const url = URL.createObjectURL(new Blob([artifact.content], { type: 'text/plain;charset=utf-8' }))
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `${filename}.${extension}`
  anchor.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div v-if="artifacts && artifacts.length > 0" class="artifact-container" data-test-role="artifact-viewer">
    <div v-for="artifact in artifacts" :key="artifact.uuid" class="artifact-item">
      <ArtifactHeader
        :artifact="artifact"
        :is-expanded="isExpanded(artifact.uuid)"
        @toggle-expand="toggleExpanded"
        @copy-content="copyContent"
        @download-content="downloadContent"
      />

      <ArtifactContent
        v-if="isExpanded(artifact.uuid)"
        :artifact="artifact"
      />
    </div>
  </div>
</template>

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
