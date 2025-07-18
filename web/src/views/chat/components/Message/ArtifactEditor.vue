<template>
  <div class="artifact-editor">
    <div class="editor-header">
      <div class="editor-title">
        <Icon :icon="getLanguageIcon(language)" class="language-icon" />
        <span>{{ title || 'Code Editor' }}</span>
        <NBadge :value="language" type="info" />
      </div>
      <div class="editor-actions">
        <NButton size="small" @click="showTemplates" type="tertiary">
          <template #icon>
            <Icon icon="ri:archive-line" />
          </template>
          Templates
        </NButton>
        <NButton size="small" @click="showHistory" type="tertiary">
          <template #icon>
            <Icon icon="ri:history-line" />
          </template>
          History
        </NButton>
        <NButton size="small" @click="formatCode" type="tertiary">
          <template #icon>
            <Icon icon="ri:code-line" />
          </template>
          Format
        </NButton>
        <NButton size="small" @click="saveAsTemplate" type="tertiary">
          <template #icon>
            <Icon icon="ri:save-line" />
          </template>
          Save Template
        </NButton>
      </div>
    </div>

    <div class="editor-content">
      <!-- Main editor -->
      <div class="editor-main">
        <div class="editor-toolbar">
          <div class="toolbar-left">
            <NButton size="tiny" @click="undo" :disabled="!canUndo">
              <template #icon>
                <Icon icon="ri:arrow-left-line" />
              </template>
              Undo
            </NButton>
            <NButton size="tiny" @click="redo" :disabled="!canRedo">
              <template #icon>
                <Icon icon="ri:arrow-right-line" />
              </template>
              Redo
            </NButton>
            <div class="toolbar-separator"></div>
            <NButton size="tiny" @click="insertSnippet('function')">
              <template #icon>
                <Icon icon="ri:function-line" />
              </template>
              Function
            </NButton>
            <NButton size="tiny" @click="insertSnippet('loop')">
              <template #icon>
                <Icon icon="ri:loop-left-line" />
              </template>
              Loop
            </NButton>
            <NButton size="tiny" @click="insertSnippet('class')">
              <template #icon>
                <Icon icon="ri:code-box-line" />
              </template>
              Class
            </NButton>
          </div>
          <div class="toolbar-right">
            <span class="cursor-position">Line {{ cursorLine }}, Col {{ cursorColumn }}</span>
            <span class="word-count">{{ wordCount }} words</span>
          </div>
        </div>

        <div class="editor-textarea-container">
          <textarea
            ref="editorTextarea"
            v-model="code"
            :placeholder="getPlaceholder(language)"
            class="editor-textarea"
            :rows="editorRows"
            @input="onCodeChange"
            @keydown="onKeyDown"
            @click="updateCursorPosition"
            @keyup="updateCursorPosition"
            @scroll="syncLineNumbers"
            spellcheck="false"
          ></textarea>
          <div class="line-numbers" ref="lineNumbers" @scroll="syncEditor">
            <div v-for="n in lineCount" :key="n" class="line-number">{{ n }}</div>
          </div>
        </div>

        <div class="editor-footer">
          <div class="footer-info">
            <span class="char-count">{{ code.length }} characters</span>
            <span class="line-count">{{ lineCount }} lines</span>
            <span v-if="lastModified" class="last-modified">
              Modified {{ formatRelativeTime(lastModified) }}
            </span>
          </div>
          <div class="footer-actions">
            <NButton size="tiny" @click="clearCode" type="error" ghost>
              <template #icon>
                <Icon icon="ri:delete-bin-line" />
              </template>
              Clear
            </NButton>
            <NButton size="tiny" @click="copyCode" type="primary" ghost>
              <template #icon>
                <Icon icon="ri:file-copy-line" />
              </template>
              Copy
            </NButton>
          </div>
        </div>
      </div>

      <!-- Side panels -->
      <div class="editor-panels">
        <!-- Templates panel -->
        <div v-if="showTemplatesPanel" class="panel templates-panel">
          <div class="panel-header">
            <h3>Code Templates</h3>
            <NButton size="tiny" text @click="showTemplatesPanel = false">
              <Icon icon="ri:close-line" />
            </NButton>
          </div>
          <div class="panel-content">
            <div class="template-search">
              <NInput v-model:value="templateSearch" placeholder="Search templates..." size="small">
                <template #prefix>
                  <Icon icon="ri:search-line" />
                </template>
              </NInput>
            </div>
            <div class="template-categories">
              <div v-for="category in filteredCategories" :key="category.id" class="category">
                <div class="category-header" @click="toggleCategory(category.id)">
                  <Icon :icon="category.icon" :style="{ color: category.color }" />
                  <span>{{ category.name }}</span>
                  <span class="template-count">({{ category.templates.length }})</span>
                  <Icon :icon="expandedCategories.has(category.id) ? 'ri:arrow-up-s-line' : 'ri:arrow-down-s-line'" />
                </div>
                <div v-if="expandedCategories.has(category.id)" class="category-templates">
                  <div v-for="template in category.templates" :key="template.id" 
                       class="template-item" 
                       @click="insertTemplate(template)">
                    <div class="template-info">
                      <div class="template-name">{{ template.name }}</div>
                      <div class="template-description">{{ template.description }}</div>
                      <div class="template-tags">
                        <NBadge v-for="tag in template.tags.slice(0, 3)" :key="tag" :value="tag" size="small" />
                      </div>
                    </div>
                    <div class="template-meta">
                      <NBadge :value="template.difficulty" :type="getDifficultyType(template.difficulty)" size="small" />
                      <span class="usage-count">{{ template.usageCount }} uses</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- History panel -->
        <div v-if="showHistoryPanel" class="panel history-panel">
          <div class="panel-header">
            <h3>Execution History</h3>
            <NButton size="tiny" text @click="showHistoryPanel = false">
              <Icon icon="ri:close-line" />
            </NButton>
          </div>
          <div class="panel-content">
            <div class="history-search">
              <NInput v-model:value="historySearch" placeholder="Search history..." size="small">
                <template #prefix>
                  <Icon icon="ri:search-line" />
                </template>
              </NInput>
            </div>
            <div class="history-filters">
              <NSelect v-model:value="historyFilter" :options="historyFilterOptions" size="small" />
            </div>
            <div class="history-list">
              <div v-for="entry in filteredHistory" :key="entry.id" 
                   class="history-item"
                   @click="loadFromHistory(entry)">
                <div class="history-info">
                  <div class="history-time">{{ formatTime(entry.timestamp) }}</div>
                  <div class="history-preview">{{ entry.code.substring(0, 60) }}...</div>
                  <div class="history-meta">
                    <NBadge :value="entry.language" size="small" />
                    <NBadge :value="entry.success ? 'success' : 'error'" 
                           :type="entry.success ? 'success' : 'error'" size="small" />
                    <span class="execution-time">{{ entry.executionTime }}ms</span>
                  </div>
                </div>
                <div class="history-tags">
                  <NBadge v-for="tag in entry.tags.slice(0, 2)" :key="tag" :value="tag" size="small" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Template Save Modal -->
    <NModal v-model:show="showSaveTemplateModal" :mask-closable="false">
      <NCard style="width: 600px" title="Save as Template">
        <div class="save-template-form">
          <NFormItem label="Template Name">
            <NInput v-model:value="newTemplate.name" placeholder="Enter template name" />
          </NFormItem>
          <NFormItem label="Description">
            <NInput v-model:value="newTemplate.description" placeholder="Enter description" type="textarea" />
          </NFormItem>
          <NFormItem label="Category">
            <NSelect v-model:value="newTemplate.category" :options="categoryOptions" />
          </NFormItem>
          <NFormItem label="Difficulty">
            <NSelect v-model:value="newTemplate.difficulty" :options="difficultyOptions" />
          </NFormItem>
          <NFormItem label="Tags">
            <NInput v-model:value="newTemplate.tagsInput" placeholder="Enter tags separated by commas" />
          </NFormItem>
        </div>
        <template #footer>
          <div class="modal-actions">
            <NButton @click="showSaveTemplateModal = false">Cancel</NButton>
            <NButton type="primary" @click="saveTemplate">Save Template</NButton>
          </div>
        </template>
      </NCard>
    </NModal>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { NButton, NInput, NSelect, NModal, NCard, NFormItem, NBadge, useMessage } from 'naive-ui'
