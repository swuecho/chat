<template>
  <div class="vfs-file-manager">
    <!-- Header with actions -->
    <div class="file-manager-header">
      <div class="breadcrumb">
        <span class="path-segment" @click="navigateTo('/')">/</span>
        <template v-for="(segment, index) in pathSegments" :key="index">
          <span class="separator">/</span>
          <span class="path-segment clickable" @click="navigateTo(getPathUpTo(index))">
            {{ segment }}
          </span>
        </template>
      </div>

      <div class="toolbar">
        <n-button-group>
          <n-button @click="showUploadDialog = true" type="primary" size="small">
            <template #icon><n-icon size="16">
                <CloudUpload />
              </n-icon></template>
            Upload
          </n-button>

          <n-button @click="createNewFolder" size="small">
            <template #icon><n-icon size="16">
                <Add />
              </n-icon></template>
            New Folder
          </n-button>

          <n-button @click="refreshCurrentPath" size="small">
            <template #icon><n-icon size="16">
                <Refresh />
              </n-icon></template>
            Refresh
          </n-button>

          <n-button @click="downloadSelected" :disabled="selectedItems.length === 0" size="small">
            <template #icon><n-icon size="16">
                <CloudDownload />
              </n-icon></template>
            Download
          </n-button>

          <n-button @click="deleteSelected" :disabled="selectedItems.length === 0" type="error" size="small">
            <template #icon><n-icon size="16">
                <Trash />
              </n-icon></template>
            Delete
          </n-button>

          <n-tooltip placement="bottom">
            <template #trigger>
              <n-button size="small" quaternary>
                <template #icon><n-icon size="16">
                    <HelpCircle />
                  </n-icon></template>
              </n-button>
            </template>
            <div>
              <strong>Keyboard Shortcuts:</strong><br />
              Delete: Remove selected items<br />
              F5: Refresh<br />
              Ctrl+A: Select all
            </div>
          </n-tooltip>
        </n-button-group>
      </div>
    </div>

    <!-- File listing -->
    <div class="file-list-container" @dragover.prevent @drop.prevent="handleDrop">
      <n-data-table :columns="fileColumns" :data="fileItems" :row-key="row => row.path"
        v-model:checked-row-keys="selectedItems" :pagination="false" size="small" :max-height="400" virtual-scroll />
    </div>

    <!-- Status bar -->
    <div class="status-bar">
      <span class="item-count">
        {{ fileItems.length }} items
        <span v-if="selectedItems.length > 0" class="selection-count">
          ({{ selectedItems.length }} selected)
        </span>
      </span>
      <span class="storage-info">{{ storageInfo }}</span>
    </div>

    <!-- Upload Dialog -->
    <n-modal v-model:show="showUploadDialog" preset="dialog" title="Upload Files">
      <template #header>
        <div class="upload-header">
          <n-icon size="20">
            <CloudUpload />
          </n-icon>
          <span>Upload Files to {{ currentPath }}</span>
        </div>
      </template>

      <div class="upload-area">
        <n-upload ref="uploadRef" multiple directory-dnd :show-file-list="true" :max="50"
          :on-before-upload="handleFileUpload" :show-cancel-button="false" :show-download-button="false"
          :show-retry-button="false">
          <n-upload-dragger>
            <div class="upload-content">
              <n-icon size="48" depth="3">
                <CloudUpload />
              </n-icon>
              <n-text depth="3">
                Click or drag files here to upload
              </n-text>
              <n-p depth="3" style="margin: 8px 0 0 0">
                Supports multiple files and folders
              </n-p>
            </div>
          </n-upload-dragger>
        </n-upload>

        <div v-if="uploadResults.length > 0" class="upload-results">
          <n-divider>Upload Results</n-divider>
          <div v-for="result in uploadResults" :key="result.filename" class="upload-result">
            <n-icon :color="result.success ? '#18a058' : '#d03050'">
              <CheckmarkCircle v-if="result.success" />
              <Close v-else />
            </n-icon>
            <span>{{ result.filename }}</span>
            <span class="result-message">{{ result.message }}</span>
          </div>
        </div>
      </div>

      <template #action>
        <n-space>
          <n-button @click="showUploadDialog = false">Close</n-button>
          <n-button @click="clearUploadResults" v-if="uploadResults.length > 0">Clear Results</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Import/Export Dialog -->
    <n-modal v-model:show="showImportExportDialog" preset="dialog" title="Import/Export">
      <n-tabs default-value="import" size="medium">
        <n-tab-pane name="import" tab="Import">
          <div class="import-section">
            <n-space vertical>
              <div>
                <n-text>Import from URL:</n-text>
                <n-input-group>
                  <n-input v-model:value="importUrl" placeholder="https://example.com/data.csv" />
                  <n-button @click="importFromURL" :loading="importing">Import</n-button>
                </n-input-group>
              </div>

              <n-divider>OR</n-divider>

              <div>
                <n-text>Import VFS Session:</n-text>
                <n-upload :show-file-list="false" accept=".vfs.json,.json" :on-before-upload="importVFSSession">
                  <n-button>Select VFS Session File</n-button>
                </n-upload>
              </div>
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="export" tab="Export">
          <div class="export-section">
            <n-space vertical>
              <div>
                <n-text>Export Current Session:</n-text>
                <n-input-group>
                  <n-input v-model:value="exportSessionName" placeholder="my-session" />
                  <n-button @click="exportVFSSession" :loading="exporting">Export Session</n-button>
                </n-input-group>
              </div>

              <n-divider>Statistics</n-divider>

              <div class="export-stats">
                <div>Files: {{ vfsStats.totalFiles }}</div>
                <div>Directories: {{ vfsStats.totalDirectories }}</div>
                <div>Total Size: {{ formatSize(vfsStats.totalSize) }}</div>
              </div>
            </n-space>
          </div>
        </n-tab-pane>
      </n-tabs>

      <template #action>
        <n-button @click="showImportExportDialog = false">Close</n-button>
      </template>
    </n-modal>

    <!-- Create Folder Dialog -->
    <n-modal v-model:show="showCreateFolderDialog" preset="dialog" title="Create New Folder">
      <n-space vertical>
        <n-text>Enter the name for the new folder:</n-text>
        <n-input v-model:value="newFolderName" placeholder="Folder name" @keydown.enter="confirmCreateFolder" />
      </n-space>
      <template #action>
        <n-space>
          <n-button @click="showCreateFolderDialog = false">Cancel</n-button>
          <n-button @click="confirmCreateFolder" type="primary">Create</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Delete Confirmation Dialog -->
    <n-modal v-model:show="showDeleteConfirmDialog" preset="dialog" title="Confirm Delete">
      <n-text>Are you sure you want to delete {{ itemsToDelete.length }} item(s)?</n-text>
      <template #action>
        <n-space>
          <n-button @click="showDeleteConfirmDialog = false">Cancel</n-button>
          <n-button @click="confirmDelete" type="error">Delete</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, h } from 'vue'
