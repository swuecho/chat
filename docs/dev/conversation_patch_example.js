/**
 * Example patch for Conversation.vue to add VFS file upload
 * 
 * This shows the minimal changes needed to add VFS upload to the chat interface
 */

// 1. Add these imports to the existing imports section
const additionalImports = `
import ChatVFSUploader from '@/components/ChatVFSUploader.vue'
import VFSProvider from '@/components/VFSProvider.vue'
`

// 2. Add these event handlers to the script section
const eventHandlers = `
// VFS event handlers
const handleVFSFileUploaded = (fileInfo) => {
  console.log('File uploaded to VFS:', fileInfo)
  nui_msg.success(\`File uploaded: \${fileInfo.filename} → \${fileInfo.path}\`)
}

const handleCodeExampleAdded = (codeInfo) => {
  // Add code examples as a system message
  const exampleMessage = \`📁 **Files uploaded successfully!**

**Python example:**
\\\`\\\`\\\`python
\${codeInfo.python}
\\\`\\\`\\\`

**JavaScript example:**
\\\`\\\`\\\`javascript
\${codeInfo.javascript}
\\\`\\\`\\\`

Your files are now available in the Virtual File System.\`

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
    },
  )
  
  nui_msg.success('Files uploaded! Code examples added to chat.')
}
`

// 3. Template modification - wrap the entire template with VFSProvider
const templateWrapper = `
<template>
  <VFSProvider>
    <div class="flex flex-col w-full h-full">
      <!-- All existing content stays the same -->
      
      <!-- Add VFS uploader in the footer section, before the input area -->
      <footer :class="footerClass">
        <div class="w-full max-w-screen-xl m-auto">
          
          <!-- NEW: VFS Upload Section -->
          <div class="vfs-upload-section">
            <ChatVFSUploader 
              :session-uuid="sessionUuid"
              @file-uploaded="handleVFSFileUploaded"
              @code-example-added="handleCodeExampleAdded"
            />
          </div>
          
          <!-- Existing input area remains unchanged -->
          <div class="flex items-center justify-between space-x-1">
            <!-- All existing buttons and input stay exactly the same -->
          </div>
        </div>
      </footer>
    </div>
  </VFSProvider>
</template>
`

// 4. Add these styles
const additionalStyles = `
<style scoped>
/* Add to existing styles */
.vfs-upload-section {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding: 8px 0;
  margin-bottom: 8px;
  border-top: 1px solid var(--border-color);
}

.vfs-upload-section::before {
  content: "📁 Upload files for code runners:";
  font-size: 12px;
  color: var(--text-color-3);
  margin-right: 12px;
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
`

// 5. Complete minimal integration example
const minimalIntegrationExample = `
// MINIMAL INTEGRATION: Just add these 3 things to Conversation.vue

// 1. Add import
import ChatVFSUploader from '@/components/ChatVFSUploader.vue'
import VFSProvider from '@/components/VFSProvider.vue'

// 2. Add event handler
const handleVFSFileUploaded = (fileInfo) => {
  nui_msg.success(\`File uploaded: \${fileInfo.filename}\`)
}

// 3. Add to template (wrap existing content with VFSProvider and add uploader)
/*
<template>
  <VFSProvider>
    <div class="flex flex-col w-full h-full">
      <!-- existing content -->
      
      <footer :class="footerClass">
        <div class="w-full max-w-screen-xl m-auto">
          <!-- Add this line before the existing input area -->
          <ChatVFSUploader :session-uuid="sessionUuid" @file-uploaded="handleVFSFileUploaded" />
          
          <!-- existing input area unchanged -->
          <div class="flex items-center justify-between space-x-1">
            <!-- existing buttons and input -->
          </div>
        </div>
      </footer>
    </div>
  </VFSProvider>
</template>
*/
`

export {
  additionalImports,
  eventHandlers,
  templateWrapper,
  additionalStyles,
  minimalIntegrationExample
}