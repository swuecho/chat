/**
 * Python Code Runner Web Worker
 * Executes Python code using Pyodide in a safe, isolated environment
 */

// Global Pyodide instance
let pyodide = null
let isInitialized = false
let initializationPromise = null

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

  // Initialize Pyodide
  async initializePyodide() {
    if (isInitialized) return pyodide
    if (initializationPromise) return initializationPromise

    initializationPromise = new Promise(async (resolve, reject) => {
      try {
        this.addOutput('info', 'Initializing Python environment...')
        
        // Load Pyodide
        importScripts('https://cdn.jsdelivr.net/pyodide/v0.24.1/full/pyodide.js')
        
        pyodide = await loadPyodide({
          indexURL: 'https://cdn.jsdelivr.net/pyodide/v0.24.1/full/',
          stdout: this.outputCapture.write.bind(this.outputCapture),
          stderr: this.outputCapture.write.bind(this.outputCapture)
        })
        
        // Set up matplotlib backend for web
        await pyodide.runPython(`
import sys
import io
import base64
import json
from contextlib import redirect_stdout, redirect_stderr

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

# Custom print function
def custom_print(*args, **kwargs):
    output = io.StringIO()
    print(*args, file=output, **kwargs)
    result = output.getvalue()
    _output_capture.write(result)
    return result

# Replace built-in print
__builtins__['print'] = custom_print
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

  // Load a Python package
  async loadPackage(packageName) {
    if (!pyodide) {
      await this.initializePyodide()
    }

    if (this.loadedPackages.has(packageName)) {
      return true
    }

    try {
      this.addOutput('info', `Loading Python package: ${packageName}`)
      await pyodide.loadPackage(packageName)
      this.loadedPackages.add(packageName)
      this.addOutput('info', `Package '${packageName}' loaded successfully`)
      return true
    } catch (error) {
      this.addOutput('error', `Failed to load package '${packageName}': ${error.message}`)
      return false
    }
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
        await this.loadPackage(pkg)
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
      const result = await pyodide.runPython(`
import sys
import traceback

try:
    # Execute user code
    exec(${JSON.stringify(code)})
    
    # Get captured output
    captured_output = _output_capture.getvalue()
    
    # Clear output buffer
    _output_capture.output = []
    
    # Return captured output
    captured_output
except Exception as e:
    error_msg = f"{type(e).__name__}: {str(e)}"
    traceback.print_exc()
    raise Exception(error_msg)
`)

      // Add any captured output
      if (result && result.trim()) {
        this.addOutput('stdout', result)
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
  const { type, code, timeout = 30000, requestId, packageName } = e.data
  
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