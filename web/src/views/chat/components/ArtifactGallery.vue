<template>
  <div class="artifact-gallery">
    <div class="gallery-header">
      <div class="gallery-title">
        <Icon icon="ri:gallery-line" class="gallery-icon" />
        <h2>Artifact Gallery</h2>
        <NBadge :value="filteredArtifacts.length" type="info" />
      </div>
      <div class="gallery-actions">
        <NButton @click="showFilters = !showFilters" size="small">
          <template #icon><Icon icon="ri:filter-line" /></template>
          Filters
        </NButton>
        <NButton @click="exportArtifacts" size="small">
          <template #icon><Icon icon="ri:download-line" /></template>
          Export
        </NButton>
      </div>
    </div>

    <div v-if="showFilters" class="filters-panel">
      <div class="filters-grid">
        <NInput v-model:value="searchQuery" placeholder="Search artifacts..." clearable>
          <template #prefix><Icon icon="ri:search-line" /></template>
        </NInput>
        <NSelect v-model:value="selectedType" :options="typeOptions" clearable />
        <NSelect v-model:value="selectedLanguage" :options="languageOptions" clearable />
        <NSelect v-model:value="selectedSession" :options="sessionOptions" clearable />
      </div>
    </div>

    <div v-if="filteredArtifacts.length === 0" class="empty-state">
      <Icon icon="ri:folder-open-line" class="empty-icon" />
      <h3>No artifacts found</h3>
      <p>Artifacts created in chat messages will appear here.</p>
    </div>

    <div v-else class="artifact-grid">
      <div v-for="artifact in filteredArtifacts" :key="artifact.id" class="artifact-card">
        <div class="card-header">
          <div class="card-type">
            <Icon :icon="getTypeIcon(artifact.type)" class="type-icon" />
            <span>{{ artifact.type }}</span>
          </div>
          <div class="card-actions">
            <NButton size="tiny" @click="previewArtifact(artifact)" circle>
              <template #icon><Icon icon="ri:eye-line" /></template>
            </NButton>
            <NButton v-if="isViewableArtifact(artifact)" size="tiny" @click="viewArtifact(artifact)" circle>
              <template #icon><Icon icon="ri:external-link-line" /></template>
            </NButton>
            <NButton size="tiny" @click="editArtifact(artifact)" circle>
              <template #icon><Icon icon="ri:edit-line" /></template>
            </NButton>
            <NButton size="tiny" @click="duplicateArtifact(artifact)" circle>
              <template #icon><Icon icon="ri:file-copy-line" /></template>
            </NButton>
            <NButton size="tiny" @click="deleteArtifact(artifact)" circle type="error">
              <template #icon><Icon icon="ri:delete-bin-line" /></template>
            </NButton>
          </div>
        </div>

        <div class="card-content">
          <h4 class="artifact-title">{{ artifact.title || 'Untitled' }}</h4>
          <div class="artifact-meta">
            <span>{{ formatDate(artifact.createdAt) }}</span>
            <span v-if="artifact.language">{{ artifact.language }}</span>
            <span v-if="artifact.sessionTitle">{{ artifact.sessionTitle }}</span>
          </div>
          <pre class="artifact-preview">{{ truncateCode(artifact.content, 180) }}</pre>
        </div>
      </div>
    </div>

    <NModal v-model:show="showPreviewModal" :mask-closable="false">
      <NCard style="width: 90vw; max-width: 1200px; max-height: 90vh" :title="previewingArtifact?.title || 'Artifact Preview'">
        <ArtifactViewer v-if="previewingArtifact" :artifacts="[toViewerArtifact(previewingArtifact)]" />
        <template #footer>
          <div class="modal-actions">
            <NButton @click="showPreviewModal = false">Close</NButton>
            <NButton v-if="previewingArtifact" type="primary" @click="editArtifact(previewingArtifact)">Edit</NButton>
          </div>
        </template>
      </NCard>
    </NModal>

    <NModal v-model:show="showEditModal" :mask-closable="false">
      <NCard style="width: 90vw; max-width: 1200px; max-height: 90vh" :title="editingArtifact?.title || 'Edit Artifact'">
        <ArtifactEditor
          v-if="editingArtifact"
          v-model="editingArtifact.content"
          :language="editingArtifact.language || 'text'"
          :title="editingArtifact.title"
          :artifact-id="editingArtifact.id"
        />
        <template #footer>
          <div class="modal-actions">
            <NButton @click="cancelEdit">Cancel</NButton>
            <NButton type="primary" @click="saveEdit">Save Changes</NButton>
          </div>
        </template>
      </NCard>
    </NModal>

    <NModal v-model:show="showViewModal" :mask-closable="false">
      <NCard style="width: 90vw; max-width: 1200px; max-height: 90vh" :title="viewingArtifact?.title || 'View Artifact'">
        <ArtifactViewer v-if="viewingArtifact" :artifacts="[toViewerArtifact(viewingArtifact)]" />
        <template #footer>
          <div class="modal-actions">
            <NButton @click="showViewModal = false">Close</NButton>
          </div>
        </template>
      </NCard>
    </NModal>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { NBadge, NButton, NCard, NInput, NModal, NSelect, useDialog, useMessage } from 'naive-ui'
