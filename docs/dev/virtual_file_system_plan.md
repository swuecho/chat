# Virtual File System (VFS) Implementation Plan

## Overview

This document outlines the implementation plan for a Virtual File System (VFS) that will enable file I/O simulation and data manipulation capabilities in both JavaScript and Python code runners. The VFS will provide a secure, isolated environment for users to work with files and data without accessing the host system.

## Architecture Goals

- **Security**: Isolated from host file system with strict validation
- **Performance**: Efficient in-memory storage with compression and caching
- **Compatibility**: Standard file API compatibility for both Python and JavaScript
- **Persistence**: Session-based storage with import/export capabilities
- **Resource Management**: Configurable limits and quota enforcement

## Phase 1: Core Architecture Design

### VFS Class Structure
```javascript
class VirtualFileSystem {
  constructor() {
    this.files = new Map()           // file path -> file data
    this.directories = new Set()     // directory paths
    this.metadata = new Map()        // file metadata (size, modified, etc.)
    this.currentDirectory = '/'      // current working directory
    this.maxFileSize = 10 * 1024 * 1024  // 10MB per file
    this.maxTotalSize = 100 * 1024 * 1024 // 100MB total
    this.fileHandlers = new Map()    // extension -> handler
    this.permissions = new Map()     // path -> permissions
  }
}
```

### File System API Design
```javascript
// Core operations
fs.writeFile(path, data, options)   // Write file with encoding support
fs.readFile(path, encoding)         // Read file with optional encoding
fs.mkdir(path, recursive)           // Create directory
fs.rmdir(path, recursive)           // Remove directory
fs.unlink(path)                     // Delete file
fs.exists(path)                     // Check if path exists
fs.stat(path)                       // Get file/directory metadata
fs.readdir(path)                    // List directory contents

// Navigation
fs.chdir(path)                      // Change current directory
fs.getcwd()                         // Get current working directory

// Advanced operations
fs.copy(src, dest)                  // Copy file or directory
fs.move(src, dest)                  // Move/rename file or directory
fs.glob(pattern)                    // Find files by pattern
fs.watch(path, callback)            // Watch for changes
```

### Path Resolution System
```javascript
class PathResolver {
  normalize(path)     // Handle ./ ../ ~ / normalization
  resolve(path)       // Convert relative to absolute paths
  validate(path)      // Security checks, prevent traversal attacks
  split(path)         // Break path into components
  join(...parts)      // Join path components safely
  dirname(path)       // Get directory portion
  basename(path)      // Get filename portion
  extname(path)       // Get file extension
}
```

## Phase 2: Core Implementation

### File Storage Strategy
- **Text Files**: UTF-8 strings with BOM detection
- **Binary Files**: Base64 encoded with MIME type detection
- **Large Files**: Chunked storage with LZ4 compression
- **Memory Management**: LRU cache with configurable size limits
- **Metadata Storage**: Created/modified timestamps, permissions, checksums

### Security Model
```javascript
class VFSSecurity {
  validatePath(path)              // Prevent directory traversal
  checkQuota(size)                // Enforce storage limits
  sanitizeFilename(name)          // Remove dangerous characters
  validateFileType(data, ext)     // Prevent type confusion attacks
  enforcePermissions(path, op)    // Check read/write permissions
  scanContent(data)               // Basic malware detection
}
```

### Resource Management
- Maximum file size: 10MB per file
- Maximum total storage: 100MB per session
- Maximum files: 1000 per session
- Path length: 260 characters maximum
- Filename restrictions: No control characters, reserved names

## Phase 3: Python Runner Integration

### Custom File Handlers
```python
import io
import os
import builtins
from pathlib import Path

class VFSFile(io.TextIOWrapper):
    """File-like object that interfaces with VFS"""
    def __init__(self, path, mode, vfs_instance):
        self.vfs = vfs_instance
        self.path = path
        self.mode = mode
        self.position = 0
        self.closed = False
        
    def read(self, size=-1):
        return self.vfs.readFile(self.path, size, self.position)
        
    def write(self, data):
        return self.vfs.writeFile(self.path, data, self.mode)
        
    def seek(self, position):
        self.position = position
        
    def tell(self):
        return self.position
```

