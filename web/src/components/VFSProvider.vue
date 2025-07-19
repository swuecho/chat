<template>
  <div class="vfs-provider">
    <slot :vfs="vfs" :importExport="importExport" :isReady="isVFSReady" />
  </div>
</template>

<script setup>
import { ref, provide, onMounted, onUnmounted } from 'vue'
import { useMessage } from 'naive-ui'

const message = useMessage()

// VFS instances
const vfs = ref(null)
const importExport = ref(null)
const isVFSReady = ref(false)

// Initialize VFS when component mounts
onMounted(async () => {
  try {
    // Dynamic import to avoid loading VFS on every page
    const [VirtualFileSystem, VFSImportExport] = await Promise.all([
      import('@/utils/virtualFileSystem.js').then(m => m.default || m.VirtualFileSystem),
      import('@/utils/vfsImportExport.js').then(m => m.default || m.VFSImportExport)
    ])

    // Initialize VFS
    vfs.value = new VirtualFileSystem({
      maxFileSize: 10 * 1024 * 1024,  // 10MB per file
      maxTotalSize: 100 * 1024 * 1024, // 100MB total
      maxFiles: 500
    })

    // Initialize Import/Export
    importExport.value = new VFSImportExport(vfs.value)

    // Create default directories
    await vfs.value.mkdir('/data', { recursive: true })
    await vfs.value.mkdir('/workspace', { recursive: true })
    await vfs.value.mkdir('/tmp', { recursive: true })
    await vfs.value.mkdir('/uploads', { recursive: true })

    // Set VFS as ready
    isVFSReady.value = true

    // Provide instances to child components
    provide('vfsInstance', vfs.value)
    provide('importExportInstance', importExport.value)

    console.log('VFS initialized successfully')
  } catch (error) {
    console.error('Failed to initialize VFS:', error)
    message.error('Failed to initialize Virtual File System')
  }
})

// Provide reactive references
provide('vfsInstance', vfs)
provide('importExportInstance', importExport)
provide('isVFSReady', isVFSReady)

// Cleanup on unmount
onUnmounted(() => {
  if (vfs.value) {
    // Could add cleanup logic here if needed
    console.log('VFS provider unmounted')
  }
})

// Expose VFS instances for parent components
defineExpose({
  vfs,
  importExport,
  isVFSReady,
  
  // Helper methods
  getVFSStats: () => {
    if (importExport.value) {
      return importExport.value.getImportStats()
    }
    return { totalFiles: 0, totalDirectories: 0, totalSize: 0 }
  },
  
  clearVFS: () => {
    if (vfs.value) {
      vfs.value.clear()
    }
  },
  
  createSampleFiles: async () => {
    if (!vfs.value) return
    
    try {
      // Create sample data files
      const sampleCSV = `name,age,city,salary
John Doe,30,New York,75000
Jane Smith,25,San Francisco,85000
Bob Johnson,35,Chicago,65000`

      const sampleJSON = {
        "users": [
          {"id": 1, "name": "Alice", "active": true},
          {"id": 2, "name": "Bob", "active": false},
          {"id": 3, "name": "Charlie", "active": true}
        ],
        "metadata": {
          "version": "1.0",
          "created": new Date().toISOString()
        }
      }

      const samplePython = `# Sample Python script for VFS
import json
import os

# Read data from VFS
if os.path.exists('/data/sample.json'):
    with open('/data/sample.json', 'r') as f:
        data = json.load(f)
    print(f"Loaded {len(data['users'])} users")
else:
    print("No data file found")

# Create some output
result = {"message": "Hello from Python VFS!", "timestamp": "2024-01-01"}
with open('/data/python_output.json', 'w') as f:
    json.dump(result, f, indent=2)

print("Python script completed successfully")`

      const sampleJS = `// Sample JavaScript script for VFS
const fs = require('fs');
const path = require('path');

// Read data from VFS
try {
  if (fs.existsSync('/data/sample.json')) {
    const data = JSON.parse(fs.readFileSync('/data/sample.json', 'utf8'));
    console.log(\`Loaded \${data.users.length} users\`);
    
    // Process data
    const activeUsers = data.users.filter(u => u.active);
    console.log(\`Found \${activeUsers.length} active users\`);
    
    // Save processed data
    fs.writeFileSync('/data/active_users.json', JSON.stringify(activeUsers, null, 2));
  } else {
    console.log('No data file found');
  }
} catch (error) {
  console.error('Error processing data:', error.message);
}

console.log('JavaScript script completed successfully');`

      // Write sample files
      await vfs.value.writeFile('/data/sample.csv', sampleCSV)
      await vfs.value.writeFile('/data/sample.json', JSON.stringify(sampleJSON, null, 2))
      await vfs.value.writeFile('/workspace/sample_script.py', samplePython)
      await vfs.value.writeFile('/workspace/sample_script.js', sampleJS)

      // Create README
      const readme = `# Virtual File System Demo

Welcome to the VFS! This directory contains sample files to demonstrate the capabilities.

## Files included:

- \`sample.csv\` - Sample CSV data with employee information
- \`sample.json\` - Sample JSON data with user information  
- \`sample_script.py\` - Python script that reads and processes VFS files
- \`sample_script.js\` - JavaScript script that reads and processes VFS files

## How to use:

1. **Upload your own files** using the "Upload to VFS" button
2. **Run the sample scripts** in the code runners to see VFS in action
3. **Create new files** by writing code that saves data to VFS paths
4. **Download results** using the file manager

## Available directories:

- \`/data\` - Store your data files (CSV, JSON, etc.)
- \`/workspace\` - Store your code and working files  
- \`/tmp\` - Store temporary files
- \`/uploads\` - Default location for uploaded files

Try running the sample scripts to see how Python and JavaScript can work with the same files!`

      await vfs.value.writeFile('/workspace/README.md', readme)

      message.success('Sample files created successfully!')
      return true
    } catch (error) {
      console.error('Failed to create sample files:', error)
      message.error('Failed to create sample files')
      return false
    }
  }
})
</script>

<style scoped>
.vfs-provider {
  height: 100%;
  width: 100%;
}
</style>