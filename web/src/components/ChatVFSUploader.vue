<template>
  <div class="chat-vfs-uploader">
    <!-- Upload Button for Chat -->
    <n-tooltip placement="top">
      <template #trigger>
        <n-button @click="handleButtonClick" size="small" circle quaternary type="primary" :loading="uploading">
          <template #icon>
            <n-icon>
              <FolderOpen />
            </n-icon>
          </template>
        </n-button>
      </template>
      Upload files to VFS for code runners
    </n-tooltip>

    <!-- Upload Modal -->
    <n-modal v-model:show="showUploadModal" preset="dialog" title="Upload Files for Code Runners"
      style="width: 90vw; max-width: 800px;">
      <template #header>
        <div class="upload-header">
          <n-icon size="20">
            <FolderOpen />
          </n-icon>
          <span>Upload Files to Virtual File System</span>
        </div>
      </template>

      <div class="upload-content">
        <n-alert type="info" style="margin-bottom: 16px;" :show-icon="false">
          <p><strong>Files uploaded here will be immediately available in Python and JavaScript code runners.</strong>
          </p>
          <p>Access them using standard file operations like <code>pd.read_csv('/data/filename.csv')</code> or
            <code>fs.readFileSync('/data/filename.csv')</code>
          </p>
        </n-alert>

        <!-- Quick Directory Selection -->
        <div class="directory-selection">
          <n-text style="margin-bottom: 8px; display: block; font-weight: 500;">Upload to:</n-text>
          <n-radio-group v-model:value="selectedDirectory">
            <n-space>
              <n-radio value="/data">
                📊 /data (CSV, JSON, datasets)
              </n-radio>
              <n-radio value="/workspace">
                💼 /workspace (code, docs)
              </n-radio>
              <n-radio value="/uploads">
                📤 /uploads (general files)
              </n-radio>
            </n-space>
          </n-radio-group>
        </div>

        <!-- Upload Area -->
        <n-upload ref="uploadRef" multiple directory-dnd :show-file-list="true" :max="10"
          :on-before-upload="handleFileUpload" :show-cancel-button="false" :show-download-button="false"
          :show-retry-button="false"
          accept=".txt,.csv,.json,.xlsx,.xls,.py,.js,.ts,.jsx,.tsx,.md,.xml,.yaml,.yml,.log,.html,.css,.sql,.png,.jpg,.jpeg,.gif,.svg,.pdf,.zip,.tar,.gz">
          <n-upload-dragger>
            <div class="upload-dragger-content">
              <n-icon size="48" depth="3">
                <CloudUpload />
              </n-icon>
              <n-text style="font-size: 16px" depth="3">
                Click or drag files here to upload
              </n-text>
              <n-p depth="3" style="margin: 8px 0 0 0">
                Supports: Data files (CSV, JSON, Excel), Code files (Python, JavaScript, TypeScript), Documents
                (Markdown,
                Text), Images (PNG, JPG, SVG), Archives (ZIP, TAR)
              </n-p>
            </div>
          </n-upload-dragger>
        </n-upload>

        <!-- Upload Results -->
        <div v-if="uploadResults.length > 0" class="upload-results">
          <n-divider>Upload Results</n-divider>
          <div class="results-list">
            <div v-for="result in uploadResults" :key="result.filename" class="upload-result">
              <n-icon :color="result.success ? '#18a058' : '#d03050'">
                <CheckmarkCircle v-if="result.success" />
                <Close v-else />
              </n-icon>
              <div class="result-details">
                <div class="filename">{{ result.filename }}</div>
                <div class="file-path">{{ result.path }}</div>
                <div class="result-message">{{ result.message }}</div>
              </div>
            </div>
          </div>

          <!-- Code Examples -->
          <div v-if="uploadResults.some(r => r.success)" class="code-examples">
            <n-divider>How to use uploaded files:</n-divider>

            <n-tabs default-value="python" size="small">
              <n-tab-pane name="python" tab="Python">
                <n-code :code="generatePythonExample()" language="python" />
                <n-button @click="copyCode(generatePythonExample())" size="tiny" style="margin-top: 8px;">
                  Copy Python Code
                </n-button>
              </n-tab-pane>

              <n-tab-pane name="javascript" tab="JavaScript">
                <n-code :code="generateJavaScriptExample()" language="javascript" />
                <n-button @click="copyCode(generateJavaScriptExample())" size="tiny" style="margin-top: 8px;">
                  Copy JavaScript Code
                </n-button>
              </n-tab-pane>
            </n-tabs>
          </div>
        </div>

        <!-- VFS Status -->
        <div class="vfs-status">
          <n-divider>Current VFS Status</n-divider>
          <div class="status-grid">
            <div class="status-item">
              <strong>{{ vfsStats.totalFiles }}</strong>
              <span>Files</span>
            </div>
            <div class="status-item">
              <strong>{{ formatSize(vfsStats.totalSize) }}</strong>
              <span>Used</span>
            </div>
            <div class="status-item">
              <strong>{{ Math.round((vfsStats.totalSize / (100 * 1024 * 1024)) * 100) }}%</strong>
              <span>Quota</span>
            </div>
          </div>
        </div>
      </div>

      <template #action>
        <n-space>
          <n-button @click="openFileManager" size="small">
            <template #icon><n-icon>
                <Folder />
              </n-icon></template>
            File Manager
          </n-button>
          <n-button @click="clearResults" v-if="uploadResults.length > 0">Clear Results</n-button>
          <n-button @click="addCodeExample" type="primary" v-if="uploadResults.some(r => r.success)">
            Add Code Example to Chat
          </n-button>
          <n-button @click="showUploadModal = false">Close</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- File Manager Modal -->
    <n-modal v-model:show="showFileManager" style="width: 90vw; max-width: 1200px;">
      <n-card title="VFS File Manager" :bordered="false" size="huge" role="dialog">
        <VFSFileManager v-if="vfsInstance && importExportInstance" :vfs-instance="vfsInstance"
          :import-export="importExportInstance" />
        <div v-else class="loading">
          <n-spin size="small" />
          <span>Loading file manager...</span>
        </div>
        <template #footer>
          <n-button @click="showFileManager = false">Close</n-button>
        </template>
      </n-card>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, inject, onMounted, watch } from 'vue'