### Python Standard Library Integration
```python
# Override built-in open() function
original_open = builtins.open

def vfs_open(file, mode='r', **kwargs):
    """VFS-aware open() replacement"""
    if _is_vfs_path(file):
        return VFSFile(file, mode, global_vfs)
    return original_open(file, mode, **kwargs)

builtins.open = vfs_open

# Patch os module functions
import os
original_os_functions = {}

def patch_os_module():
    """Patch os module to work with VFS"""
    original_os_functions['listdir'] = os.listdir
    original_os_functions['makedirs'] = os.makedirs
    original_os_functions['path.exists'] = os.path.exists
    
    os.listdir = lambda path: global_vfs.readdir(path)
    os.makedirs = lambda path, **kwargs: global_vfs.mkdir(path, True)
    os.path.exists = lambda path: global_vfs.exists(path)
    
    # Patch pathlib.Path for modern Python code
    # Patch csv, json, pickle modules for data access
```

### Data Science Library Support
```python
# Pandas integration
import pandas as pd
original_read_csv = pd.read_csv
original_to_csv = pd.DataFrame.to_csv

pd.read_csv = lambda filepath, **kwargs: original_read_csv(
    global_vfs.readFile(filepath) if _is_vfs_path(filepath) else filepath, 
    **kwargs
)

# NumPy, SciPy, Matplotlib file operations
# Jupyter notebook file operations
```

## Phase 4: JavaScript Runner Integration

### Node.js-style fs Module
```javascript
const fs = {
  // Promise-based async versions
  readFile: async (path, options = {}) => {
    const encoding = options.encoding || options
    return await vfs.readFile(path, encoding)
  },
  
  writeFile: async (path, data, options = {}) => {
    return await vfs.writeFile(path, data, options)
  },
  
  mkdir: async (path, options = {}) => {
    return await vfs.mkdir(path, options.recursive)
  },
  
  readdir: async (path, options = {}) => {
    return await vfs.readdir(path, options)
  },
  
  // Synchronous versions
  readFileSync: (path, options = {}) => {
    const encoding = options.encoding || options
    return vfs.readFileSync(path, encoding)
  },
  
  writeFileSync: (path, data, options = {}) => {
    return vfs.writeFileSync(path, data, options)
  },
  
  mkdirSync: (path, options = {}) => {
    return vfs.mkdirSync(path, options.recursive)
  },
  
  readdirSync: (path, options = {}) => {
    return vfs.readdirSync(path, options)
  },
  
  // Stream support
  createReadStream: (path, options = {}) => {
    return new VFSReadableStream(path, options)
  },
  
  createWriteStream: (path, options = {}) => {
    return new VFSWritableStream(path, options)
  },
  
  // Additional utilities
  existsSync: (path) => vfs.exists(path),
  statSync: (path) => vfs.stat(path),
  unlinkSync: (path) => vfs.unlink(path),
  rmdirSync: (path, options = {}) => vfs.rmdir(path, options.recursive)
}
```

### Stream Implementation
```javascript
class VFSReadableStream extends ReadableStream {
  constructor(path, options = {}) {
    super({
      start(controller) {
        // Initialize stream with VFS file data
      },
      pull(controller) {
        // Read chunks from VFS
      },
      cancel() {
        // Cleanup
      }
    })
  }
}

class VFSWritableStream extends WritableStream {
  constructor(path, options = {}) {
    super({
      write(chunk, controller) {
        // Write chunk to VFS
      },
      close() {
        // Finalize file in VFS
      },
      abort(reason) {
        // Cleanup on error
      }
    })
  }
}
```

## Phase 5: Data Format Support

### File Type Handlers
```javascript
const FileHandlers = {
  // Text formats
  '.txt': new TextHandler(),
  '.csv': new CSVHandler(),
  '.json': new JSONHandler(),
  '.xml': new XMLHandler(),
  '.md': new MarkdownHandler(),
  '.yaml': new YAMLHandler(),
  '.toml': new TOMLHandler(),
  
  // Data formats  
  '.xlsx': new ExcelHandler(),
  '.parquet': new ParquetHandler(),
  '.sqlite': new SQLiteHandler(),
  '.h5': new HDF5Handler(),
  
  // Binary formats
  '.png': new ImageHandler(),
  '.jpg': new ImageHandler(),
  '.gif': new ImageHandler(),
  '.pdf': new PDFHandler(),
  '.zip': new ZipHandler(),
  '.tar': new TarHandler(),
  
  // Programming languages
  '.py': new PythonHandler(),
  '.js': new JavaScriptHandler(),
  '.html': new HTMLHandler(),
  '.css': new CSSHandler()
}

class CSVHandler {
  async parse(data, options = {}) {
    // Parse CSV with configurable delimiter, headers, etc.
    const delimiter = options.delimiter || ','
    const hasHeaders = options.headers !== false
    // Return structured data
  }
  
  async stringify(data, options = {}) {
    // Convert structured data to CSV string
  }
}

class JSONHandler {
  async parse(data) {
    return JSON.parse(data)
  }
  
  async stringify(data, options = {}) {
    const indent = options.indent || 2
    return JSON.stringify(data, null, indent)
  }
}
```

