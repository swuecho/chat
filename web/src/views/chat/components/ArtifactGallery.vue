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
          <template #icon>
            <Icon icon="ri:filter-line" />
          </template>
          Filters
        </NButton>
        <NButton @click="showStats = !showStats" size="small">
          <template #icon>
            <Icon icon="ri:bar-chart-line" />
          </template>
          Statistics
        </NButton>
        <NButton @click="exportArtifacts" size="small">
          <template #icon>
            <Icon icon="ri:download-line" />
          </template>
          Export
        </NButton>
      </div>
    </div>

    <!-- Filters Panel -->
    <div v-if="showFilters" class="filters-panel">
      <div class="filters-grid">
        <div class="filter-group">
          <label>Search</label>
          <NInput v-model:value="searchQuery" placeholder="Search artifacts..." clearable>
            <template #prefix>
              <Icon icon="ri:search-line" />
            </template>
          </NInput>
        </div>
        <div class="filter-group">
          <label>Type</label>
          <NSelect v-model:value="selectedType" :options="typeOptions" clearable />
        </div>
        <div class="filter-group">
          <label>Language</label>
          <NSelect v-model:value="selectedLanguage" :options="languageOptions" clearable />
        </div>
        <div class="filter-group">
          <label>Date Range</label>
          <NSelect v-model:value="selectedDateRange" :options="dateRangeOptions" clearable />
        </div>
        <div class="filter-group">
          <label>Session</label>
          <NSelect v-model:value="selectedSession" :options="sessionOptions" clearable />
        </div>
        <div class="filter-group">
          <label>Sort By</label>
          <NSelect v-model:value="sortBy" :options="sortOptions" />
        </div>
        <div class="filter-group">
          <label>View</label>
          <NButtonGroup>
            <NButton :type="viewMode === 'grid' ? 'primary' : 'default'" @click="viewMode = 'grid'">
              <template #icon>
                <Icon icon="ri:grid-line" />
              </template>
              Grid
            </NButton>
            <NButton :type="viewMode === 'list' ? 'primary' : 'default'" @click="viewMode = 'list'">
              <template #icon>
                <Icon icon="ri:list-check" />
              </template>
              List
            </NButton>
          </NButtonGroup>
        </div>
      </div>
    </div>

    <!-- Statistics Panel -->
    <div v-if="showStats" class="stats-panel">
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-value">{{ galleryStats.totalArtifacts }}</div>
          <div class="stat-label">Total Artifacts</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ galleryStats.totalExecutions }}</div>
          <div class="stat-label">Total Executions</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ Math.round(galleryStats.averageExecutionTime) }}ms</div>
          <div class="stat-label">Avg Execution Time</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ Math.round(galleryStats.successRate * 100) }}%</div>
          <div class="stat-label">Success Rate</div>
        </div>
      </div>
      <div class="stats-charts">
        <div class="chart-container">
          <h4>Artifacts by Type</h4>
          <div class="type-chart">
            <div v-for="(count, type) in galleryStats.typeBreakdown" :key="type" class="type-bar">
              <div class="type-label">{{ type }}</div>
              <div class="type-progress">
                <div class="type-fill" :style="{ width: `${(count / galleryStats.totalArtifacts) * 100}%` }"></div>
              </div>
              <div class="type-count">{{ count }}</div>
            </div>
          </div>
        </div>
        <div class="chart-container">
          <h4>Language Distribution</h4>
          <div class="language-chart">
            <div v-for="(count, lang) in galleryStats.languageBreakdown" :key="lang" class="language-item">
              <NBadge :value="lang" />
              <span class="language-count">{{ count }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Artifact Grid/List -->
    <div class="gallery-content">
      <div v-if="filteredArtifacts.length === 0" class="empty-state">
        <Icon icon="ri:folder-open-line" class="empty-icon" />
        <h3>No artifacts found</h3>
        <p>Try adjusting your filters or create some artifacts to get started.</p>
      </div>
      
      <div v-else-if="viewMode === 'grid'" class="artifact-grid">
        <div v-for="artifact in paginatedArtifacts" :key="artifact.id" class="artifact-card">
          <div class="card-header">
            <div class="card-type">
              <Icon :icon="getTypeIcon(artifact.type)" class="type-icon" />
              <span>{{ artifact.type }}</span>
            </div>
            <div class="card-actions">
              <NButton size="tiny" @click="previewArtifactFn(artifact)" circle>
                <template #icon>
                  <Icon icon="ri:eye-line" />
                </template>
              </NButton>
              <NButton v-if="isExecutableArtifact(artifact)" size="tiny" @click="runArtifact(artifact)" circle type="primary">
                <template #icon>
                  <Icon icon="ri:play-line" />
                </template>
              </NButton>
              <NButton v-if="isViewableArtifact(artifact)" size="tiny" @click="viewArtifact(artifact)" circle type="info">
                <template #icon>
                  <Icon icon="ri:external-link-line" />
                </template>
              </NButton>
              <NButton size="tiny" @click="editArtifact(artifact)" circle>
                <template #icon>
                  <Icon icon="ri:edit-line" />
                </template>
              </NButton>
              <NButton size="tiny" @click="duplicateArtifact(artifact)" circle>
                <template #icon>
                  <Icon icon="ri:file-copy-line" />
                </template>
              </NButton>
              <NButton size="tiny" @click="deleteArtifact(artifact)" circle type="error">
                <template #icon>
                  <Icon icon="ri:delete-bin-line" />
                </template>
              </NButton>
            </div>
          </div>
          
          <div class="card-content">
            <h4 class="artifact-title">{{ artifact.title || 'Untitled' }}</h4>
            <div class="artifact-meta">
              <div class="meta-item">
                <Icon icon="ri:calendar-line" />
                <span>{{ formatDate(artifact.createdAt) }}</span>
              </div>
              <div v-if="artifact.language" class="meta-item">
                <Icon icon="ri:code-line" />
                <span>{{ artifact.language }}</span>
              </div>
              <div v-if="artifact.executionCount" class="meta-item">
                <Icon icon="ri:play-line" />
                <span>{{ artifact.executionCount }} runs</span>
              </div>
              <div v-if="artifact.sessionTitle" class="meta-item">
                <Icon icon="ri:chat-1-line" />
                <span>{{ artifact.sessionTitle }}</span>
              </div>
            </div>
            <div class="artifact-preview">
              <pre class="code-preview">{{ truncateCode(artifact.content, 100) }}</pre>
            </div>
          </div>
          
          <div class="card-footer">
            <div class="artifact-tags">
              <NBadge v-for="tag in artifact.tags?.slice(0, 3)" :key="tag" :value="tag" size="small" />
            </div>
            <div class="artifact-rating">
              <Icon icon="ri:star-line" class="star-icon" />
              <span>{{ artifact.rating || 'N/A' }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <div v-else class="artifact-list">
        <div v-for="artifact in paginatedArtifacts" :key="artifact.id" class="artifact-row">
          <div class="row-main">
            <div class="row-type">
              <Icon :icon="getTypeIcon(artifact.type)" class="type-icon" />
              <span>{{ artifact.type }}</span>
            </div>
            <div class="row-content">
              <h4 class="artifact-title">{{ artifact.title || 'Untitled' }}</h4>
              <div class="artifact-description">
                {{ truncateCode(artifact.content, 150) }}
              </div>
            </div>
            <div class="row-meta">
              <div class="meta-item">
                <Icon icon="ri:calendar-line" />
                <span>{{ formatDate(artifact.createdAt) }}</span>
              </div>
              <div v-if="artifact.language" class="meta-item">
                <Icon icon="ri:code-line" />
                <span>{{ artifact.language }}</span>
              </div>
              <div v-if="artifact.executionCount" class="meta-item">
                <Icon icon="ri:play-line" />
                <span>{{ artifact.executionCount }} runs</span>
              </div>
              <div v-if="artifact.sessionTitle" class="meta-item">
                <Icon icon="ri:chat-1-line" />
                <span>{{ artifact.sessionTitle }}</span>
              </div>
            </div>
            <div class="row-actions">
              <NButton size="small" @click="previewArtifactFn(artifact)">
                <template #icon>
                  <Icon icon="ri:eye-line" />
                </template>
                Preview
              </NButton>
              <NButton v-if="isExecutableArtifact(artifact)" size="small" @click="runArtifact(artifact)" type="primary">
                <template #icon>
                  <Icon icon="ri:play-line" />
                </template>
                Run
              </NButton>
              <NButton v-if="isViewableArtifact(artifact)" size="small" @click="viewArtifact(artifact)" type="info">
                <template #icon>
                  <Icon icon="ri:external-link-line" />
                </template>
                View
              </NButton>
              <NButton size="small" @click="editArtifact(artifact)">
                <template #icon>
                  <Icon icon="ri:edit-line" />
                </template>
                Edit
              </NButton>
              <NButton size="small" @click="duplicateArtifact(artifact)">
                <template #icon>
                  <Icon icon="ri:file-copy-line" />
                </template>
                Duplicate
              </NButton>
              <NButton size="small" @click="deleteArtifact(artifact)" type="error">
                <template #icon>
                  <Icon icon="ri:delete-bin-line" />
                </template>
                Delete
              </NButton>
            </div>
          </div>
          <div v-if="artifact.tags?.length" class="row-tags">
            <NBadge v-for="tag in artifact.tags" :key="tag" :value="tag" size="small" />
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="gallery-pagination">
      <NPagination 
        v-model:page="currentPage" 
        :page-count="totalPages" 
        :page-size="pageSize" 
        show-size-picker
        :page-sizes="[12, 24, 48, 96]"
        @update:page-size="onPageSizeChange"
      />
    </div>

    <!-- Preview Modal -->
    <NModal v-model:show="showPreviewModal" :mask-closable="false">
      <NCard style="width: 90vw; max-width: 1200px; max-height: 90vh" :title="previewArtifact?.title || 'Artifact Preview'">
        <div v-if="previewArtifact" class="preview-content">
          <div class="preview-header">
            <div class="preview-meta">
              <NBadge :value="previewArtifact.type" />
              <NBadge v-if="previewArtifact.language" :value="previewArtifact.language" />
              <span class="preview-date">{{ formatDate(previewArtifact.createdAt) }}</span>
            </div>
            <div class="preview-actions">
              <NButton size="small" @click="copyArtifactContent(previewArtifact)">
                <template #icon>
                  <Icon icon="ri:file-copy-line" />
                </template>
                Copy
              </NButton>
              <NButton size="small" @click="downloadArtifact(previewArtifact)">
                <template #icon>
                  <Icon icon="ri:download-line" />
                </template>
                Download
              </NButton>
            </div>
          </div>
          <div class="preview-body">
            <ArtifactViewer :artifacts="[previewArtifact]" />
          </div>
        </div>
        <template #footer>
          <div class="modal-actions">
            <NButton @click="showPreviewModal = false">Close</NButton>
            <NButton type="primary" @click="previewArtifact && editArtifact(previewArtifact)">Edit</NButton>
          </div>
        </template>
      </NCard>
    </NModal>

    <!-- Edit Modal -->
    <NModal v-model:show="showEditModal" :mask-closable="false">
      <NCard style="width: 90vw; max-width: 1200px; max-height: 90vh" :title="editingArtifact?.title || 'Edit Artifact'">
        <div v-if="editingArtifact" class="edit-content">
          <ArtifactEditor 
            v-model="editingArtifact.content" 
            :language="editingArtifact.language || 'javascript'"
            :title="editingArtifact.title"
            :artifact-id="editingArtifact.id"
          />
        </div>
        <template #footer>
          <div class="modal-actions">
            <NButton @click="cancelEdit">Cancel</NButton>
            <NButton type="primary" @click="saveEdit">Save Changes</NButton>
          </div>
        </template>
      </NCard>
    </NModal>

    <!-- Run Modal -->
    <NModal v-model:show="showRunModal" :mask-closable="false">
      <NCard style="width: 90vw; max-width: 1000px; max-height: 90vh" :title="runningArtifact?.title || 'Run Artifact'">
        <div v-if="runningArtifact" class="run-content">
          <div class="run-header">
            <div class="run-meta">
              <NBadge :value="runningArtifact.type" />
              <NBadge v-if="runningArtifact.language" :value="runningArtifact.language" />
            </div>
            <div class="run-actions">
              <NButton 
                size="small" 
                type="primary" 
                @click="executeArtifact" 
                :disabled="isRunning" 
                :loading="isRunning">
                <template #icon>
                  <Icon icon="ri:play-line" />
                </template>
                {{ isRunning ? 'Running...' : 'Run Code' }}
              </NButton>
              <NButton size="small" @click="clearResults" :disabled="!executionResults.length">
                <template #icon>
                  <Icon icon="ri:delete-bin-line" />
                </template>
                Clear
              </NButton>
            </div>
          </div>
          <div class="run-body">
            <div class="code-preview">
              <pre><code>{{ runningArtifact.content }}</code></pre>
            </div>
            <div v-if="executionResults.length" class="execution-results">
              <h4>Output:</h4>
              <div class="results-container">
                <div v-for="result in executionResults" :key="result.id" class="result-item" :class="result.type">
                  <span class="result-type">{{ result.type }}</span>
                  <span class="result-content">{{ result.content }}</span>
                  <span class="result-time">{{ formatTime(result.timestamp) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <template #footer>
          <div class="modal-actions">
            <NButton @click="showRunModal = false">Close</NButton>
          </div>
        </template>
      </NCard>
    </NModal>

    <!-- View Modal -->
    <NModal v-model:show="showViewModal" :mask-closable="false">
      <NCard style="width: 90vw; max-width: 1200px; max-height: 90vh" :title="viewingArtifact?.title || 'View Artifact'">
        <div v-if="viewingArtifact" class="view-content">
          <div class="view-header">
            <div class="view-meta">
              <NBadge :value="viewingArtifact.type" />
              <NBadge v-if="viewingArtifact.language" :value="viewingArtifact.language" />
            </div>
            <div class="view-actions">
              <NButton size="small" @click="copyArtifactContent(viewingArtifact)">
                <template #icon>
                  <Icon icon="ri:file-copy-line" />
                </template>
                Copy
              </NButton>
            </div>
          </div>
          <div class="view-body">
            <ArtifactViewer :artifacts="[convertToViewerFormat(viewingArtifact)]" />
          </div>
        </div>
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
import { ref, computed, onMounted, watch } from 'vue'
import { NButton, NInput, NSelect, NButtonGroup, NBadge, NPagination, NModal, NCard, useMessage, useDialog } from 'naive-ui'
import { Icon } from '@iconify/vue'
import ArtifactViewer from './Message/ArtifactViewer.vue'
import ArtifactEditor from './Message/ArtifactEditor.vue'
import { useMessageStore, useSessionStore } from '@/store'
import { getCodeRunner, type ExecutionResult } from '@/services/codeRunner'

interface Artifact {
  uuid: string
  id: string
  title: string
  content: string
  type: string
  language?: string
  createdAt: string
  updatedAt?: string
  tags?: string[]
  rating?: number
  executionCount?: number
  sessionUuid?: string
  messageUuid?: string
  sessionTitle?: string
}

const message = useMessage()
const dialog = useDialog()
const messageStore = useMessageStore()
const sessionStore = useSessionStore()

// UI State
const showFilters = ref(false)
const showStats = ref(false)
const viewMode = ref<'grid' | 'list'>('grid')
const showPreviewModal = ref(false)
const showEditModal = ref(false)
const showRunModal = ref(false)
const showViewModal = ref(false)

// Filter State
const searchQuery = ref('')
const selectedType = ref('')
const selectedLanguage = ref('')
const selectedDateRange = ref('')
const selectedSession = ref('')
const sortBy = ref('createdAt')

// Data State
const artifacts = ref<Artifact[]>([])
const previewArtifact = ref<Artifact | null>(null)
const editingArtifact = ref<Artifact | null>(null)
const originalArtifact = ref<Artifact | null>(null)
const runningArtifact = ref<Artifact | null>(null)
const viewingArtifact = ref<Artifact | null>(null)
const executionResults = ref<ExecutionResult[]>([])
const isRunning = ref(false)

// Code runner instance
const codeRunner = getCodeRunner()

// Pagination State
const currentPage = ref(1)
const pageSize = ref(24)

// Filter Options
const typeOptions = computed(() => [
  { label: 'All Types', value: '' },
  { label: 'Code', value: 'code' },
  { label: 'Executable Code', value: 'executable-code' },
  { label: 'HTML', value: 'html' },
  { label: 'SVG', value: 'svg' },
  { label: 'JSON', value: 'json' },
  { label: 'Mermaid', value: 'mermaid' }
])

const languageOptions = computed(() => {
  const languages = [...new Set(artifacts.value.map(a => a.language).filter(Boolean))]
  return [
    { label: 'All Languages', value: '' },
    ...languages.map(lang => ({ label: lang, value: lang }))
  ]
})

const sessionOptions = computed(() => {
  const sessions = [...new Set(artifacts.value.map(a => a.sessionTitle).filter(Boolean))]
  return [
    { label: 'All Sessions', value: '' },
    ...sessions.map(session => ({ label: session, value: session }))
  ]
})

const dateRangeOptions = [
  { label: 'All Time', value: '' },
  { label: 'Last 24 Hours', value: '1d' },
  { label: 'Last Week', value: '7d' },
  { label: 'Last Month', value: '30d' },
  { label: 'Last Year', value: '365d' }
]

const sortOptions = [
  { label: 'Created Date (Newest)', value: 'createdAt' },
  { label: 'Created Date (Oldest)', value: 'createdAt-asc' },
  { label: 'Updated Date', value: 'updatedAt' },
  { label: 'Title', value: 'title' },
  { label: 'Type', value: 'type' },
  { label: 'Most Executed', value: 'executionCount' },
  { label: 'Highest Rated', value: 'rating' }
]

// Computed Properties
const filteredArtifacts = computed(() => {
  let filtered = artifacts.value

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(artifact => 
      artifact.title.toLowerCase().includes(query) ||
      artifact.content.toLowerCase().includes(query) ||
      artifact.type.toLowerCase().includes(query) ||
      artifact.language?.toLowerCase().includes(query) ||
      artifact.tags?.some(tag => tag.toLowerCase().includes(query))
    )
  }

  // Type filter
  if (selectedType.value) {
    filtered = filtered.filter(artifact => artifact.type === selectedType.value)
  }

  // Language filter
  if (selectedLanguage.value) {
    filtered = filtered.filter(artifact => artifact.language === selectedLanguage.value)
  }

  // Session filter
  if (selectedSession.value) {
    filtered = filtered.filter(artifact => artifact.sessionTitle === selectedSession.value)
  }

  // Date range filter
  if (selectedDateRange.value) {
    const now = new Date()
    const days = parseInt(selectedDateRange.value.replace('d', ''))
    const cutoff = new Date(now.getTime() - days * 24 * 60 * 60 * 1000)
    filtered = filtered.filter(artifact => new Date(artifact.createdAt) >= cutoff)
  }

  // Sort
  filtered.sort((a, b) => {
    switch (sortBy.value) {
      case 'createdAt':
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      case 'createdAt-asc':
        return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
      case 'updatedAt':
        return new Date(b.updatedAt || b.createdAt).getTime() - new Date(a.updatedAt || a.createdAt).getTime()
      case 'title':
        return a.title.localeCompare(b.title)
      case 'type':
        return a.type.localeCompare(b.type)
      case 'executionCount':
        return (b.executionCount || 0) - (a.executionCount || 0)
      case 'rating':
        return (b.rating || 0) - (a.rating || 0)
      default:
        return 0
    }
  })

  return filtered
})

const totalPages = computed(() => Math.ceil(filteredArtifacts.value.length / pageSize.value))

const paginatedArtifacts = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredArtifacts.value.slice(start, end)
})

const galleryStats = computed(() => {
  const totalArtifacts = artifacts.value.length
  const typeBreakdown: Record<string, number> = {}
  const languageBreakdown: Record<string, number> = {}
  let totalExecutions = 0
  let totalExecutionTime = 0
  let successfulExecutions = 0

  artifacts.value.forEach(artifact => {
    // Type breakdown
    typeBreakdown[artifact.type] = (typeBreakdown[artifact.type] || 0) + 1
    
    // Language breakdown
    if (artifact.language) {
      languageBreakdown[artifact.language] = (languageBreakdown[artifact.language] || 0) + 1
    }
    
    // Execution stats
    if (artifact.executionCount) {
      totalExecutions += artifact.executionCount
      // Add more execution statistics as needed
    }
  })

  return {
    totalArtifacts,
    totalExecutions,
    averageExecutionTime: totalExecutions > 0 ? totalExecutionTime / totalExecutions : 0,
    successRate: totalExecutions > 0 ? successfulExecutions / totalExecutions : 0,
    typeBreakdown,
    languageBreakdown
  }
})

// Methods
const getTypeIcon = (type: string) => {
  const icons: Record<string, string> = {
    'code': 'ri:code-line',
    'executable-code': 'ri:play-circle-line',
    'html': 'ri:html5-line',
    'svg': 'ri:image-line',
    'json': 'ri:file-code-line',
    'mermaid': 'ri:flow-chart',
    'markdown': 'ri:markdown-line'
  }
  return icons[type] || 'ri:file-line'
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const truncateCode = (code: string, maxLength: number) => {
  if (code.length <= maxLength) return code
  return code.substring(0, maxLength) + '...'
}

const previewArtifactFn = (artifact: Artifact) => {
  previewArtifact.value = artifact
  showPreviewModal.value = true
}

const editArtifact = (artifact: Artifact) => {
  editingArtifact.value = { ...artifact }
  originalArtifact.value = artifact
  showEditModal.value = true
  showPreviewModal.value = false
}

const duplicateArtifact = (artifact: Artifact) => {
  const duplicate: Artifact = {
    ...artifact,
    uuid: generateId(),
    id: generateId(),
    title: `${artifact.title} (Copy)`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    sessionUuid: undefined, // Don't associate with original session
    messageUuid: undefined,
    sessionTitle: undefined
  }
  artifacts.value.unshift(duplicate)
  message.success('Artifact duplicated successfully')
}

const deleteArtifact = (artifact: Artifact) => {
  dialog.warning({
    title: 'Delete Artifact',
    content: `Are you sure you want to delete "${artifact.title}"? This action cannot be undone.`,
    positiveText: 'Delete',
    negativeText: 'Cancel',
    onPositiveClick: () => {
      // Remove from message store
      if (artifact.sessionUuid && artifact.messageUuid) {
        const sessionMessages = messageStore.getChatSessionDataByUuid(artifact.sessionUuid)
        if (sessionMessages) {
          const messageIndex = sessionMessages.findIndex(msg => msg.uuid === artifact.messageUuid)
          if (messageIndex !== -1) {
            const message = sessionMessages[messageIndex]
            if (message.artifacts) {
              const artifactIndex = message.artifacts.findIndex(a => a.uuid === artifact.uuid)
              if (artifactIndex !== -1) {
                message.artifacts.splice(artifactIndex, 1)
              }
            }
          }
        }
      }
      
      // Remove from local artifacts
      const index = artifacts.value.findIndex(a => a.id === artifact.id)
      if (index !== -1) {
        artifacts.value.splice(index, 1)
        message.success('Artifact deleted successfully')
      }
    }
  })
}

const copyArtifactContent = async (artifact: Artifact) => {
  try {
    await navigator.clipboard.writeText(artifact.content)
    message.success('Content copied to clipboard')
  } catch (error) {
    message.error('Failed to copy content')
  }
}

const downloadArtifact = (artifact: Artifact) => {
  const extension = getFileExtension(artifact.type, artifact.language)
  const filename = `${artifact.title.replace(/[^a-zA-Z0-9]/g, '_')}.${extension}`
  const blob = new Blob([artifact.content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
  message.success('Artifact downloaded successfully')
}

const getFileExtension = (type: string, language?: string) => {
  if (language) {
    const extensions: Record<string, string> = {
      'javascript': 'js',
      'typescript': 'ts',
      'python': 'py',
      'html': 'html',
      'css': 'css',
      'json': 'json'
    }
    return extensions[language] || 'txt'
  }
  
  const typeExtensions: Record<string, string> = {
    'html': 'html',
    'svg': 'svg',
    'json': 'json',
    'mermaid': 'mmd',
    'markdown': 'md'
  }
  return typeExtensions[type] || 'txt'
}

const saveEdit = () => {
  if (editingArtifact.value && originalArtifact.value) {
    // Update the artifact in the message store
    if (editingArtifact.value.sessionUuid && editingArtifact.value.messageUuid) {
      const sessionMessages = messageStore.getChatSessionDataByUuid(editingArtifact.value.sessionUuid)
      if (sessionMessages) {
        const messageIndex = sessionMessages.findIndex(msg => msg.uuid === editingArtifact.value!.messageUuid)
        if (messageIndex !== -1) {
          const message = sessionMessages[messageIndex]
          if (message.artifacts) {
            const artifactIndex = message.artifacts.findIndex(a => a.uuid === editingArtifact.value!.uuid)
            if (artifactIndex !== -1) {
              message.artifacts[artifactIndex] = {
                ...message.artifacts[artifactIndex],
                title: editingArtifact.value!.title,
                content: editingArtifact.value!.content,
                language: editingArtifact.value!.language
              }
            }
          }
        }
      }
    }
    
    // Update local artifact
    Object.assign(originalArtifact.value, {
      ...editingArtifact.value,
      updatedAt: new Date().toISOString()
    })
    showEditModal.value = false
    message.success('Artifact saved successfully')
  }
}

const cancelEdit = () => {
  editingArtifact.value = null
  originalArtifact.value = null
  showEditModal.value = false
}

// New methods for run/view functionality
const isExecutableArtifact = (artifact: Artifact) => {
  return artifact.type === 'executable-code' || 
         (artifact.type === 'code' && artifact.language && 
          codeRunner.isLanguageSupported(artifact.language))
}

const isViewableArtifact = (artifact: Artifact) => {
  return artifact.type === 'html' || artifact.type === 'svg' || 
         artifact.type === 'mermaid' || artifact.type === 'json'
}

const runArtifact = (artifact: Artifact) => {
  runningArtifact.value = artifact
  executionResults.value = []
  showRunModal.value = true
}

const viewArtifact = (artifact: Artifact) => {
  viewingArtifact.value = artifact
  showViewModal.value = true
}

const executeArtifact = async () => {
  if (!runningArtifact.value || !runningArtifact.value.language) {
    message.error('No artifact or language specified')
    return
  }

  isRunning.value = true
  try {
    const results = await codeRunner.execute(
      runningArtifact.value.language,
      runningArtifact.value.content,
      runningArtifact.value.id
    )
    executionResults.value = results
    
    const hasError = results.some(result => result.type === 'error')
    if (hasError) {
      message.error('Code execution completed with errors')
    } else {
      message.success('Code executed successfully')
    }
  } catch (error) {
    message.error('Code execution failed: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    isRunning.value = false
  }
}

const clearResults = () => {
  executionResults.value = []
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleTimeString()
}

const convertToViewerFormat = (artifact: Artifact): Chat.Artifact => {
  return {
    uuid: artifact.uuid,
    type: artifact.type,
    title: artifact.title,
    content: artifact.content,
    language: artifact.language || undefined,
    isExecutable: isExecutableArtifact(artifact),
    executionResults: []
  }
}

const exportArtifacts = () => {
  const dataStr = JSON.stringify(filteredArtifacts.value, null, 2)
  const blob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `artifacts_${new Date().toISOString().split('T')[0]}.json`
  a.click()
  URL.revokeObjectURL(url)
  message.success('Artifacts exported successfully')
}

const onPageSizeChange = (newSize: number) => {
  pageSize.value = newSize
  currentPage.value = 1
}

const generateId = () => {
  return `artifact_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`
}

// Load artifacts from real conversation data
const loadArtifacts = () => {
  const allArtifacts: Artifact[] = []
  
  // Iterate through all chat sessions
  sessionStore.getAllSessions().forEach(session => {
    const messages = messageStore.getChatSessionDataByUuid(session.uuid) || []
    
    // Extract artifacts from each message
    messages.forEach(msg => {
      if (msg.artifacts && msg.artifacts.length > 0) {
        msg.artifacts.forEach(artifact => {
          allArtifacts.push({
            uuid: artifact.uuid,
            id: artifact.uuid, // Use uuid as id for compatibility
            title: artifact.title || 'Untitled Artifact',
            content: artifact.content,
            type: artifact.type,
            language: artifact.language,
            createdAt: msg.dateTime,
            updatedAt: msg.dateTime,
            sessionUuid: session.uuid,
            messageUuid: msg.uuid,
            sessionTitle: session.title,
            // Add default values for optional fields
            tags: extractTagsFromContent(artifact.content, artifact.type),
            rating: undefined,
            executionCount: artifact.executionResults?.length || 0
          })
        })
      }
    })
  })
  
  artifacts.value = allArtifacts
}

// Helper function to extract tags from content and type
const extractTagsFromContent = (content: string, type: string): string[] => {
  const tags: string[] = [type]
  
  // Add language-specific tags based on content
  if (content.includes('import ')) tags.push('imports')
  if (content.includes('function ')) tags.push('functions')
  if (content.includes('class ')) tags.push('classes')
  if (content.includes('async ')) tags.push('async')
  if (content.includes('await ')) tags.push('await')
  if (content.includes('export ')) tags.push('exports')
  if (content.includes('console.log')) tags.push('logging')
  if (content.includes('useState') || content.includes('useEffect')) tags.push('react')
  if (content.includes('plt.') || content.includes('matplotlib')) tags.push('visualization')
  if (content.includes('np.') || content.includes('numpy')) tags.push('numpy')
  if (content.includes('pd.') || content.includes('pandas')) tags.push('pandas')
  
  return tags
}

onMounted(() => {
  loadArtifacts()
})

// Watch for changes in message store and reload artifacts
watch(
  () => messageStore.sessions,
  () => {
    loadArtifacts()
  },
  { deep: true }
)

// Watch for changes in session history (new sessions)
watch(
  () => sessionStore.sessions,
  () => {
    loadArtifacts()
  },
  { deep: true }
)
</script>

<style scoped>
.artifact-gallery {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--artifact-content-bg);
  border-radius: 8px;
  overflow: hidden;
}

.gallery-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
}

.gallery-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.gallery-title h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
}

.gallery-icon {
  font-size: 24px;
  color: var(--primary-color);
}

.gallery-actions {
  display: flex;
  gap: 8px;
}

.filters-panel {
  padding: 16px 20px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.filter-group label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color);
}

.stats-panel {
  padding: 16px 20px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
  padding: 16px;
  background: var(--artifact-content-bg);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--primary-color);
}