import { useMessage, NSpin, NTabPane, NCard, NModal, NButton, NIcon, NText, NAlert, NDivider, NUpload, NUploadDragger, NCode, NRadio, NRadioGroup, NSpace, NTooltip, NTabs, NP } from 'naive-ui'
import {
  FolderOpen,
  CloudUpload,
  CheckmarkCircle,
  Close,
  Folder
} from '@vicons/ionicons5'
import VFSFileManager from './VFSFileManager.vue'
import { getCodeRunner } from '@/services/codeRunner'

// Props
const props = defineProps({
  sessionUuid: {
    type: String,
    required: true
  }
})

// Emits
const emit = defineEmits(['fileUploaded', 'codeExampleAdded'])

const message = useMessage()

// Inject VFS instances (should be provided by VFSProvider)
const vfsInstance = inject('vfsInstance', ref(null))
const importExportInstance = inject('importExportInstance', ref(null))
const isVFSReady = inject('isVFSReady', ref(false))

// Reactive state
const showUploadModal = ref(false)
const showFileManager = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const selectedDirectory = ref('/data')
const uploadResults = ref([])
const uploadRef = ref()

const vfsStats = ref({
  totalFiles: 0,
  totalDirectories: 0,
  totalSize: 0
})

// Watch for VFS readiness
watch([vfsInstance, importExportInstance], () => {
  if (vfsInstance.value && importExportInstance.value) {
    initializeVFS()
  }
}, { immediate: true })

// Methods
const handleButtonClick = () => {
  showUploadModal.value = true
}

