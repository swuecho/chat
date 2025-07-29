# 🚀 Quick Integration Guide: Add File Upload to Chat

This guide shows you exactly how to add VFS file upload to your chat session in 3 simple steps.

## 📂 Files to Add

First, make sure you have these new components in your project:

```
web/src/components/
├── VFSProvider.vue          ✅ (already created)
├── ChatVFSUploader.vue      ✅ (already created)  
├── VFSFileManager.vue       ✅ (already created)
└── VFSIntegration.vue       ✅ (already created)

web/src/utils/
├── virtualFileSystem.js     ✅ (already created)
└── vfsImportExport.js       ✅ (already created)
```

## 🔧 Step 1: Modify Conversation.vue

Add these **3 lines** to your `web/src/views/chat/components/Conversation.vue`:

### Add Imports (at the top with other imports):
```typescript
// Add these two lines to existing imports
import ChatVFSUploader from '@/components/ChatVFSUploader.vue'
import VFSProvider from '@/components/VFSProvider.vue'
```

### Add Event Handler (in script setup section):
```typescript
// Add this event handler with your other functions
const handleVFSFileUploaded = (fileInfo: any) => {
  nui_msg.success(`📁 File uploaded: ${fileInfo.filename}`)
  
  // Optional: Add a helpful message to chat
  const helpMessage = `File uploaded to VFS! Use this code to access it:

**Python:**
\`\`\`python
# For ${fileInfo.filename}
import pandas as pd  # if CSV
data = pd.read_csv('${fileInfo.path}')
print(data.head())
\`\`\`

**JavaScript:**
\`\`\`javascript
// For ${fileInfo.filename}
const fs = require('fs');
const content = fs.readFileSync('${fileInfo.path}', 'utf8');
console.log(content);
\`\`\``

  addChat(
    sessionUuid,
    {
      uuid: uuidv7(),
      dateTime: nowISO(),
      text: helpMessage,
      inversion: false,
      error: false,
      loading: false,
    },
  )
}
```

### Modify Template (wrap and add uploader):

Find your existing template and make these changes:

**BEFORE:**
```vue
<template>
  <div class="flex flex-col w-full h-full">
    <!-- existing content -->
  </div>
</template>
```

**AFTER:**
```vue
<template>
  <VFSProvider>
    <div class="flex flex-col w-full h-full">
      <!-- ALL existing content stays exactly the same -->
      
      <!-- Just add this one line in your footer, before the input area -->
      <footer :class="footerClass">
        <div class="w-full max-w-screen-xl m-auto">
          
          <!-- ADD THIS LINE -->
          <div class="mb-2 flex justify-end">
            <ChatVFSUploader 
              :session-uuid="sessionUuid" 
              @file-uploaded="handleVFSFileUploaded" 
            />
          </div>
          
          <!-- All existing buttons and input stay the same -->
          <div class="flex items-center justify-between space-x-1">
            <!-- existing content unchanged -->
          </div>
        </div>
      </footer>
    </div>
  </VFSProvider>
</template>
```

## 🎯 Step 2: Test It!

1. **Start your dev server**
2. **Go to any chat session** 
3. **Look for the folder icon button** in the bottom right of the chat
4. **Click it and upload a CSV or JSON file**
5. **The system will show code examples** for accessing the file
6. **Copy and run the code** in a code runner!

## 📋 Complete Example

Here's what a complete integration looks like in your Conversation.vue:

```vue
<script lang='ts' setup>
// Existing imports...
import { NAutoComplete, NButton, NInput, NModal, NSpin, useDialog, useMessage } from 'naive-ui'
// ... other existing imports ...

// ADD THESE TWO IMPORTS
import ChatVFSUploader from '@/components/ChatVFSUploader.vue'
import VFSProvider from '@/components/VFSProvider.vue'

// ... all your existing code stays the same ...

// ADD THIS EVENT HANDLER
const handleVFSFileUploaded = (fileInfo: any) => {
  nui_msg.success(`📁 File uploaded: ${fileInfo.filename}`)
  
  const helpMessage = `File **${fileInfo.filename}** uploaded to VFS at \`${fileInfo.path}\`

Try this code to access it:

**Python:**
\`\`\`python
# Read your uploaded file
${fileInfo.filename.endsWith('.csv') ? 
  `import pandas as pd\ndf = pd.read_csv('${fileInfo.path}')\nprint(df.head())` :
  fileInfo.filename.endsWith('.json') ?
  `import json\nwith open('${fileInfo.path}', 'r') as f:\n    data = json.load(f)\nprint(data)` :
  `with open('${fileInfo.path}', 'r') as f:\n    content = f.read()\nprint(content)`
}
\`\`\`

Your file is now available in the Virtual File System! 🚀`

  addChat(
    sessionUuid,
    {
      uuid: uuidv7(),
      dateTime: nowISO(),
      text: helpMessage,
      inversion: false,
      error: false,
      loading: false,
    },
  )
}

// ... rest of existing code unchanged ...
</script>

<template>
  <!-- WRAP everything with VFSProvider -->
  <VFSProvider>
    <div class="flex flex-col w-full h-full">
      <!-- All existing content stays exactly the same -->
      <div>
        <UploadModal :sessionUuid="sessionUuid" :showUploadModal="showUploadModal"
          @update:showUploadModal="showUploadModal = $event" />
      </div>
      <HeaderMobile v-if="isMobile" @add-chat="handleAdd" @snapshot="handleSnapshot" @toggle="showModal = true" />
      <main class="flex-1 overflow-hidden">
        <!-- ... all existing main content ... -->
      </main>
      
      <!-- In footer, add VFS uploader -->
      <footer :class="footerClass">
        <div class="w-full max-w-screen-xl m-auto">
          
          <!-- ADD THIS VFS UPLOAD SECTION -->
          <div class="mb-2 flex justify-end items-center gap-2 pb-2 border-b border-gray-200 dark:border-gray-700">
            <span class="text-xs text-gray-500">Upload files for code runners:</span>
            <ChatVFSUploader 
              :session-uuid="sessionUuid" 
              @file-uploaded="handleVFSFileUploaded" 
            />
          </div>
          
          <!-- All existing input area stays exactly the same -->
          <div class="flex items-center justify-between space-x-1">
            <!-- ... all existing buttons and input unchanged ... -->
          </div>
        </div>
      </footer>
    </div>
  </VFSProvider>
</template>
```

## 🎉 That's It!

You now have:

✅ **File Upload Button** - Users can upload files to VFS  
✅ **Auto Code Generation** - System shows how to use uploaded files  
✅ **Cross-Language Support** - Files work in both Python and JavaScript  
✅ **Session Integration** - Files persist throughout the chat session  
✅ **File Manager** - Browse and download VFS files  

## 💡 Usage Examples

After integration, users can:

1. **Upload** `sales.csv` → VFS stores it at `/data/sales.csv`
2. **Get code examples** automatically in chat
3. **Run Python code:**
   ```python
   import pandas as pd
   df = pd.read_csv('/data/sales.csv')
   df['profit_margin'] = df['profit'] / df['sales'] * 100
   df.to_csv('/data/sales_with_margins.csv', index=False)
   ```
4. **Run JavaScript code:**
   ```javascript
   const fs = require('fs');
   const data = fs.readFileSync('/data/sales_with_margins.csv', 'utf8');
   console.log('Processed data:', data.split('\n').length, 'rows');
   ```
5. **Download results** via the file manager

The VFS creates a complete data processing workflow within your chat interface! 🚀