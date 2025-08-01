/**
 * JavaScript Code Runner Web Worker
 * Executes JavaScript code in a safe, isolated environment with library support
 */

// Simplified VFS implementation for JavaScript Runner
class SimpleVFS {
  constructor() {
    this.files = new Map()
    this.directories = new Set(['/'])
    this.currentDirectory = '/workspace'
    
    // Create default directories
    this.directories.add('/data')
    this.directories.add('/tmp') 
    this.directories.add('/workspace')
  }
  
  writeFile(path, data, options = {}) {
    const normalizedPath = this.normalize(path)
    this.ensureDirectoryExists(this.dirname(normalizedPath))
    this.files.set(normalizedPath, data)
    return normalizedPath
  }
  
  readFile(path, encoding = 'utf8') {
    const normalizedPath = this.normalize(path)
    if (!this.files.has(normalizedPath)) {
      throw new Error(`File not found: ${path}`)
    }
    return this.files.get(normalizedPath)
  }
  
  exists(path) {
    const normalizedPath = this.normalize(path)
    return this.files.has(normalizedPath) || this.directories.has(normalizedPath)
  }
  
  mkdir(path, options = {}) {
    const normalizedPath = this.normalize(path)
    if (options.recursive) {
      const parts = normalizedPath.split('/').filter(Boolean)
      let currentPath = '/'
      for (const part of parts) {
        currentPath = currentPath === '/' ? '/' + part : currentPath + '/' + part
        this.directories.add(currentPath)
      }
    } else {
      this.directories.add(normalizedPath)
    }
    return normalizedPath
  }
  
  readdir(path) {
    const normalizedPath = this.normalize(path)
    if (!this.directories.has(normalizedPath)) {
      throw new Error(`Directory not found: ${path}`)
    }
    
    const items = []
    const prefix = normalizedPath === '/' ? '/' : normalizedPath + '/'
    
    // Find immediate children
    for (const dir of this.directories) {
      if (dir !== normalizedPath && dir.startsWith(prefix)) {
        const relative = dir.slice(prefix.length)
        if (!relative.includes('/')) {
          items.push(relative)
        }
      }
    }
    
    for (const file of this.files.keys()) {
      if (file.startsWith(prefix)) {
        const relative = file.slice(prefix.length)
        if (!relative.includes('/')) {
          items.push(relative)
        }
      }
    }
    
    return items.sort()
  }
  
  stat(path) {
    const normalizedPath = this.normalize(path)
    if (this.directories.has(normalizedPath)) {
      return { isDirectory: true, isFile: false }
    }
    if (this.files.has(normalizedPath)) {
      return { isDirectory: false, isFile: true }
    }
    throw new Error(`Path not found: ${path}`)
  }
  
  unlink(path) {
    const normalizedPath = this.normalize(path)
    if (!this.files.has(normalizedPath)) {
      throw new Error(`File not found: ${path}`)
    }
    this.files.delete(normalizedPath)
  }
  
  rmdir(path) {
    const normalizedPath = this.normalize(path)
    if (!this.directories.has(normalizedPath)) {
      throw new Error(`Directory not found: ${path}`)
    }
    this.directories.delete(normalizedPath)
  }
  
  chdir(path) {
    const normalizedPath = this.normalize(path)
    if (!this.directories.has(normalizedPath)) {
      throw new Error(`Directory not found: ${path}`)
    }
    this.currentDirectory = normalizedPath
  }
  
  getcwd() {
    return this.currentDirectory
  }
  
  normalize(path) {
    if (!path || path === '.') return this.currentDirectory
    if (!path.startsWith('/')) {
      path = this.currentDirectory + '/' + path
    }
    
    const parts = path.split('/').filter(Boolean)
    const resolved = []
    
    for (const part of parts) {
      if (part === '.') continue
      if (part === '..') {
        resolved.pop()
      } else {
        resolved.push(part)
      }
    }
    
    return '/' + resolved.join('/')
  }
  