import { Icon } from '@iconify/vue'
import { useCodeTemplates } from '@/services/codeTemplates'
import { useExecutionHistory } from '@/services/executionHistory'

interface Props {
  modelValue: string
  language: string
  title?: string
  artifactId?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: string]
  'run': []
}>()

const message = useMessage()
const { categories, searchTemplates, addTemplate, incrementUsage } = useCodeTemplates()
const { getArtifactHistory, searchHistory } = useExecutionHistory()

// Editor state
const code = ref(props.modelValue)
const editorTextarea = ref<HTMLTextAreaElement>()
const lineNumbers = ref<HTMLDivElement>()
const editorRows = ref(20)
const lastModified = ref<string>()

// History state
const history = ref<string[]>([])
const historyIndex = ref(-1)
const maxHistorySize = 50

// UI state
const showTemplatesPanel = ref(false)
const showHistoryPanel = ref(false)
const showSaveTemplateModal = ref(false)
const expandedCategories = ref<Set<string>>(new Set())

// Search and filter state
const templateSearch = ref('')
const historySearch = ref('')
const historyFilter = ref('all')

// Cursor state
const cursorLine = ref(1)
const cursorColumn = ref(1)

// Template creation state
const newTemplate = ref({
  name: '',
  description: '',
  category: 'basics',
  difficulty: 'beginner' as const,
  tagsInput: ''
})