import {
  useMessage,
  NModal,
  NButton,
  NButtonGroup,
  NIcon,
  NDataTable,
  NDivider,
  NUpload,
  NUploadDragger,
  NText,
  NP,
  NSpace,
  NTabs,
  NTabPane,
  NInput,
  NInputGroup,
  NTooltip
} from 'naive-ui'
import {
  CloudUpload,
  CloudDownload,
  Add,
  Refresh,
  Trash,
  CheckmarkCircle,
  Close,
  Folder,
  Document,
  HelpCircle
} from '@vicons/ionicons5'

// Props
const props = defineProps({
  vfsInstance: {
    type: Object,
    required: true
  },
  importExport: {
    type: Object,
    required: true
  }
})

// Reactive data
const message = useMessage()
const currentPath = ref('/workspace')
const fileItems = ref([])
const selectedItems = ref([])
const showUploadDialog = ref(false)
const showImportExportDialog = ref(false)
const showCreateFolderDialog = ref(false)
const showDeleteConfirmDialog = ref(false)
const uploadResults = ref([])
const importing = ref(false)
const exporting = ref(false)
const importUrl = ref('')
const exportSessionName = ref('my-session')
const newFolderName = ref('')
const itemsToDelete = ref([])
const vfsStats = ref({
  totalFiles: 0,
  totalDirectories: 0,
  totalSize: 0
})