### Data Import/Export System
```javascript
class DataImporter {
  async importFromURL(url, path, options = {}) {
    // Fetch remote file and store in VFS
    const response = await fetch(url)
    const data = await response.arrayBuffer()
    return await vfs.writeFile(path, data, { binary: true })
  }
  
  async importFromFile(file, path) {
    // Handle browser File API uploads
    const reader = new FileReader()
    return new Promise((resolve, reject) => {
      reader.onload = async (e) => {
        await vfs.writeFile(path, e.target.result)
        resolve(path)
      }
      reader.onerror = reject
      reader.readAsArrayBuffer(file)
    })
  }
  
  async importCSV(csvText, path, options = {}) {
    const handler = new CSVHandler()
    const data = await handler.parse(csvText, options)
    await vfs.writeFile(path, JSON.stringify(data))
    return data
  }
  
  async importJSON(jsonText, path) {
    const data = JSON.parse(jsonText)
    await vfs.writeFile(path, jsonText)
    return data
  }
  
  // Export functions
  async exportToDownload(path, filename) {
    const data = await vfs.readFile(path, 'binary')
    const blob = new Blob([data])
    const url = URL.createObjectURL(blob)
    
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    
    URL.revokeObjectURL(url)
  }
  
  async exportToZip(paths, zipName) {
    // Create ZIP file containing multiple VFS files
    const JSZip = await import('jszip')
    const zip = new JSZip()
    
    for (const path of paths) {
      const data = await vfs.readFile(path, 'binary')
      const relativePath = path.startsWith('/') ? path.slice(1) : path
      zip.file(relativePath, data)
    }
    
    const zipBlob = await zip.generateAsync({ type: 'blob' })
    this.downloadBlob(zipBlob, zipName)
  }
  
  async exportToDataURL(path) {
    const data = await vfs.readFile(path, 'binary')
    const mimeType = this.detectMimeType(path)
    return `data:${mimeType};base64,${btoa(data)}`
  }
}
```

## Phase 6: Advanced Features

### File System Utilities
```javascript
class FileSystemUtils {
  // Search and filtering
  async find(pattern, options = {}) {
    // Find files by glob pattern
    const recursive = options.recursive !== false
    const maxDepth = options.maxDepth || 100
    const caseInsensitive = options.caseInsensitive || false
    
    return vfs.glob(pattern, { recursive, maxDepth, caseInsensitive })
  }
  
  async grep(pattern, files, options = {}) {
    // Search within files for text patterns
    const results = []
    const regex = new RegExp(pattern, options.flags || 'gi')
    
    for (const file of files) {
      const content = await vfs.readFile(file, 'utf8')
      const matches = [...content.matchAll(regex)]
      if (matches.length > 0) {
        results.push({ file, matches })
      }
    }
    
    return results
  }
  
  // File operations
  async compress(path, algorithm = 'gzip') {
    const data = await vfs.readFile(path, 'binary')
    const compressed = await this.compressData(data, algorithm)
    const compressedPath = `${path}.${algorithm}`
    await vfs.writeFile(compressedPath, compressed, { binary: true })
    return compressedPath
  }
  
  async decompress(path) {
    const data = await vfs.readFile(path, 'binary')
    const algorithm = this.detectCompressionType(path)
    const decompressed = await this.decompressData(data, algorithm)
    const originalPath = path.replace(new RegExp(`\.${algorithm}$`), '')
    await vfs.writeFile(originalPath, decompressed, { binary: true })
    return originalPath
  }
  
  async checksum(path, algorithm = 'sha256') {
    const data = await vfs.readFile(path, 'binary')
    const hash = await crypto.subtle.digest(algorithm.toUpperCase(), data)
    return Array.from(new Uint8Array(hash))
      .map(b => b.toString(16).padStart(2, '0'))
      .join('')
  }
  
  // Batch operations
  async bulkCopy(srcPattern, destDir, options = {}) {
    const files = await this.find(srcPattern)
    const results = []
    
    for (const file of files) {
      const basename = vfs.path.basename(file)
      const destPath = vfs.path.join(destDir, basename)
      await vfs.copy(file, destPath)
      results.push({ src: file, dest: destPath })
    }
    
    return results
  }
  
  async bulkDelete(pattern, options = {}) {
    const files = await this.find(pattern)
    const confirm = options.confirm !== false
    
    if (confirm && files.length > 10) {
      // Safety check for bulk deletion
      throw new Error(`Bulk delete would affect ${files.length} files. Use {confirm: false} to proceed.`)
    }
    
    for (const file of files) {
      await vfs.unlink(file)
    }
    
    return files
  }
  
  async bulkRename(pattern, replacement, options = {}) {
    const files = await this.find(pattern)
    const results = []
    
    for (const file of files) {
      const newName = file.replace(new RegExp(pattern), replacement)
      await vfs.move(file, newName)
      results.push({ old: file, new: newName })
    }
    
    return results
  }
}
```