// Computed properties
const lineCount = computed(() => code.value.split('\n').length)
const wordCount = computed(() => code.value.trim().split(/\s+/).filter(word => word.length > 0).length)
const canUndo = computed(() => historyIndex.value > 0)
const canRedo = computed(() => historyIndex.value < history.value.length - 1)

const filteredCategories = computed(() => {
  const languageCategories = categories.value.filter(cat => 
    cat.templates.some(t => t.language === props.language)
  )
  
  if (!templateSearch.value) return languageCategories
  
  return languageCategories.map(category => ({
    ...category,
    templates: category.templates.filter(t => 
      t.language === props.language &&
      (t.name.toLowerCase().includes(templateSearch.value.toLowerCase()) ||
       t.description.toLowerCase().includes(templateSearch.value.toLowerCase()) ||
       t.tags.some(tag => tag.toLowerCase().includes(templateSearch.value.toLowerCase())))
    )
  })).filter(cat => cat.templates.length > 0)
})

const filteredHistory = computed(() => {
  if (!props.artifactId) return []
  
  let entries = getArtifactHistory(props.artifactId, 50)
  
  if (historySearch.value) {
    entries = entries.filter(entry => 
      entry.code.toLowerCase().includes(historySearch.value.toLowerCase()) ||
      entry.tags.some(tag => tag.toLowerCase().includes(historySearch.value.toLowerCase()))
    )
  }
  
  if (historyFilter.value !== 'all') {
    entries = entries.filter(entry => {
      switch (historyFilter.value) {
        case 'success':
          return entry.success
        case 'error':
          return !entry.success
        case 'recent':
          return new Date(entry.timestamp).getTime() > Date.now() - 24 * 60 * 60 * 1000
        default:
          return true
      }
    })
  }
  
  return entries
})

const historyFilterOptions = [
  { label: 'All', value: 'all' },
  { label: 'Successful', value: 'success' },
  { label: 'Errors', value: 'error' },
  { label: 'Recent (24h)', value: 'recent' }
]

const categoryOptions = computed(() => 
  categories.value.map(cat => ({ label: cat.name, value: cat.id }))
)

const difficultyOptions = [
  { label: 'Beginner', value: 'beginner' },
  { label: 'Intermediate', value: 'intermediate' },
  { label: 'Advanced', value: 'advanced' }
]

// Watch for prop changes
watch(() => props.modelValue, (newValue) => {
  if (newValue !== code.value) {
    code.value = newValue
  }
})

