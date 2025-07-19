<template>
  <div class="vfs-integration">
    <!-- VFS Provider wraps everything and provides VFS instances -->
    <VFSProvider v-slot="{ vfs, importExport, isReady }">
      <!-- VFS Controls - Always visible when VFS is ready -->
      <div v-if="isReady" class="vfs-controls">
        <div class="vfs-toolbar">
          <!-- Upload Button -->
          <VFSFileUploader 
            ref="uploaderRef"
            @file-uploaded="handleFileUploaded"
            @vfs-cleared="handleVFSCleared"
            @vfs-session-imported="handleSessionImported"
          />
          
          <!-- Quick Actions -->
          <n-button-group size="small">
            <n-button @click="openFileManager" ghost>
              <template #icon><n-icon><Folder /></n-icon></template>
              Files
            </n-button>
            
            <n-button @click="createSampleFiles" ghost>
              <template #icon><n-icon><DocumentText /></n-icon></template>
              Samples
            </n-button>
            
            <n-button @click="showVFSInfo = true" ghost>
              <template #icon><n-icon><InformationCircle /></n-icon></template>
              Info
            </n-button>
          </n-button-group>
        </div>

        <!-- VFS Status Bar (collapsible) -->
        <div v-if="showStatusBar" class="vfs-status-bar">
          <div class="status-items">
            <div class="status-item">
              <n-icon><Document /></n-icon>
              <span>{{ vfsStats.totalFiles }} files</span>
            </div>
            <div class="status-item">
              <n-icon><Archive /></n-icon>
              <span>{{ formatSize(vfsStats.totalSize) }}</span>
            </div>
            <div class="status-item">
              <n-icon><Folder /></n-icon>
              <span>{{ vfsStats.totalDirectories }} dirs</span>
            </div>
          </div>
          
          <n-button @click="showStatusBar = false" text size="tiny">
            <n-icon><Close /></n-icon>
          </n-button>
        </div>

        <!-- File Manager Modal -->
        <n-modal v-model:show="showFileManager" style="width: 90vw; max-width: 1200px;">
          <n-card title="Virtual File System Manager" :bordered="false" size="huge" role="dialog">
            <template #header-extra>
              <n-button @click="refreshVFSStats" size="small" quaternary>
                <template #icon><n-icon><Refresh /></n-icon></template>
                Refresh
              </n-button>
            </template>
            
            <VFSFileManager 
              ref="fileManagerRef"
              :vfs-instance="vfs"
              :import-export="importExport"
            />
            
            <template #footer>
              <n-space>
                <n-button @click="exportVFSSession">Export Session</n-button>
                <n-button @click="showFileManager = false">Close</n-button>
              </n-space>
            </template>
          </n-card>
        </n-modal>

        <!-- VFS Info Modal -->
        <n-modal v-model:show="showVFSInfo" preset="dialog" title="Virtual File System">
          <div class="vfs-info-content">
            <n-alert type="info" style="margin-bottom: 16px;">
              <p>The Virtual File System (VFS) allows you to work with files in both Python and JavaScript code runners.</p>
            </n-alert>

            <n-tabs default-value="usage" size="medium">
              <n-tab-pane name="usage" tab="How to Use">
                <div class="usage-content">
                  <h4>Getting Started:</h4>
                  <ol>
                    <li><strong>Upload files</strong> using the "Upload to VFS" button</li>
                    <li><strong>Access files</strong> in your code using standard file operations</li>
                    <li><strong>Create new files</strong> by writing code that saves to VFS paths</li>
                    <li><strong>Download results</strong> using the file manager</li>
                  </ol>

                  <h4>Python Example:</h4>
                  <n-code language="python" :code="pythonExample" />

                  <h4>JavaScript Example:</h4>
                  <n-code language="javascript" :code="javascriptExample" />

                  <h4>Available Directories:</h4>
                  <ul>
                    <li><code>/data</code> - Store your data files (CSV, JSON, etc.)</li>
                    <li><code>/workspace</code> - Store your code and working files</li>
                    <li><code>/tmp</code> - Store temporary files</li>
                    <li><code>/uploads</code> - Default location for uploaded files</li>
                  </ul>
                </div>
              </n-tab-pane>

              <n-tab-pane name="stats" tab="Statistics">
                <div class="stats-content">
                  <div class="stats-grid">
                    <div class="stat-card">
                      <div class="stat-value">{{ vfsStats.totalFiles }}</div>
                      <div class="stat-label">Total Files</div>
                    </div>
                    <div class="stat-card">
                      <div class="stat-value">{{ vfsStats.totalDirectories }}</div>
                      <div class="stat-label">Directories</div>
                    </div>
                    <div class="stat-card">
                      <div class="stat-value">{{ formatSize(vfsStats.totalSize) }}</div>
                      <div class="stat-label">Storage Used</div>
                    </div>
                  </div>

                  <div v-if="Object.keys(vfsStats.fileTypes || {}).length > 0" class="file-types">
                    <h4>File Types:</h4>
                    <div class="file-type-list">
                      <div v-for="(count, ext) in vfsStats.fileTypes" :key="ext" class="file-type-item">
                        <span class="extension">{{ ext || 'no extension' }}</span>
                        <span class="count">{{ count }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </n-tab-pane>
            </n-tabs>
          </div>

          <template #action>
            <n-space>
              <n-button @click="createSampleFiles">Create Sample Files</n-button>
              <n-button @click="showVFSInfo = false">Close</n-button>
            </n-space>
          </template>
        </n-modal>
      </div>

      <!-- Loading state -->
      <div v-else class="vfs-loading">
        <n-spin size="small" />
        <span>Initializing Virtual File System...</span>
      </div>
    </VFSProvider>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { 
  Folder, 
  Document, 
  DocumentText,
  InformationCircle,
  Archive,
  Close,
  Refresh
} from '@vicons/ionicons5'
import VFSProvider from './VFSProvider.vue'
import VFSFileUploader from './VFSFileUploader.vue'
import VFSFileManager from './VFSFileManager.vue'

const message = useMessage()

// Component state
const showFileManager = ref(false)
const showVFSInfo = ref(false)
const showStatusBar = ref(true)
const uploaderRef = ref()
const fileManagerRef = ref()
const vfsProviderRef = ref()

const vfsStats = ref({
  totalFiles: 0,
  totalDirectories: 0,
  totalSize: 0,
  fileTypes: {}
})

// Code examples for the info modal
const pythonExample = `# Python VFS Example
import pandas as pd
import json

# Read uploaded CSV file
df = pd.read_csv('/data/sales.csv')

# Process the data
summary = {
    'total_sales': df['amount'].sum(),
    'avg_sale': df['amount'].mean(),
    'top_customer': df.loc[df['amount'].idxmax(), 'customer']
}

# Save results
with open('/data/sales_summary.json', 'w') as f:
    json.dump(summary, f, indent=2)

print(f"Processed {len(df)} sales records")
print(f"Total sales: ${summary['total_sales']:,.2f}")`

const javascriptExample = `// JavaScript VFS Example
const fs = require('fs');

// Read the processed data from Python
const summary = JSON.parse(fs.readFileSync('/data/sales_summary.json', 'utf8'));

// Create a report
const report = {
    ...summary,
    report_date: new Date().toISOString(),
    status: 'completed'
};

// Save the report
fs.writeFileSync('/data/final_report.json', JSON.stringify(report, null, 2));

console.log('Report generated successfully');
console.log('Files available:', fs.readdirSync('/data'));`

// Methods
const refreshVFSStats = () => {
  if (vfsProviderRef.value) {
    vfsStats.value = vfsProviderRef.value.getVFSStats()
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
  refreshVFSStats()
}

const createSampleFiles = async () => {
  if (vfsProviderRef.value) {
    const success = await vfsProviderRef.value.createSampleFiles()
    if (success) {
      refreshVFSStats()
      emit('sampleFilesCreated')
    }
  }
}

const exportVFSSession = async () => {
  if (uploaderRef.value && uploaderRef.value.exportVFS) {
    await uploaderRef.value.exportVFS()
  }
}

// Event handlers
const handleFileUploaded = (fileInfo) => {
  refreshVFSStats()
  message.success(`File uploaded: ${fileInfo.filename}`)
  emit('fileUploaded', fileInfo)
}

const handleVFSCleared = () => {
  refreshVFSStats()
  emit('vfsCleared')
}

const handleSessionImported = (sessionInfo) => {
  refreshVFSStats()
  emit('sessionImported', sessionInfo)
}

// Lifecycle
onMounted(() => {
  refreshVFSStats()
})

// Emits
const emit = defineEmits(['fileUploaded', 'vfsCleared', 'sessionImported', 'sampleFilesCreated'])

// Expose methods for parent component
defineExpose({
  openFileManager,
  openUploadModal: () => uploaderRef.value?.openUploadModal(),
  createSampleFiles,
  refreshStats: refreshVFSStats,
  showInfo: () => { showVFSInfo.value = true }
})
</script>

<style scoped>
.vfs-integration {
  width: 100%;
}

.vfs-controls {
  width: 100%;
}

.vfs-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.vfs-status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: var(--code-block-bg);
  border-radius: 4px;
  margin: 8px 0;
  border: 1px solid var(--border-color);
}

.status-items {
  display: flex;
  gap: 16px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-color-3);
}

