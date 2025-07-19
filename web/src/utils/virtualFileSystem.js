/**
 * Virtual File System (VFS) Implementation
 * Provides secure, isolated file system operations for code runners
 */

class VirtualFileSystem {
  constructor(options = {}) {
    // Core storage
    this.files = new Map()           // file path -> file data
    this.directories = new Set(['/']).add('/') // directory paths
    this.metadata = new Map()        // file metadata (size, modified, etc.)
    this.currentDirectory = '/'      // current working directory
    
    // Resource limits
    this.maxFileSize = options.maxFileSize || 10 * 1024 * 1024  // 10MB per file
    this.maxTotalSize = options.maxTotalSize || 100 * 1024 * 1024 // 100MB total
    this.maxFiles = options.maxFiles || 1000  // Maximum number of files
    
    // Components
    this.pathResolver = new PathResolver()
    this.security = new VFSSecurity(this)
    this.utils = new FileSystemUtils(this)
    this.persistence = new VFSPersistence(this)
    
    // Import/Export will be initialized externally
    this.importExport = null
    
    // Initialize root directory
    this.directories.add('/')
    this._updateDirectoryMetadata('/')
  }

  // ============ CORE FILE OPERATIONS ============

  async writeFile(filePath, data, options = {}) {
    try {
      const normalizedPath = this.pathResolver.normalize(filePath)
      this.security.validatePath(normalizedPath)
      
      // Convert data to appropriate format
      let fileData, size
      if (options.binary || data instanceof ArrayBuffer || data instanceof Uint8Array) {
        // Handle binary data
        if (data instanceof ArrayBuffer) {
          fileData = new Uint8Array(data)
        } else if (data instanceof Uint8Array) {
          fileData = data
        } else {
          fileData = new TextEncoder().encode(data)
        }
        size = fileData.byteLength
      } else {
        // Handle text data
        fileData = String(data)
        size = new TextEncoder().encode(fileData).byteLength
      }
      
      // Check quotas
      const currentSize = this._getTotalSize()
      this.security.checkQuota(currentSize, size)
      
      // Ensure parent directory exists
      const parentDir = this.pathResolver.dirname(normalizedPath)
      if (!this.directories.has(parentDir)) {
        await this.mkdir(parentDir, { recursive: true })
      }
      
      // Store file
      this.files.set(normalizedPath, fileData)
      this._updateFileMetadata(normalizedPath, {
        size: size,
        type: options.binary ? 'binary' : 'text',
        encoding: options.encoding || 'utf8',
        mtime: new Date(),
        created: this.metadata.get(normalizedPath)?.created || new Date()
      })
      
      return normalizedPath
    } catch (error) {
      throw new Error(`Failed to write file ${filePath}: ${error.message}`)
    }
  }

  async readFile(filePath, encoding = 'utf8') {
    try {
      const normalizedPath = this.pathResolver.normalize(filePath)
      this.security.validatePath(normalizedPath)
      
      if (!this.files.has(normalizedPath)) {
        throw new Error(`File not found: ${filePath}`)
      }
      
      const fileData = this.files.get(normalizedPath)
      const metadata = this.metadata.get(normalizedPath)
      
      // Return data in requested format
      if (encoding === 'binary' || encoding === null) {
        return fileData
      } else if (metadata?.type === 'binary' && fileData instanceof Uint8Array) {
        return new TextDecoder(encoding).decode(fileData)
      } else {
        return String(fileData)
      }
    } catch (error) {
      throw new Error(`Failed to read file ${filePath}: ${error.message}`)
    }
  }

  async mkdir(dirPath, options = {}) {
    try {
      const normalizedPath = this.pathResolver.normalize(dirPath)
      this.security.validatePath(normalizedPath)
      
      if (this.directories.has(normalizedPath)) {
        if (!options.recursive) {
          throw new Error(`Directory already exists: ${dirPath}`)
        }
        return normalizedPath
      }
      
      // Create parent directories if recursive
      if (options.recursive) {
        const parts = normalizedPath.split('/').filter(Boolean)
        let currentPath = '/'
        
        for (const part of parts) {
          currentPath = this.pathResolver.join(currentPath, part)
          if (!this.directories.has(currentPath)) {
            this.directories.add(currentPath)
            this._updateDirectoryMetadata(currentPath)
          }
        }
      } else {
        // Check parent exists
        const parentDir = this.pathResolver.dirname(normalizedPath)
        if (!this.directories.has(parentDir)) {
          throw new Error(`Parent directory does not exist: ${parentDir}`)
        }
        
        this.directories.add(normalizedPath)
        this._updateDirectoryMetadata(normalizedPath)
      }
      
      return normalizedPath
    } catch (error) {
      throw new Error(`Failed to create directory ${dirPath}: ${error.message}`)
    }
  }