  dirname(path) {
    const normalized = this.normalize(path)
    if (normalized === '/') return '/'
    const lastSlash = normalized.lastIndexOf('/')
    return lastSlash === 0 ? '/' : normalized.slice(0, lastSlash)
  }
  
  ensureDirectoryExists(path) {
    if (!this.directories.has(path)) {
      this.mkdir(path, { recursive: true })
    }
  }
}

// Global VFS instance
let vfs = null

class SafeJSRunner {
  constructor() {
    this.output = []
    this.loadedLibraries = new Set()
    this.libraryCache = new Map()
    this.executionStats = {
      startTime: 0,
      memoryUsage: 0,
      operationCount: 0,
      maxOperations: 100000, // Prevent infinite loops
      maxMemory: 50 * 1024 * 1024 // 50MB limit (approximate)
    }
    this.setupConsole()
    this.setupVFS()
    this.setupFS()
  }

  setupVFS() {
    // Initialize Virtual File System
    if (!vfs) {
      vfs = new SimpleVFS()
      this.addOutput('info', 'Virtual file system initialized')
    }
  }

  setupFS() {
    // Create Node.js-style fs module
    this.fs = {
      // Synchronous versions
      readFileSync: (path, options = {}) => {
        const encoding = options.encoding || options
        return vfs.readFile(path, encoding)
      },
      
      writeFileSync: (path, data, options = {}) => {
        return vfs.writeFile(path, data, options)
      },
      
      existsSync: (path) => {
        return vfs.exists(path)
      },
      
      mkdirSync: (path, options = {}) => {
        return vfs.mkdir(path, options)
      },
      
      readdirSync: (path, options = {}) => {
        return vfs.readdir(path)
      },
      
      statSync: (path) => {
        return vfs.stat(path)
      },
      
      unlinkSync: (path) => {
        return vfs.unlink(path)
      },
      
      rmdirSync: (path, options = {}) => {
        return vfs.rmdir(path)
      },
      
      // Async versions (Promise-based)
      readFile: async (path, options = {}) => {
        const encoding = options.encoding || options
        return vfs.readFile(path, encoding)
      },
      
      writeFile: async (path, data, options = {}) => {
        return vfs.writeFile(path, data, options)
      },
      
      mkdir: async (path, options = {}) => {
        return vfs.mkdir(path, options)
      },
      
      readdir: async (path, options = {}) => {
        return vfs.readdir(path)
      },
      
      stat: async (path) => {
        return vfs.stat(path)
      },
      
      unlink: async (path) => {
        return vfs.unlink(path)
      },
      
      rmdir: async (path, options = {}) => {
        return vfs.rmdir(path)
      },
      
      // Additional utilities
      promises: {
        readFile: async (path, options = {}) => {
          const encoding = options.encoding || options
          return vfs.readFile(path, encoding)
        },
        
        writeFile: async (path, data, options = {}) => {
          return vfs.writeFile(path, data, options)
        },
        
        mkdir: async (path, options = {}) => {
          return vfs.mkdir(path, options)
        },
        
        readdir: async (path, options = {}) => {
          return vfs.readdir(path)
        },
        
        stat: async (path) => {
          return vfs.stat(path)
        },
        
        unlink: async (path) => {
          return vfs.unlink(path)
        },
        
        rmdir: async (path, options = {}) => {
          return vfs.rmdir(path)
        }
      }
    }
    
    // Path utilities
    this.path = {
      join: (...parts) => {
        const joined = parts.join('/')
        return vfs.normalize(joined)
      },
      
      dirname: (path) => {
        return vfs.dirname(path)
      },
      
      basename: (path) => {
        const normalized = vfs.normalize(path)
        const lastSlash = normalized.lastIndexOf('/')
        return normalized.slice(lastSlash + 1)
      },
      
      extname: (path) => {
        const base = this.path.basename(path)
        const lastDot = base.lastIndexOf('.')
        if (lastDot === -1 || lastDot === 0) return ''
        return base.slice(lastDot)
      },
      
      resolve: (...paths) => {
        return vfs.normalize(paths.join('/'))
      },
      
      relative: (from, to) => {
        // Simple relative path implementation
        const fromParts = vfs.normalize(from).split('/').filter(Boolean)
        const toParts = vfs.normalize(to).split('/').filter(Boolean)
        
        let commonLength = 0
        for (let i = 0; i < Math.min(fromParts.length, toParts.length); i++) {
          if (fromParts[i] === toParts[i]) {
            commonLength++
          } else {
            break
          }
        }
        
        const upLevels = fromParts.length - commonLength
        const relativeParts = Array(upLevels).fill('..').concat(toParts.slice(commonLength))
        
        return relativeParts.join('/') || '.'
      },
      
      isAbsolute: (path) => {
        return path.startsWith('/')
      }
    }
  }

