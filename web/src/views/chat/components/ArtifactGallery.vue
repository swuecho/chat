<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { NBadge, NButton, NCard, NInput, NModal, NPagination, NSelect, NSpin, useDialog, useMessage } from 'naive-ui'
import { Icon } from '@iconify/vue'
import ArtifactViewer from './Message/ArtifactViewerBase.vue'
import ArtifactEditor from './Message/ArtifactEditor.vue'
import { useMessageStore, useSessionStore } from '@/store'
import { deleteArtifact as deleteArtifactRequest, duplicateArtifact as duplicateArtifactRequest, listArtifacts, updateArtifact } from '@/api/generated_client'
import { useWorkspaceRouting } from '@/hooks/useWorkspaceRouting'
import { t } from '@/locales'

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

const emit = defineEmits<{ close: [] }>()

const message = useMessage()
const dialog = useDialog()
const messageStore = useMessageStore()
const sessionStore = useSessionStore()
const { navigateToSession } = useWorkspaceRouting()

const showFilters = ref(false)
const showPreviewModal = ref(false)
const showEditModal = ref(false)

const searchQuery = ref('')
const selectedType = ref('')
const selectedLanguage = ref('')
const selectedSession = ref('')
const currentPage = ref(1)
const pageSize = 24
const totalArtifacts = ref(0)
const loading = ref(false)

const artifacts = ref<ArtifactRecord[]>([])
const previewingArtifact = ref<ArtifactRecord | null>(null)
const editingArtifact = ref<ArtifactRecord | null>(null)
const originalArtifact = ref<ArtifactRecord | null>(null)
const canSaveEdit = computed(() => Boolean(
  editingArtifact.value?.title.trim()
  && new Blob([editingArtifact.value.title]).size <= 200
  && new Blob([editingArtifact.value.content]).size <= 1024 * 1024,
))

const typeOptions = computed(() => [
  { label: t('artifact.allTypes'), value: '' },
  ...['code', 'html', 'svg', 'mermaid', 'json', 'markdown'].map(type => ({ label: type, value: type })),
])

const languageOptions = computed(() => [
  { label: t('artifact.allLanguages'), value: '' },
  ...['javascript', 'typescript', 'python', 'html', 'svg', 'mermaid', 'json', 'markdown', 'text'].map(language => ({
    label: language,
    value: language,
  })),
])

const sessionOptions = computed(() => [
  { label: t('artifact.allSessions'), value: '' },
  ...sessionStore.getAllSessions().map(session => ({
    label: session.title,
    value: session.uuid,
  })),
])

