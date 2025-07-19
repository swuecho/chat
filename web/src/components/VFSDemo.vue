<template>
  <div class="vfs-demo">
    <n-card title="Virtual File System Demo" size="large">
      <template #header-extra>
        <n-button @click="resetDemo" size="small">Reset Demo</n-button>
      </template>

      <div class="demo-content">
        <!-- Step 1: File Upload -->
        <div class="demo-step">
          <h3>Step 1: Upload a File</h3>
          <p>Upload a CSV file to see the VFS in action:</p>
          
          <VFSProvider v-slot="{ vfs, importExport, isReady }">
            <div v-if="!isReady" class="loading">
              <n-spin size="small" />
              <span>Initializing VFS...</span>
            </div>
            
            <div v-else class="upload-section">
              <input 
                ref="fileInputRef" 
                type="file" 
                accept=".csv,.json,.txt" 
                @change="handleFileUpload"
                style="display: none"
              />
              
              <n-button @click="$refs.fileInputRef?.click()" type="primary">
                <template #icon><n-icon><CloudUpload /></n-icon></template>
                Upload Sample File
              </n-button>
              
              <n-button @click="createSampleData" style="margin-left: 8px;">
                <template #icon><n-icon><DocumentText /></n-icon></template>
                Create Sample Data
              </n-button>
              
              <!-- File Status -->
              <div v-if="uploadedFile" class="file-status">
                <n-alert type="success" style="margin-top: 12px;">
                  <p><strong>File uploaded:</strong> {{ uploadedFile.name }}</p>
                  <p><strong>Path:</strong> <code>{{ uploadedFile.path }}</code></p>
                  <p><strong>Size:</strong> {{ formatSize(uploadedFile.size) }}</p>
                </n-alert>
              </div>
            </div>
          </VFSProvider>
        </div>

        <!-- Step 2: Python Processing -->
        <div class="demo-step">
          <h3>Step 2: Process with Python</h3>
          <p>Run this Python code to process your uploaded file:</p>
          
          <n-code 
            :code="pythonCode" 
            language="python" 
            style="margin: 12px 0;"
          />
          
          <n-button @click="copyCode('python')" size="small">
            <template #icon><n-icon><Copy /></n-icon></template>
            Copy Python Code
          </n-button>
        </div>

        <!-- Step 3: JavaScript Processing -->
        <div class="demo-step">
          <h3>Step 3: Process with JavaScript</h3>
          <p>Run this JavaScript code to further process the data:</p>
          
          <n-code 
            :code="javascriptCode" 
            language="javascript" 
            style="margin: 12px 0;"
          />
          
          <n-button @click="copyCode('javascript')" size="small">
            <template #icon><n-icon><Copy /></n-icon></template>
            Copy JavaScript Code
          </n-button>
        </div>

        <!-- Step 4: Results -->
        <div class="demo-step">
          <h3>Step 4: View Results</h3>
          <p>After running the code, you can:</p>
          
          <ul class="result-list">
            <li>✅ Check the processed data in <code>/data/processed_sales.json</code></li>
            <li>✅ View the summary report in <code>/data/summary_report.json</code></li>
            <li>✅ Download files using the file manager</li>
            <li>✅ Upload more files and repeat the process</li>
          </ul>
          
          <div class="demo-actions">
            <VFSFileUploader />
          </div>
        </div>

        <!-- VFS Status -->
        <div class="demo-step">
          <h3>VFS Status</h3>
          <VFSProvider v-slot="{ vfs, importExport, isReady }">
            <div v-if="isReady" class="vfs-status">
              <div class="status-grid">
                <div class="status-item">
                  <strong>{{ getVFSStats(importExport).totalFiles }}</strong>
                  <span>Files</span>
                </div>
                <div class="status-item">
                  <strong>{{ getVFSStats(importExport).totalDirectories }}</strong>
                  <span>Directories</span>
                </div>
                <div class="status-item">
                  <strong>{{ formatSize(getVFSStats(importExport).totalSize) }}</strong>
                  <span>Storage</span>
                </div>
              </div>
              
              <n-button @click="openFileManager" size="small" style="margin-top: 12px;">
                <template #icon><n-icon><Folder /></n-icon></template>
                Open File Manager
              </n-button>
            </div>
          </VFSProvider>
        </div>
      </div>
    </n-card>

    <!-- File Manager Modal -->
    <n-modal v-model:show="showFileManager" style="width: 90vw; max-width: 1200px;">
      <n-card title="VFS File Manager" :bordered="false" size="huge">
        <VFSProvider v-slot="{ vfs, importExport, isReady }">
          <VFSFileManager 
            v-if="isReady"
            :vfs-instance="vfs"
            :import-export="importExport"
          />
          <div v-else class="loading">
            <n-spin />
            <span>Loading file manager...</span>
          </div>
        </VFSProvider>
        
        <template #footer>
          <n-button @click="showFileManager = false">Close</n-button>
        </template>
      </n-card>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { 
  CloudUpload, 
  DocumentText, 
  Copy, 
  Folder
} from '@vicons/ionicons5'
import VFSProvider from './VFSProvider.vue'
import VFSFileUploader from './VFSFileUploader.vue'
import VFSFileManager from './VFSFileManager.vue'

