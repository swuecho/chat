/**
 * VFS Data Import/Export System
 * Handles file uploads, downloads, and data format conversions
 */

class VFSImportExport {
  constructor(vfs) {
    this.vfs = vfs
    this.supportedFormats = {
      'text': ['.txt', '.md', '.csv', '.json', '.xml', '.yaml', '.yml', '.log'],
      'data': ['.csv', '.json', '.xlsx', '.tsv'],
      'code': ['.js', '.py', '.html', '.css', '.sql'],
      'binary': ['.png', '.jpg', '.jpeg', '.gif', '.pdf', '.zip', '.tar']
    }
  }

  // ============ FILE UPLOAD FUNCTIONALITY ============

  async uploadFile(file, targetPath = null) {
    try {
      // Validate file object
      if (!file || !(file instanceof File) && !(file instanceof Blob)) {
        throw new Error(`Invalid file object: expected File or Blob, got ${typeof file}`)
      }

      // Generate target path if not provided
      if (!targetPath) {
        targetPath = `/data/${this.sanitizeFilename(file.name)}`
      }

      // Detect file type and handle appropriately
      const fileExtension = this.getFileExtension(file.name)
      const isTextFile = this.isTextFile(fileExtension)

      let fileData
      if (isTextFile) {
        fileData = await this.readFileAsText(file)
      } else {
        fileData = await this.readFileAsBinary(file)
      }

      // Store in VFS
      await this.vfs.writeFile(targetPath, fileData, {
        binary: !isTextFile,
        originalName: file.name,
        size: file.size,
        type: file.type,
        lastModified: new Date(file.lastModified)
      })

      return {
        success: true,
        path: targetPath,
        size: file.size,
        type: isTextFile ? 'text' : 'binary',
        message: `File uploaded successfully to ${targetPath}`
      }
    } catch (error) {
      return {
        success: false,
        error: error.message,
        message: `Failed to upload ${file.name}: ${error.message}`
      }
    }
  }

  async uploadMultipleFiles(files, targetDirectory = '/data') {
    const results = []

    // Ensure target directory exists
    await this.vfs.mkdir(targetDirectory, { recursive: true })

    for (const file of files) {
      const targetPath = `${targetDirectory}/${this.sanitizeFilename(file.name)}`
      const result = await this.uploadFile(file, targetPath)
      results.push({
        filename: file.name,
        ...result
      })
    }

    return results
  }

  // ============ FILE DOWNLOAD FUNCTIONALITY ============

  async downloadFile(vfsPath, downloadName = null) {
    try {
      if (!await this.vfs.exists(vfsPath)) {
        throw new Error(`File not found: ${vfsPath}`)
      }

      const stat = await this.vfs.stat(vfsPath)
      if (stat.isDirectory) {
        throw new Error(`Cannot download directory: ${vfsPath}`)
      }

      const fileData = await this.vfs.readFile(vfsPath, 'binary')
      const filename = downloadName || this.vfs.pathResolver.basename(vfsPath)

      this.triggerDownload(fileData, filename, this.getMimeType(filename))

      return {
        success: true,
        filename: filename,
        size: stat.size,
        message: `Downloaded ${filename} successfully`
      }
    } catch (error) {
      return {
        success: false,
        error: error.message,
        message: `Failed to download ${vfsPath}: ${error.message}`
      }
    }
  }

  async downloadDirectory(vfsPath, zipName = null) {
    try {
      if (!await this.vfs.exists(vfsPath)) {
        throw new Error(`Directory not found: ${vfsPath}`)
      }

      const stat = await this.vfs.stat(vfsPath)
      if (!stat.isDirectory) {
        throw new Error(`Path is not a directory: ${vfsPath}`)
      }

      // Collect all files in directory recursively
      const files = await this.collectDirectoryFiles(vfsPath)
      const zipFilename = zipName || `${this.vfs.pathResolver.basename(vfsPath)}.zip`

      // Create ZIP file
      const zipBlob = await this.createZipFromFiles(files, vfsPath)
      this.triggerDownload(zipBlob, zipFilename, 'application/zip')

      return {
        success: true,
        filename: zipFilename,
        fileCount: files.length,
        message: `Downloaded ${files.length} files as ${zipFilename}`
      }
    } catch (error) {
      return {
        success: false,
        error: error.message,
        message: `Failed to download directory ${vfsPath}: ${error.message}`
      }
    }
  }

  // ============ DATA FORMAT CONVERSION ============