  async readdir(dirPath = this.currentDirectory) {
    try {
      const normalizedPath = this.pathResolver.normalize(dirPath)
      this.security.validatePath(normalizedPath)
      
      if (!this.directories.has(normalizedPath)) {
        throw new Error(`Directory not found: ${dirPath}`)
      }
      
      const items = []
      
      // Find immediate children (files and directories)
      const pathPrefix = normalizedPath === '/' ? '/' : normalizedPath + '/'
      
      // Add child directories
      for (const dir of this.directories) {
        if (dir !== normalizedPath && dir.startsWith(pathPrefix)) {
          const relativePath = dir.slice(pathPrefix.length)
          if (!relativePath.includes('/')) { // Immediate child only
            items.push(relativePath)
          }
        }
      }
      
      // Add files
      for (const file of this.files.keys()) {
        if (file.startsWith(pathPrefix)) {
          const relativePath = file.slice(pathPrefix.length)
          if (!relativePath.includes('/')) { // Immediate child only
            items.push(relativePath)
          }
        }
      }
      
      return items.sort()
    } catch (error) {
      throw new Error(`Failed to read directory ${dirPath}: ${error.message}`)
    }
  }

  async stat(itemPath) {
    try {
      const normalizedPath = this.pathResolver.normalize(itemPath)
      this.security.validatePath(normalizedPath)
      
      // Check if it's a directory
      if (this.directories.has(normalizedPath)) {
        const metadata = this.metadata.get(normalizedPath) || {}
        return {
          isDirectory: true,
          isFile: false,
          size: 0,
          mtime: metadata.mtime || new Date(),
          created: metadata.created || new Date(),
          path: normalizedPath
        }
      }
      
      // Check if it's a file
      if (this.files.has(normalizedPath)) {
        const metadata = this.metadata.get(normalizedPath) || {}
        return {
          isDirectory: false,
          isFile: true,
          size: metadata.size || 0,
          type: metadata.type || 'text',
          encoding: metadata.encoding || 'utf8',
          mtime: metadata.mtime || new Date(),
          created: metadata.created || new Date(),
          path: normalizedPath
        }
      }
      
      throw new Error(`Path not found: ${itemPath}`)
    } catch (error) {
      throw new Error(`Failed to stat ${itemPath}: ${error.message}`)
    }
  }

  async exists(itemPath) {
    try {
      const normalizedPath = this.pathResolver.normalize(itemPath)
      this.security.validatePath(normalizedPath)
      return this.directories.has(normalizedPath) || this.files.has(normalizedPath)
    } catch (error) {
      return false
    }
  }

  async unlink(filePath) {
    try {
      const normalizedPath = this.pathResolver.normalize(filePath)
      this.security.validatePath(normalizedPath)
      
      if (!this.files.has(normalizedPath)) {
        throw new Error(`File not found: ${filePath}`)
      }
      
      this.files.delete(normalizedPath)
      this.metadata.delete(normalizedPath)
      
      return normalizedPath
    } catch (error) {
      throw new Error(`Failed to delete file ${filePath}: ${error.message}`)
    }
  }

  async rmdir(dirPath, options = {}) {
    try {
      const normalizedPath = this.pathResolver.normalize(dirPath)
      this.security.validatePath(normalizedPath)
      
      if (normalizedPath === '/') {
        throw new Error('Cannot delete root directory')
      }
      
      if (!this.directories.has(normalizedPath)) {
        throw new Error(`Directory not found: ${dirPath}`)
      }
      
      // Check if directory is empty (unless recursive)
      if (!options.recursive) {
        const items = await this.readdir(normalizedPath)
        if (items.length > 0) {
          throw new Error(`Directory not empty: ${dirPath}`)
        }
      } else {
        // Recursively delete contents
        const items = await this.readdir(normalizedPath)
        for (const item of items) {
          const itemPath = this.pathResolver.join(normalizedPath, item)
          const itemStat = await this.stat(itemPath)
          
          if (itemStat.isDirectory) {
            await this.rmdir(itemPath, { recursive: true })
          } else {
            await this.unlink(itemPath)
          }
        }
      }
      
      this.directories.delete(normalizedPath)
      this.metadata.delete(normalizedPath)
      
      return normalizedPath
    } catch (error) {
      throw new Error(`Failed to remove directory ${dirPath}: ${error.message}`)
    }
  }