const filteredArtifacts = computed(() => {
  return artifacts.value
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

const toViewerArtifact = (artifact: ArtifactRecord): Chat.Artifact => ({
  uuid: artifact.uuid,
  type: artifact.type,
  title: artifact.title,
  content: artifact.content,
  language: artifact.language,
})

const getSourceMessage = (artifact: ArtifactRecord) => {
  if (!artifact.sessionUuid || !artifact.messageUuid)
    return undefined
  return messageStore.getChatSessionDataByUuid(artifact.sessionUuid)?.find(entry => entry.uuid === artifact.messageUuid)
}

const previewArtifact = (artifact: ArtifactRecord) => {
  previewingArtifact.value = artifact
  showPreviewModal.value = true
}

const editArtifact = (artifact: ArtifactRecord) => {
  previewingArtifact.value = null
  showPreviewModal.value = false
  originalArtifact.value = artifact
  editingArtifact.value = { ...artifact }
  showEditModal.value = true
}

const downloadArtifact = (artifact: ArtifactRecord) => {
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

const openSourceChat = async (artifact: ArtifactRecord) => {
  if (artifact.sessionUuid) {
    await navigateToSession(artifact.sessionUuid)
    emit('close')
  }
}

const duplicateArtifact = async (artifact: ArtifactRecord) => {
  try {
    const result = await duplicateArtifactRequest({ path: { uuid: artifact.uuid } })
    const targetMessage = getSourceMessage(artifact)
    if (targetMessage) {
      targetMessage.artifacts = [...(targetMessage.artifacts || []), {
        uuid: result.uuid,
        type: artifact.type,
        title: `${artifact.title} (Copy)`,
        content: artifact.content,
        language: artifact.language,
      }]
    }
    await loadArtifacts()
    message.success(t('artifact.duplicateSuccess'))
  }
  catch {
    message.error(t('artifact.duplicateFailed'))
  }
}

const deleteArtifact = (artifact: ArtifactRecord) => {
  dialog.warning({
    title: t('artifact.delete'),
    content: t('artifact.deleteConfirm', { title: artifact.title }),
    positiveText: t('artifact.delete'),
    negativeText: t('artifact.cancel'),
    onPositiveClick: async () => {
      try {
        await deleteArtifactRequest({ path: { uuid: artifact.uuid } })
        const targetMessage = getSourceMessage(artifact)
        if (targetMessage?.artifacts)
          targetMessage.artifacts = targetMessage.artifacts.filter(entry => entry.uuid !== artifact.uuid)
        await loadArtifacts()
        message.success(t('artifact.deleteSuccess'))
      }
      catch {
        message.error(t('artifact.deleteFailed'))
      }
    },
  })
}

const saveEdit = async () => {
  if (!editingArtifact.value || !originalArtifact.value)
    return

  try {
    await updateArtifact({
      path: { uuid: editingArtifact.value.uuid },
      body: {
        title: editingArtifact.value.title,
        content: editingArtifact.value.content,
        language: editingArtifact.value.language || '',
      },
    })
    const targetMessage = getSourceMessage(editingArtifact.value)
    const targetArtifact = targetMessage?.artifacts?.find(entry => entry.uuid === editingArtifact.value?.uuid)
    if (targetArtifact) {
      targetArtifact.title = editingArtifact.value.title
      targetArtifact.content = editingArtifact.value.content
      targetArtifact.language = editingArtifact.value.language
    }
  }
  catch {
    message.error(t('artifact.saveFailed'))
    return
  }

  Object.assign(originalArtifact.value, {
    ...editingArtifact.value,
    updatedAt: new Date().toISOString(),
  })

  showEditModal.value = false
  editingArtifact.value = null
  originalArtifact.value = null
  await loadArtifacts()
  message.success(t('artifact.saveSuccess'))
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
  anchor.download = `artifacts_page_${currentPage.value}_${new Date().toISOString().split('T')[0]}.json`
  anchor.click()
  URL.revokeObjectURL(url)
}

const loadArtifacts = async () => {
  loading.value = true
  try {
    const result = await listArtifacts({
      query: {
        search: searchQuery.value.trim() || undefined,
        type: selectedType.value || undefined,
        language: selectedLanguage.value || undefined,
        sessionUuid: selectedSession.value || undefined,
        limit: pageSize,
        offset: (currentPage.value - 1) * pageSize,
      },
    })
    if (result.total > 0 && result.items.length === 0 && currentPage.value > 1) {
      currentPage.value--
      return
    }
    totalArtifacts.value = result.total
    artifacts.value = result.items.map(artifact => ({ ...artifact, id: artifact.uuid }))
  }
  catch {
    message.error(t('artifact.loadFailed'))
  }
  finally {
    loading.value = false
  }
}

watch(
  currentPage,
  loadArtifacts,
  { immediate: true },
)

watch([selectedType, selectedLanguage, selectedSession], () => {
  if (currentPage.value !== 1)
    currentPage.value = 1
  else
    loadArtifacts()
})

let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(searchQuery, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    if (currentPage.value !== 1)
      currentPage.value = 1
    else
      loadArtifacts()
  }, 250)
})
</script>