// Computed properties
const pathSegments = computed(() => {
  return currentPath.value.split('/').filter(Boolean)
})

const storageInfo = computed(() => {
  const info = props.vfsInstance.getStorageInfo()
  return `${info.usage.size} storage used, ${info.usage.files} files`
})

const fileColumns = [
  {
    type: 'selection'
  },
  {
    title: 'Name',
    key: 'name',
    render: (row) => {
      return h('div', { class: 'file-name' }, [
        h(NIcon, {
          size: 16,
          style: { marginRight: '8px' }
        }, {
          default: () => row.isDirectory ? h(Folder) : h(Document)
        }),
        h('span', {
          class: row.isDirectory ? 'directory-name clickable' : 'file-name',
          onClick: row.isDirectory ? () => navigateTo(row.path) : undefined
        }, row.name)
      ])
    }
  },
  {
    title: 'Size',
    key: 'size',
    sorter: (a, b) => (a.size || 0) - (b.size || 0),
    render: (row) => row.isDirectory ? '—' : formatSize(row.size || 0)
  },
  {
    title: 'Modified',
    key: 'modified',
    sorter: (a, b) => (new Date(a.mtime || 0)) - (new Date(b.mtime || 0)),
    render: (row) => row.mtime ? new Date(row.mtime).toLocaleString() : '—'
  },
  {
    title: 'Actions',
    key: 'actions',
    render: (row) => {
      return h('div', { class: 'file-actions' }, [
        h(NButton, {
          size: 'tiny',
          onClick: () => downloadItem(row.path)
        }, { default: () => 'Download' }),
        h(NButton, {
          size: 'tiny',
          type: 'error',
          style: { marginLeft: '8px' },
          onClick: () => deleteItem(row.path)
        }, { default: () => 'Delete' })
      ])
    }
  }
]

// Methods
const refreshCurrentPath = async () => {
  try {
    const items = await props.vfsInstance.readdir(currentPath.value)
    fileItems.value = []

    for (const item of items) {
      const itemPath = props.vfsInstance.pathResolver.join(currentPath.value, item)
      const stat = await props.vfsInstance.stat(itemPath)

      fileItems.value.push({
        name: item,
        path: itemPath,
        isDirectory: stat.isDirectory,
        isFile: stat.isFile,
        size: stat.size,
        mtime: stat.mtime
      })
    }

    // Sort: directories first, then files
    fileItems.value.sort((a, b) => {
      if (a.isDirectory && !b.isDirectory) return -1
      if (!a.isDirectory && b.isDirectory) return 1
      return a.name.localeCompare(b.name)
    })

    // Update VFS stats
    updateVFSStats()
  } catch (error) {
    message.error(`Failed to read directory: ${error.message}`)
  }
}

const navigateTo = (path) => {
  currentPath.value = path
  selectedItems.value = []
  refreshCurrentPath()
}

const getPathUpTo = (index) => {
  const segments = pathSegments.value.slice(0, index + 1)
  return '/' + segments.join('/')
}

const createNewFolder = () => {
  newFolderName.value = ''
  showCreateFolderDialog.value = true
}

const confirmCreateFolder = async () => {
  if (!newFolderName.value.trim()) {
    message.error('Please enter a folder name')
    return
  }

  try {
    const folderPath = props.vfsInstance.pathResolver.join(currentPath.value, newFolderName.value.trim())
    await props.vfsInstance.mkdir(folderPath)
    message.success(`Created folder: ${newFolderName.value}`)
    showCreateFolderDialog.value = false
    refreshCurrentPath()
  } catch (error) {
    message.error(`Failed to create folder: ${error.message}`)
  }
}

