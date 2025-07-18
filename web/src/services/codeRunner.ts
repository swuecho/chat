/**
 * Code Runner Service
 * Handles execution of JavaScript code in a safe Web Worker environment
 */

export interface ExecutionResult {
  id: string
  type: 'log' | 'error' | 'return' | 'stdout' | 'warn' | 'info' | 'debug' | 'canvas' | 'matplotlib'
  content: string
  timestamp: string
  execution_time_ms?: number
}

export interface LibraryInfo {
  [key: string]: string
}

export interface CanvasOperation {
  type: 'canvas'
  width: number
  height: number
  operations: any[]
}

export interface ExecutionResponse {
  results: ExecutionResult[]
  status: 'success' | 'error' | 'timeout'
  execution_time_ms: number
}

export class CodeRunner {
  private jsWorker: Worker | null = null
  private pyWorker: Worker | null = null
  private requestCounter = 0
  private pendingRequests = new Map<string, {
    resolve: (value: ExecutionResult[]) => void
    reject: (error: Error) => void
    timeout?: NodeJS.Timeout
  }>()

  constructor() {
    this.initializeWorkers()
  }

  private initializeWorkers() {
    try {
      // Initialize JavaScript worker
      this.jsWorker = new Worker('/workers/jsRunner.js')
      this.jsWorker.onmessage = this.handleWorkerMessage.bind(this)
      this.jsWorker.onerror = this.handleWorkerError.bind(this)
      
      // Initialize Python worker
      this.pyWorker = new Worker('/workers/pyRunner.js')
      this.pyWorker.onmessage = this.handleWorkerMessage.bind(this)
      this.pyWorker.onerror = this.handleWorkerError.bind(this)
    } catch (error) {
      console.error('Failed to initialize workers:', error)
    }
  }

  private handleWorkerMessage(e: MessageEvent) {
    const { type, data, requestId } = e.data

    const pendingRequest = this.pendingRequests.get(requestId)
    if (!pendingRequest) return

    // Clear timeout
    if (pendingRequest.timeout) {
      clearTimeout(pendingRequest.timeout)
    }

    // Remove from pending requests
    this.pendingRequests.delete(requestId)

    if (type === 'results') {
      pendingRequest.resolve(data)
    } else if (type === 'libraryLoaded') {
      pendingRequest.resolve(data)
    } else if (type === 'availableLibraries') {
      pendingRequest.resolve(data)
    } else if (type === 'packageLoaded') {
      pendingRequest.resolve(data)
    } else if (type === 'availablePackages') {
      pendingRequest.resolve(data)
    } else if (type === 'initialized') {
      pendingRequest.resolve(data)
    } else if (type === 'error') {
      pendingRequest.reject(new Error(data.message || 'Unknown execution error'))
    }
  }

  private handleWorkerError(error: ErrorEvent) {
    console.error('Worker error:', error)
    
    // Reject all pending requests
    for (const [requestId, request] of this.pendingRequests) {
      if (request.timeout) {
        clearTimeout(request.timeout)
      }
      request.reject(new Error('Worker crashed: ' + error.message))
    }
    this.pendingRequests.clear()

    // Reinitialize workers
    this.dispose()
    this.initializeWorkers()
  }

  private generateRequestId(): string {
    return `req_${++this.requestCounter}_${Date.now()}`
  }

  /**
   * Execute JavaScript code
   */
  async executeJavaScript(code: string, timeoutMs = 10000): Promise<ExecutionResult[]> {
    if (!this.jsWorker) {
      throw new Error('JavaScript worker not available')
    }

    const requestId = this.generateRequestId()

    return new Promise<ExecutionResult[]>((resolve, reject) => {
      // Set up timeout
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId)
        reject(new Error(`Code execution timed out after ${timeoutMs}ms`))
      }, timeoutMs)

      // Store pending request
      this.pendingRequests.set(requestId, {
        resolve,
        reject,
        timeout
      })