  // ============ NAVIGATION OPERATIONS ============

  chdir(dirPath) {
    const normalizedPath = this.pathResolver.normalize(dirPath)
    this.security.validatePath(normalizedPath)
    
    if (!this.directories.has(normalizedPath)) {
      throw new Error(`Directory not found: ${dirPath}`)
    }
    
    this.currentDirectory = normalizedPath
    return normalizedPath
  }

  getcwd() {
    return this.currentDirectory
  }

  // ============ ADVANCED OPERATIONS ============

  async copy(srcPath, destPath) {
    try {
      const srcNormalized = this.pathResolver.normalize(srcPath)
      const destNormalized = this.pathResolver.normalize(destPath)
      
      this.security.validatePath(srcNormalized)
      this.security.validatePath(destNormalized)
      
      const srcStat = await this.stat(srcNormalized)
      
      if (srcStat.isFile) {
        const data = await this.readFile(srcNormalized, 'binary')
        const metadata = this.metadata.get(srcNormalized)
        await this.writeFile(destNormalized, data, { 
          binary: metadata?.type === 'binary',
          encoding: metadata?.encoding 
        })
      } else if (srcStat.isDirectory) {
        await this.mkdir(destNormalized, { recursive: true })
        const items = await this.readdir(srcNormalized)
        
        for (const item of items) {
          const srcItem = this.pathResolver.join(srcNormalized, item)
          const destItem = this.pathResolver.join(destNormalized, item)
          await this.copy(srcItem, destItem)
        }
      }
      
      return destNormalized
    } catch (error) {
      throw new Error(`Failed to copy ${srcPath} to ${destPath}: ${error.message}`)
    }
  }

  async move(srcPath, destPath) {
    try {
      await this.copy(srcPath, destPath)
      
      const srcStat = await this.stat(srcPath)
      if (srcStat.isDirectory) {
        await this.rmdir(srcPath, { recursive: true })
      } else {
        await this.unlink(srcPath)
      }
      
      return destPath
    } catch (error) {
      throw new Error(`Failed to move ${srcPath} to ${destPath}: ${error.message}`)
    }
  }

  async glob(pattern, options = {}) {
    const matches = []
    const regex = this._globToRegex(pattern)
    
    // Search files
    for (const filePath of this.files.keys()) {
      if (regex.test(filePath)) {
        matches.push(filePath)
      }
    }
    
    // Search directories if requested
    if (options.includeDirectories !== false) {
      for (const dirPath of this.directories) {
        if (dirPath !== '/' && regex.test(dirPath)) {
          matches.push(dirPath)
        }
      }
    }
    
    return matches.sort()
  }

  // ============ UTILITY METHODS ============

  getStorageInfo() {
    const totalSize = this._getTotalSize()
    const fileCount = this.files.size
    const dirCount = this.directories.size
    
    return {
      totalSize,
      fileCount,
      dirCount,
      maxFileSize: this.maxFileSize,
      maxTotalSize: this.maxTotalSize,
      maxFiles: this.maxFiles,
      usage: {
        size: (totalSize / this.maxTotalSize * 100).toFixed(1) + '%',
        files: (fileCount / this.maxFiles * 100).toFixed(1) + '%'
      }
    }
  }

  clear() {
    this.files.clear()
    this.directories.clear()
    this.metadata.clear()
    this.currentDirectory = '/'
    this.directories.add('/')
    this._updateDirectoryMetadata('/')
  }

  // ============ PRIVATE METHODS ============

  _getTotalSize() {
    let total = 0
    for (const metadata of this.metadata.values()) {
      total += metadata.size || 0
    }
    return total
  }

  _updateFileMetadata(path, metadata) {
    const existing = this.metadata.get(path) || {}
    this.metadata.set(path, { ...existing, ...metadata })
  }

  _updateDirectoryMetadata(path) {
    const existing = this.metadata.get(path) || {}
    this.metadata.set(path, {
      ...existing,
      size: 0,
      mtime: new Date(),
      created: existing.created || new Date()
    })
  }

  _globToRegex(pattern) {
    // Convert glob pattern to regex
    let regex = pattern
      .replace(/\./g, '\\.')
      .replace(/\*/g, '[^/]*')
      .replace(/\*\*/g, '.*')
      .replace(/\?/g, '[^/]')
    
    return new RegExp(`^${regex}$`)
  }
}