// Watch for code changes
watch(code, (newCode) => {
  emit('update:modelValue', newCode)
  lastModified.value = new Date().toISOString()
})

// Methods
const onCodeChange = (event: Event) => {
  const target = event.target as HTMLTextAreaElement
  const newCode = target.value
  
  // Add to history if significant change
  if (newCode !== code.value && newCode.length > 0) {
    addToHistory(code.value)
  }
  
  code.value = newCode
}

const onKeyDown = (event: KeyboardEvent) => {
  // Handle tab key
  if (event.key === 'Tab') {
    event.preventDefault()
    const textarea = event.target as HTMLTextAreaElement
    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    
    const newCode = code.value.substring(0, start) + '  ' + code.value.substring(end)
    code.value = newCode
    
    nextTick(() => {
      textarea.selectionStart = textarea.selectionEnd = start + 2
    })
  }
  
  // Handle Ctrl+Z (undo)
  if (event.ctrlKey && event.key === 'z') {
    event.preventDefault()
    undo()
  }
  
  // Handle Ctrl+Y (redo)
  if (event.ctrlKey && event.key === 'y') {
    event.preventDefault()
    redo()
  }
  
  // Handle Ctrl+Enter (run)
  if (event.ctrlKey && event.key === 'Enter') {
    event.preventDefault()
    emit('run')
  }
}

const updateCursorPosition = () => {
  const textarea = editorTextarea.value
  if (!textarea) return
  
  const textBeforeCursor = textarea.value.substring(0, textarea.selectionStart)
  const lines = textBeforeCursor.split('\n')
  cursorLine.value = lines.length
  cursorColumn.value = lines[lines.length - 1].length + 1
}

const syncLineNumbers = () => {
  if (lineNumbers.value && editorTextarea.value) {
    lineNumbers.value.scrollTop = editorTextarea.value.scrollTop
  }
}

const syncEditor = () => {
  if (lineNumbers.value && editorTextarea.value) {
    editorTextarea.value.scrollTop = lineNumbers.value.scrollTop
  }
}

const addToHistory = (codeToAdd: string) => {
  if (codeToAdd.trim() === '') return
  
  // Remove duplicates
  const existingIndex = history.value.indexOf(codeToAdd)
  if (existingIndex !== -1) {
    history.value.splice(existingIndex, 1)
  }
  
  // Add to beginning
  history.value.unshift(codeToAdd)
  
  // Limit history size
  if (history.value.length > maxHistorySize) {
    history.value.splice(maxHistorySize)
  }
  
  historyIndex.value = 0
}

const undo = () => {
  if (canUndo.value) {
    historyIndex.value++
    code.value = history.value[historyIndex.value]
  }
}

const redo = () => {
  if (canRedo.value) {
    historyIndex.value--
    code.value = history.value[historyIndex.value]
  }
}

const formatCode = () => {
  // Basic code formatting
  try {
    if (props.language === 'json') {
      const parsed = JSON.parse(code.value)
      code.value = JSON.stringify(parsed, null, 2)
      message.success('Code formatted successfully')
    } else {
      // Basic indentation fix
      const lines = code.value.split('\n')
      let indentLevel = 0
      const formattedLines = lines.map(line => {
        const trimmed = line.trim()
        if (trimmed.includes('}') || trimmed.includes(']') || trimmed.includes(')')) {
          indentLevel = Math.max(0, indentLevel - 1)
        }
        const formatted = '  '.repeat(indentLevel) + trimmed
        if (trimmed.includes('{') || trimmed.includes('[') || trimmed.includes('(')) {
          indentLevel++
        }
        return formatted
      })
      code.value = formattedLines.join('\n')
      message.success('Code formatted successfully')
    }
  } catch (error) {
    message.error('Failed to format code')
  }
}

const clearCode = () => {
  addToHistory(code.value)
  code.value = ''
}

const copyCode = async () => {
  try {
    await navigator.clipboard.writeText(code.value)
    message.success('Code copied to clipboard')
  } catch (error) {
    message.error('Failed to copy code')
  }
}

