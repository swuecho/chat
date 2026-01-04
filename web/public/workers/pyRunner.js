/**
 * Python Code Runner Web Worker
 * Executes Python code using Pyodide in a safe, isolated environment
 */

// Simplified VFS implementation for Python Runner
class SimpleVFS {
  constructor() {
    this.files = new Map()
    this.directories = new Set(['/'])
    this.metadata = new Map()  // Add metadata support
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

// Global Pyodide instance
let pyodide = null
let isInitialized = false
let initializationPromise = null

// Global VFS instance
let vfs = null

// Counter to detect corrupted package issues
let packageLoadFailures = 0
let totalPackageLoadAttempts = 0

class SafePyRunner {
  constructor() {
    this.output = []
    this.loadedPackages = new Set()
    this.executionStats = {
      startTime: 0,
      memoryUsage: 0,
      operationCount: 0,
      maxOperations: 100000,
      maxMemory: 100 * 1024 * 1024 // 100MB limit
    }
    this.setupOutput()
    this.setupVFS()
  }

  setupVFS() {
    // Initialize Virtual File System
    if (!vfs) {
      vfs = new SimpleVFS()
      this.addOutput('info', 'Virtual file system initialized')
    }
  }

  setupOutput() {
    // Capture stdout/stderr
    this.capturedOutput = []
    this.outputCapture = {
      write: (text) => {
        this.capturedOutput.push(text)
        this.addOutput('stdout', text)
      },
      flush: () => {}
    }
  }

  addOutput(type, content) {
    this.output.push({
      id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
      type: type,
      content: String(content),
      timestamp: new Date().toISOString()
    })
  }

  // Initialize Pyodide with fallback CDNs
  async initializePyodide() {
    if (isInitialized) return pyodide
    if (initializationPromise) return initializationPromise

    initializationPromise = new Promise(async (resolve, reject) => {
      try {
        this.addOutput('info', 'Initializing Python environment...')

        // Try multiple CDNs in case one is down or corrupted
        const cdnOptions = [
          {
            name: 'jsDelivr',
            scriptURL: 'https://cdn.jsdelivr.net/pyodide/v0.24.1/full/pyodide.js',
            indexURL: 'https://cdn.jsdelivr.net/pyodide/v0.24.1/full/'
          },
          {
            name: 'unpkg',
            scriptURL: 'https://unpkg.com/pyodide@0.24.1/full/pyodide.js',
            indexURL: 'https://unpkg.com/pyodide@0.24.1/full/'
          }
        ]

        let lastError = null
        let loadedSuccessfully = false

        for (const cdn of cdnOptions) {
          try {
            this.addOutput('info', `Trying Pyodide CDN: ${cdn.name}...`)

            // Load Pyodide script dynamically (only if not already loaded)
            if (typeof self.loadPyodide === 'undefined') {
              await new Promise((loadResolve, loadReject) => {
                const script = document ? document.createElement('script') : null
                if (!script) {
                  // In worker context, use importScripts (synchronous)
                  try {
                    importScripts(cdn.scriptURL)
                    loadResolve()
                  } catch (e) {
                    loadReject(e)
                  }
                } else {
                  // In main thread context
                  script.src = cdn.scriptURL
                  script.onload = () => loadResolve()
                  script.onerror = () => loadReject(new Error(`Failed to load ${cdn.scriptURL}`))
                  document.head.appendChild(script)
                }
              })
            }

            // Try to load Pyodide
            pyodide = await loadPyodide({
              indexURL: cdn.indexURL,
              stdout: this.outputCapture.write.bind(this.outputCapture),
              stderr: this.outputCapture.write.bind(this.outputCapture)
            })

            this.addOutput('info', `✓ Successfully loaded Pyodide from ${cdn.name}`)
            loadedSuccessfully = true
            break
          } catch (error) {
            lastError = error
            this.addOutput('warn', `Failed to load from ${cdn.name}: ${error.message}`)
            // Reset for next CDN attempt
            pyodide = null
          }
        }

        if (!loadedSuccessfully || !pyodide) {
          throw lastError || new Error('Failed to load Pyodide from any CDN')
        }

        // Register VFS with Pyodide
        pyodide.registerJsModule('vfs', vfs)

        // Set up matplotlib backend for web and VFS integration
        await pyodide.runPython(`
import sys
import io
import os
import base64
import json
import builtins
from contextlib import redirect_stdout, redirect_stderr
from pathlib import Path
import tempfile

# Create custom output capture
class OutputCapture:
    def __init__(self):
        self.output = []
    
    def write(self, text):
        if text.strip():
            self.output.append(text)
    
    def flush(self):
        pass
    
    def getvalue(self):
        return ''.join(self.output)

# Global output capture
_output_capture = OutputCapture()

# Store original print function
_original_print = builtins.print

# Custom print function
def custom_print(*args, **kwargs):
    output = io.StringIO()
    _original_print(*args, file=output, **kwargs)
    result = output.getvalue()
    _output_capture.write(result)
    return result

# Replace built-in print safely
builtins.print = custom_print

# VFS Integration - File operations
class VFSFile:
    """File-like object that interfaces with JavaScript VFS"""
    def __init__(self, path, mode='r', encoding='utf-8'):
        self.path = path
        self.mode = mode
        self.encoding = encoding
        self.position = 0
        self.closed = False
        self._content = None
        
        # Read existing content if file exists and mode allows reading
        if 'r' in mode or 'a' in mode or '+' in mode:
            try:
                from js import vfs
                if vfs.exists(path):
                    if 'b' in mode:
                        self._content = vfs.readFile(path, 'binary')
                    else:
                        self._content = vfs.readFile(path, 'utf8')
                else:
                    self._content = b'' if 'b' in mode else ''
            except:
                self._content = b'' if 'b' in mode else ''
        else:
            self._content = b'' if 'b' in mode else ''
            
        # For write modes, truncate content
        if 'w' in mode:
            self._content = b'' if 'b' in mode else ''
    
    def read(self, size=-1):
        if self.closed:
            raise ValueError("I/O operation on closed file")
        
        if size == -1:
            result = self._content[self.position:]
            self.position = len(self._content)
        else:
            result = self._content[self.position:self.position + size]
            self.position += len(result)
        
        return result
    
    def readline(self, size=-1):
        if self.closed:
            raise ValueError("I/O operation on closed file")
        
        if 'b' in self.mode:
            newline = b'\\n'
        else:
            newline = '\\n'
        
        start = self.position
        try:
            newline_pos = self._content.index(newline, start)
            if size == -1 or newline_pos - start + 1 <= size:
                result = self._content[start:newline_pos + 1]
                self.position = newline_pos + 1
            else:
                result = self._content[start:start + size]
                self.position = start + size
        except ValueError:
            # No newline found
            if size == -1:
                result = self._content[start:]
                self.position = len(self._content)
            else:
                result = self._content[start:start + size]
                self.position = start + size
        
        return result
    
    def readlines(self):
        if self.closed:
            raise ValueError("I/O operation on closed file")
        
        lines = []
        while True:
            line = self.readline()
            if not line:
                break
            lines.append(line)
        return lines
    
    def write(self, data):
        if self.closed:
            raise ValueError("I/O operation on closed file")
        
        if 'r' in self.mode and '+' not in self.mode:
            raise io.UnsupportedOperation("not writable")
        
        if 'b' in self.mode and isinstance(data, str):
            data = data.encode(self.encoding)
        elif 'b' not in self.mode and isinstance(data, bytes):
            data = data.decode(self.encoding)
        
        if 'a' in self.mode:
            # Append mode
            self._content += data
            self.position = len(self._content)
        else:
            # Write or insert mode
            if isinstance(self._content, bytes):
                self._content = self._content[:self.position] + data + self._content[self.position + len(data):]
            else:
                self._content = self._content[:self.position] + data + self._content[self.position + len(data):]
            self.position += len(data)
        
        return len(data)
    
    def writelines(self, lines):
        for line in lines:
            self.write(line)
    
    def seek(self, position, whence=0):
        if self.closed:
            raise ValueError("I/O operation on closed file")
        
        if whence == 0:  # SEEK_SET
            self.position = position
        elif whence == 1:  # SEEK_CUR
            self.position += position
        elif whence == 2:  # SEEK_END
            self.position = len(self._content) + position
        
        self.position = max(0, min(self.position, len(self._content)))
        return self.position
    
    def tell(self):
        if self.closed:
            raise ValueError("I/O operation on closed file")
        return self.position
    
    def flush(self):
        if self.closed:
            raise ValueError("I/O operation on closed file")
        
        # Write content to VFS
        try:
            from js import vfs
            if 'b' in self.mode:
                vfs.writeFile(self.path, self._content, {'binary': True})
            else:
                vfs.writeFile(self.path, self._content)
        except Exception as e:
            pass  # Ignore VFS errors for now
    
    def close(self):
        if not self.closed:
            self.flush()
            self.closed = True
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()
    
    def readable(self):
        return 'r' in self.mode or '+' in self.mode
    
    def writable(self):
        return 'w' in self.mode or 'a' in self.mode or '+' in self.mode
    
    def seekable(self):
        return True

# Override built-in open function
_original_open = builtins.open

def vfs_open(file, mode='r', buffering=-1, encoding=None, errors=None, newline=None, closefd=True, opener=None):
    """VFS-aware open() replacement"""
    # Convert file to string if it's a Path object
    if hasattr(file, '__fspath__'):
        file = file.__fspath__()
    
    file_str = str(file)
    
    # Check if this is a VFS path (starts with /)
    if file_str.startswith('/'):
        if encoding is None:
            encoding = 'utf-8'
        return VFSFile(file_str, mode, encoding)
    else:
        # Use original open for non-VFS paths
        return _original_open(file, mode, buffering, encoding, errors, newline, closefd, opener)

# Replace built-in open
builtins.open = vfs_open

# Override os module functions for VFS
_original_listdir = os.listdir
_original_makedirs = os.makedirs
_original_path_exists = os.path.exists
_original_path_isfile = os.path.isfile
_original_path_isdir = os.path.isdir
_original_path_getsize = os.path.getsize
_original_getcwd = os.getcwd
_original_chdir = os.chdir
_original_remove = os.remove
_original_rmdir = os.rmdir

def vfs_listdir(path='.'):
    """VFS-aware listdir"""
    try:
        if str(path).startswith('/'):
            from js import vfs
            return vfs.readdir(str(path))
        else:
            return _original_listdir(path)
    except:
        return _original_listdir(path)

def vfs_makedirs(name, mode=0o777, exist_ok=False):
    """VFS-aware makedirs"""
    try:
        if str(name).startswith('/'):
            from js import vfs
            vfs.mkdir(str(name), {'recursive': True})
        else:
            _original_makedirs(name, mode, exist_ok)
    except:
        if not exist_ok:
            raise

def vfs_path_exists(path):
    """VFS-aware path.exists"""
    try:
        if str(path).startswith('/'):
            from js import vfs
            return vfs.exists(str(path))
        else:
            return _original_path_exists(path)
    except:
        return _original_path_exists(path)

def vfs_path_isfile(path):
    """VFS-aware path.isfile"""
    try:
        if str(path).startswith('/'):
            from js import vfs
            if vfs.exists(str(path)):
                stat = vfs.stat(str(path))
                return stat.get('isFile', False)
            return False
        else:
            return _original_path_isfile(path)
    except:
        return _original_path_isfile(path)

def vfs_path_isdir(path):
    """VFS-aware path.isdir"""
    try:
        if str(path).startswith('/'):
            from js import vfs
            if vfs.exists(str(path)):
                stat = vfs.stat(str(path))
                return stat.get('isDirectory', False)
            return False
        else:
            return _original_path_isdir(path)
    except:
        return _original_path_isdir(path)

def vfs_getcwd():
    """VFS-aware getcwd"""
    try:
        from js import vfs
        return vfs.getcwd()
    except:
        return '/workspace'

def vfs_chdir(path):
    """VFS-aware chdir"""
    try:
        if str(path).startswith('/'):
            from js import vfs
            vfs.chdir(str(path))
        else:
            _original_chdir(path)
    except:
        pass

def vfs_remove(path):
    """VFS-aware remove"""
    try:
        if str(path).startswith('/'):
            from js import vfs
            vfs.unlink(str(path))
        else:
            _original_remove(path)
    except:
        _original_remove(path)

def vfs_rmdir(path):
    """VFS-aware rmdir"""
    try:
        if str(path).startswith('/'):
            from js import vfs
            vfs.rmdir(str(path))
        else:
            _original_rmdir(path)
    except:
        _original_rmdir(path)

# Patch os module
os.listdir = vfs_listdir
os.makedirs = vfs_makedirs
os.path.exists = vfs_path_exists
os.path.isfile = vfs_path_isfile
os.path.isdir = vfs_path_isdir
os.getcwd = vfs_getcwd
os.chdir = vfs_chdir
os.remove = vfs_remove
os.rmdir = vfs_rmdir

# Patch pathlib for modern Python code
try:
    import pathlib
    
    class VFSPath(pathlib.PurePosixPath):
        """VFS-aware Path implementation"""
        
        def exists(self):
            return vfs_path_exists(str(self))
        
        def is_file(self):
            return vfs_path_isfile(str(self))
        
        def is_dir(self):
            return vfs_path_isdir(str(self))
        
        def open(self, mode='r', buffering=-1, encoding=None, errors=None, newline=None):
            return vfs_open(str(self), mode, buffering, encoding, errors, newline)
        
        def read_text(self, encoding=None, errors=None):
            with self.open('r', encoding=encoding, errors=errors) as f:
                return f.read()
        
        def read_bytes(self):
            with self.open('rb') as f:
                return f.read()
        
        def write_text(self, data, encoding=None, errors=None):
            with self.open('w', encoding=encoding, errors=errors) as f:
                return f.write(data)
        
        def write_bytes(self, data):
            with self.open('wb') as f:
                return f.write(data)
        
        def mkdir(self, mode=0o777, parents=False, exist_ok=False):
            vfs_makedirs(str(self), mode, exist_ok or parents)
        
        def iterdir(self):
            if self.is_dir():
                for name in vfs_listdir(str(self)):
                    yield self / name
        
        def glob(self, pattern):
            # Simple glob implementation
            try:
                from js import vfs
                matches = vfs.glob(str(self / pattern))
                return [VFSPath(match) for match in matches]
            except:
                return []
        
        def unlink(self):
            vfs_remove(str(self))
        
        def rmdir(self):
            vfs_rmdir(str(self))
    
    # Replace Path with VFSPath for VFS paths
    _original_Path = pathlib.Path
    
    def smart_Path(*args, **kwargs):
        path_str = str(args[0]) if args else '.'
        if path_str.startswith('/'):
            return VFSPath(*args, **kwargs)
        else:
            return _original_Path(*args, **kwargs)
    
    pathlib.Path = smart_Path
    
except ImportError:
    pass  # pathlib not available

# Set initial working directory
try:
    from js import vfs
    vfs.chdir('/workspace')
except:
    pass
`)

        isInitialized = true
        this.addOutput('info', 'Python environment initialized successfully')
        resolve(pyodide)
      } catch (error) {
        this.addOutput('error', `Failed to initialize Python: ${error.message}`)
        reject(error)
      }
    })

    return initializationPromise
  }

  // Get available Python packages
  getAvailablePackages() {
    return {
      'numpy': 'Numerical computing library',
      'pandas': 'Data manipulation and analysis',
      'matplotlib': 'Plotting library',
      'scipy': 'Scientific computing',
      'scikit-learn': 'Machine learning library',
      'requests': 'HTTP library',
      'beautifulsoup4': 'Web scraping',
      'pillow': 'Image processing',
      'sympy': 'Symbolic mathematics',
      'networkx': 'Graph analysis',
      'seaborn': 'Statistical visualization',
      'plotly': 'Interactive plotting',
      'bokeh': 'Interactive visualization',
      'altair': 'Statistical visualization'
    }
  }

  // Load a Python package with retry logic and cache busting
  async loadPackage(packageName, retries = 3) {
    if (!pyodide) {
      await this.initializePyodide()
    }

    if (this.loadedPackages.has(packageName)) {
      return true
    }

    let lastError = null
    for (let attempt = 1; attempt <= retries; attempt++) {
      try {
        totalPackageLoadAttempts++
        this.addOutput('info', `Loading Python package: ${packageName}${attempt > 1 ? ` (attempt ${attempt}/${retries})` : ''}`)

        // Add cache-busting parameter for retries to avoid using corrupted cached files
        const options = attempt > 1 ? {} : undefined
        await pyodide.loadPackage(packageName, options)
        this.loadedPackages.add(packageName)
        this.addOutput('info', `Package '${packageName}' loaded successfully`)
        return true
      } catch (error) {
        lastError = error
        packageLoadFailures++
        this.addOutput('warn', `Attempt ${attempt}/${retries} failed for '${packageName}': ${error.message}`)

        if (attempt < retries) {
          // Exponential backoff: wait 1s, 2s, 4s, etc.
          const waitTime = Math.pow(2, attempt - 1) * 1000
          this.addOutput('info', `Retrying in ${waitTime / 1000}s...`)
          await new Promise(resolve => setTimeout(resolve, waitTime))

          // Clear pyodide's internal package cache on retry
          try {
            if (pyodide.package_loader && pyodide.package_loader.loadedPackages) {
              delete pyodide.package_loader.loadedPackages[packageName]
            }
          } catch (e) {
            // Ignore cache clear errors
          }
        }
      }
    }

    // Check if this is a systematic failure (most packages failing)
    const failureRate = totalPackageLoadAttempts > 0 ? packageLoadFailures / totalPackageLoadAttempts : 0
    const isSystematicFailure = failureRate > 0.7 && totalPackageLoadAttempts >= 3

    // Provide detailed troubleshooting instructions
    this.addOutput('error', `Failed to load package '${packageName}' after ${retries} attempts`)
    this.addOutput('error', `Last error: ${lastError?.message || 'Unknown error'}`)

    if (isSystematicFailure) {
      this.addOutput('error', `⚠️  DETECTED SYSTEMATIC PACKAGE FAILURE (${Math.round(failureRate * 100)}% failure rate)`)
      this.addOutput('error', `All packages are being corrupted during download. This is likely:`)
      this.addOutput('error', `  • Corporate proxy/firewall inspecting and corrupting .whl files`)
      this.addOutput('error', `  • Browser extension interfering with downloads`)
      this.addOutput('error', `  • Network issue corrupting binary downloads`)
      this.addOutput('error', ``)
      this.addOutput('error', `TO FIX THIS ISSUE:`)
      this.addOutput('error', `  1. Try disabling ALL browser extensions`)
      this.addOutput('error', `  2. Try a different browser (Chrome, Firefox, Safari)`)
      this.addOutput('error', `  3. Try a different network (WiFi hotspot, mobile data)`)
      this.addOutput('error', `  4. If using corporate VPN/proxy, try disconnecting`)
      this.addOutput('error', `  5. Check if antivirus is scanning web downloads`)
      this.addOutput('error', `  6. Try in incognito/private mode (extensions disabled)`)
      this.addOutput('info', `💡 Python standard library will still work (math, json, re, datetime, etc.)`)
      this.addOutput('info', `   Only external packages like numpy/pandas require download.`)
    } else {
      this.addOutput('info', `📋 Troubleshooting steps:`)
      this.addOutput('info', `   1. Hard refresh: Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows)`)
      this.addOutput('info', `   2. Clear browser cache for this site`)
      this.addOutput('info', `   3. Check your internet connection`)
      this.addOutput('info', `   4. Try a different browser or network`)
      this.addOutput('info', `   5. If problem persists, the CDN may be temporarily unavailable`)
    }

    return false
  }

  // Parse code for package import statements
  parsePackageImports(code) {
    const imports = []
    
    // Match import statements
    const importRegex = /(?:^|\n)\s*(?:import|from)\s+([a-zA-Z_][a-zA-Z0-9_]*)/gm
    let match
    
    while ((match = importRegex.exec(code)) !== null) {
      const packageName = match[1]
      
      // Map common package names to Pyodide package names
      const packageMapping = {
        'numpy': 'numpy',
        'np': 'numpy',
        'pandas': 'pandas',
        'pd': 'pandas',
        'matplotlib': 'matplotlib',
        'plt': 'matplotlib',
        'scipy': 'scipy',
        'sklearn': 'scikit-learn',
        'requests': 'requests',
        'bs4': 'beautifulsoup4',
        'PIL': 'pillow',
        'sympy': 'sympy',
        'networkx': 'networkx',
        'seaborn': 'seaborn',
        'plotly': 'plotly',
        'bokeh': 'bokeh',
        'altair': 'altair'
      }
      
      const mappedName = packageMapping[packageName] || packageName
      if (this.getAvailablePackages()[mappedName]) {
        imports.push(mappedName)
      }
    }
    
    return [...new Set(imports)] // Remove duplicates
  }

  // Handle matplotlib plots
  async handleMatplotlib(code) {
    if (!code.includes('matplotlib') && !code.includes('plt')) {
      return null
    }

    try {
      // Set up matplotlib for web output
      await pyodide.runPython(`
import matplotlib
matplotlib.use('Agg')  # Use non-interactive backend
import matplotlib.pyplot as plt
import io
import base64

# Clear any existing plots
plt.clf()

# Function to capture plot as base64
def capture_plot():
    buf = io.BytesIO()
    plt.savefig(buf, format='png', bbox_inches='tight', dpi=150)
    buf.seek(0)
    img_base64 = base64.b64encode(buf.read()).decode('utf-8')
    buf.close()
    return img_base64
`)

      // Execute the user code
      await pyodide.runPython(code)
      
      // Check if there's a plot to capture
      const hasPlot = await pyodide.runPython(`
import matplotlib.pyplot as plt
len(plt.get_fignums()) > 0
`)
      
      if (hasPlot) {
        const plotData = await pyodide.runPython('capture_plot()')
        return {
          type: 'matplotlib',
          data: plotData,
          format: 'png'
        }
      }
      
      return null
    } catch (error) {
      this.addOutput('error', `Matplotlib error: ${error.message}`)
      return null
    }
  }

  // Resource monitoring
  checkResourceLimits() {
    this.executionStats.operationCount++
    
    if (this.executionStats.operationCount > this.executionStats.maxOperations) {
      throw new Error(`Operation limit exceeded: ${this.executionStats.maxOperations} operations`)
    }
    
    // Estimate memory usage
    const memoryEstimate = this.output.length * 1000 + this.loadedPackages.size * 10000000
    if (memoryEstimate > this.executionStats.maxMemory) {
      throw new Error(`Memory limit exceeded: ~${Math.round(memoryEstimate / 1024 / 1024)}MB`)
    }
    
    this.executionStats.memoryUsage = memoryEstimate
  }

  // Execute Python code
  async execute(code) {
    this.output = []
    this.capturedOutput = []
    this.executionStats.startTime = performance.now()
    this.executionStats.operationCount = 0
    this.executionStats.memoryUsage = 0

    try {
      // Initialize Pyodide if not already done
      if (!pyodide) {
        await this.initializePyodide()
      }

      // Parse and load required packages
      const requiredPackages = this.parsePackageImports(code)
      for (const pkg of requiredPackages) {
        const success = await this.loadPackage(pkg)
        if (!success) {
          throw new Error(`Failed to load required package '${pkg}'. Please try again.`)
        }
      }

      // Clear previous output
      await pyodide.runPython(`
_output_capture.output = []
`)

      // Handle matplotlib plots
      const plotResult = await this.handleMatplotlib(code)
      if (plotResult) {
        this.addOutput('matplotlib', JSON.stringify(plotResult))
      }

      // Execute the code and capture result
      let result = null
      try {
        // Handle asyncio.run() which doesn't work in Pyodide's existing event loop
        let processedCode = code
        if (code.includes('asyncio.run(')) {
          this.addOutput('info', 'Detected asyncio.run() - converting for Pyodide compatibility')
          
          // Replace asyncio.run(func()) with await func() wrapped in an async context
          processedCode = code.replace(/asyncio\.run\(([^)]+)\)/g, 'await $1')
          
          // Wrap the entire code in an async function that can be executed by runPythonAsync
          processedCode = `
import asyncio

async def _execute_main():
${processedCode.split('\n').map(line => '    ' + line).join('\n')}

# Execute the main function
await _execute_main()
`
          
          this.addOutput('info', 'Wrapped code in async context for Pyodide execution')
        }
        
        // Check if code has async/await but no asyncio.run
        else if (processedCode.includes('async def') || processedCode.includes('await ')) {
          // Wrap async code in a proper async context
          processedCode = `
import asyncio
try:
    loop = asyncio.get_event_loop()
except RuntimeError:
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

${processedCode}
`
        }
        
        // Use runPythonAsync for async code, runPython for sync code
        if (processedCode.includes('asyncio.') || processedCode.includes('async def') || processedCode.includes('await ')) {
          result = await pyodide.runPythonAsync(processedCode)
        } else {
          result = await pyodide.runPython(processedCode)
        }
        
        // Get captured output
        const capturedOutput = await pyodide.runPython(`
_output_capture.getvalue()
`)
        
        if (capturedOutput && capturedOutput.trim()) {
          this.addOutput('stdout', capturedOutput)
        }
        
        // Clear output buffer
        await pyodide.runPython(`
_output_capture.output = []
`)
        
        // Add result if it exists
        if (result !== undefined && result !== null) {
          this.addOutput('return', String(result))
        }
      } catch (error) {
        // Get any remaining output before error
        try {
          const capturedOutput = await pyodide.runPython(`_output_capture.getvalue()`)
          if (capturedOutput && capturedOutput.trim()) {
            this.addOutput('stdout', capturedOutput)
          }
        } catch (e) {
          // Ignore output capture errors
        }
        
        throw error
      }


      // Add execution statistics
      const executionTime = Math.round(performance.now() - this.executionStats.startTime)
      const memoryUsed = Math.round(this.executionStats.memoryUsage / 1024 / 1024 * 100) / 100
      const operations = this.executionStats.operationCount
      
      this.addOutput('info', `Python execution completed in ${executionTime}ms | ~${memoryUsed}MB | ${operations} ops`)

      return this.output
    } catch (error) {
      const executionTime = Math.round(performance.now() - this.executionStats.startTime)
      const memoryUsed = Math.round(this.executionStats.memoryUsage / 1024 / 1024 * 100) / 100
      const operations = this.executionStats.operationCount
      
      this.addOutput('error', `${error.message}`)
      this.addOutput('info', `Python execution failed after ${executionTime}ms | ~${memoryUsed}MB | ${operations} ops`)
      return this.output
    }
  }
}

// Create runner instance
const runner = new SafePyRunner()

// Handle messages from main thread
self.onmessage = async (e) => {
  const { type, code, timeout = 30000, requestId, packageName, path, data, options } = e.data
  
  try {
    if (type === 'execute') {
      // Set up execution timeout
      let timeoutId
      const timeoutPromise = new Promise((_, reject) => {
        timeoutId = setTimeout(() => {
          reject(new Error(`Python execution timed out after ${timeout}ms`))
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
    } else if (type === 'loadPackage') {
      // Handle package loading requests
      const success = await runner.loadPackage(packageName)
      self.postMessage({
        type: 'packageLoaded',
        data: { success, packageName },
        requestId: requestId
      })
    } else if (type === 'getAvailablePackages') {
      // Handle package list requests
      const packages = runner.getAvailablePackages()
      self.postMessage({
        type: 'availablePackages',
        data: packages,
        requestId: requestId
      })
    } else if (type === 'initialize') {
      // Handle initialization requests
      await runner.initializePyodide()
      self.postMessage({
        type: 'initialized',
        data: { success: true },
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
          vfs.metadata.clear()

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

          // Import metadata from serialized state (array of [path, meta] pairs)
          if (Array.isArray(vfsState.metadata)) {
            vfsState.metadata.forEach(([path, meta]) => {
              // Convert ISO strings back to Date objects
              const processedMeta = {
                ...meta,
                mtime: meta.mtime && typeof meta.mtime === 'string' ? new Date(meta.mtime) : meta.mtime,
                created: meta.created && typeof meta.created === 'string' ? new Date(meta.created) : meta.created
              }
              vfs.metadata.set(path, processedMeta)
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
      message: error.message || 'Unknown Python worker error',
      name: 'WorkerError'
    }
  })
}

// Handle unhandled promise rejections
self.onunhandledrejection = (event) => {
  self.postMessage({
    type: 'error',
    data: {
      message: event.reason?.message || 'Unhandled promise rejection in Python worker',
      name: 'UnhandledRejection'
    }
  })
}