const initializeVFS = async () => {
  if (!vfsInstance.value) return

  try {
    // Ensure directories exist
    await vfsInstance.value.mkdir('/data', { recursive: true })
    await vfsInstance.value.mkdir('/workspace', { recursive: true })
    await vfsInstance.value.mkdir('/uploads', { recursive: true })

    updateVFSStats()
  } catch (error) {
    console.error('Failed to initialize VFS directories:', error)
  }
}

const handleFileUpload = async (file) => {
  if (!vfsInstance.value || !importExportInstance.value) {
    message.error('VFS not available. Please refresh the page.')
    return false
  }

  uploading.value = true

  try {
    // Extract the actual File object from Naive UI's structure
    // Naive UI structure: { file: { file: File, ... }, fileList: [] }
    const actualFile = file.file.file

    if (!actualFile || !(actualFile instanceof File)) {
      throw new Error(`Invalid file structure: expected file.file.file to be a File object, got ${typeof actualFile}`)
    }

    // Generate target path
    const targetPath = `${selectedDirectory.value}/${actualFile.name}`

    // Upload file to VFS
    const result = await importExportInstance.value.uploadFile(actualFile, targetPath)

    uploadResults.value.push({
      filename: actualFile.name,
      path: targetPath,
      ...result
    })

    if (result.success) {
      message.success(`Uploaded ${actualFile.name} to VFS`)
      updateVFSStats()

      // Sync VFS to code runners after successful upload
      try {
        const codeRunner = getCodeRunner()
        await codeRunner.syncVFSToWorkers()
        console.log('VFS synchronized to code runners after file upload')
      } catch (error) {
        console.warn('Failed to sync VFS to code runners:', error)
      }

      // Emit event for parent component
      emit('fileUploaded', {
        filename: actualFile.name,
        path: targetPath,
        size: actualFile.size,
        sessionUuid: props.sessionUuid
      })
    } else {
      message.error(result.message)
    }
  } catch (error) {
    const filename = file?.file?.file?.name || 'unknown'
    const errorResult = {
      filename: filename,
      path: `${selectedDirectory.value}/${filename}`,
      success: false,
      message: error.message
    }

    uploadResults.value.push(errorResult)
    message.error(`Failed to upload ${filename}: ${error.message}`)
  } finally {
    uploading.value = false
  }

  return false // Prevent default upload behavior
}