const showTemplates = () => {
  showTemplatesPanel.value = true
  showHistoryPanel.value = false
}

const showHistory = () => {
  showHistoryPanel.value = true
  showTemplatesPanel.value = false
}

const toggleCategory = (categoryId: string) => {
  if (expandedCategories.value.has(categoryId)) {
    expandedCategories.value.delete(categoryId)
  } else {
    expandedCategories.value.add(categoryId)
  }
}

const insertTemplate = (template: any) => {
  addToHistory(code.value)
  code.value = template.code
  incrementUsage(template.id)
  showTemplatesPanel.value = false
  message.success(`Template "${template.name}" inserted`)
}

const loadFromHistory = (entry: any) => {
  addToHistory(code.value)
  code.value = entry.code
  showHistoryPanel.value = false
  message.success('Code loaded from history')
}

const insertSnippet = (type: string) => {
  const snippets = {
    function: {
      javascript: 'function functionName() {\n  // TODO: Implement function\n  return null;\n}',
      python: 'def function_name():\n    """TODO: Implement function"""\n    pass'
    },
    loop: {
      javascript: 'for (let i = 0; i < array.length; i++) {\n  // TODO: Process array[i]\n}',
      python: 'for item in items:\n    # TODO: Process item\n    pass'
    },
    class: {
      javascript: 'class ClassName {\n  constructor() {\n    // TODO: Initialize\n  }\n}',
      python: 'class ClassName:\n    def __init__(self):\n        """TODO: Initialize"""\n        pass'
    }
  }
  
  const snippet = snippets[type]?.[props.language]
  if (snippet) {
    const textarea = editorTextarea.value
    if (textarea) {
      const start = textarea.selectionStart
      const end = textarea.selectionEnd
      const newCode = code.value.substring(0, start) + snippet + code.value.substring(end)
      code.value = newCode
      
      nextTick(() => {
        textarea.selectionStart = textarea.selectionEnd = start + snippet.length
        textarea.focus()
      })
    }
  }
}

const saveAsTemplate = () => {
  newTemplate.value.name = ''
  newTemplate.value.description = ''
  newTemplate.value.category = 'basics'
  newTemplate.value.difficulty = 'beginner'
  newTemplate.value.tagsInput = ''
  showSaveTemplateModal.value = true
}

const saveTemplate = () => {
  if (!newTemplate.value.name.trim()) {
    message.error('Template name is required')
    return
  }
  
  const tags = newTemplate.value.tagsInput
    .split(',')
    .map(tag => tag.trim())
    .filter(tag => tag.length > 0)
  
  const templateId = addTemplate({
    name: newTemplate.value.name,
    description: newTemplate.value.description,
    language: props.language,
    code: code.value,
    category: newTemplate.value.category,
    difficulty: newTemplate.value.difficulty,
    tags,
    author: 'User'
  })
  
  showSaveTemplateModal.value = false
  message.success(`Template "${newTemplate.value.name}" saved successfully`)
}

const getLanguageIcon = (language: string) => {
  const icons = {
    javascript: 'ri:javascript-line',
    typescript: 'ri:typescript-line',
    python: 'ri:python-line',
    json: 'ri:file-code-line',
    html: 'ri:html5-line',
    css: 'ri:css3-line'
  }
  return icons[language] || 'ri:code-line'
}

const getPlaceholder = (language: string) => {
  const placeholders = {
    javascript: 'Enter JavaScript code...',
    typescript: 'Enter TypeScript code...',
    python: 'Enter Python code...',
    json: 'Enter JSON data...',
    html: 'Enter HTML markup...',
    css: 'Enter CSS styles...'
  }
  return placeholders[language] || 'Enter code...'
}

const getDifficultyType = (difficulty: string) => {
  const types = {
    beginner: 'success',
    intermediate: 'warning',
    advanced: 'error'
  }
  return types[difficulty] || 'default'
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString()
}

const formatRelativeTime = (timestamp: string) => {
  const now = new Date()
  const time = new Date(timestamp)
  const diffMs = now.getTime() - time.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)
  
  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  return `${diffDays}d ago`
}