  setupConsole() {
    // Capture console methods and redirect output
    this.console = {
      log: (...args) => this.addOutput('log', this.formatArgs(args)),
      error: (...args) => this.addOutput('error', this.formatArgs(args)),
      warn: (...args) => this.addOutput('warn', this.formatArgs(args)),
      info: (...args) => this.addOutput('info', this.formatArgs(args)),
      debug: (...args) => this.addOutput('debug', this.formatArgs(args)),
      clear: () => {
        this.output = []
        this.addOutput('info', 'Console cleared')
      }
    }
  }

  formatArgs(args) {
    return args.map(arg => {
      if (typeof arg === 'object' && arg !== null) {
        try {
          return JSON.stringify(arg, null, 2)
        } catch (e) {
          return String(arg)
        }
      }
      return String(arg)
    }).join(' ')
  }

  addOutput(type, content) {
    this.output.push({
      id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
      type: type,
      content: String(content),
      timestamp: new Date().toISOString()
    })
  }

  createSafeTimeout(fn, ms) {
    // Limit timeout duration to prevent infinite delays
    const maxMs = Math.min(ms, 5000) // Max 5 seconds
    return setTimeout(fn, maxMs)
  }

  createSafeInterval(fn, ms) {
    // Limit interval frequency to prevent overwhelming execution
    const minMs = Math.max(ms, 100) // Min 100ms
    return setInterval(fn, minMs)
  }

  // Available libraries with their CDN URLs
  getAvailableLibraries() {
    return {
      'lodash': 'https://cdn.jsdelivr.net/npm/lodash@4.17.21/lodash.min.js',
      'd3': 'https://d3js.org/d3.v7.min.js',
      'chart.js': 'https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.min.js',
      'moment': 'https://cdn.jsdelivr.net/npm/moment@2.29.4/moment.min.js',
      'axios': 'https://cdn.jsdelivr.net/npm/axios@1.6.0/dist/axios.min.js',
      'rxjs': 'https://cdn.jsdelivr.net/npm/rxjs@7.8.1/dist/bundles/rxjs.umd.min.js',
      'p5': 'https://cdn.jsdelivr.net/npm/p5@1.7.0/lib/p5.min.js',
      'three': 'https://cdn.jsdelivr.net/npm/three@0.158.0/build/three.min.js',
      'fabric': 'https://cdn.jsdelivr.net/npm/fabric@5.3.0/dist/fabric.min.js'
    }
  }