.stat-label {
  font-size: 12px;
  color: var(--text-color-secondary);
  margin-top: 4px;
}

.stats-charts {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.chart-container h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: var(--text-color);
}

.type-chart {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.type-bar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.type-label {
  min-width: 80px;
  font-size: 12px;
  color: var(--text-color);
}

.type-progress {
  flex: 1;
  height: 8px;
  background: var(--border-color);
  border-radius: 4px;
  overflow: hidden;
}

.type-fill {
  height: 100%;
  background: var(--primary-color);
  transition: width 0.3s ease;
}

.type-count {
  min-width: 30px;
  font-size: 12px;
  color: var(--text-color-secondary);
  text-align: right;
}

.language-chart {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.language-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: var(--artifact-content-bg);
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.language-count {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.gallery-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-color-secondary);
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-state h3 {
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.artifact-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.artifact-card {
  background: var(--artifact-content-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.artifact-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
}

.card-type {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-color);
}

.type-icon {
  font-size: 16px;
  color: var(--primary-color);
}

.card-actions {
  display: flex;
  gap: 4px;
}

.card-content {
  padding: 16px;
}

.artifact-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color);
}

.artifact-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 12px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.artifact-preview {
  margin-top: 12px;
}

.code-preview {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  color: var(--text-color);
  background: var(--artifact-header-bg);
  padding: 8px;
  border-radius: 4px;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: pre;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--artifact-header-bg);
  border-top: 1px solid var(--border-color);
}