.vfs-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  color: var(--text-color-3);
  font-size: 14px;
}

.vfs-info-content {
  max-height: 60vh;
  overflow-y: auto;
}

.usage-content h4 {
  margin: 16px 0 8px 0;
  color: var(--text-color);
}

.usage-content ol, .usage-content ul {
  margin: 8px 0;
  padding-left: 20px;
}

.usage-content li {
  margin: 4px 0;
}

.usage-content code {
  background: var(--code-block-bg);
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Fira Code', monospace;
}

.stats-content {
  padding: 16px 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  text-align: center;
  padding: 16px;
  background: var(--code-block-bg);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--primary-color);
}

.stat-label {
  font-size: 12px;
  color: var(--text-color-3);
  margin-top: 4px;
}

.file-types h4 {
  margin-bottom: 12px;
}

.file-type-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
}

.file-type-item {
  display: flex;
  justify-content: space-between;
  padding: 4px 8px;
  background: var(--hover-color);
  border-radius: 4px;
  font-size: 12px;
}

.extension {
  font-family: 'Fira Code', monospace;
  color: var(--text-color);
}

.count {
  color: var(--primary-color);
  font-weight: 500;
}

@media (max-width: 768px) {
  .vfs-toolbar {
    flex-direction: column;
    gap: 8px;
    align-items: stretch;
  }

  .status-items {
    flex-direction: column;
    gap: 4px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>