const handleDrop = async (event) => {
  const files = Array.from(event.dataTransfer.files)
  if (files.length === 0) return

  for (const file of files) {
    // Create the same structure that naive-ui upload provides
    await handleFileUpload({ file: file })
  }
}

const handleFileUpload = async (file) => {
  try {
    // Extract the actual File object from Naive UI's structure
    // Based on ChatVFSUploader debug: file.file.file contains the actual File object
    const actualFile = file.file.file

    if (!actualFile || !(actualFile instanceof File)) {
      throw new Error(`Invalid file structure: expected file.file.file to be a File object, got ${typeof actualFile}`)
    }

    const result = await props.importExport.uploadFile(actualFile,
      `${currentPath.value}/${actualFile.name}`)

    uploadResults.value.push({
      filename: actualFile.name,
      ...result
    })

    if (result.success) {
      message.success(`Uploaded ${actualFile.name}`)
      refreshCurrentPath()
    } else {
      message.error(result.message)
    }
  } catch (error) {
    const filename = file?.file?.file?.name || 'unknown'
    uploadResults.value.push({
      filename: filename,
      success: false,
      message: error.message
    })
    message.error(`Failed to upload ${filename}: ${error.message}`)
  }

  return false // Prevent default upload behavior
}

const downloadSelected = async () => {
  if (selectedItems.value.length === 0) return

  for (const itemPath of selectedItems.value) {
    try {
      const stat = await props.vfsInstance.stat(itemPath)
      if (stat.isDirectory) {
        await props.importExport.downloadDirectory(itemPath)
      } else {
        await props.importExport.downloadFile(itemPath)
      }
    } catch (error) {
      message.error(`Failed to download ${itemPath}: ${error.message}`)
    }
  }
}

const downloadItem = async (itemPath) => {
  try {
    const stat = await props.vfsInstance.stat(itemPath)
    if (stat.isDirectory) {
      const result = await props.importExport.downloadDirectory(itemPath)
      if (result.success) {
        message.success(result.message)
      } else {
        message.error(result.message)
      }
    } else {
      const result = await props.importExport.downloadFile(itemPath)
      if (result.success) {
        message.success(result.message)
      } else {
        message.error(result.message)
      }
    }
  } catch (error) {
    message.error(`Download failed: ${error.message}`)
  }
}

const deleteSelected = () => {
  if (selectedItems.value.length === 0) return

  itemsToDelete.value = [...selectedItems.value]
  showDeleteConfirmDialog.value = true
}

const confirmDelete = async () => {
  for (const itemPath of itemsToDelete.value) {
    await deleteItem(itemPath, false)
  }

  selectedItems.value = []
  showDeleteConfirmDialog.value = false
  refreshCurrentPath()
}

const deleteItem = async (itemPath, refresh = true) => {
  try {
    const stat = await props.vfsInstance.stat(itemPath)
    if (stat.isDirectory) {
      await props.vfsInstance.rmdir(itemPath, { recursive: true })
    } else {
      await props.vfsInstance.unlink(itemPath)
    }

    message.success(`Deleted ${props.vfsInstance.pathResolver.basename(itemPath)}`)
    if (refresh) refreshCurrentPath()
  } catch (error) {
    message.error(`Failed to delete: ${error.message}`)
  }
}

const clearUploadResults = () => {
  uploadResults.value = []
}

const importFromURL = async () => {
  if (!importUrl.value) {
    message.warning('Please enter a URL')
    return
  }

  importing.value = true
  try {
    const filename = props.importExport.extractFilenameFromURL(importUrl.value)
    const targetPath = `${currentPath.value}/${filename}`

    const result = await props.importExport.importFromURL(importUrl.value, targetPath)

    if (result.success) {
      message.success(result.message)
      refreshCurrentPath()
      importUrl.value = ''
    } else {
      message.error(result.message)
    }
  } catch (error) {
    message.error(`Import failed: ${error.message}`)
  } finally {
    importing.value = false
  }
}

