# Code Runner Feature Implementation Plan

## Overview

This document outlines the implementation plan for adding interactive code execution capabilities to the chat application's artifact system. The code runner will allow users to execute JavaScript, Python, and other languages directly within chat artifacts, providing real-time output and interactive programming experiences.

## Goals

- **Educational**: Enable live coding tutorials and learning experiences
- **Prototyping**: Quick algorithm testing and experimentation
- **Data Visualization**: Interactive charts, graphs, and visual demonstrations
- **Development**: Code examples that users can run and modify

## Architecture Decision

### Selected Approach: Browser-Based Execution

**Reasoning**:
- No server resources required
- Real-time execution with minimal latency
- Scales with user count (computation on client)
- Supports rich DOM manipulation and visualization
- Easier deployment and maintenance

**Supported Languages**:
1. **JavaScript/TypeScript** - Native browser execution in Web Workers
2. **Python** - Pyodide (Python in WebAssembly)
3. **Future**: HTML/CSS live preview, SQL (via SQLite WASM)

## Technical Implementation

### 1. Database Schema Extensions

```sql
-- Add execution results to chat messages
ALTER TABLE chat_message ADD COLUMN execution_results JSONB DEFAULT '[]';

-- Index for searching executable artifacts
CREATE INDEX idx_chat_message_execution ON chat_message 
USING GIN (execution_results) 
WHERE execution_results != '[]';
```

**Execution Result Structure**:
```json
{
  "artifact_uuid": "string",
  "execution_id": "string",
  "timestamp": "ISO8601",
  "language": "javascript|python",
  "output": [
    {
      "type": "log|error|return|stdout",
      "content": "string",
      "timestamp": "ISO8601"
    }
  ],
  "execution_time_ms": "number",
  "status": "success|error|timeout"
}
```

### 2. Frontend Architecture

#### Enhanced Artifact Types

**New Artifact Type**: `executable-code`
- Extends existing `code` artifacts
- Includes execution metadata
- Supports multiple output formats

```typescript
// web/src/typings/chat.d.ts
interface ExecutableArtifact extends Artifact {
  type: 'executable-code'
  language: 'javascript' | 'python' | 'typescript'
  isExecutable: true
  executionResults?: ExecutionResult[]
}

interface ExecutionResult {
  id: string
  type: 'log' | 'error' | 'return' | 'stdout'
  content: string
  timestamp: string
}
```

#### Core Components

**1. Enhanced ArtifactViewer.vue**

