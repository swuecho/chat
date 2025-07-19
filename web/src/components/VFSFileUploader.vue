<template>
  <div class="vfs-file-uploader">
    <!-- VFS Upload Button -->
    <n-button @click="showVFSUploadModal = true" type="primary" size="small">
      <template #icon>
        <n-icon><CloudUpload /></n-icon>
      </template>
      Upload to VFS
    </n-button>

    <!-- VFS Upload Modal -->
    <n-modal v-model:show="showVFSUploadModal" preset="dialog" title="Upload Files to Virtual File System">
      <template #header>
        <div class="upload-header">
          <n-icon size="20"><CloudUpload /></n-icon>
          <span>Upload Files to VFS</span>
        </div>
      </template>
      
      <div class="vfs-upload-content">
        <!-- Target Directory Selection -->
        <div class="target-directory">
          <n-text>Target Directory:</n-text>
          <n-select 
            v-model:value="targetDirectory" 
            :options="directoryOptions"
            placeholder="Select or type directory path"
            filterable
            tag
          />
        </div>

        <!-- Upload Area -->
        <div class="upload-area">
          <n-upload
            ref="uploadRef"
            multiple
            directory-dnd
            :show-file-list="true"
            :max="20"
            :on-before-upload="handleVFSFileUpload"
            :show-cancel-button="false"
            :show-download-button="false"
            :show-retry-button="false"
            accept=".txt,.csv,.json,.py,.js,.md,.xml,.yaml,.yml,.log,.html,.css,.sql"
          >
            <n-upload-dragger>
              <div class="upload-content">
                <n-icon size="48" depth="3"><CloudUpload /></n-icon>
                <n-text depth="3">
                  Click or drag files here to upload to VFS
                </n-text>
                <n-p depth="3" style="margin: 8px 0 0 0">
                  Supports text files, data files, and code files
                </n-p>
                <n-p depth="3" style="margin: 4px 0 0 0; font-size: 12px;">
                  Files will be available in both Python and JavaScript code runners
                </n-p>
              </div>
            </n-upload-dragger>
          </n-upload>
        </div>

        <!-- Upload Results -->
        <div v-if="uploadResults.length > 0" class="upload-results">
          <n-divider>Upload Results</n-divider>
          <div class="results-list">
            <div v-for="result in uploadResults" :key="result.filename" class="upload-result">
              <n-icon :color="result.success ? '#18a058' : '#d03050'">
                <CheckCircle v-if="result.success" />
                <CloseCircle v-else />
              </n-icon>
              <div class="result-details">
                <div class="filename">{{ result.filename }}</div>
                <div class="result-message">{{ result.message }}</div>
                <div v-if="result.success" class="file-path">→ {{ result.path }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- VFS Status -->
        <div class="vfs-status">
          <n-divider>VFS Status</n-divider>
          <div class="status-grid">
            <div class="status-item">
              <n-text depth="3">Files:</n-text>
              <n-text>{{ vfsStats.totalFiles }}</n-text>
            </div>
            <div class="status-item">
              <n-text depth="3">Storage:</n-text>
              <n-text>{{ formatSize(vfsStats.totalSize) }}</n-text>
            </div>
            <div class="status-item">
              <n-text depth="3">Directories:</n-text>
              <n-text>{{ vfsStats.totalDirectories }}</n-text>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="quick-actions">
          <n-button-group>
            <n-button @click="openFileManager" size="small">
              <template #icon><n-icon><Folder /></n-icon></template>
              File Manager
            </n-button>
            <n-button @click="clearVFS" size="small" type="warning">
              <template #icon><n-icon><Trash /></n-icon></template>
              Clear VFS
            </n-button>
            <n-button @click="exportVFS" size="small">
              <template #icon><n-icon><Download /></n-icon></template>
              Export Session
            </n-button>
          </n-button-group>
        </div>
      </div>

      <template #action>
        <n-space>
          <n-button @click="clearUploadResults" v-if="uploadResults.length > 0">Clear Results</n-button>
          <n-button @click="showVFSUploadModal = false">Close</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- File Manager Modal -->
    <n-modal v-model:show="showFileManager" style="width: 90vw; max-width: 1200px;">
      <n-card title="VFS File Manager" :bordered="false" size="huge" role="dialog">
        <VFSFileManager 
          ref="fileManagerRef"
          :vfs-instance="vfsInstance"
          :import-export="importExportInstance"
        />
        <template #footer>
          <n-button @click="showFileManager = false">Close</n-button>
        </template>
      </n-card>
    </n-modal>

    <!-- Import VFS Session -->
    <input
      ref="fileInputRef"
      type="file"
      accept=".vfs.json,.json"
      style="display: none"
      @change="handleVFSSessionImport"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useMessage } from 'naive-ui'
import { 
  CloudUpload, 
  CheckCircle,
  CloseCircle,
  Folder,
  Trash,
  Download
} from '@vicons/ionicons5'
import VFSFileManager from './VFSFileManager.vue'

// Inject VFS instances (these should be provided by a parent component)
const vfsInstance = inject('vfsInstance')
const importExportInstance = inject('importExportInstance')

const message = useMessage()