const importVFSSession = async (file) => {
  try {
    // Extract the actual File object from Naive UI's structure
    const actualFile = file.file.file

    if (!actualFile || !(actualFile instanceof File)) {
      throw new Error(`Invalid file structure: expected file.file.file to be a File object, got ${typeof actualFile}`)
    }

    const result = await props.importExport.importVFSSession(actualFile)

    if (result.success) {
      message.success(result.message)
      refreshCurrentPath()
      showImportExportDialog.value = false
    } else {
      message.error(result.message)
    }
  } catch (error) {
    message.error(`Import failed: ${error.message}`)
  }

  return false // Prevent default upload behavior
}

const exportVFSSession = async () => {
  if (!exportSessionName.value) {
    message.warning('Please enter a session name')
    return
  }

  exporting.value = true
  try {
    const result = await props.importExport.exportVFSSession(exportSessionName.value)

    if (result.success) {
      message.success(result.message)
    } else {
      message.error(result.message)
    }
  } catch (error) {
    message.error(`Export failed: ${error.message}`)
  } finally {
    exporting.value = false
  }
}

const updateVFSStats = () => {
  vfsStats.value = props.importExport.getImportStats()
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

// Keyboard shortcuts  
const handleKeyDown = (event) => {
  if (event.key === 'Delete' && selectedItems.value.length > 0) {
    deleteSelected()
  } else if (event.key === 'F5') {
    event.preventDefault()
    refreshCurrentPath()
  } else if (event.ctrlKey && event.key === 'a') {
    event.preventDefault()
    selectedItems.value = fileItems.value.map(item => item.path)
  }
}

// Lifecycle
onMounted(() => {
  refreshCurrentPath()
  document.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
})

// Watch for VFS changes
watch(() => props.vfsInstance.files.size, () => {
  updateVFSStats()
})

// Expose methods for parent component
defineExpose({
  refreshCurrentPath,
  navigateTo,
  openImportExportDialog: () => { showImportExportDialog.value = true }
})
</script>

<style scoped>
.vfs-file-manager {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
}

.file-manager-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--code-block-bg);
  border-bottom: 1px solid var(--border-color);
}

.breadcrumb {
  display: flex;
  align-items: center;
  font-family: 'Fira Code', monospace;
  font-size: 14px;
}

.path-segment {
  color: var(--text-color);
  padding: 2px 4px;
  border-radius: 3px;
}

.path-segment.clickable {
  cursor: pointer;
  color: var(--primary-color);
}

.path-segment.clickable:hover {
  background: var(--hover-color);
}

.separator {
  color: var(--text-color-3);
  margin: 0 2px;
}

.file-list-container {
  max-height: 400px;
  overflow: auto;
}

.file-name {
  display: flex;
  align-items: center;
}

.directory-name.clickable {
  cursor: pointer;
  color: var(--primary-color);
}

.directory-name.clickable:hover {
  text-decoration: underline;
}

.file-actions {
  display: flex;
  gap: 4px;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: var(--code-block-bg);
  border-top: 1px solid var(--border-color);
  font-size: 12px;
  color: var(--text-color-3);
}

.selection-count {
  color: var(--primary-color);
  font-weight: 500;
}

.upload-area {
  min-height: 200px;
}

.upload-content {
  text-align: center;
  padding: 40px 20px;
}

.upload-results {
  margin-top: 16px;
}

.upload-result {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  font-size: 14px;
}

.result-message {
  color: var(--text-color-3);
  font-size: 12px;
}

.import-section,
.export-section {
  padding: 16px 0;
}

.export-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
  padding: 8px;
  background: var(--code-block-bg);
  border-radius: 4px;
  font-family: 'Fira Code', monospace;
  font-size: 12px;
}

.upload-header {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>