const message = useMessage()
const fileInputRef = ref()
const uploadedFile = ref(null)
const showFileManager = ref(false)

// Demo code examples
const pythonCode = `import pandas as pd
import json
import os

# Check if uploaded file exists
if os.path.exists('/data/sample.csv'):
    print("📁 Reading uploaded file...")
    
    # Read the CSV file
    df = pd.read_csv('/data/sample.csv')
    print(f"📊 Loaded {len(df)} rows of data")
    print("🔍 Data preview:")
    print(df.head())
    
    # Process the data
    if 'sales' in df.columns and 'profit' in df.columns:
        df['profit_margin'] = (df['profit'] / df['sales'] * 100).round(2)
        print("💰 Calculated profit margins")
    
    # Save processed data
    df.to_json('/data/processed_sales.json', orient='records', indent=2)
    print("💾 Saved processed data to /data/processed_sales.json")
    
    # Create summary
    summary = {
        'total_rows': len(df),
        'columns': list(df.columns),
        'avg_sales': df['sales'].mean() if 'sales' in df.columns else 0,
        'max_profit': df['profit'].max() if 'profit' in df.columns else 0,
        'processing_timestamp': pd.Timestamp.now().isoformat()
    }
    
    with open('/data/python_summary.json', 'w') as f:
        json.dump(summary, f, indent=2)
    
    print("✅ Python processing complete!")
    print(f"📈 Average sales: {summary['avg_sales']:.2f}")
    
else:
    print("❌ No file found at /data/sample.csv")
    print("💡 Please upload a CSV file first or create sample data")`