  async convertToFormat(vfsPath, targetFormat, outputPath = null) {
    try {
      const sourceData = await this.vfs.readFile(vfsPath, 'utf8')
      const sourceExtension = this.getFileExtension(vfsPath)

      if (!outputPath) {
        const baseName = this.vfs.pathResolver.basename(vfsPath).replace(/\.[^.]+$/, '')
        outputPath = `${this.vfs.pathResolver.dirname(vfsPath)}/${baseName}.${targetFormat}`
      }

      let convertedData
      switch (targetFormat.toLowerCase()) {
        case 'json':
          convertedData = await this.convertToJSON(sourceData, sourceExtension)
          break
        case 'csv':
          convertedData = await this.convertToCSV(sourceData, sourceExtension)
          break
        case 'xlsx':
          convertedData = await this.convertToExcel(sourceData, sourceExtension)
          break
        case 'xml':
          convertedData = await this.convertToXML(sourceData, sourceExtension)
          break
        default:
          throw new Error(`Unsupported target format: ${targetFormat}`)
      }

      await this.vfs.writeFile(outputPath, convertedData)

      return {
        success: true,
        inputPath: vfsPath,
        outputPath: outputPath,
        format: targetFormat,
        message: `Converted ${vfsPath} to ${targetFormat} format`
      }
    } catch (error) {
      return {
        success: false,
        error: error.message,
        message: `Failed to convert ${vfsPath}: ${error.message}`
      }
    }
  }

  // ============ BULK OPERATIONS ============

  async importFromURL(url, targetPath, options = {}) {
    try {
      const response = await fetch(url)
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const filename = targetPath || this.extractFilenameFromURL(url)
      const contentType = response.headers.get('content-type') || ''

      let data
      if (contentType.includes('text') || contentType.includes('json') || contentType.includes('xml')) {
        data = await response.text()
      } else {
        data = await response.arrayBuffer()
      }

      await this.vfs.writeFile(filename, data, {
        binary: !(contentType.includes('text') || contentType.includes('json')),
        source: 'url',
        originalURL: url
      })

      return {
        success: true,
        path: filename,
        url: url,
        size: data.length || data.byteLength,
        message: `Imported from ${url} to ${filename}`
      }
    } catch (error) {
      return {
        success: false,
        error: error.message,
        message: `Failed to import from ${url}: ${error.message}`
      }
    }
  }

  async exportVFSSession(sessionName = 'vfs-session') {
    try {
      const sessionData = {
        metadata: {
          exportDate: new Date().toISOString(),
          sessionName: sessionName,
          version: '1.0'
        },
        files: {},
        directories: Array.from(this.vfs.directories)
      }

      // Export all files
      for (const [path, _] of this.vfs.files) {
        try {
          const data = await this.vfs.readFile(path, 'binary')
          const metadata = this.vfs.metadata.get(path)

          sessionData.files[path] = {
            data: this.arrayBufferToBase64(data),
            metadata: metadata,
            encoding: 'base64'
          }
        } catch (error) {
          console.warn(`Failed to export file ${path}:`, error)
        }
      }

      const sessionJSON = JSON.stringify(sessionData, null, 2)
      const filename = `${sessionName}-${new Date().toISOString().slice(0, 10)}.vfs.json`

      this.triggerDownload(sessionJSON, filename, 'application/json')

      return {
        success: true,
        filename: filename,
        fileCount: Object.keys(sessionData.files).length,
        message: `Exported VFS session as ${filename}`
      }
    } catch (error) {
      return {
        success: false,
        error: error.message,
        message: `Failed to export VFS session: ${error.message}`
      }
    }
  }

  async importVFSSession(file) {
    try {
      const sessionText = await this.readFileAsText(file)
      const sessionData = JSON.parse(sessionText)

      if (!sessionData.metadata || !sessionData.files) {
        throw new Error('Invalid VFS session file format')
      }

      // Clear existing VFS (optional - could ask user)
      this.vfs.clear()

      // Restore directories
      if (sessionData.directories) {
        for (const dir of sessionData.directories) {
          this.vfs.directories.add(dir)
        }
      }

      // Restore files
      let importedCount = 0
      for (const [path, fileInfo] of Object.entries(sessionData.files)) {
        try {
          let data
          if (fileInfo.encoding === 'base64') {
            data = this.base64ToArrayBuffer(fileInfo.data)
          } else {
            data = fileInfo.data
          }

          await this.vfs.writeFile(path, data, {
            binary: fileInfo.metadata?.type === 'binary',
            ...fileInfo.metadata
          })
          importedCount++
        } catch (error) {
          console.warn(`Failed to import file ${path}:`, error)
        }
      }

      return {
        success: true,
        importedFiles: importedCount,
        sessionName: sessionData.metadata.sessionName,
        exportDate: sessionData.metadata.exportDate,
        message: `Imported ${importedCount} files from VFS session`
      }
    } catch (error) {
      return {
        success: false,
        error: error.message,
        message: `Failed to import VFS session: ${error.message}`
      }
    }
  }

