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
          <div class="stat-value">{{ stats.totalArtifacts }}</div>
          <div class="stat-label">Total Artifacts</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats.totalExecutions }}</div>
          <div class="stat-label">Total Executions</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ Math.round(stats.averageExecutionTime) }}ms</div>
          <div class="stat-label">Avg Execution Time</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ Math.round(stats.successRate * 100) }}%</div>
          <div class="stat-label">Success Rate</div>
        </div>
      </div>
      <div class="stats-charts">
        <div class="chart-container">
          <h4>Artifacts by Type</h4>
          <div class="type-chart">
            <div v-for="(count, type) in stats.typeBreakdown" :key="type" class="type-bar">
              <div class="type-label">{{ type }}</div>
              <div class="type-progress">
                <div class="type-fill" :style="{ width: `${(count / stats.totalArtifacts) * 100}%` }"></div>
              </div>
              <div class="type-count">{{ count }}</div>
            </div>
          </div>
        </div>
        <div class="chart-container">
          <h4>Language Distribution</h4>
          <div class="language-chart">
            <div v-for="(count, lang) in stats.languageBreakdown" :key="lang" class="language-item">
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
              <NButton size="tiny" @click="previewArtifact(artifact)" circle>
                <template #icon>
                  <Icon icon="ri:eye-line" />
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
            </div>
            <div class="row-actions">
              <NButton size="small" @click="previewArtifact(artifact)">
                <template #icon>
                  <Icon icon="ri:eye-line" />
                </template>
                Preview
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
            <NButton type="primary" @click="editArtifact(previewArtifact)">Edit</NButton>
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
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { NButton, NInput, NSelect, NButtonGroup, NBadge, NPagination, NModal, NCard, useMessage, useDialog } from 'naive-ui'
import { Icon } from '@iconify/vue'
import ArtifactViewer from './Message/ArtifactViewer.vue'
import ArtifactEditor from './Message/ArtifactEditor.vue'
import { useExecutionHistory } from '@/services/executionHistory'

interface Artifact {
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
}

const message = useMessage()
const dialog = useDialog()
const { stats } = useExecutionHistory()

// UI State
const showFilters = ref(false)
const showStats = ref(false)
const viewMode = ref<'grid' | 'list'>('grid')
const showPreviewModal = ref(false)
const showEditModal = ref(false)

// Filter State
const searchQuery = ref('')
const selectedType = ref('')
const selectedLanguage = ref('')
const selectedDateRange = ref('')
const sortBy = ref('createdAt')

// Data State
const artifacts = ref<Artifact[]>([])
const previewArtifact = ref<Artifact | null>(null)
const editingArtifact = ref<Artifact | null>(null)
const originalArtifact = ref<Artifact | null>(null)

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
  const icons = {
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

const previewArtifact = (artifact: Artifact) => {
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
    id: generateId(),
    title: `${artifact.title} (Copy)`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
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
    const extensions = {
      'javascript': 'js',
      'typescript': 'ts',
      'python': 'py',
      'html': 'html',
      'css': 'css',
      'json': 'json'
    }
    return extensions[language] || 'txt'
  }
  
  const typeExtensions = {
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
  return `artifact_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

// Load artifacts (placeholder - integrate with actual data source)
const loadArtifacts = () => {
  // This would typically load from a real data source
  // For now, we'll use mock data
  artifacts.value = [
    {
      id: '1',
      title: 'Hello World Example',
      content: 'console.log("Hello, World!")',
      type: 'executable-code',
      language: 'javascript',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
      tags: ['beginner', 'example'],
      rating: 5,
      executionCount: 15
    },
    {
      id: '2',
      title: 'Data Visualization',
      content: 'import matplotlib.pyplot as plt\nimport numpy as np\n\nx = np.linspace(0, 10, 100)\ny = np.sin(x)\n\nplt.plot(x, y)\nplt.show()',
      type: 'executable-code',
      language: 'python',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
      tags: ['visualization', 'matplotlib'],
      rating: 4.5,
      executionCount: 8
    },
    // Add more mock artifacts as needed
  ]
}

onMounted(() => {
  loadArtifacts()
})
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
</style>