```vue
<template>
  <div class="artifact-viewer" :class="{ 'executable': isExecutable }">
    <!-- Existing code display -->
    <div class="artifact-header">
      <span class="artifact-title">{{ artifact.title }}</span>
      <div class="artifact-actions">
        <button v-if="isExecutable" 
                @click="runCode" 
                :disabled="running"
                class="run-button">
          <Icon name="play" v-if="!running" />
          <Icon name="spinner" v-else spinning />
          {{ running ? 'Running...' : 'Run Code' }}
        </button>
        <button @click="toggleExpanded">
          {{ expanded ? 'Collapse' : 'Expand' }}
        </button>
        <button @click="copyContent">Copy</button>
      </div>
    </div>

    <div v-if="expanded" class="artifact-content">
      <!-- Code editor/viewer -->
      <div class="code-content">
        <CodeEditor 
          v-if="editable"
          v-model="editableContent"
          :language="artifact.language"
          :readonly="!editable"
        />
        <CodeViewer 
          v-else
          :content="artifact.content"
          :language="artifact.language"
        />
      </div>

      <!-- Execution controls -->
      <div v-if="isExecutable" class="execution-controls">
        <div class="control-bar">
          <button @click="runCode" :disabled="running" class="primary">
            Run Code
          </button>
          <button @click="clearOutput" :disabled="!hasOutput">
            Clear Output
          </button>
          <button @click="toggleEditor" class="secondary">
            {{ editable ? 'View Mode' : 'Edit Mode' }}
          </button>
        </div>
        
        <div class="execution-info" v-if="lastExecution">
          <span class="execution-time">
            Executed in {{ lastExecution.execution_time_ms }}ms
          </span>
          <span class="execution-status" :class="lastExecution.status">
            {{ lastExecution.status }}
          </span>
        </div>
      </div>

      <!-- Output area -->
      <div v-if="hasOutput" class="execution-output">
        <div class="output-header">
          <span class="output-title">Output</span>
          <button @click="clearOutput" class="clear-btn">×</button>
        </div>
        <div class="output-content">
          <div v-for="result in currentOutput" 
               :key="result.id" 
               class="output-line"
               :class="result.type">
            <span class="output-type">{{ result.type }}</span>
            <span class="output-content">{{ result.content }}</span>
            <span class="output-time">{{ formatTime(result.timestamp) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { CodeRunner } from '@/services/codeRunner'
import { ExecutableArtifact, ExecutionResult } from '@/typings/chat'

const props = defineProps<{
  artifact: ExecutableArtifact
}>()

const codeRunner = new CodeRunner()
const running = ref(false)
const expanded = ref(false)
const editable = ref(false)
const editableContent = ref(props.artifact.content)
const currentOutput = ref<ExecutionResult[]>([])

const isExecutable = computed(() => 
  props.artifact.type === 'executable-code' && 
  ['javascript', 'python', 'typescript'].includes(props.artifact.language)
)

const hasOutput = computed(() => currentOutput.value.length > 0)

const lastExecution = computed(() => {
  const results = props.artifact.executionResults || []
  return results[results.length - 1]
})

async function runCode() {
  if (running.value) return
  
  running.value = true
  currentOutput.value = []
  
  try {
    const code = editable.value ? editableContent.value : props.artifact.content
    const results = await codeRunner.execute(props.artifact.language, code)
    currentOutput.value = results
    
    // Emit execution event for parent components
    emit('execution-complete', {
      artifact: props.artifact,
      results: results
    })
  } catch (error) {
    currentOutput.value = [{
      id: Date.now().toString(),
      type: 'error',
      content: error.message,
      timestamp: new Date().toISOString()
    }]
  } finally {
    running.value = false
  }
}

function clearOutput() {
  currentOutput.value = []
}

function toggleEditor() {
  editable.value = !editable.value
}

function formatTime(timestamp: string) {
  return new Date(timestamp).toLocaleTimeString()
}

onMounted(() => {
  // Auto-expand executable artifacts
  if (isExecutable.value) {
    expanded.value = true
  }
})
</script>
```

**2. Code Runner Service**

```typescript
// web/src/services/codeRunner.ts
import { ExecutionResult } from '@/typings/chat'

export class CodeRunner {
  private jsWorker: Worker | null = null
  private pythonRunner: PythonRunner | null = null

  async execute(language: string, code: string): Promise<ExecutionResult[]> {
    const startTime = performance.now()
    
    try {
      let results: ExecutionResult[]
      
      switch (language) {
        case 'javascript':
        case 'typescript':
          results = await this.executeJavaScript(code)
          break
        case 'python':
          results = await this.executePython(code)
          break
        default:
          throw new Error(`Unsupported language: ${language}`)
      }
      
      // Add execution time to results
      const executionTime = performance.now() - startTime
      results.forEach(result => {
        result.execution_time_ms = executionTime
      })
      
      return results
    } catch (error) {
      return [{
        id: Date.now().toString(),
        type: 'error',
        content: error.message,
        timestamp: new Date().toISOString(),
        execution_time_ms: performance.now() - startTime
      }]
    }
  }

  private async executeJavaScript(code: string): Promise<ExecutionResult[]> {
    return new Promise((resolve, reject) => {
      if (!this.jsWorker) {
        this.jsWorker = new Worker('/workers/jsRunner.js')
      }

      const timeout = setTimeout(() => {
        this.jsWorker?.terminate()
        this.jsWorker = null
        reject(new Error('Code execution timed out'))
      }, 10000) // 10 second timeout

      this.jsWorker.onmessage = (e) => {
        clearTimeout(timeout)
        const { type, data } = e.data
        
        if (type === 'results') {
          resolve(data)
        } else if (type === 'error') {
          reject(new Error(data.message))
        }
      }

      this.jsWorker.onerror = (error) => {
        clearTimeout(timeout)
        reject(error)
      }

      this.jsWorker.postMessage({ code, timeout: 10000 })
    })
  }

  private async executePython(code: string): Promise<ExecutionResult[]> {
    if (!this.pythonRunner) {
      this.pythonRunner = new PythonRunner()
      await this.pythonRunner.initialize()
    }

    return this.pythonRunner.execute(code)
  }

  dispose() {
    if (this.jsWorker) {
      this.jsWorker.terminate()
      this.jsWorker = null
    }
    if (this.pythonRunner) {
      this.pythonRunner.dispose()
      this.pythonRunner = null
    }
  }
}
```