  // Load a library from CDN
  async loadLibrary(libraryName) {
    if (this.loadedLibraries.has(libraryName)) {
      return true // Already loaded
    }

    const libraries = this.getAvailableLibraries()
    const url = libraries[libraryName.toLowerCase()]
    
    if (!url) {
      throw new Error(`Library '${libraryName}' is not available. Available libraries: ${Object.keys(libraries).join(', ')}`)
    }

    try {
      // Check cache first
      if (this.libraryCache.has(url)) {
        const cachedCode = this.libraryCache.get(url)
        eval(cachedCode)
        this.loadedLibraries.add(libraryName)
        this.addOutput('info', `Library '${libraryName}' loaded from cache`)
        return true
      }

      // Fetch library code
      const response = await fetch(url)
      if (!response.ok) {
        throw new Error(`Failed to load library: ${response.status} ${response.statusText}`)
      }

      const libraryCode = await response.text()
      
      // Cache the library code
      this.libraryCache.set(url, libraryCode)
      
      // Execute library code in global scope
      eval(libraryCode)
      
      this.loadedLibraries.add(libraryName)
      this.addOutput('info', `Library '${libraryName}' loaded successfully`)
      return true
    } catch (error) {
      this.addOutput('error', `Failed to load library '${libraryName}': ${error.message}`)
      return false
    }
  }

  // Parse code for library import statements
  parseLibraryImports(code) {
    const importRegex = /\/\/\s*@import\s+(\w+)/gi
    const imports = []
    let match

    while ((match = importRegex.exec(code)) !== null) {
      imports.push(match[1])
    }

    return imports
  }

  // Create a virtual canvas for graphics operations
  createVirtualCanvas(width = 400, height = 300) {
    const canvas = {
      width: width,
      height: height,
      data: [],
      
      getContext: (type) => {
        if (type === '2d') {
          return {
            // Basic 2D context methods
            fillStyle: '#000000',
            strokeStyle: '#000000',
            lineWidth: 1,
            
            fillRect: function(x, y, w, h) {
              canvas.data.push({ 
                type: 'fillRect', 
                x, y, width: w, height: h, 
                style: this.fillStyle 
              })
            },
            
            strokeRect: function(x, y, w, h) {
              canvas.data.push({ 
                type: 'strokeRect', 
                x, y, width: w, height: h, 
                style: this.strokeStyle,
                lineWidth: this.lineWidth
              })
            },
            
            beginPath: () => canvas.data.push({ type: 'beginPath' }),
            closePath: () => canvas.data.push({ type: 'closePath' }),
            
            moveTo: (x, y) => canvas.data.push({ type: 'moveTo', x, y }),
            lineTo: (x, y) => canvas.data.push({ type: 'lineTo', x, y }),
            
            arc: (x, y, radius, startAngle, endAngle) => {
              canvas.data.push({ type: 'arc', x, y, radius, startAngle, endAngle })
            },
            
            fill: function() { canvas.data.push({ type: 'fill', style: this.fillStyle }) },
            stroke: function() { canvas.data.push({ type: 'stroke', style: this.strokeStyle, lineWidth: this.lineWidth }) },
            
            clearRect: (x, y, w, h) => canvas.data.push({ type: 'clearRect', x, y, width: w, height: h }),
            
            // Text methods
            fillText: function(text, x, y) { canvas.data.push({ type: 'fillText', text, x, y, style: this.fillStyle }) },
            strokeText: function(text, x, y) { canvas.data.push({ type: 'strokeText', text, x, y, style: this.strokeStyle }) }
          }
        }
        return null
      },
      
      toDataURL: () => {
        // Return canvas operations as JSON for later rendering
        return JSON.stringify({
          type: 'canvas',
          width: canvas.width,
          height: canvas.height,
          operations: canvas.data
        })
      }
    }
    
    return canvas
  }

  // Memory and performance monitoring
  checkResourceLimits() {
    // Increment operation count
    this.executionStats.operationCount++
    
    // Check operation limit
    if (this.executionStats.operationCount > this.executionStats.maxOperations) {
      throw new Error(`Operation limit exceeded: ${this.executionStats.maxOperations} operations`)
    }
    
    // Estimate memory usage (rough approximation)
    const memoryEstimate = this.output.length * 1000 + 
                          this.loadedLibraries.size * 100000 + 
                          this.libraryCache.size * 200000
    
    if (memoryEstimate > this.executionStats.maxMemory) {
      throw new Error(`Memory limit exceeded: ~${Math.round(memoryEstimate / 1024 / 1024)}MB`)
    }
    
    this.executionStats.memoryUsage = memoryEstimate
  }