// ============ PATH RESOLVER ============

class PathResolver {
  normalize(path) {
    if (!path) return '/'
    
    // Convert to string and handle basic cases
    path = String(path).replace(/\\/g, '/')
    
    if (!path.startsWith('/')) {
      path = '/' + path
    }
    
    // Split into parts and resolve . and ..
    const parts = path.split('/').filter(Boolean)
    const resolved = []
    
    for (const part of parts) {
      if (part === '.') {
        continue
      } else if (part === '..') {
        resolved.pop()
      } else {
        resolved.push(part)
      }
    }
    
    return '/' + resolved.join('/')
  }

  resolve(path) {
    return this.normalize(path)
  }

  dirname(path) {
    const normalized = this.normalize(path)
    if (normalized === '/') return '/'
    
    const lastSlash = normalized.lastIndexOf('/')
    if (lastSlash === 0) return '/'
    
    return normalized.slice(0, lastSlash)
  }

  basename(path) {
    const normalized = this.normalize(path)
    if (normalized === '/') return '/'
    
    const lastSlash = normalized.lastIndexOf('/')
    return normalized.slice(lastSlash + 1)
  }

  extname(path) {
    const base = this.basename(path)
    const lastDot = base.lastIndexOf('.')
    
    if (lastDot === -1 || lastDot === 0) return ''
    return base.slice(lastDot)
  }

  join(...parts) {
    const joined = parts.join('/')
    return this.normalize(joined)
  }

  split(path) {
    const normalized = this.normalize(path)
    if (normalized === '/') return ['/']
    
    return normalized.split('/').filter(Boolean)
  }

  validate(path) {
    // Basic validation - detailed validation in VFSSecurity
    if (typeof path !== 'string') {
      throw new Error('Path must be a string')
    }
    
    if (path.length > 260) {
      throw new Error('Path too long')
    }
    
    return true
  }
}

// ============ SECURITY CLASS ============

class VFSSecurity {
  constructor(vfs) {
    this.vfs = vfs
    this.maxPathLength = 260
    this.maxFilenameLength = 255
    this.forbiddenChars = /[<>:"|?*\x00-\x1f]/
    this.reservedNames = new Set([
      'CON', 'PRN', 'AUX', 'NUL', 'COM1', 'COM2', 'COM3', 'COM4', 'COM5', 
      'COM6', 'COM7', 'COM8', 'COM9', 'LPT1', 'LPT2', 'LPT3', 'LPT4', 
      'LPT5', 'LPT6', 'LPT7', 'LPT8', 'LPT9'
    ])
  }

  validatePath(path) {
    // Prevent directory traversal
    if (path.includes('..')) {
      throw new Error('Path traversal detected')
    }
    
    if (path.length > this.maxPathLength) {
      throw new Error(`Path too long: ${path.length} > ${this.maxPathLength}`)
    }
    
    // Check for dangerous patterns
    if (path.match(/\/\.{2,}\//)) {
      throw new Error('Invalid path pattern detected')
    }
    
    return true
  }

  checkQuota(currentSize, additionalSize) {
    if (additionalSize > this.vfs.maxFileSize) {
      throw new Error(`File too large: ${this.formatSize(additionalSize)} > ${this.formatSize(this.vfs.maxFileSize)}`)
    }
    
    const totalSize = currentSize + additionalSize
    if (totalSize > this.vfs.maxTotalSize) {
      throw new Error(`Storage quota exceeded: ${this.formatSize(totalSize)} > ${this.formatSize(this.vfs.maxTotalSize)}`)
    }
    
    if (this.vfs.files.size >= this.vfs.maxFiles) {
      throw new Error(`File count limit exceeded: ${this.vfs.files.size} >= ${this.vfs.maxFiles}`)
    }
    
    return true
  }

  sanitizeFilename(name) {
    let sanitized = name.replace(this.forbiddenChars, '_')
    
    const baseName = sanitized.split('.')[0].toUpperCase()
    if (this.reservedNames.has(baseName)) {
      sanitized = `_${sanitized}`
    }
    
    if (sanitized.length > this.maxFilenameLength) {
      const ext = this.vfs.pathResolver.extname(sanitized)
      const maxBase = this.maxFilenameLength - ext.length
      const base = sanitized.slice(0, maxBase)
      sanitized = base + ext
    }
    
    return sanitized
  }

  formatSize(bytes) {
    const units = ['B', 'KB', 'MB', 'GB']
    let size = bytes
    let unitIndex = 0
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }
    
    return `${size.toFixed(1)}${units[unitIndex]}`
  }
}

// ============ FILE SYSTEM UTILITIES ============

class FileSystemUtils {
  constructor(vfs) {
    this.vfs = vfs
  }