**3. JavaScript Worker**

```typescript
// public/workers/jsRunner.js
class SafeJSRunner {
  constructor() {
    this.output = []
    this.setupConsole()
  }

  setupConsole() {
    // Capture console methods
    this.console = {
      log: (...args) => this.addOutput('log', args.join(' ')),
      error: (...args) => this.addOutput('error', args.join(' ')),
      warn: (...args) => this.addOutput('warn', args.join(' ')),
      info: (...args) => this.addOutput('info', args.join(' '))
    }
  }

  addOutput(type, content) {
    this.output.push({
      id: Date.now().toString() + Math.random(),
      type: type,
      content: String(content),
      timestamp: new Date().toISOString()
    })
  }

  async execute(code) {
    this.output = []
    
    try {
      // Create safe execution context
      const safeGlobals = {
        console: this.console,
        Math: Math,
        Date: Date,
        Array: Array,
        Object: Object,
        String: String,
        Number: Number,
        Boolean: Boolean,
        JSON: JSON,
        setTimeout: (fn, ms) => setTimeout(fn, Math.min(ms, 5000)),
        setInterval: (fn, ms) => setInterval(fn, Math.max(ms, 100))
      }

      // Execute code in safe context
      const result = new Function(
        ...Object.keys(safeGlobals), 
        `
        "use strict";
        ${code}
        `
      )(...Object.values(safeGlobals))

      // Add return value if exists
      if (result !== undefined) {
        this.addOutput('return', JSON.stringify(result, null, 2))
      }

      return this.output
    } catch (error) {
      this.addOutput('error', error.message)
      return this.output
    }
  }
}

const runner = new SafeJSRunner()

self.onmessage = async (e) => {
  const { code, timeout } = e.data
  
  try {
    const results = await runner.execute(code)
    self.postMessage({ type: 'results', data: results })
  } catch (error) {
    self.postMessage({ 
      type: 'error', 
      data: { message: error.message } 
    })
  }
}
```

**4. Python Runner (Pyodide)**

```typescript
// web/src/services/pythonRunner.ts
export class PythonRunner {
  private pyodide: any = null
  private initialized = false

  async initialize() {
    if (this.initialized) return

    // Dynamic import to avoid bundling
    const { loadPyodide } = await import('pyodide')
    
    this.pyodide = await loadPyodide({
      indexURL: 'https://cdn.jsdelivr.net/pyodide/'
    })

    // Install common packages
    await this.pyodide.loadPackage(['numpy', 'matplotlib', 'pandas'])
    
    this.initialized = true
  }

  async execute(code: string): Promise<ExecutionResult[]> {
    if (!this.initialized) {
      await this.initialize()
    }

    const output: ExecutionResult[] = []
    
    try {
      // Capture stdout
      this.pyodide.runPython(`
        import sys
        from io import StringIO
        
        # Capture stdout
        old_stdout = sys.stdout
        sys.stdout = captured_output = StringIO()
      `)

      // Execute user code
      const result = this.pyodide.runPython(code)

      // Get captured output
      const stdout = this.pyodide.runPython(`
        sys.stdout = old_stdout
        captured_output.getvalue()
      `)

      // Add stdout if any
      if (stdout.trim()) {
        output.push({
          id: Date.now().toString(),
          type: 'stdout',
          content: stdout,
          timestamp: new Date().toISOString()
        })
      }

      // Add return value if exists
      if (result !== undefined && result !== null) {
        output.push({
          id: Date.now().toString() + '1',
          type: 'return',
          content: String(result),
          timestamp: new Date().toISOString()
        })
      }

      return output
    } catch (error) {
      return [{
        id: Date.now().toString(),
        type: 'error',
        content: error.message,
        timestamp: new Date().toISOString()
      }]
    }
  }

  dispose() {
    // Cleanup if needed
    this.pyodide = null
    this.initialized = false
  }
}
```