  // Create monitored timeout with resource checking
  createMonitoredTimeout(fn, ms) {
    const maxMs = Math.min(ms, 5000)
    return setTimeout(() => {
      try {
        this.checkResourceLimits()
        fn()
      } catch (error) {
        this.addOutput('error', `Timeout callback error: ${error.message}`)
      }
    }, maxMs)
  }

  // Create monitored interval with resource checking
  createMonitoredInterval(fn, ms) {
    const minMs = Math.max(ms, 100)
    const intervalId = setInterval(() => {
      try {
        this.checkResourceLimits()
        fn()
      } catch (error) {
        this.addOutput('error', `Interval callback error: ${error.message}`)
        clearInterval(intervalId)
      }
    }, minMs)
    return intervalId
  }

  // Enhanced execution environment with monitoring
  createSecureFunction(code) {
    // Direct execution without nested eval to preserve console context
    return code
  }

  async execute(code) {
    this.output = []
    this.executionStats.startTime = performance.now()
    this.executionStats.operationCount = 0
    this.executionStats.memoryUsage = 0
    
    try {
      // Parse and load any required libraries
      const requiredLibraries = this.parseLibraryImports(code)
      for (const library of requiredLibraries) {
        await this.loadLibrary(library)
      }

      // Create safe execution context with enhanced globals
      const safeGlobals = {
        // Console methods
        console: this.console,
        
        // Safe built-in objects
        Math: Math,
        Date: Date,
        Array: Array,
        Object: Object,
        String: String,
        Number: Number,
        Boolean: Boolean,
        JSON: JSON,
        RegExp: RegExp,
        
        // Enhanced timer functions with monitoring
        setTimeout: this.createMonitoredTimeout.bind(this),
        setInterval: this.createMonitoredInterval.bind(this),
        clearTimeout: clearTimeout,
        clearInterval: clearInterval,
        
        // Performance monitoring
        performance: { now: performance.now.bind(performance) },
        
        // Safe Promise support
        Promise: Promise,
        
        // Error handling
        Error: Error,
        TypeError: TypeError,
        ReferenceError: ReferenceError,
        SyntaxError: SyntaxError,
        
        // Utility functions
        isNaN: isNaN,
        isFinite: isFinite,
        parseInt: parseInt,
        parseFloat: parseFloat,
        
        // Canvas and graphics support
        createCanvas: this.createVirtualCanvas.bind(this),
        
        // Library management
        loadLibrary: this.loadLibrary.bind(this),
        getAvailableLibraries: this.getAvailableLibraries.bind(this),
        
        // File system support
        fs: this.fs,
        require: (module) => {
          if (module === 'fs') return this.fs
          if (module === 'path') return this.path
          throw new Error(`Module '${module}' not found`)
        },
        
        // Process simulation
        process: {
          cwd: () => vfs.getcwd(),
          chdir: (path) => vfs.chdir(path),
          env: {},
          argv: ['node'],
          version: 'v18.0.0',
          platform: 'browser'
        },
        
        // Resource monitoring (internal use)
        checkResourceLimits: this.checkResourceLimits.bind(this),
        
        // Enhanced crypto support
        crypto: {
          getRandomValues: crypto.getRandomValues.bind(crypto),
          randomUUID: crypto.randomUUID ? crypto.randomUUID.bind(crypto) : () => {
            return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
              const r = Math.random() * 16 | 0
              const v = c == 'x' ? r : (r & 0x3 | 0x8)
              return v.toString(16)
            })
          }
        },
        
        // Text encoding/decoding
        TextEncoder: typeof TextEncoder !== 'undefined' ? TextEncoder : undefined,
        TextDecoder: typeof TextDecoder !== 'undefined' ? TextDecoder : undefined,
        
        // Enhanced URL support
        URL: typeof URL !== 'undefined' ? URL : undefined,
        URLSearchParams: typeof URLSearchParams !== 'undefined' ? URLSearchParams : undefined,
        
        // Prevent access to dangerous globals
        eval: undefined,
        Function: undefined,
        window: undefined,
        global: undefined,
        self: undefined,
        document: undefined,
        XMLHttpRequest: undefined,
        fetch: undefined,
        WebSocket: undefined,
        Worker: undefined,
        SharedWorker: undefined,
        ServiceWorker: undefined,
        localStorage: undefined,
        sessionStorage: undefined,
        indexedDB: undefined,
        location: undefined,
        navigator: undefined,
        history: undefined
      }