### Session Persistence
```javascript
class VFSPersistence {
  constructor(vfs) {
    this.vfs = vfs
    this.storageKey = 'vfs_session'
    this.autoSaveInterval = null
  }
  
  async saveSession(name = 'default') {
    // Serialize VFS state to JSON
    const sessionData = {
      version: '1.0',
      timestamp: new Date().toISOString(),
      files: Object.fromEntries(this.vfs.files),
      directories: Array.from(this.vfs.directories),
      metadata: Object.fromEntries(this.vfs.metadata),
      currentDirectory: this.vfs.currentDirectory
    }
    
    // Compress session data
    const compressed = await this.compressSession(sessionData)
    
    // Store in IndexedDB for persistence
    await this.storeInIndexedDB(name, compressed)
    
    return { name, size: compressed.length, timestamp: sessionData.timestamp }
  }
  
  async loadSession(name = 'default') {
    // Load from IndexedDB
    const compressed = await this.loadFromIndexedDB(name)
    if (!compressed) {
      throw new Error(`Session '${name}' not found`)
    }
    
    // Decompress and parse
    const sessionData = await this.decompressSession(compressed)
    
    // Restore VFS state
    this.vfs.files = new Map(Object.entries(sessionData.files))
    this.vfs.directories = new Set(sessionData.directories)
    this.vfs.metadata = new Map(Object.entries(sessionData.metadata))
    this.vfs.currentDirectory = sessionData.currentDirectory
    
    return sessionData
  }
  
  async exportSession(name = 'default') {
    // Export session as downloadable ZIP
    const sessionData = await this.saveSession(name)
    const blob = new Blob([JSON.stringify(sessionData)], { 
      type: 'application/json' 
    })
    
    const filename = `vfs_session_${name}_${new Date().toISOString().slice(0, 19)}.json`
    this.downloadBlob(blob, filename)
    
    return filename
  }
  
  async importSession(file) {
    // Import session from uploaded file
    const text = await file.text()
    const sessionData = JSON.parse(text)
    
    // Validate session data structure
    this.validateSessionData(sessionData)
    
    // Restore VFS state
    await this.restoreSessionData(sessionData)
    
    return sessionData
  }
  
  enableAutoSave(interval = 30000) {
    // Automatically save session every 30 seconds
    if (this.autoSaveInterval) {
      clearInterval(this.autoSaveInterval)
    }
    
    this.autoSaveInterval = setInterval(async () => {
      try {
        await this.saveSession('autosave')
      } catch (error) {
        console.warn('Auto-save failed:', error)
      }
    }, interval)
  }
  
  disableAutoSave() {
    if (this.autoSaveInterval) {
      clearInterval(this.autoSaveInterval)
      this.autoSaveInterval = null
    }
  }
  
  async listSessions() {
    // List all saved sessions
    const db = await this.openIndexedDB()
    const transaction = db.transaction(['sessions'], 'readonly')
    const store = transaction.objectStore('sessions')
    const keys = await store.getAllKeys()
    
    const sessions = []
    for (const key of keys) {
      const session = await store.get(key)
      sessions.push({
        name: key,
        timestamp: session.timestamp,
        size: session.size
      })
    }
    
    return sessions
  }
  
  async deleteSession(name) {
    // Delete a saved session
    const db = await this.openIndexedDB()
    const transaction = db.transaction(['sessions'], 'readwrite')
    const store = transaction.objectStore('sessions')
    await store.delete(name)
  }
}
```