import { Icon } from '@iconify/vue'
import ArtifactViewer from './Message/ArtifactViewer.vue'
import ArtifactEditor from './Message/ArtifactEditor.vue'
import { useMessageStore, useSessionStore } from '@/store'

interface ArtifactRecord {
  uuid: string
  id: string
  title: string
  content: string
  type: string
  language?: string
  createdAt: string
  updatedAt?: string
  sessionUuid?: string
  messageUuid?: string
  sessionTitle?: string
}

const message = useMessage()
const dialog = useDialog()
const messageStore = useMessageStore()
const sessionStore = useSessionStore()

const showFilters = ref(false)
const showPreviewModal = ref(false)
const showEditModal = ref(false)
const showViewModal = ref(false)

const searchQuery = ref('')
const selectedType = ref('')
const selectedLanguage = ref('')
const selectedSession = ref('')

const artifacts = ref<ArtifactRecord[]>([])
const previewingArtifact = ref<ArtifactRecord | null>(null)
const viewingArtifact = ref<ArtifactRecord | null>(null)
const editingArtifact = ref<ArtifactRecord | null>(null)
const originalArtifact = ref<ArtifactRecord | null>(null)

const typeOptions = computed(() => [
  { label: 'All Types', value: '' },
  ...[...new Set(artifacts.value.map(artifact => artifact.type))].map(type => ({ label: type, value: type }))
])

const languageOptions = computed(() => [
  { label: 'All Languages', value: '' },
  ...[...new Set(artifacts.value.map(artifact => artifact.language).filter(Boolean))].map(language => ({
    label: language,
    value: language
  }))
])

const sessionOptions = computed(() => [
  { label: 'All Sessions', value: '' },
  ...[...new Set(artifacts.value.map(artifact => artifact.sessionTitle).filter(Boolean))].map(sessionTitle => ({
    label: sessionTitle,
    value: sessionTitle
  }))
])

const filteredArtifacts = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()

  return artifacts.value.filter(artifact => {
    if (selectedType.value && artifact.type !== selectedType.value) return false
    if (selectedLanguage.value && artifact.language !== selectedLanguage.value) return false
    if (selectedSession.value && artifact.sessionTitle !== selectedSession.value) return false

    if (!query) return true

    return [
      artifact.title,
      artifact.content,
      artifact.type,
      artifact.language || '',
      artifact.sessionTitle || ''
    ].some(value => value.toLowerCase().includes(query))
  })
})

const getTypeIcon = (type: string) => {
  const icons: Record<string, string> = {
    code: 'ri:code-line',
    html: 'ri:html5-line',
    svg: 'ri:image-line',
    json: 'ri:file-code-line',
    mermaid: 'ri:flow-chart',
    markdown: 'ri:markdown-line',
  }
  return icons[type] || 'ri:file-line'
}

const formatDate = (value: string) => {
  const date = new Date(value)
  return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
}

const truncateCode = (code: string, limit: number) => (
  code.length <= limit ? code : `${code.slice(0, limit)}...`
)

const isViewableArtifact = (artifact: ArtifactRecord) => (
  ['html', 'svg', 'mermaid', 'json', 'markdown'].includes(artifact.type)
)

const toViewerArtifact = (artifact: ArtifactRecord): Chat.Artifact => ({
  uuid: artifact.uuid,
  type: artifact.type,
  title: artifact.title,
  content: artifact.content,
  language: artifact.language,
})

const previewArtifact = (artifact: ArtifactRecord) => {
  previewingArtifact.value = artifact
  showPreviewModal.value = true
}

const viewArtifact = (artifact: ArtifactRecord) => {
  viewingArtifact.value = artifact
  showViewModal.value = true
}

const editArtifact = (artifact: ArtifactRecord) => {
  previewingArtifact.value = null
  showPreviewModal.value = false
  originalArtifact.value = artifact
  editingArtifact.value = { ...artifact }
  showEditModal.value = true
}

const duplicateArtifact = (artifact: ArtifactRecord) => {
  artifacts.value.unshift({
    ...artifact,
    uuid: `${artifact.uuid}-copy-${Date.now()}`,
    id: `${artifact.id}-copy-${Date.now()}`,
    title: `${artifact.title} (Copy)`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    sessionUuid: undefined,
    messageUuid: undefined,
    sessionTitle: undefined,
  })
  message.success('Artifact duplicated successfully')
}