      // Execute code in safe context with enhanced monitoring
      const secureCode = this.createSecureFunction(code)
      
      // Set up periodic resource checking
      const checkInterval = setInterval(() => {
        try {
          this.checkResourceLimits()
        } catch (error) {
          clearInterval(checkInterval)
          throw error
        }
      }, 100)
      
      let result
      try {
        result = new Function(
          ...Object.keys(safeGlobals), 
          `
          try {
            return (function() {
              ${secureCode}
            })();
          } catch (error) {
            console.error('Runtime Error: ' + error.message);
            throw error;
          }
          `
        )(...Object.values(safeGlobals))
      } finally {
        clearInterval(checkInterval)
      }

      // Add return value if it exists and is not undefined
      if (result !== undefined) {
        let formattedResult = result
        
        // Handle different return types
        if (typeof result === 'object' && result !== null) {
          // Check if it's a canvas operation result
          if (result.toDataURL && typeof result.toDataURL === 'function') {
            try {
              const canvasData = result.toDataURL()
              this.addOutput('canvas', canvasData)
              return this.output
            } catch (e) {
              this.addOutput('error', `Canvas error: ${e.message}`)
            }
          }
          
          // Check if it's already a canvas data object
          if (result.type === 'canvas' && result.operations) {
            this.addOutput('canvas', JSON.stringify(result))
            return this.output
          }
          
          // Handle other objects
          try {
            formattedResult = JSON.stringify(result, null, 2)
          } catch (e) {
            formattedResult = String(result)
          }
        }
        
        this.addOutput('return', formattedResult)
      }

      // Add execution statistics
      const executionTime = Math.round(performance.now() - this.executionStats.startTime)
      const memoryUsed = Math.round(this.executionStats.memoryUsage / 1024 / 1024 * 100) / 100
      const operations = this.executionStats.operationCount
      
      this.addOutput('info', `Execution completed in ${executionTime}ms | ~${memoryUsed}MB | ${operations} ops`)

      return this.output
    } catch (error) {
      // Handle syntax and runtime errors
      const executionTime = Math.round(performance.now() - this.executionStats.startTime)
      const memoryUsed = Math.round(this.executionStats.memoryUsage / 1024 / 1024 * 100) / 100
      const operations = this.executionStats.operationCount
      
      this.addOutput('error', `${error.name}: ${error.message}`)
      this.addOutput('info', `Execution failed after ${executionTime}ms | ~${memoryUsed}MB | ${operations} ops`)
      return this.output
    }
  }
}

// Create runner instance
const runner = new SafeJSRunner()