## Phase 7: Security & Validation

### Security Implementation
```javascript
class VFSSecurity {
  constructor() {
    this.maxPathLength = 260
    this.maxFilenameLength = 255
    this.forbiddenChars = /[<>:"|?*\x00-\x1f]/
    this.reservedNames = ['CON', 'PRN', 'AUX', 'NUL', 'COM1', 'COM2', 'COM3', 'COM4', 'COM5', 'COM6', 'COM7', 'COM8', 'COM9', 'LPT1', 'LPT2', 'LPT3', 'LPT4', 'LPT5', 'LPT6', 'LPT7', 'LPT8', 'LPT9']
  }
  
  validatePath(path) {
    // Prevent directory traversal attacks
    if (path.includes('..')) {
      throw new Error('Path traversal detected')
    }
    
    if (path.length > this.maxPathLength) {
      throw new Error(`Path too long: ${path.length} > ${this.maxPathLength}`)
    }
    
    // Normalize path separators
    const normalizedPath = path.replace(/\\/g, '/')
    
    // Check for dangerous patterns
    if (normalizedPath.match(/\/\.{2,}\//)) {
      throw new Error('Invalid path pattern detected')
    }
    
    return normalizedPath
  }
  
  checkQuota(currentSize, additionalSize) {
    // Enforce storage limits
    const totalSize = currentSize + additionalSize
    
    if (additionalSize > this.maxFileSize) {
      throw new Error(`File too large: ${additionalSize} > ${this.maxFileSize}`)
    }
    
    if (totalSize > this.maxTotalSize) {
      throw new Error(`Storage quota exceeded: ${totalSize} > ${this.maxTotalSize}`)
    }
    
    return true
  }
  
  sanitizeFilename(name) {
    // Remove dangerous characters from filenames
    let sanitized = name.replace(this.forbiddenChars, '_')
    
    // Check against reserved names
    const baseName = sanitized.split('.')[0].toUpperCase()
    if (this.reservedNames.includes(baseName)) {
      sanitized = `_${sanitized}`
    }
    
    // Ensure filename isn't too long
    if (sanitized.length > this.maxFilenameLength) {
      const ext = sanitized.split('.').pop()
      const base = sanitized.slice(0, this.maxFilenameLength - ext.length - 1)
      sanitized = `${base}.${ext}`
    }
    
    return sanitized
  }
  
  validateFileType(data, extension) {
    // Prevent type confusion attacks
    const detectedType = this.detectFileType(data)
    const expectedType = this.getExpectedType(extension)
    
    if (detectedType && expectedType && detectedType !== expectedType) {
      console.warn(`File type mismatch: expected ${expectedType}, detected ${detectedType}`)
    }
    
    // Check for executable content
    if (this.containsExecutableContent(data)) {
      throw new Error('Executable content detected')
    }
    
    return true
  }
  
  detectFileType(data) {
    // Basic file type detection by magic bytes
    const bytes = new Uint8Array(data.slice(0, 16))
    
    // PNG
    if (bytes[0] === 0x89 && bytes[1] === 0x50 && bytes[2] === 0x4E && bytes[3] === 0x47) {
      return 'png'
    }
    
    // JPEG
    if (bytes[0] === 0xFF && bytes[1] === 0xD8 && bytes[2] === 0xFF) {
      return 'jpeg'
    }
    
    // PDF
    if (bytes[0] === 0x25 && bytes[1] === 0x50 && bytes[2] === 0x44 && bytes[3] === 0x46) {
      return 'pdf'
    }
    
    // ZIP
    if (bytes[0] === 0x50 && bytes[1] === 0x4B) {
      return 'zip'
    }
    
    return null
  }
  
  containsExecutableContent(data) {
    // Basic check for executable content patterns
    const text = typeof data === 'string' ? data : new TextDecoder().decode(data)
    
    // Check for common script patterns
    const dangerousPatterns = [
      /<script/i,
      /javascript:/i,
      /vbscript:/i,
      /data:/i,
      /eval\s*\(/i,
      /function\s*\(/i,
      /setTimeout\s*\(/i,
      /setInterval\s*\(/i
    ]
    
    return dangerousPatterns.some(pattern => pattern.test(text))
  }
}
```