const deleteArtifact = (artifact: ArtifactRecord) => {
  dialog.warning({
    title: 'Delete Artifact',
    content: `Delete "${artifact.title}"?`,
    positiveText: 'Delete',
    negativeText: 'Cancel',
    onPositiveClick: () => {
      if (artifact.sessionUuid && artifact.messageUuid) {
        const sessionMessages = messageStore.getChatSessionDataByUuid(artifact.sessionUuid)
        const targetMessage = sessionMessages?.find(entry => entry.uuid === artifact.messageUuid)
        if (targetMessage?.artifacts) {
          targetMessage.artifacts = targetMessage.artifacts.filter(entry => entry.uuid !== artifact.uuid)
        }
      }

      artifacts.value = artifacts.value.filter(entry => entry.id !== artifact.id)
      message.success('Artifact deleted successfully')
    }
  })
}

const saveEdit = () => {
  if (!editingArtifact.value || !originalArtifact.value) return

  if (editingArtifact.value.sessionUuid && editingArtifact.value.messageUuid) {
    const sessionMessages = messageStore.getChatSessionDataByUuid(editingArtifact.value.sessionUuid)
    const targetMessage = sessionMessages?.find(entry => entry.uuid === editingArtifact.value?.messageUuid)
    const targetArtifact = targetMessage?.artifacts?.find(entry => entry.uuid === editingArtifact.value?.uuid)
    if (targetArtifact) {
      targetArtifact.title = editingArtifact.value.title
      targetArtifact.content = editingArtifact.value.content
      targetArtifact.language = editingArtifact.value.language
    }
  }

  Object.assign(originalArtifact.value, {
    ...editingArtifact.value,
    updatedAt: new Date().toISOString(),
  })

  showEditModal.value = false
  editingArtifact.value = null
  originalArtifact.value = null
  loadArtifacts()
  message.success('Artifact saved successfully')
}

const cancelEdit = () => {
  showEditModal.value = false
  editingArtifact.value = null
  originalArtifact.value = null
}

const exportArtifacts = () => {
  const blob = new Blob([JSON.stringify(filteredArtifacts.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `artifacts_${new Date().toISOString().split('T')[0]}.json`
  anchor.click()
  URL.revokeObjectURL(url)
}

const loadArtifacts = () => {
  const nextArtifacts: ArtifactRecord[] = []

  sessionStore.getAllSessions().forEach(session => {
    const messages = messageStore.getChatSessionDataByUuid(session.uuid) || []
    messages.forEach(chatMessage => {
      chatMessage.artifacts?.forEach(artifact => {
        nextArtifacts.push({
          uuid: artifact.uuid,
          id: artifact.uuid,
          title: artifact.title || 'Untitled Artifact',
          content: artifact.content,
          type: artifact.type,
          language: artifact.language,
          createdAt: chatMessage.dateTime,
          updatedAt: chatMessage.dateTime,
          sessionUuid: session.uuid,
          messageUuid: chatMessage.uuid,
          sessionTitle: session.title,
        })
      })
    })
  })

  artifacts.value = nextArtifacts.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
}

watch(
  () => [
    sessionStore.getAllSessions().length,
    Object.values(messageStore.chat).reduce((sum, entries) => sum + entries.length, 0),
  ],
  loadArtifacts,
  { immediate: true }
)
</script>

<style scoped>
.artifact-gallery {
  padding: 1rem;
}

.gallery-header,
.gallery-title,
.gallery-actions,
.filters-grid,
.card-header,
.card-actions,
.modal-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.gallery-header {
  justify-content: space-between;
  margin-bottom: 1rem;
}

.gallery-title h2 {
  margin: 0;
}

.filters-panel {
  margin-bottom: 1rem;
}

.filters-grid {
  flex-wrap: wrap;
}

.filters-grid > * {
  min-width: 220px;
  flex: 1;
}

.artifact-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1rem;
}

.artifact-card {
  border: 1px solid #e5e7eb;
  border-radius: 0.75rem;
  background: #fff;
  overflow: hidden;
}

.card-header,
.card-content {
  padding: 1rem;
}

.card-header {
  justify-content: space-between;
  border-bottom: 1px solid #e5e7eb;
}

.artifact-meta {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  font-size: 0.875rem;
  color: #6b7280;
  margin-bottom: 0.75rem;
}

.artifact-preview {
  margin: 0;
  max-height: 220px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.empty-state {
  text-align: center;
  padding: 4rem 1rem;
  color: #6b7280;
}

.empty-icon,
.gallery-icon,
.type-icon {
  font-size: 1.25rem;
}

.modal-actions {
  justify-content: flex-end;
}
</style>