const updateVFSStats = () => {
  if (importExportInstance.value) {
    vfsStats.value = importExportInstance.value.getImportStats()
  }
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const generatePythonExample = () => {
  const successfulUploads = uploadResults.value.filter(r => r.success)
  if (successfulUploads.length === 0) return ''

  const examples = successfulUploads.map(upload => {
    const { filename, path } = upload
    const ext = filename.split('.').pop().toLowerCase()

    switch (ext) {
      case 'csv':
        return `# Read CSV file
import pandas as pd
df = pd.read_csv('${path}')
print(f"Loaded {len(df)} rows from ${filename}")
print(df.head())`

      case 'json':
        return `# Read JSON file
import json
with open('${path}', 'r') as f:
    data = json.load(f)
print(f"Loaded JSON data from ${filename}")
print(f"Keys: {list(data.keys()) if isinstance(data, dict) else 'Array data'}")`

      case 'xlsx':
        return `# Read Excel file
import pandas as pd
df = pd.read_excel('${path}')
print(f"Loaded {len(df)} rows from ${filename}")
print(df.info())`

      case 'txt':
        return `# Read text file
with open('${path}', 'r') as f:
    content = f.read()
print(f"Read {len(content)} characters from ${filename}")
print(content[:200] + "..." if len(content) > 200 else content)`

      default:
        return `# Read file
with open('${path}', 'r') as f:
    content = f.read()
print(f"Read file: ${filename}")
print(f"Size: {len(content)} characters")`
    }
  }).join('\n\n')

  return examples
}

const generateJavaScriptExample = () => {
  const successfulUploads = uploadResults.value.filter(r => r.success)
  if (successfulUploads.length === 0) return ''

  const examples = successfulUploads.map(upload => {
    const { filename, path } = upload
    const ext = filename.split('.').pop().toLowerCase()

    switch (ext) {
      case 'csv':
        return `// Read CSV file
const fs = require('fs');
const csvContent = fs.readFileSync('${path}', 'utf8');
const lines = csvContent.split('\\n');
const headers = lines[0].split(',');
console.log(\`Loaded CSV ${filename} with \${lines.length - 1} rows\`);
console.log('Headers:', headers);`

      case 'json':
        return `// Read JSON file
const fs = require('fs');
const data = JSON.parse(fs.readFileSync('${path}', 'utf8'));
console.log(\`Loaded JSON data from ${filename}\`);
console.log('Data:', data);`

      case 'txt':
        return `// Read text file
const fs = require('fs');
const content = fs.readFileSync('${path}', 'utf8');
console.log(\`Read \${content.length} characters from ${filename}\`);
console.log(content.substring(0, 200) + (content.length > 200 ? '...' : ''));`

      default:
        return `// Read file
const fs = require('fs');
const content = fs.readFileSync('${path}', 'utf8');
console.log(\`Read file: ${filename}\`);
console.log(\`Size: \${content.length} characters\`);`
    }
  }).join('\n\n')

  return examples
}

const copyCode = async (code) => {
  try {
    await navigator.clipboard.writeText(code)
    message.success('Code copied to clipboard')
  } catch (error) {
    message.error('Failed to copy code')
  }
}

const clearResults = () => {
  uploadResults.value = []
}

const openFileManager = () => {
  showFileManager.value = true
  showUploadModal.value = false
}

const addCodeExample = () => {
  const pythonCode = generatePythonExample()
  const jsCode = generateJavaScriptExample()

  emit('codeExampleAdded', {
    python: pythonCode,
    javascript: jsCode,
    uploadedFiles: uploadResults.value.filter(r => r.success),
    sessionUuid: props.sessionUuid
  })

  showUploadModal.value = false
  message.success('Code examples ready to use!')
}

// Lifecycle
onMounted(() => {
  if (vfsInstance.value && importExportInstance.value) {
    initializeVFS()
  }
})

// Expose methods for parent component
defineExpose({
  openUploadModal: () => { showUploadModal.value = true },
  openFileManager: () => { showFileManager.value = true },
  getUploadedFiles: () => uploadResults.value.filter(r => r.success),
  refreshStats: updateVFSStats
})
</script>

<style scoped>
.chat-vfs-uploader {
  display: inline-block;
}

.upload-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.upload-content {
  max-height: 70vh;
  overflow-y: auto;
  min-height: 400px;
  padding: 8px;
}

.directory-selection {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--code-block-bg);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.upload-dragger-content {
  text-align: center;
  padding: 60px 40px;
  min-height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 16px;
}

.upload-results {
  margin-top: 16px;
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

.file-path {
  font-size: 11px;
  color: var(--primary-color);
  margin-top: 2px;
  font-family: monospace;
}

.result-message {
  font-size: 12px;
  color: var(--text-color-3);
  margin-top: 2px;
}

.code-examples {
  margin-top: 16px;
}

.code-examples .n-code {
  max-height: 300px;
  overflow-y: auto;
}

.vfs-status {
  margin-top: 16px;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 12px;
  background: var(--code-block-bg);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.status-item {
  text-align: center;
}

.status-item strong {
  display: block;
  font-size: 20px;
  color: var(--primary-color);
  margin-bottom: 4px;
}

.status-item span {
  font-size: 12px;
  color: var(--text-color-3);
}

.loading {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
  padding: 40px;
  color: var(--text-color-3);
}

@media (max-width: 768px) {
  .status-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .upload-dragger-content {
    padding: 40px 20px;
    min-height: 150px;
  }

  .upload-content {
    min-height: 300px;
    padding: 4px;
  }

  .directory-selection {
    padding: 8px;
  }
}
</style>