### 3. Backend Integration

#### Artifact Detection Enhancement

```go
// api/chat_main_service.go
func extractArtifacts(content string) []Artifact {
    // Existing artifact extraction logic...
    
    // Add executable code detection
    executableCodeRegex := regexp.MustCompile(`(?s)```(\w+)\s*<!--\s*executable:\s*([^>]+)\s*-->\s*\n(.*?)\n\s*````)
    executableMatches := executableCodeRegex.FindAllStringSubmatch(content, -1)
    
    for _, match := range executableMatches {
        if len(match) >= 4 {
            language := match[1]
            title := strings.TrimSpace(match[2])
            code := strings.TrimSpace(match[3])
            
            // Only certain languages are executable
            if isExecutableLanguage(language) {
                artifacts = append(artifacts, Artifact{
                    UUID:     generateUUID(),
                    Type:     "executable-code",
                    Title:    title,
                    Content:  code,
                    Language: language,
                })
            }
        }
    }
    
    return artifacts
}

func isExecutableLanguage(lang string) bool {
    executableLangs := []string{"javascript", "python", "typescript", "js", "py", "ts"}
    for _, execLang := range executableLangs {
        if strings.EqualFold(lang, execLang) {
            return true
        }
    }
    return false
}
```

#### Execution Results Storage

```go
// api/models.go
type ExecutionResult struct {
    ArtifactUUID     string                 `json:"artifact_uuid"`
    ExecutionID      string                 `json:"execution_id"`
    Timestamp        time.Time              `json:"timestamp"`
    Language         string                 `json:"language"`
    Output           []ExecutionOutputLine  `json:"output"`
    ExecutionTimeMs  int64                  `json:"execution_time_ms"`
    Status           string                 `json:"status"` // success, error, timeout
}

type ExecutionOutputLine struct {
    Type      string    `json:"type"`      // log, error, return, stdout
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

#### API Endpoints

```go
// api/chat_execution_handler.go
func (h *ChatHandler) SaveExecutionResult(w http.ResponseWriter, r *http.Request) {
    // POST /api/chat/executions
    // Save execution results to database
}