## Phase 8: User Experience

### File Browser UI Component
```javascript
class FileBrowser {
  constructor(vfs) {
    this.vfs = vfs
    this.currentPath = '/'
    this.selectedItems = new Set()
    this.viewMode = 'list' // 'list' or 'grid'
  }
  
  render() {
    return `
      <div class="file-browser">
        <div class="toolbar">
          <button onclick="this.navigateUp()">↑ Up</button>
          <button onclick="this.createFolder()">📁 New Folder</button>
          <button onclick="this.uploadFile()">📤 Upload</button>
          <button onclick="this.downloadSelected()">📥 Download</button>
          <button onclick="this.deleteSelected()">🗑️ Delete</button>
          <input type="text" class="path-input" value="${this.currentPath}" onchange="this.navigateTo(this.value)">
        </div>
        
        <div class="breadcrumb">
          ${this.renderBreadcrumb()}
        </div>
        
        <div class="file-list ${this.viewMode}">
          ${this.renderFileList()}
        </div>
        
        <div class="status-bar">
          ${this.renderStatusBar()}
        </div>
      </div>
    `
  }
  
  async renderFileList() {
    const items = await this.vfs.readdir(this.currentPath)
    const itemsWithStats = await Promise.all(
      items.map(async item => ({
        name: item,
        path: this.vfs.path.join(this.currentPath, item),
        stat: await this.vfs.stat(this.vfs.path.join(this.currentPath, item))
      }))
    )
    
    return itemsWithStats.map(item => `
      <div class="file-item ${item.stat.isDirectory ? 'directory' : 'file'}" 
           onclick="this.selectItem('${item.path}')"
           ondblclick="this.openItem('${item.path}')">
        <div class="icon">${item.stat.isDirectory ? '📁' : this.getFileIcon(item.name)}</div>
        <div class="name">${item.name}</div>
        <div class="size">${item.stat.isDirectory ? '' : this.formatSize(item.stat.size)}</div>
        <div class="modified">${this.formatDate(item.stat.mtime)}</div>
      </div>
    `).join('')
  }
  
  getFileIcon(filename) {
    const ext = filename.split('.').pop().toLowerCase()
    const iconMap = {
      'txt': '📄', 'md': '📝', 'json': '📋', 'csv': '📊',
      'js': '📜', 'py': '🐍', 'html': '🌐', 'css': '🎨',
      'png': '🖼️', 'jpg': '🖼️', 'gif': '🖼️', 'pdf': '📕',
      'zip': '📦', 'tar': '📦', 'xlsx': '📊', 'sql': '🗃️'
    }
    return iconMap[ext] || '📄'
  }
  
  async uploadFile() {
    const input = document.createElement('input')
    input.type = 'file'
    input.multiple = true
    input.onchange = async (e) => {
      const files = Array.from(e.target.files)
      for (const file of files) {
        const path = this.vfs.path.join(this.currentPath, file.name)
        await this.vfs.importFromFile(file, path)
      }
      this.refresh()
    }
    input.click()
  }
  
  async downloadSelected() {
    if (this.selectedItems.size === 0) return
    
    if (this.selectedItems.size === 1) {
      const path = Array.from(this.selectedItems)[0]
      const filename = this.vfs.path.basename(path)
      await this.vfs.exportToDownload(path, filename)
    } else {
      // Multiple files - create ZIP
      const paths = Array.from(this.selectedItems)
      const zipName = `files_${new Date().toISOString().slice(0, 10)}.zip`
      await this.vfs.exportToZip(paths, zipName)
    }
  }
  
  async createFolder() {
    const name = prompt('Enter folder name:')
    if (name) {
      const path = this.vfs.path.join(this.currentPath, name)
      await this.vfs.mkdir(path)
      this.refresh()
    }
  }
  
  async deleteSelected() {
    if (this.selectedItems.size === 0) return
    
    const confirmed = confirm(`Delete ${this.selectedItems.size} item(s)?`)
    if (confirmed) {
      for (const path of this.selectedItems) {
        const stat = await this.vfs.stat(path)
        if (stat.isDirectory) {
          await this.vfs.rmdir(path, true)
        } else {
          await this.vfs.unlink(path)
        }
      }
      this.selectedItems.clear()
      this.refresh()
    }
  }
}
```

### Code Examples for Users

