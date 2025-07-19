# Chat Session VFS Integration

This guide shows how to add VFS file upload functionality to chat sessions so users can upload files and use them directly in code runners.

## Integration Steps

### 1. Add VFS Provider to Chat Layout

First, wrap your chat layout with the VFS provider in `web/src/views/chat/layout/Layout.vue`:

```vue
<script setup>
// Add VFS imports
import VFSProvider from '@/components/VFSProvider.vue'
// ... existing imports
</script>

<template>
  <VFSProvider>
    <div class="h-full flex">
      <!-- Existing layout content -->
      <NLayout class="z-40 transition" :has-sider="true">
        <!-- ... existing layout -->
      </NLayout>
    </div>
  </VFSProvider>
</template>
```

### 2. Modify Conversation Component

Update `web/src/views/chat/components/Conversation.vue`:

```vue
<script lang='ts' setup>
// Add VFS imports
import ChatVFSUploader from '@/components/ChatVFSUploader.vue'
import VFSProvider from '@/components/VFSProvider.vue'

// ... existing imports and code ...

// Add VFS event handlers
const handleVFSFileUploaded = (fileInfo: any) => {
  console.log('File uploaded to VFS:', fileInfo)
  nui_msg.success(`File uploaded: ${fileInfo.filename} → ${fileInfo.path}`)
}

const handleCodeExampleAdded = (codeInfo: any) => {
  console.log('Code examples generated:', codeInfo)
  
  // Option 1: Add the code as a system message in chat
  const exampleMessage = `📁 **Files uploaded successfully!**

**Python example:**
\`\`\`python
${codeInfo.python}
\`\`\`

**JavaScript example:**
\`\`\`javascript
${codeInfo.javascript}
\`\`\`

Your files are now available in the Virtual File System and can be accessed using the paths shown above.`

  // Add system message to chat
  addChat(
    sessionUuid,
    {
      uuid: uuidv7(),
      dateTime: nowISO(),
      text: exampleMessage,
      inversion: false,
      error: false,
      loading: false,
      isSystem: true, // Mark as system message
    },
  )
  
  // Option 2: Pre-fill the input with code (alternative)
  // prompt.value = codeInfo.python // or codeInfo.javascript
}

// ... rest of existing code ...
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <!-- Existing upload modal -->
    <div>
      <UploadModal :sessionUuid="sessionUuid" :showUploadModal="showUploadModal"
        @update:showUploadModal="showUploadModal = $event" />
    </div>
    
    <!-- ... existing header and main content ... -->

    <footer :class="footerClass">
      <div class="w-full max-w-screen-xl m-auto">
        <!-- Add VFS uploader above the input area -->
        <div class="vfs-upload-section mb-2">
          <ChatVFSUploader 
            :session-uuid="sessionUuid"
            @file-uploaded="handleVFSFileUploaded"
            @code-example-added="handleCodeExampleAdded"
          />
        </div>
        
        <div class="flex items-center justify-between space-x-1">
          <!-- ... existing buttons ... -->
          
          <!-- Modify the original upload button to distinguish it -->
          <button class="!-ml-8 z-10" @click="showUploadModal = true" title="Upload to Server (for chat context)">
            <span class="text-xl text-[#4b9e5f]">
              <SvgIcon icon="clarity:attachment-line" />
            </span>
          </button>
          
          <!-- ... rest of existing input area ... -->
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.vfs-upload-section {
  display: flex;
  justify-content: flex-end;
  padding: 8px 0;
  border-top: 1px solid var(--border-color);
  margin-top: 8px;
}

/* Optional: Add visual distinction */
.vfs-upload-section::before {
  content: "📁 Upload files for code runners:";
  font-size: 12px;
  color: var(--text-color-3);
  margin-right: auto;
  display: flex;
  align-items: center;
}

@media (max-width: 768px) {
  .vfs-upload-section {
    justify-content: center;
  }
  
  .vfs-upload-section::before {
    display: none;
  }
}
</style>
```

### 3. Alternative: Simpler Integration

For a cleaner integration, you can add the VFS uploader as a button next to the existing upload button:

```vue
<template>
  <!-- In the footer button area -->
  <div class="flex items-center justify-between space-x-1">
    <!-- ... existing buttons ... -->
    
    <!-- Upload buttons group -->
    <div class="upload-buttons-group">
      <!-- Original upload (for chat context) -->
      <n-tooltip placement="top">
        <template #trigger>
          <button class="upload-btn" @click="showUploadModal = true">
            <span class="text-xl text-[#4b9e5f]">
              <SvgIcon icon="clarity:attachment-line" />
            </span>
          </button>
        </template>
        Upload files for chat context
      </n-tooltip>
      
      <!-- VFS upload (for code runners) -->
      <ChatVFSUploader 
        :session-uuid="sessionUuid"
        @file-uploaded="handleVFSFileUploaded"
        @code-example-added="handleCodeExampleAdded"
      />
    </div>
    
    <!-- ... rest of existing input and send button ... -->
  </div>
</template>

<style scoped>
.upload-buttons-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.upload-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  transition: background-color 0.2s ease;
}

.upload-btn:hover {
  background-color: var(--hover-color);
}
</style>
```

## Usage Flow

### 1. User uploads files via VFS uploader:
- Click the VFS upload button (folder icon)
- Select files (CSV, JSON, Excel, text files, etc.)
- Choose target directory (`/data`, `/workspace`, `/uploads`)
- Files are uploaded to the Virtual File System

### 2. System generates code examples:
- Automatically generates Python and JavaScript examples
- Shows how to access the uploaded files
- User can copy the code or add it to chat

### 3. User runs code in chat:
```python
# Python example (auto-generated)
import pandas as pd
df = pd.read_csv('/data/sales.csv')
print(f"Loaded {len(df)} rows from sales.csv")
print(df.head())
```

```javascript
// JavaScript example (auto-generated)  
const fs = require('fs');
const csvContent = fs.readFileSync('/data/sales.csv', 'utf8');
const lines = csvContent.split('\n');
console.log(`Loaded CSV sales.csv with ${lines.length - 1} rows`);
```

### 4. Files persist throughout the session:
- Files remain available in VFS during the entire chat session
- Can be accessed by subsequent code runners
- Can be downloaded via file manager

## Key Features

### Dual Upload System:
- **Server Upload** (existing): Files for chat context, AI can see content
- **VFS Upload** (new): Files for code execution, available in runners

### Smart Code Generation:
- Detects file types and generates appropriate code
- Provides both Python and JavaScript examples
- Shows exact file paths for easy copy-paste

### Session Integration:
- Files tied to specific chat session
- Automatic code example insertion into chat
- Visual feedback and success messages

### User Experience:
- Clear distinction between upload types
- Helpful tooltips and guidance
- Immediate code examples
- File manager for browsing/downloading

## Complete Example Message Flow

1. **User uploads** `sales_data.csv` via VFS uploader
2. **System responds** with auto-generated message:
   ```
   📁 Files uploaded successfully!
   
   Python example:
   ```python
   import pandas as pd
   df = pd.read_csv('/data/sales_data.csv')
   print(f"Loaded {len(df)} rows from sales_data.csv")
   ```
   
   JavaScript example:
   ```javascript
   const fs = require('fs');
   const csvContent = fs.readFileSync('/data/sales_data.csv', 'utf8');
   ```
   ```
3. **User copies and runs** the Python code
4. **User processes data** and saves results to VFS
5. **User downloads** processed files via file manager

This creates a complete data processing workflow within the chat interface!