func (h *ChatHandler) GetExecutionHistory(w http.ResponseWriter, r *http.Request) {
    // GET /api/chat/executions/:messageUUID
    // Get execution history for a message
}
```

### 4. Security Considerations

#### JavaScript Security
- **Web Workers**: No DOM access, isolated execution
- **Timeout Limits**: 10-second maximum execution time
- **Memory Limits**: Worker termination prevents memory leaks
- **API Restrictions**: No network access, limited global objects
- **Safe Globals**: Whitelist of allowed APIs only

#### Python Security
- **Pyodide Sandbox**: Runs in WebAssembly, isolated from system
- **Package Restrictions**: Only pre-approved packages
- **Resource Limits**: Memory and execution time constraints
- **No File System**: No access to local files

#### General Security
- **Content Validation**: Sanitize all user inputs
- **Rate Limiting**: Limit executions per user/session
- **Error Handling**: Safe error messages without system info
- **Audit Logging**: Track all code executions

### 5. Performance Optimizations

#### Loading Strategy
- **Lazy Loading**: Load runners only when needed
- **Worker Pooling**: Reuse workers for multiple executions
- **Caching**: Cache Pyodide and common libraries
- **Progressive Loading**: Load features incrementally

#### Resource Management
- **Memory Monitoring**: Track and limit memory usage
- **Cleanup**: Proper disposal of workers and resources
- **Debouncing**: Limit rapid execution requests
- **Background Loading**: Preload runners in background

### 6. User Experience Enhancements

#### Visual Feedback
- **Loading States**: Show progress during execution
- **Status Indicators**: Success/error/timeout states
- **Execution Time**: Display performance metrics
- **Output Formatting**: Syntax highlighting for results

#### Interactive Features
- **Code Editing**: Inline editing with syntax highlighting
- **Auto-completion**: Basic code completion
- **Error Highlighting**: Visual error indicators
- **Execution History**: Show previous runs

### 7. Implementation Phases

#### Phase 1: Basic JavaScript Runner (Week 1)
- [ ] Web Worker implementation
- [ ] Basic JavaScript execution
- [ ] Console output capture
- [ ] Error handling
- [ ] UI integration

#### Phase 2: Enhanced JavaScript (Week 2)
- [ ] Advanced security measures
- [ ] Timeout and memory controls
- [ ] Better error reporting
- [ ] Execution history

#### Phase 3: Python Support (Week 3)
- [ ] Pyodide integration
- [ ] Python execution environment
- [ ] Package management
- [ ] Performance optimization

#### Phase 4: Advanced Features (Week 4)
- [ ] Code editing capabilities
- [ ] Library loading
- [ ] Visualization support
- [ ] Export/sharing features

### 8. Testing Strategy

#### Unit Tests
- Code execution accuracy
- Security boundary testing
- Error handling scenarios
- Performance benchmarks

#### Integration Tests
- Full artifact workflow
- Database persistence
- API endpoint testing
- UI interaction testing

#### Security Tests
- Sandbox escape attempts
- Resource exhaustion tests
- Malicious code detection
- Cross-site scripting prevention

### 9. Deployment Considerations

#### Bundle Size
- **Pyodide**: ~3MB initial download
- **Workers**: Minimal overhead
- **Lazy Loading**: Only load when needed

#### Browser Compatibility
- **Web Workers**: Supported in all modern browsers
- **WebAssembly**: Required for Python (95%+ browser support)
- **Graceful Degradation**: Fallback for unsupported browsers

#### CDN Strategy
- **Pyodide CDN**: Use official CDN for reliability
- **Worker Files**: Serve from application domain
- **Caching**: Aggressive caching for static assets

### 10. Future Enhancements

#### Additional Languages
- **SQL**: SQLite WebAssembly for database queries
- **R**: R WebAssembly for statistical computing
- **Go**: TinyGo WebAssembly compilation
- **Rust**: Rust WebAssembly support

#### Advanced Features
- **Collaborative Editing**: Multiple users editing same code
- **Version Control**: Track code changes over time
- **Package Management**: Install custom packages
- **Debugging Tools**: Step-through debugging
- **Performance Profiling**: Execution analysis

#### Integration Features
- **GitHub Integration**: Save/load from repositories
- **Notebook Export**: Export to Jupyter notebooks
- **Sharing**: Public executable artifact galleries
- **Embedding**: Embed in external websites

## Success Metrics

- **Execution Speed**: JavaScript < 100ms, Python < 1s for simple code
- **Security**: Zero successful sandbox escapes
- **Reliability**: 99.9% successful executions
- **User Adoption**: 50% of code artifacts become executable
- **Performance**: No impact on chat loading time

## Conclusion

The Code Runner feature will transform the chat application into a powerful interactive development environment. By supporting both JavaScript and Python execution with strong security measures, it opens up possibilities for education, prototyping, and data analysis directly within the chat interface.

The phased implementation approach ensures steady progress while maintaining system stability and security. The browser-based architecture provides excellent scalability and performance while keeping infrastructure costs minimal.