  // ============ HELPER METHODS ============

  async readFileAsText(file) {
    return new Promise((resolve, reject) => {
      // Validate that file is a Blob/File object
      if (!(file instanceof Blob) && !(file instanceof File)) {
        reject(new Error(`Expected File or Blob, but received: ${typeof file} - ${file?.constructor?.name || 'unknown'}`))
        return
      }

      const reader = new FileReader()
      reader.onload = e => resolve(e.target.result)
      reader.onerror = e => reject(new Error('Failed to read file as text'))
      reader.readAsText(file)
    })
  }

  async readFileAsBinary(file) {
    return new Promise((resolve, reject) => {
      // Validate that file is a Blob/File object
      if (!(file instanceof Blob) && !(file instanceof File)) {
        reject(new Error(`Expected File or Blob, but received: ${typeof file} - ${file?.constructor?.name || 'unknown'}`))
        return
      }

      const reader = new FileReader()
      reader.onload = e => resolve(e.target.result)
      reader.onerror = e => reject(new Error('Failed to read file as binary'))
      reader.readAsArrayBuffer(file)
    })
  }

  triggerDownload(data, filename, mimeType = 'application/octet-stream') {
    let blob
    if (data instanceof ArrayBuffer) {
      blob = new Blob([data], { type: mimeType })
    } else if (typeof data === 'string') {
      blob = new Blob([data], { type: mimeType })
    } else {
      blob = data // Assume it's already a Blob
    }

    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.style.display = 'none'

    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)