<template>
  <div class="artifact-gallery">
    <div class="gallery-header">
      <div class="gallery-title">
        <Icon icon="ri:gallery-line" class="gallery-icon" />
        <h2>{{ $t('artifact.gallery') }}</h2>
        <NBadge :value="totalArtifacts" type="info" />
      </div>
      <div class="gallery-actions">
        <NButton size="small" @click="showFilters = !showFilters">
          <template #icon>
            <Icon icon="ri:filter-line" />
          </template>
          {{ $t('artifact.filters') }}
        </NButton>
        <NButton size="small" @click="exportArtifacts">
          <template #icon>
            <Icon icon="ri:download-line" />
          </template>
          {{ $t('artifact.exportPage') }}
        </NButton>
      </div>
    </div>

    <div v-if="showFilters" class="filters-panel">
      <div class="filters-grid">
        <NInput v-model:value="searchQuery" :placeholder="$t('artifact.search')" clearable>
          <template #prefix>
            <Icon icon="ri:search-line" />
          </template>
        </NInput>
        <NSelect v-model:value="selectedType" :options="typeOptions" clearable />
        <NSelect v-model:value="selectedLanguage" :options="languageOptions" clearable />
        <NSelect v-model:value="selectedSession" :options="sessionOptions" clearable />
      </div>
    </div>

    <NSpin :show="loading">
      <div v-if="filteredArtifacts.length === 0" class="empty-state">
        <Icon icon="ri:folder-open-line" class="empty-icon" />
        <h3>{{ $t('artifact.empty') }}</h3>
        <p>{{ $t('artifact.emptyDescription') }}</p>
      </div>

      <div v-else class="artifact-grid">
        <div v-for="artifact in filteredArtifacts" :key="artifact.id" class="artifact-card">
          <div class="card-header">
            <div class="card-type">
              <Icon :icon="getTypeIcon(artifact.type)" class="type-icon" />
              <span>{{ artifact.type }}</span>
            </div>
            <div class="card-actions">
              <NButton size="tiny" circle :title="$t('artifact.preview')" :aria-label="$t('artifact.preview')" @click="previewArtifact(artifact)">
                <template #icon>
                  <Icon icon="ri:eye-line" />
                </template>
              </NButton>
              <NButton size="tiny" circle :title="$t('artifact.openChat')" :aria-label="$t('artifact.openChat')" @click="openSourceChat(artifact)">
                <template #icon>
                  <Icon icon="ri:external-link-line" />
                </template>
              </NButton>
              <NButton size="tiny" circle :title="$t('artifact.edit')" :aria-label="$t('artifact.edit')" @click="editArtifact(artifact)">
                <template #icon>
                  <Icon icon="ri:edit-line" />
                </template>
              </NButton>
              <NButton size="tiny" circle :title="$t('artifact.duplicate')" :aria-label="$t('artifact.duplicate')" @click="duplicateArtifact(artifact)">
                <template #icon>
                  <Icon icon="ri:file-copy-line" />
                </template>
              </NButton>
              <NButton size="tiny" circle :title="$t('artifact.download')" :aria-label="$t('artifact.download')" @click="downloadArtifact(artifact)">
                <template #icon>
                  <Icon icon="ri:download-line" />
                </template>
              </NButton>
              <NButton size="tiny" circle type="error" :title="$t('artifact.delete')" :aria-label="$t('artifact.delete')" @click="deleteArtifact(artifact)">
                <template #icon>
                  <Icon icon="ri:delete-bin-line" />
                </template>
              </NButton>
            </div>
          </div>

          <div class="card-content">
            <h4 class="artifact-title">
              {{ artifact.title || $t('artifact.untitled') }}
            </h4>
            <div class="artifact-meta">
              <span>{{ formatDate(artifact.createdAt) }}</span>
              <span v-if="artifact.language">{{ artifact.language }}</span>
              <span v-if="artifact.sessionTitle">{{ artifact.sessionTitle }}</span>
            </div>
            <pre class="artifact-preview">{{ truncateCode(artifact.content, 180) }}</pre>
          </div>
        </div>
      </div>

      <div v-if="totalArtifacts > pageSize" class="pagination-row">
        <NPagination v-model:page="currentPage" :page-size="pageSize" :item-count="totalArtifacts" />
      </div>
    </NSpin>

    <NModal v-model:show="showPreviewModal" :mask-closable="false">
      <NCard class="artifact-modal" :title="previewingArtifact?.title || $t('artifact.preview')">
        <ArtifactViewer v-if="previewingArtifact" :artifacts="[toViewerArtifact(previewingArtifact)]" />
        <template #footer>
          <div class="modal-actions">
            <NButton @click="showPreviewModal = false">
              {{ $t('artifact.close') }}
            </NButton>
            <NButton v-if="previewingArtifact" type="primary" @click="editArtifact(previewingArtifact)">
              {{ $t('artifact.edit') }}
            </NButton>
          </div>
        </template>
      </NCard>
    </NModal>

    <NModal v-model:show="showEditModal" :mask-closable="false">
      <NCard class="artifact-modal" :title="editingArtifact?.title || $t('artifact.edit')">
        <ArtifactEditor
          v-if="editingArtifact"
          v-model="editingArtifact.content"
          :language="editingArtifact.language || 'text'"
          :title="editingArtifact.title"
          @update:language="editingArtifact.language = $event"
          @update:title="editingArtifact.title = $event"
        />
        <template #footer>
          <div class="modal-actions">
            <NButton @click="cancelEdit">
              {{ $t('artifact.cancel') }}
            </NButton>
            <NButton type="primary" :disabled="!canSaveEdit" @click="saveEdit">
              {{ $t('artifact.save') }}
            </NButton>
          </div>
        </template>
      </NCard>
    </NModal>
  </div>
</template>

<style scoped>
.artifact-gallery {
  padding: 1rem;
}

.artifact-modal {
  width: min(1200px, 94vw);
  max-height: 92vh;
  overflow: auto;
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

.pagination-row {
  display: flex;
  justify-content: center;
  margin-top: 1.25rem;
}

.empty-icon,
.gallery-icon,
.type-icon {
  font-size: 1.25rem;
}

@media (max-width: 640px) {
  .artifact-gallery {
    padding: 0.75rem;
  }

  .gallery-header,
  .card-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .artifact-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .card-actions {
    flex-wrap: wrap;
  }
}

.modal-actions {
  justify-content: flex-end;
}
</style>