  async find(pattern, options = {}) {
    return await this.vfs.glob(pattern, options)
  }

  async grep(pattern, files, options = {}) {
    const results = []
    const regex = new RegExp(pattern, options.flags || 'gi')
    
    for (const file of files) {
      try {
        const content = await this.vfs.readFile(file, 'utf8')
        const matches = [...content.matchAll(regex)]
        
        if (matches.length > 0) {
          results.push({
            file,
            matches: matches.map(match => ({
              text: match[0],
              index: match.index,
              groups: match.slice(1)
            }))
          })
        }
      } catch (error) {
        // Skip files that can't be read as text
        continue
      }
    }
    
    return results
  }

  async bulkCopy(srcPattern, destDir, options = {}) {
    const files = await this.find(srcPattern)
    const results = []
    
    await this.vfs.mkdir(destDir, { recursive: true })
    
    for (const file of files) {
      const basename = this.vfs.pathResolver.basename(file)
      const destPath = this.vfs.pathResolver.join(destDir, basename)
      
      try {
        await this.vfs.copy(file, destPath)
        results.push({ src: file, dest: destPath, success: true })
      } catch (error) {
        results.push({ src: file, dest: destPath, success: false, error: error.message })
      }
    }
    
    return results
  }

  async bulkDelete(pattern, options = {}) {
    const files = await this.find(pattern)
    
    if (!options.force && files.length > 10) {
      throw new Error(`Bulk delete would affect ${files.length} files. Use {force: true} to proceed.`)
    }
    
    const results = []
    
    for (const file of files) {
      try {
        const stat = await this.vfs.stat(file)
        if (stat.isDirectory) {
          await this.vfs.rmdir(file, { recursive: true })
        } else {
          await this.vfs.unlink(file)
        }
        results.push({ path: file, success: true })
      } catch (error) {
        results.push({ path: file, success: false, error: error.message })
      }
    }
    
    return results
  }
}

// ============ PERSISTENCE CLASS ============

class VFSPersistence {
  constructor(vfs) {
    this.vfs = vfs
    this.storageKey = 'vfs_session'
  }

  async saveSession(name = 'default') {
    const sessionData = {
      version: '1.0',
      timestamp: new Date().toISOString(),
      files: Object.fromEntries(this.vfs.files),
      directories: Array.from(this.vfs.directories),
      metadata: Object.fromEntries(this.vfs.metadata),
      currentDirectory: this.vfs.currentDirectory
    }
    
    try {
      const jsonData = JSON.stringify(sessionData)
      localStorage.setItem(`${this.storageKey}_${name}`, jsonData)
      
      return {
        name,
        size: jsonData.length,
        timestamp: sessionData.timestamp
      }
    } catch (error) {
      throw new Error(`Failed to save session: ${error.message}`)
    }
  }

  async loadSession(name = 'default') {
    try {
      const jsonData = localStorage.getItem(`${this.storageKey}_${name}`)
      if (!jsonData) {
        throw new Error(`Session '${name}' not found`)
      }
      
      const sessionData = JSON.parse(jsonData)
      
      // Restore VFS state
      this.vfs.files = new Map(Object.entries(sessionData.files))
      this.vfs.directories = new Set(sessionData.directories)
      this.vfs.metadata = new Map(Object.entries(sessionData.metadata))
      this.vfs.currentDirectory = sessionData.currentDirectory
      
      return sessionData
    } catch (error) {
      throw new Error(`Failed to load session: ${error.message}`)
    }
  }

  listSessions() {
    const sessions = []
    
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key.startsWith(this.storageKey + '_')) {
        const name = key.slice((this.storageKey + '_').length)
        const data = localStorage.getItem(key)
        
        try {
          const sessionData = JSON.parse(data)
          sessions.push({
            name,
            timestamp: sessionData.timestamp,
            size: data.length
          })
        } catch (error) {
          // Skip corrupted sessions
          continue
        }
      }
    }
    
    return sessions.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
  }

  deleteSession(name) {
    const key = `${this.storageKey}_${name}`
    if (localStorage.getItem(key)) {
      localStorage.removeItem(key)
      return true
    }
    return false
  }
}

// Export the VFS class
export default VirtualFileSystem
export { VirtualFileSystem }