    // Clean up the URL object
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  }

  getFileExtension(filename) {
    const lastDot = filename.lastIndexOf('.')
    return lastDot === -1 ? '' : filename.slice(lastDot).toLowerCase()
  }

  isTextFile(extension) {
    return this.supportedFormats.text.includes(extension) ||
      this.supportedFormats.code.includes(extension)
  }

  sanitizeFilename(filename) {
    return filename.replace(/[<>:"|?*\x00-\x1f]/g, '_')
  }

  getMimeType(filename) {
    const ext = this.getFileExtension(filename)
    const mimeTypes = {
      '.txt': 'text/plain',
      '.csv': 'text/csv',
      '.json': 'application/json',
      '.xml': 'application/xml',
      '.html': 'text/html',
      '.css': 'text/css',
      '.js': 'application/javascript',
      '.py': 'text/x-python',
      '.md': 'text/markdown',
      '.png': 'image/png',
      '.jpg': 'image/jpeg',
      '.jpeg': 'image/jpeg',
      '.gif': 'image/gif',
      '.pdf': 'application/pdf',
      '.zip': 'application/zip'
    }
    return mimeTypes[ext] || 'application/octet-stream'
  }

  extractFilenameFromURL(url) {
    try {
      const urlObj = new URL(url)
      const pathname = urlObj.pathname
      const filename = pathname.split('/').pop()
      return filename || 'downloaded-file'
    } catch (error) {
      return 'downloaded-file'
    }
  }

  async collectDirectoryFiles(dirPath) {
    const files = []

    const collectRecursive = async (currentPath) => {
      const items = await this.vfs.readdir(currentPath)

      for (const item of items) {
        const itemPath = this.vfs.pathResolver.join(currentPath, item)
        const stat = await this.vfs.stat(itemPath)

        if (stat.isFile) {
          files.push(itemPath)
        } else if (stat.isDirectory) {
          await collectRecursive(itemPath)
        }
      }
    }

    await collectRecursive(dirPath)
    return files
  }

  async createZipFromFiles(filePaths, basePath) {
    // Simple ZIP creation - in a real implementation, you'd use a ZIP library
    const files = {}

    for (const filePath of filePaths) {
      const relativePath = filePath.startsWith(basePath)
        ? filePath.slice(basePath.length + 1)
        : filePath

      const data = await this.vfs.readFile(filePath, 'binary')
      files[relativePath] = data
    }

    // For now, return a simple archive format
    // In production, use a proper ZIP library like JSZip
    const archiveData = JSON.stringify(files, null, 2)
    return new Blob([archiveData], { type: 'application/json' })
  }

  // ============ DATA FORMAT CONVERTERS ============

  async convertToJSON(sourceData, sourceExtension) {
    switch (sourceExtension) {
      case '.csv':
      case '.tsv':
        return this.csvToJSON(sourceData, sourceExtension === '.tsv' ? '\t' : ',')
      case '.xml':
        return this.xmlToJSON(sourceData)
      default:
        throw new Error(`Cannot convert ${sourceExtension} to JSON`)
    }
  }

  async convertToCSV(sourceData, sourceExtension) {
    switch (sourceExtension) {
      case '.json':
        return this.jsonToCSV(sourceData)
      case '.tsv':
        return sourceData.replace(/\t/g, ',')
      default:
        throw new Error(`Cannot convert ${sourceExtension} to CSV`)
    }
  }

  csvToJSON(csvData, delimiter = ',') {
    const lines = csvData.split('\n').filter(line => line.trim())
    if (lines.length === 0) return '[]'

    const headers = lines[0].split(delimiter).map(h => h.trim())
    const records = []

    for (let i = 1; i < lines.length; i++) {
      const values = lines[i].split(delimiter)
      const record = {}

      headers.forEach((header, index) => {
        record[header] = values[index]?.trim() || ''
      })

      records.push(record)
    }

    return JSON.stringify(records, null, 2)
  }

  jsonToCSV(jsonData) {
    const data = JSON.parse(jsonData)
    if (!Array.isArray(data) || data.length === 0) {
      throw new Error('JSON must be an array of objects')
    }

    const headers = Object.keys(data[0])
    const csvLines = [headers.join(',')]

    for (const record of data) {
      const values = headers.map(header => {
        const value = record[header] || ''
        // Escape commas and quotes
        return value.toString().includes(',') ? `"${value}"` : value
      })
      csvLines.push(values.join(','))
    }

    return csvLines.join('\n')
  }

  xmlToJSON(xmlData) {
    // Basic XML to JSON conversion
    // In production, use a proper XML parser
    try {
      const parser = new DOMParser()
      const xmlDoc = parser.parseFromString(xmlData, 'text/xml')
      const result = this.xmlNodeToObject(xmlDoc.documentElement)
      return JSON.stringify(result, null, 2)
    } catch (error) {
      throw new Error(`Failed to parse XML: ${error.message}`)
    }
  }

  xmlNodeToObject(node) {
    const result = {}

    // Handle attributes
    if (node.attributes && node.attributes.length > 0) {
      result['@attributes'] = {}
      for (const attr of node.attributes) {
        result['@attributes'][attr.name] = attr.value
      }
    }

    // Handle child nodes
    const children = Array.from(node.childNodes)
    const textContent = children
      .filter(child => child.nodeType === Node.TEXT_NODE)
      .map(child => child.textContent.trim())
      .filter(text => text)
      .join(' ')

    if (textContent) {
      result['#text'] = textContent
    }

    const elementChildren = children.filter(child => child.nodeType === Node.ELEMENT_NODE)
    for (const child of elementChildren) {
      const childName = child.nodeName
      const childValue = this.xmlNodeToObject(child)

      if (result[childName]) {
        if (!Array.isArray(result[childName])) {
          result[childName] = [result[childName]]
        }
        result[childName].push(childValue)
      } else {
        result[childName] = childValue
      }
    }

    return result
  }

  // ============ UTILITY METHODS ============

  arrayBufferToBase64(buffer) {
    const bytes = new Uint8Array(buffer)
    let binary = ''
    for (let i = 0; i < bytes.byteLength; i++) {
      binary += String.fromCharCode(bytes[i])
    }
    return btoa(binary)
  }

  base64ToArrayBuffer(base64) {
    const binary = atob(base64)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    return bytes.buffer
  }

  getImportStats() {
    const stats = {
      totalFiles: this.vfs.files.size,
      totalDirectories: this.vfs.directories.size,
      totalSize: 0,
      fileTypes: {}
    }

    for (const [path, _] of this.vfs.files) {
      const metadata = this.vfs.metadata.get(path)
      const size = metadata?.size || 0
      stats.totalSize += size

      const ext = this.getFileExtension(path)
      stats.fileTypes[ext] = (stats.fileTypes[ext] || 0) + 1
    }

    return stats
  }
}

// Export for use in other modules
export default VFSImportExport
export { VFSImportExport }