onMounted(() => {
  // Initialize with current code
  if (code.value) {
    addToHistory(code.value)
  }
  
  // Auto-expand first category
  if (categories.value.length > 0) {
    expandedCategories.value.add(categories.value[0].id)
  }
})
</script>

<style scoped>
.artifact-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--artifact-content-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
}

.editor-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--text-color);
}

.language-icon {
  font-size: 18px;
  color: var(--primary-color);
}

.editor-actions {
  display: flex;
  gap: 8px;
}

.editor-content {
  display: flex;
  flex: 1;
  min-height: 0;
}

.editor-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
  font-size: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-separator {
  width: 1px;
  height: 16px;
  background: var(--border-color);
  margin: 0 8px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-color-secondary);
}

.editor-textarea-container {
  position: relative;
  flex: 1;
  display: flex;
}

.editor-textarea {
  flex: 1;
  padding: 16px 16px 16px 60px;
  border: none;
  outline: none;
  background: var(--artifact-content-bg);
  color: var(--text-color);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  resize: none;
  overflow-y: auto;
  white-space: pre;
  overflow-wrap: normal;
}

.line-numbers {
  position: absolute;
  left: 0;
  top: 0;
  width: 50px;
  height: 100%;
  background: var(--artifact-header-bg);
  border-right: 1px solid var(--border-color);
  padding: 16px 8px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: var(--text-color-secondary);
  text-align: right;
  user-select: none;
  overflow: hidden;
}

.line-number {
  height: 21px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.editor-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--artifact-header-bg);
  border-top: 1px solid var(--border-color);
  font-size: 12px;
}

.footer-info {
  display: flex;
  gap: 16px;
  color: var(--text-color-secondary);
}

.footer-actions {
  display: flex;
  gap: 8px;
}

.editor-panels {
  display: flex;
  flex-direction: column;
  width: 350px;
  border-left: 1px solid var(--border-color);
}

.panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--artifact-content-bg);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--artifact-header-bg);
  border-bottom: 1px solid var(--border-color);
}

.panel-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color);
}

.panel-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.template-search,
.history-search {
  margin-bottom: 16px;
}

.template-categories {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.category {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: var(--artifact-header-bg);
  cursor: pointer;
  transition: background-color 0.2s;
}

.category-header:hover {
  background: var(--hover-color);
}

.template-count {
  margin-left: auto;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.category-templates {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.template-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 12px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background-color 0.2s;
}

.template-item:hover {
  background: var(--hover-color);
}

.template-item:last-child {
  border-bottom: none;
}

.template-info {
  flex: 1;
  min-width: 0;
}

.template-name {
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 4px;
}

.template-description {
  font-size: 12px;
  color: var(--text-color-secondary);
  margin-bottom: 6px;
}

.template-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.template-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  margin-left: 8px;
}

.usage-count {
  font-size: 11px;
  color: var(--text-color-secondary);
}

.history-filters {
  margin-bottom: 16px;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.history-item:hover {
  background: var(--hover-color);
}

.history-info {
  flex: 1;
  min-width: 0;
}

.history-time {
  font-size: 12px;
  color: var(--text-color-secondary);
  margin-bottom: 4px;
}

.history-preview {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  color: var(--text-color);
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-meta {
  display: flex;
  gap: 8px;
  align-items: center;
}

.execution-time {
  font-size: 11px;
  color: var(--text-color-secondary);
}

.history-tags {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-left: 8px;
}

.save-template-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* Dark mode adjustments */
[data-theme='dark'] .editor-textarea {
  background: #1a1a1a;
  color: #e0e0e0;
}

[data-theme='dark'] .line-numbers {
  background: #2d2d2d;
  color: #888;
}

/* Responsive design */
@media (max-width: 768px) {
  .editor-panels {
    width: 100%;
    height: 300px;
  }
  
  .editor-content {
    flex-direction: column;
  }
  
  .editor-main {
    min-height: 400px;
  }
  
  .editor-header {
    padding: 8px 12px;
  }
  
  .editor-actions {
    gap: 4px;
  }
  
  .toolbar-right {
    display: none;
  }
}
</style>