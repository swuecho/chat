# VFS Integration Example

This document shows how to integrate the Virtual File System upload functionality into the existing chat interface.

## Step 1: Add VFS Components to Chat

### Import VFS Components

Add these imports to your `Conversation.vue` component:

```typescript
// Add to existing imports
import VFSIntegration from '@/components/VFSIntegration.vue'
```

### Add VFS Integration to Template

Add the VFS integration component near the upload button area. Here's how to modify the footer section:

```vue
<template>
  <div class="flex flex-col w-full h-full">
    <!-- Existing content... -->
    
    <footer :class="footerClass">
      <div class="w-full max-w-screen-xl m-auto">
        <!-- Add VFS Integration above the input area -->
        <div class="vfs-section mb-2">
          <VFSIntegration 
            @file-uploaded="handleVFSFileUploaded"
            @sample-files-created="handleSampleFilesCreated"
          />
        </div>
        
        <div class="flex items-center justify-between space-x-1">
          <!-- Existing buttons and input... -->
          
          <!-- Modify the upload button to show both options -->
          <div class="upload-options">
            <!-- Original upload button -->
            <button class="!-ml-8 z-10" @click="showUploadModal = true" title="Upload to Server">
              <span class="text-xl text-[#4b9e5f]">
                <SvgIcon icon="clarity:attachment-line" />
              </span>
            </button>
            
            <!-- VFS upload is now handled by VFSIntegration component above -->
          </div>
          
          <!-- Rest of existing template... -->
        </div>
      </div>
    </footer>
  </div>
</template>
```

### Add Event Handlers

Add these methods to your component:

```typescript
// Add to script setup
const handleVFSFileUploaded = (fileInfo: any) => {
  console.log('File uploaded to VFS:', fileInfo)
  nui_msg.success(`File uploaded to VFS: ${fileInfo.filename}`)
}

const handleSampleFilesCreated = () => {
  console.log('Sample files created in VFS')
  nui_msg.success('Sample files created! Try running the example scripts.')
}
```

## Step 2: Complete Integration Example

Here's a complete example of how to add VFS to an existing chat component:

```vue
<script lang='ts' setup>
// Existing imports...
import VFSIntegration from '@/components/VFSIntegration.vue'

// Existing code...

// Add VFS event handlers
const handleVFSFileUploaded = (fileInfo: any) => {
  nui_msg.success(`File uploaded to VFS: ${fileInfo.filename}`)
  
  // Optionally add a message to the chat suggesting how to use the file
  if (fileInfo.filename.endsWith('.csv')) {
    const suggestion = `File uploaded! You can now access it in your code:

Python:
\`\`\`python
import pandas as pd
df = pd.read_csv('${fileInfo.path}')
print(df.head())
\`\`\`

JavaScript:
\`\`\`javascript
const fs = require('fs');
const content = fs.readFileSync('${fileInfo.path}', 'utf8');
console.log('File content:', content);
\`\`\`
`
    
    // Add this suggestion as a system message (optional)
    addChat(
      sessionUuid,
      {
        uuid: uuidv7(),
        dateTime: nowISO(),
        text: suggestion,
        inversion: false,
        error: false,
        loading: false,
        isSystem: true, // Mark as system message
      },
    )
  }
}

const handleSampleFilesCreated = () => {
  nui_msg.success('Sample files created! Try running the example scripts.')
  
  // Optionally add a message about the sample files
  const sampleInfo = `Sample files have been created in your Virtual File System! 

Try these examples:

**Python example:**
\`\`\`python
# Run this to see VFS in action
exec(open('/workspace/sample_script.py').read())
\`\`\`

**JavaScript example:**
\`\`\`javascript
// Run this to see VFS in action
eval(require('fs').readFileSync('/workspace/sample_script.js', 'utf8'))
\`\`\`

Use the "Files" button to browse all available files.`

  addChat(
    sessionUuid,
    {
      uuid: uuidv7(),
      dateTime: nowISO(),
      text: sampleInfo,
      inversion: false,
      error: false,
      loading: false,
      isSystem: true,
    },
  )
}
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <!-- Existing header and main content... -->
    
    <footer :class="footerClass">
      <div class="w-full max-w-screen-xl m-auto">
        <!-- VFS Integration Section -->
        <div class="vfs-section mb-3 p-2 bg-gray-50 dark:bg-gray-800 rounded-lg border">
          <VFSIntegration 
            @file-uploaded="handleVFSFileUploaded"
            @sample-files-created="handleSampleFilesCreated"
          />
        </div>
        
        <!-- Existing input area... -->
        <div class="flex items-center justify-between space-x-1">
          <!-- All existing buttons and input -->
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.vfs-section {
  transition: all 0.2s ease;
}

.vfs-section:hover {
  background-color: var(--hover-color);
}
</style>
```

## Step 3: Alternative Minimal Integration

If you prefer a minimal integration, you can add just the upload button:

```vue
<script setup>
import VFSFileUploader from '@/components/VFSFileUploader.vue'

// In your existing button area
const vfsUploaderRef = ref()

const openVFSUpload = () => {
  vfsUploaderRef.value?.openUploadModal()
}
</script>

<template>
  <!-- Add VFS Provider at the root level -->
  <VFSProvider>
    <div class="flex flex-col w-full h-full">
      <!-- Existing content... -->
      
      <!-- In your button area, add VFS upload button -->
      <div class="flex items-center space-x-2">
        <!-- Existing upload button -->
        <button @click="showUploadModal = true" title="Upload to Server">
          <SvgIcon icon="clarity:attachment-line" />
        </button>
        
        <!-- VFS upload button -->
        <VFSFileUploader ref="vfsUploaderRef" />
      </div>
    </div>
  </VFSProvider>
</template>
```

## Step 4: User Experience Flow

With VFS integration, users can:

1. **Upload Files**: Click "Upload to VFS" button
2. **Select Directory**: Choose `/data`, `/workspace`, `/tmp`, or `/uploads`
3. **Drag & Drop**: Upload multiple files at once
4. **Use in Code**: Access files immediately in Python/JavaScript
5. **Download Results**: Get processed files back

### Example User Workflow:

1. User uploads `sales.csv` to `/data/sales.csv`
2. User runs Python code:
   ```python
   import pandas as pd
   df = pd.read_csv('/data/sales.csv')
   df['profit_margin'] = df['profit'] / df['revenue'] 
   df.to_csv('/data/analyzed_sales.csv', index=False)
   ```
3. User runs JavaScript code:
   ```javascript
   const fs = require('fs');
   const data = fs.readFileSync('/data/analyzed_sales.csv', 'utf8');
   console.log('Analysis complete, rows:', data.split('\n').length);
   ```
4. User downloads the processed file via the file manager

## Benefits

- **Seamless Integration**: Files work across Python and JavaScript
- **No Server Storage**: Everything stays in browser memory
- **Rich File Operations**: Full file system API support
- **Data Processing Workflows**: Upload → Process → Download
- **Session Management**: Save/restore entire file collections

This integration transforms the chat interface into a complete data processing environment where users can upload their data, process it with AI-generated code, and download the results.