#### Python Examples
```python
# Basic file operations
with open('/data/sales.csv', 'w') as f:
    f.write('name,amount,date\n')
    f.write('John,100,2024-01-01\n')
    f.write('Jane,200,2024-01-02\n')

# Read and process data
import pandas as pd
df = pd.read_csv('/data/sales.csv')
print(df.describe())

# Save processed data
df.to_json('/data/sales_summary.json')
df.to_parquet('/data/sales.parquet')

# Work with images
from PIL import Image
import matplotlib.pyplot as plt

# Create and save a plot
plt.figure(figsize=(10, 6))
plt.bar(df['name'], df['amount'])
plt.title('Sales by Person')
plt.savefig('/data/sales_chart.png')

# File system operations
import os
os.makedirs('/data/processed', exist_ok=True)
os.listdir('/data')

# Archive operations
import zipfile
with zipfile.ZipFile('/data/backup.zip', 'w') as zf:
    zf.write('/data/sales.csv', 'sales.csv')
    zf.write('/data/sales_chart.png', 'chart.png')
```

#### JavaScript Examples
```javascript
// Node.js-style file operations
const fs = require('fs')

// Write JSON data
const data = { users: [{ name: 'John', age: 30 }, { name: 'Jane', age: 25 }] }
fs.writeFileSync('/data/users.json', JSON.stringify(data, null, 2))

// Read and process
const userData = JSON.parse(fs.readFileSync('/data/users.json', 'utf8'))
const avgAge = userData.users.reduce((sum, user) => sum + user.age, 0) / userData.users.length

// CSV processing
const csvContent = fs.readFileSync('/data/sales.csv', 'utf8')
const rows = csvContent.split('\n').map(row => row.split(','))
const headers = rows[0]
const records = rows.slice(1).map(row => {
  const record = {}
  headers.forEach((header, i) => record[header] = row[i])
  return record
})

// File system utilities
const path = require('path')
const files = fs.readdirSync('/data')
const csvFiles = files.filter(file => path.extname(file) === '.csv')

// Async operations with Promises
async function processFiles() {
  const files = await fs.readdir('/data')
  
  for (const file of files) {
    if (file.endsWith('.json')) {
      const content = await fs.readFile(`/data/${file}`, 'utf8')
      const data = JSON.parse(content)
      console.log(`${file}: ${Object.keys(data).length} properties`)
    }
  }
}

// Stream processing for large files
const readStream = fs.createReadStream('/data/large_file.txt')
const writeStream = fs.createWriteStream('/data/processed_file.txt')

readStream.on('data', chunk => {
  const processed = chunk.toString().toUpperCase()
  writeStream.write(processed)
})
```

## Implementation Schedule

### Week 1-2: Foundation
- Core VFS class implementation
- Path resolver and security validator
- Basic file operations (read, write, mkdir, etc.)
- Unit tests for core functionality

### Week 3-4: Python Integration
- VFS file handlers for Python
- Built-in function overrides (open, os module)
- Pandas/NumPy integration
- Python-specific testing

### Week 5-6: JavaScript Integration
- Node.js-style fs module
- Stream implementations
- Library compatibility testing
- JavaScript-specific testing

### Week 7-8: Data Format Support
- File type handlers (CSV, JSON, Excel, etc.)
- Import/export functionality
- Compression and decompression
- Format conversion utilities

### Week 9-10: Advanced Features
- File system utilities (search, batch operations)
- Session persistence
- Performance optimization
- Security hardening

### Week 11-12: User Experience
- File browser UI component
- Documentation and examples
- Integration testing
- Performance benchmarking

## Success Metrics

- **Security**: Zero successful path traversal or code injection attacks
- **Performance**: File operations complete within 100ms for files < 1MB
- **Compatibility**: 95% compatibility with standard file API usage patterns
- **Reliability**: 99.9% uptime for core file operations
- **User Adoption**: Used in 50%+ of Python/JavaScript code executions

## Risk Mitigation

- **Memory Limits**: Strict quota enforcement and garbage collection
- **Security Vulnerabilities**: Regular security audits and penetration testing
- **Performance Issues**: Profiling and optimization at each phase
- **Compatibility Problems**: Extensive testing with real-world code examples
- **User Experience**: Regular user feedback and usability testing

This VFS implementation will provide a secure, performant, and user-friendly file system for code execution environments, enabling rich data manipulation workflows while maintaining isolation from the host system.