// Handle messages from main thread
self.onmessage = async (e) => {
  const { type, code, timeout = 10000, requestId, libraryName, path, data, options } = e.data
  
  try {
    if (type === 'execute') {
      // Set up execution timeout
      let timeoutId
      const timeoutPromise = new Promise((_, reject) => {
        timeoutId = setTimeout(() => {
          reject(new Error(`Code execution timed out after ${timeout}ms`))
        }, timeout)
      })
      
      // Execute code with timeout
      const executionPromise = runner.execute(code)
      
      const results = await Promise.race([executionPromise, timeoutPromise])
      
      // Clear timeout if execution completed in time
      clearTimeout(timeoutId)
      
      // Send results back to main thread
      self.postMessage({ 
        type: 'results', 
        data: results,
        requestId: requestId
      })
    } else if (type === 'loadLibrary') {
      // Handle library loading requests
      const success = await runner.loadLibrary(libraryName)
      self.postMessage({
        type: 'libraryLoaded',
        data: { success, libraryName },
        requestId: requestId
      })
    } else if (type === 'getAvailableLibraries') {
      // Handle library list requests
      const libraries = runner.getAvailableLibraries()
      self.postMessage({
        type: 'availableLibraries',
        data: libraries,
        requestId: requestId
      })
    } else if (type === 'vfs') {
      // Handle VFS operations
      const { operation } = e.data
      let result
      
      try {
        switch (operation) {
          case 'writeFile':
            result = vfs.writeFile(path, data, options)
            break
          case 'readFile':
            result = vfs.readFile(path, options?.encoding)
            break
          case 'exists':
            result = vfs.exists(path)
            break
          case 'mkdir':
            result = vfs.mkdir(path, options)
            break
          case 'readdir':
            result = vfs.readdir(path)
            break
          case 'stat':
            result = vfs.stat(path)
            break
          case 'unlink':
            result = vfs.unlink(path)
            break
          case 'rmdir':
            result = vfs.rmdir(path)
            break
          case 'chdir':
            result = vfs.chdir(path)
            break
          case 'getcwd':
            result = vfs.getcwd()
            break
          default:
            throw new Error(`Unknown VFS operation: ${operation}`)
        }
        
        self.postMessage({
          type: 'vfsResult',
          data: result,
          requestId: requestId
        })
      } catch (error) {
        self.postMessage({
          type: 'vfsError',
          data: { message: error.message },
          requestId: requestId
        })
      }
    } else if (type === 'syncVFS') {
      // Handle VFS synchronization from main thread
      try {
        const { vfsState } = e.data
        if (vfsState && vfsState.files && vfsState.directories) {
          // Clear current VFS state
          vfs.files.clear()
          vfs.directories.clear()
          
          // Import files from serialized state (array of [path, content] pairs)
          if (Array.isArray(vfsState.files)) {
            vfsState.files.forEach(([path, content]) => {
              vfs.files.set(path, content)
            })
          }
          
          // Import directories from serialized state (array of paths)
          if (Array.isArray(vfsState.directories)) {
            vfsState.directories.forEach(dir => {
              vfs.directories.add(dir)
            })
          }
          
          // Update current directory
          if (vfsState.currentDirectory) {
            vfs.currentDirectory = vfsState.currentDirectory
          }
          
          runner.addOutput('info', `VFS synchronized: ${vfs.files.size} files, ${vfs.directories.size} directories`)
        }
      } catch (error) {
        runner.addOutput('error', `VFS sync failed: ${error.message}`)
      }
    } else {
      // Backward compatibility - treat as execute
      const executionPromise = runner.execute(code)
      const results = await executionPromise
      
      self.postMessage({ 
        type: 'results', 
        data: results,
        requestId: requestId
      })
    }
  } catch (error) {
    // Handle timeout or other errors
    self.postMessage({ 
      type: 'error', 
      data: { 
        message: error.message,
        name: error.name || 'ExecutionError'
      },
      requestId: requestId
    })
  }
}

// Handle worker errors
self.onerror = (error) => {
  self.postMessage({
    type: 'error',
    data: {
      message: error.message || 'Unknown worker error',
      name: 'WorkerError'
    }
  })
}

// Handle unhandled promise rejections
self.onunhandledrejection = (event) => {
  self.postMessage({
    type: 'error',
    data: {
      message: event.reason?.message || 'Unhandled promise rejection',
      name: 'UnhandledRejection'
    }
  })
}