// Reactive data
const showVFSUploadModal = ref(false)
const showFileManager = ref(false)
const targetDirectory = ref('/data')
const uploadResults = ref([])
const uploadRef = ref()
const fileInputRef = ref()
const fileManagerRef = ref()

const vfsStats = ref({
  totalFiles: 0,
  totalDirectories: 0,
  totalSize: 0
})

// Directory options for selection
const directoryOptions = computed(() => [
  { label: '/data - Data files', value: '/data' },
  { label: '/workspace - Working directory', value: '/workspace' },
  { label: '/tmp - Temporary files', value: '/tmp' },
  { label: '/uploads - Uploaded files', value: '/uploads' }
])

// Methods
const handleVFSFileUpload = async (file) => {
  if (!vfsInstance || !importExportInstance) {
    message.error('VFS not available. Please ensure code runners are initialized.')
    return false
  }

  try {
    // Create target directory if it doesn't exist
    await vfsInstance.mkdir(targetDirectory.value, { recursive: true })
    
    // Generate target path
    const targetPath = `${targetDirectory.value}/${file.file.name}`
    
    // Upload file to VFS
    const result = await importExportInstance.uploadFile(file.file, targetPath)
    
    uploadResults.value.push({
      filename: file.file.name,
      ...result
    })
    
    if (result.success) {
      message.success(`Uploaded ${file.file.name} to VFS`)
      updateVFSStats()
      
      // Emit event for parent component
      emit('fileUploaded', {
        filename: file.file.name,
        path: result.path,
        size: file.file.size
      })
    } else {
      message.error(result.message)
    }
  } catch (error) {
    const errorResult = {
      filename: file.file.name,
      success: false,
      message: error.message
    }
    
    uploadResults.value.push(errorResult)
    message.error(`Upload failed: ${error.message}`)
  }
  
  return false // Prevent default upload behavior
}

const clearUploadResults = () => {
  uploadResults.value = []
}

const updateVFSStats = () => {
  if (importExportInstance) {
    vfsStats.value = importExportInstance.getImportStats()
  }
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const openFileManager = () => {
  showFileManager.value = true
  showVFSUploadModal.value = false
}

const clearVFS = async () => {
  if (!vfsInstance) return
  
  const confirmed = confirm('Are you sure you want to clear all VFS files? This action cannot be undone.')
  if (confirmed) {
    vfsInstance.clear()
    uploadResults.value = []
    updateVFSStats()
    message.success('VFS cleared successfully')
    
    emit('vfsCleared')
  }
}

const exportVFS = async () => {
  if (!importExportInstance) return
  
  try {
    const sessionName = `vfs-session-${new Date().toISOString().slice(0, 10)}`
    const result = await importExportInstance.exportVFSSession(sessionName)
    
    if (result.success) {
      message.success(result.message)
    } else {
      message.error(result.message)
    }
  } catch (error) {
    message.error(`Export failed: ${error.message}`)
  }
}

const handleVFSSessionImport = async (event) => {
  const file = event.target.files[0]
  if (!file || !importExportInstance) return
  
  try {
    const result = await importExportInstance.importVFSSession(file)
    
    if (result.success) {
      message.success(result.message)
      updateVFSStats()
      emit('vfsSessionImported', result)
    } else {
      message.error(result.message)
    }
  } catch (error) {
    message.error(`Import failed: ${error.message}`)
  }
  
  // Reset file input
  event.target.value = ''
}

const importVFSSession = () => {
  fileInputRef.value?.click()
}

// Lifecycle
onMounted(() => {
  updateVFSStats()
})

// Emits
const emit = defineEmits(['fileUploaded', 'vfsCleared', 'vfsSessionImported'])

// Expose methods for parent component
defineExpose({
  openUploadModal: () => { showVFSUploadModal.value = true },
  openFileManager: () => { showFileManager.value = true },
  importVFSSession,
  updateStats: updateVFSStats
})
</script>

<style scoped>
.vfs-file-uploader {
  display: inline-block;
}

.upload-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.vfs-upload-content {
  max-height: 70vh;
  overflow-y: auto;
}

.target-directory {
  margin-bottom: 16px;
}

.target-directory .n-text {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
}

.upload-area {
  margin-bottom: 16px;
}

.upload-content {
  text-align: center;
  padding: 40px 20px;
}

.upload-results {
  margin-bottom: 16px;
}

.results-list {
  max-height: 200px;
  overflow-y: auto;
}

.upload-result {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
}

.upload-result:last-child {
  border-bottom: none;
}

.result-details {
  flex: 1;
}

.filename {
  font-weight: 500;
  color: var(--text-color);
}

.result-message {
  font-size: 12px;
  color: var(--text-color-3);
  margin-top: 2px;
}

.file-path {
  font-size: 11px;
  color: var(--primary-color);
  margin-top: 2px;
  font-family: monospace;
}

.vfs-status {
  margin-bottom: 16px;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 12px;
  background: var(--code-block-bg);
  border-radius: 6px;
}

.status-item {
  text-align: center;
}

.status-item .n-text:first-child {
  display: block;
  margin-bottom: 4px;
}

.quick-actions {
  display: flex;
  justify-content: center;
}

@media (max-width: 768px) {
  .status-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }
  
  .upload-content {
    padding: 20px 10px;
  }
}
</style>