.artifact-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.artifact-rating {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.star-icon {
  color: #ffd700;
}

.artifact-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.artifact-row {
  background: var(--artifact-content-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;
}

.artifact-row:hover {
  background: var(--hover-color);
}

.row-main {
  display: flex;
  align-items: center;
  gap: 16px;
}

.row-type {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 120px;
  font-size: 12px;
  color: var(--text-color);
}

.row-content {
  flex: 1;
  min-width: 0;
}

.artifact-description {
  font-size: 12px;
  color: var(--text-color-secondary);
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.row-meta {
  display: flex;
  gap: 16px;
  min-width: 200px;
}

.row-actions {
  display: flex;
  gap: 8px;
}

.row-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.gallery-pagination {
  display: flex;
  justify-content: center;
  padding: 16px 20px;
  background: var(--artifact-header-bg);
  border-top: 1px solid var(--border-color);
}

.preview-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.preview-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-date {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.preview-actions {
  display: flex;
  gap: 8px;
}

.preview-body {
  min-height: 400px;
}

.edit-content {
  height: 600px;
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* Responsive design */
@media (max-width: 768px) {
  .gallery-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .gallery-actions {
    justify-content: center;
  }
  
  .filters-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .artifact-grid {
    grid-template-columns: 1fr;
  }
  
  .row-main {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .row-meta {
    justify-content: space-between;
  }
  
  .row-actions {
    justify-content: center;
  }
}
/* Run Modal Styles */
.run-content, .view-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-height: 70vh;
  overflow-y: auto;
}

.run-header, .view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.run-meta, .view-meta {
  display: flex;
  gap: 8px;
  align-items: center;
}

.run-actions, .view-actions {
  display: flex;
  gap: 8px;
}

.run-body, .view-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.code-preview {
  background: var(--code-bg);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
}

.code-preview pre {
  margin: 0;
  padding: 16px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-color);
  overflow-x: auto;
  white-space: pre;
  max-height: 300px;
  overflow-y: auto;
}

.execution-results {
  background: var(--artifact-content-bg);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 16px;
}

.execution-results h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color);
}

.results-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 300px;
  overflow-y: auto;
}

.result-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  border: 1px solid var(--border-color);
}

.result-item.log {
  background: rgba(59, 130, 246, 0.05);
  border-color: rgba(59, 130, 246, 0.2);
}

.result-item.error {
  background: rgba(239, 68, 68, 0.05);
  border-color: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.result-item.warn {
  background: rgba(245, 158, 11, 0.05);
  border-color: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.result-item.info {
  background: rgba(34, 197, 94, 0.05);
  border-color: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.result-item.return {
  background: rgba(168, 85, 247, 0.05);
  border-color: rgba(168, 85, 247, 0.2);
  color: #a855f7;
}

.result-type {
  flex-shrink: 0;
  font-weight: 500;
  text-transform: uppercase;
  font-size: 10px;
  padding: 2px 4px;
  border-radius: 3px;
  background: var(--tag-bg);
  color: var(--text-color-secondary);
  min-width: 50px;
  text-align: center;
}

.result-content {
  flex: 1;
  word-break: break-word;
  white-space: pre-wrap;
  color: var(--text-color);
}

.result-time {
  flex-shrink: 0;
  font-size: 10px;
  color: var(--text-color-secondary);
  opacity: 0.7;
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>