const javascriptCode = `const fs = require('fs');
const path = require('path');

console.log("🚀 Starting JavaScript processing...");

try {
    // Check if Python processed data exists
    if (fs.existsSync('/data/processed_sales.json')) {
        console.log("📁 Reading Python-processed data...");
        
        // Read the processed data
        const processedData = JSON.parse(fs.readFileSync('/data/processed_sales.json', 'utf8'));
        console.log(\`📊 Loaded \${processedData.length} processed records\`);
        
        // Further analysis
        const totalSales = processedData.reduce((sum, record) => {
            return sum + (parseFloat(record.sales) || 0);
        }, 0);
        
        const avgProfitMargin = processedData.reduce((sum, record) => {
            return sum + (parseFloat(record.profit_margin) || 0);
        }, 0) / processedData.length;
        
        // Create final report
        const report = {
            summary: {
                total_records: processedData.length,
                total_sales: totalSales.toFixed(2),
                average_profit_margin: avgProfitMargin.toFixed(2) + '%',
                top_performer: processedData.reduce((top, current) => {
                    return (current.sales > top.sales) ? current : top;
                }, processedData[0])
            },
            data: processedData,
            generated_at: new Date().toISOString(),
            generated_by: 'JavaScript VFS Demo'
        };
        
        // Save final report
        fs.writeFileSync('/data/final_report.json', JSON.stringify(report, null, 2));
        console.log("💾 Saved final report to /data/final_report.json");
        
        console.log("📈 Analysis Results:");
        console.log(\`   💰 Total Sales: $\${report.summary.total_sales}\`);
        console.log(\`   📊 Average Profit Margin: \${report.summary.average_profit_margin}\`);
        console.log(\`   🏆 Top Performer: \${report.summary.top_performer?.name || 'N/A'}\`);
        
        console.log("✅ JavaScript processing complete!");
        
    } else {
        console.log("❌ No processed data found");
        console.log("💡 Please run the Python code first");
    }
    
    // List all files in VFS
    console.log("\\n📂 Files in VFS:");
    const dataFiles = fs.readdirSync('/data');
    dataFiles.forEach(file => {
        console.log(\`   📄 \${file}\`);
    });
    
} catch (error) {
    console.error("❌ Error in JavaScript processing:", error.message);
}`

// Methods
const handleFileUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return
  
  try {
    // This would integrate with VFS in a real implementation
    uploadedFile.value = {
      name: file.name,
      path: `/data/${file.name}`,
      size: file.size
    }
    
    message.success(`File "${file.name}" uploaded to VFS`)
  } catch (error) {
    message.error(`Upload failed: ${error.message}`)
  }
  
  // Reset input
  event.target.value = ''
}

const createSampleData = () => {
  // Simulate creating sample data
  const sampleData = `name,sales,profit
Alice Johnson,75000,15000
Bob Smith,82000,16400
Carol Brown,68000,13600
David Wilson,91000,18200
Emma Davis,77000,15400`

  uploadedFile.value = {
    name: 'sample.csv',
    path: '/data/sample.csv',
    size: sampleData.length
  }
  
  message.success('Sample data created in VFS')
}

const copyCode = async (type) => {
  const code = type === 'python' ? pythonCode : javascriptCode
  
  try {
    await navigator.clipboard.writeText(code)
    message.success(`${type} code copied to clipboard`)
  } catch (error) {
    message.error('Failed to copy code')
  }
}

const resetDemo = () => {
  uploadedFile.value = null
  message.info('Demo reset')
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const getVFSStats = (importExport) => {
  return importExport?.getImportStats() || { totalFiles: 0, totalDirectories: 0, totalSize: 0 }
}

const openFileManager = () => {
  showFileManager.value = true
}
</script>

<style scoped>
.vfs-demo {
  max-width: 1000px;
  margin: 0 auto;
}

.demo-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.demo-step {
  padding: 20px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--card-color);
}

.demo-step h3 {
  margin: 0 0 12px 0;
  color: var(--primary-color);
  font-size: 18px;
  font-weight: 600;
}

.demo-step p {
  margin: 0 0 12px 0;
  color: var(--text-color-2);
}

.upload-section {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.file-status {
  width: 100%;
}

.loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-color-3);
}

.result-list {
  margin: 12px 0;
  padding-left: 20px;
}

.result-list li {
  margin: 8px 0;
  color: var(--text-color);
}

.result-list code {
  background: var(--code-block-bg);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Fira Code', monospace;
  font-size: 12px;
}

.demo-actions {
  margin-top: 16px;
  display: flex;
  gap: 12px;
}

.vfs-status {
  padding: 16px;
  background: var(--code-block-bg);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 12px;
}

.status-item {
  text-align: center;
}

.status-item strong {
  display: block;
  font-size: 24px;
  color: var(--primary-color);
  margin-bottom: 4px;
}

.status-item span {
  font-size: 12px;
  color: var(--text-color-3);
}

@media (max-width: 768px) {
  .demo-step {
    padding: 16px;
  }
  
  .status-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }
  
  .upload-section {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>