      // Send code to worker
      this.jsWorker!.postMessage({
        type: 'execute',
        code,
        timeout: timeoutMs,
        requestId
      })
    })
  }

  /**
   * Execute Python code
   */
  async executePython(code: string, timeoutMs = 30000): Promise<ExecutionResult[]> {
    if (!this.pyWorker) {
      throw new Error('Python worker not available')
    }

    const requestId = this.generateRequestId()

    return new Promise<ExecutionResult[]>((resolve, reject) => {
      // Set up timeout
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId)
        reject(new Error(`Python execution timed out after ${timeoutMs}ms`))
      }, timeoutMs)

      // Store pending request
      this.pendingRequests.set(requestId, {
        resolve,
        reject,
        timeout
      })

      // Send code to worker
      this.pyWorker!.postMessage({
        type: 'execute',
        code,
        timeout: timeoutMs,
        requestId
      })
    })
  }

  /**
   * Execute code based on language
   */
  async execute(language: string, code: string): Promise<ExecutionResult[]> {
    const startTime = performance.now()
    
    try {
      let results: ExecutionResult[]
      
      switch (language.toLowerCase()) {
        case 'javascript':
        case 'js':
        case 'typescript':
        case 'ts':
          results = await this.executeJavaScript(code)
          break
        case 'python':
        case 'py':
          results = await this.executePython(code)
          break
        default:
          throw new Error(`Unsupported language: ${language}`)
      }
      
      // Add execution time to results
      const executionTime = Math.round(performance.now() - startTime)
      results.forEach(result => {
        result.execution_time_ms = executionTime
      })
      
      return results
    } catch (error) {
      const executionTime = Math.round(performance.now() - startTime)
      return [{
        id: Date.now().toString(),
        type: 'error',
        content: error instanceof Error ? error.message : String(error),
        timestamp: new Date().toISOString(),
        execution_time_ms: executionTime
      }]
    }
  }

  /**
   * Check if a language is supported for execution
   */
  isLanguageSupported(language: string): boolean {
    const supportedLanguages = ['javascript', 'js', 'typescript', 'ts', 'python', 'py']
    return supportedLanguages.includes(language.toLowerCase())
  }

  /**
   * Check if a code artifact is executable
   */
  isExecutable(artifact: { type: string; language?: string }): boolean {
    if (artifact.type !== 'code' && artifact.type !== 'executable-code') return false
    if (!artifact.language) return false
    return this.isLanguageSupported(artifact.language)
  }

  /**
   * Load a JavaScript library
   */
  async loadLibrary(libraryName: string): Promise<boolean> {
    if (!this.jsWorker) {
      throw new Error('JavaScript worker not available')
    }

    const requestId = this.generateRequestId()

    return new Promise<boolean>((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId)
        reject(new Error('Library loading timed out'))
      }, 30000) // 30 second timeout for library loading

      this.pendingRequests.set(requestId, {
        resolve: (data: any) => resolve(data.success),
        reject,
        timeout
      })

      this.jsWorker!.postMessage({
        type: 'loadLibrary',
        libraryName,
        requestId
      })
    })
  }

  /**
   * Get available libraries
   */
  async getAvailableLibraries(): Promise<LibraryInfo> {
    if (!this.jsWorker) {
      throw new Error('JavaScript worker not available')
    }

    const requestId = this.generateRequestId()

    return new Promise<LibraryInfo>((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId)
        reject(new Error('Getting libraries timed out'))
      }, 5000)

      this.pendingRequests.set(requestId, {
        resolve,
        reject,
        timeout
      })

      this.jsWorker!.postMessage({
        type: 'getAvailableLibraries',
        requestId
      })
    })
  }

  /**
   * Load a Python package
   */
  async loadPythonPackage(packageName: string): Promise<boolean> {
    if (!this.pyWorker) {
      throw new Error('Python worker not available')
    }

    const requestId = this.generateRequestId()

    return new Promise<boolean>((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId)
        reject(new Error('Package loading timed out'))
      }, 60000) // 60 second timeout for Python package loading

      this.pendingRequests.set(requestId, {
        resolve: (data: any) => resolve(data.success),
        reject,
        timeout
      })

      this.pyWorker!.postMessage({
        type: 'loadPackage',
        packageName,
        requestId
      })
    })
  }

  /**
   * Get available Python packages
   */
  async getAvailablePythonPackages(): Promise<LibraryInfo> {
    if (!this.pyWorker) {
      throw new Error('Python worker not available')
    }

    const requestId = this.generateRequestId()

    return new Promise<LibraryInfo>((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId)
        reject(new Error('Getting packages timed out'))
      }, 5000)

      this.pendingRequests.set(requestId, {
        resolve,
        reject,
        timeout
      })

      this.pyWorker!.postMessage({
        type: 'getAvailablePackages',
        requestId
      })
    })
  }

  /**
   * Initialize Python environment
   */
  async initializePython(): Promise<boolean> {
    if (!this.pyWorker) {
      throw new Error('Python worker not available')
    }

    const requestId = this.generateRequestId()

    return new Promise<boolean>((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId)
        reject(new Error('Python initialization timed out'))
      }, 30000) // 30 second timeout for initialization

      this.pendingRequests.set(requestId, {
        resolve: (data: any) => resolve(data.success),
        reject,
        timeout
      })

      this.pyWorker!.postMessage({
        type: 'initialize',
        requestId
      })
    })
  }

  /**
   * Check if canvas output is supported
   */
  isCanvasSupported(): boolean {
    return true
  }

  /**
   * Render canvas operations to actual canvas element
   */
  renderCanvasToElement(canvasData: string, canvasElement: HTMLCanvasElement): boolean {
    try {
      const data = JSON.parse(canvasData) as CanvasOperation
      if (data.type !== 'canvas') return false

      const ctx = canvasElement.getContext('2d')
      if (!ctx) return false

      // Set canvas size
      canvasElement.width = data.width
      canvasElement.height = data.height

      // Clear canvas
      ctx.clearRect(0, 0, data.width, data.height)

      // Execute operations
      for (const op of data.operations) {
        switch (op.type) {
          case 'fillRect':
            if (op.style) ctx.fillStyle = op.style
            ctx.fillRect(op.x, op.y, op.width, op.height)
            break
          case 'strokeRect':
            if (op.style) ctx.strokeStyle = op.style
            if (op.lineWidth) ctx.lineWidth = op.lineWidth
            ctx.strokeRect(op.x, op.y, op.width, op.height)
            break
          case 'beginPath':
            ctx.beginPath()
            break
          case 'closePath':
            ctx.closePath()
            break
          case 'moveTo':
            ctx.moveTo(op.x, op.y)
            break
          case 'lineTo':
            ctx.lineTo(op.x, op.y)
            break
          case 'arc':
            ctx.arc(op.x, op.y, op.radius, op.startAngle, op.endAngle)
            break
          case 'fill':
            if (op.style) ctx.fillStyle = op.style
            ctx.fill()
            break
          case 'stroke':
            if (op.style) ctx.strokeStyle = op.style
            if (op.lineWidth) ctx.lineWidth = op.lineWidth
            ctx.stroke()
            break
          case 'clearRect':
            ctx.clearRect(op.x, op.y, op.width, op.height)
            break
          case 'fillText':
            if (op.style) ctx.fillStyle = op.style
            ctx.fillText(op.text, op.x, op.y)
            break
          case 'strokeText':
            if (op.style) ctx.strokeStyle = op.style
            ctx.strokeText(op.text, op.x, op.y)
            break
        }
      }

      return true
    } catch (error) {
      console.error('Failed to render canvas:', error)
      return false
    }
  }

  /**
   * Render matplotlib plot to image element
   */
  renderMatplotlibToElement(plotData: string, imgElement: HTMLImageElement): boolean {
    try {
      const data = JSON.parse(plotData)
      if (data.type !== 'matplotlib') return false

      imgElement.src = `data:image/png;base64,${data.data}`
      return true
    } catch (error) {
      console.error('Failed to render matplotlib plot:', error)
      return false
    }
  }

  /**
   * Get execution capabilities info
   */
  getCapabilities() {
    return {
      javascript: {
        supported: true,
        features: [
          'console output', 
          'return values', 
          'error handling', 
          'timeouts',
          'library loading',
          'canvas graphics',
          'enhanced APIs'
        ],
        limitations: ['no DOM access', 'no direct network requests', 'no file system'],
        libraries: [
          'lodash', 'd3', 'chart.js', 'moment', 'axios', 'rxjs', 'p5', 'three', 'fabric'
        ]
      },
      python: {
        supported: true,
        features: [
          'print output',
          'matplotlib plots',
          'scientific computing',
          'data analysis',
          'package loading',
          'error handling',
          'timeouts',
          'memory monitoring'
        ],
        limitations: ['no file system', 'no direct network requests', 'limited to Pyodide packages'],
        packages: [
          'numpy', 'pandas', 'matplotlib', 'scipy', 'scikit-learn', 'requests', 
          'beautifulsoup4', 'pillow', 'sympy', 'networkx', 'seaborn', 'plotly', 'bokeh', 'altair'
        ]
      }
    }
  }

  /**
   * Dispose of resources
   */
  dispose() {
    // Clear all pending requests
    for (const [requestId, request] of this.pendingRequests) {
      if (request.timeout) {
        clearTimeout(request.timeout)
      }
      request.reject(new Error('CodeRunner disposed'))
    }
    this.pendingRequests.clear()

    // Terminate workers
    if (this.jsWorker) {
      this.jsWorker.terminate()
      this.jsWorker = null
    }
    if (this.pyWorker) {
      this.pyWorker.terminate()
      this.pyWorker = null
    }
  }
}

// Global instance for sharing across components
let globalCodeRunner: CodeRunner | null = null

export function getCodeRunner(): CodeRunner {
  if (!globalCodeRunner) {
    globalCodeRunner = new CodeRunner()
  }
  return globalCodeRunner
}

export function disposeCodeRunner() {
  if (globalCodeRunner) {
    globalCodeRunner.dispose()
    globalCodeRunner = null
  }
}