<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import MarkdownIt from 'markdown-it'
import mdKatex from '@vscode/markdown-it-katex'
import hljs from 'highlight.js'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'
import { escapeBrackets, escapeDollarNumber } from '@/utils/string'
import type { ThinkingRendererProps, ThinkingRenderOptions } from './types/thinking'

const props = withDefaults(defineProps<ThinkingRendererProps>(), {
  options: () => ({
    enableMarkdown: true,
    enableCollapsible: true,
    defaultExpanded: true,
    showBorder: true,
    borderColor: 'border-lime-600',
    maxLines: 20,
    enableCopy: true
  })
})

const emit = defineEmits<{
  toggle: [expanded: boolean]
  copy: [content: string]
}>()

const { isMobile } = useBasicLayout()
const isExpanded = ref(props.options.defaultExpanded ?? true)
const isCopied = ref(false)

// Watch for changes in defaultExpanded prop to stay in sync
watch(() => props.options.defaultExpanded, (newVal) => {
  if (newVal !== undefined) {
    isExpanded.value = newVal
  }
})

const mdi = new MarkdownIt({
  html: false,
  linkify: true,
  highlight(code, language) {
    const validLang = !!(language && hljs.getLanguage(language))
    if (validLang) {
      const lang = language ?? ''
      return highlightBlock(hljs.highlight(lang, code, true).value, lang)
    }
    return highlightBlock(hljs.highlightAuto(code).value, '')
  },
})

mdi.use(mdKatex, { blockClass: 'katexmath-block rounded-md p-[10px]', errorColor: ' #cc0000' })

const wrapClass = computed(() => {
  return [
    'text-wrap',
    'min-w-[20px]',
    'rounded-md',
    isMobile.value ? 'p-2' : 'p-3',
    props.options.showBorder ? 'border-l-2' : '',
    props.options.borderColor || 'border-lime-600',
    'dark:border-white',
    'bg-gray-50',
    'dark:bg-gray-800',
    'transition-all',
    'duration-200',
    props.class || ''
  ]
})

const renderedContent = computed(() => {
  if (!props.options.enableMarkdown) {
    return props.content.content
  }
  
  const escapedText = escapeBrackets(escapeDollarNumber(props.content.content))
  return mdi.render(escapedText)
})

const shouldShowCollapse = computed(() => {
  if (!props.options.enableCollapsible) return false
  const lines = props.content.content.split('\n').length
  return lines > (props.options.maxLines || 20)
})

const toggleExpanded = () => {
  isExpanded.value = !isExpanded.value
  emit('toggle', isExpanded.value)
}

const copyContent = async () => {
  try {
    await navigator.clipboard.writeText(props.content.content)
    isCopied.value = true
    emit('copy', props.content.content)
    setTimeout(() => {
      isCopied.value = false
    }, 2000)
  } catch (error) {
    console.error('Failed to copy thinking content:', error)
  }
}

function highlightBlock(str: string, lang?: string) {
  return `<pre class="code-block-wrapper"><div class="code-block-header"><span class="code-block-header__lang">${lang}</span><span class="code-block-header__copy">${t('chat.copyCode')}</span></div><code class="hljs code-block-body ${lang}">${str}</code></pre>`
}
</script>

<template>
  <div class="text-black relative leading-relaxed break-words" :class="wrapClass">
    <div class="flex items-center justify-between mb-2">
      <div class="flex items-center space-x-2">
        <span class="text-sm font-medium text-gray-600 dark:text-gray-400">
          💭 Thinking
        </span>
        <span v-if="content.createdAt" class="text-xs text-gray-500 dark:text-gray-500">
          {{ new Date(content.createdAt).toLocaleTimeString() }}
        </span>
      </div>
      
      <div class="flex items-center space-x-1">
        <button
          v-if="options.enableCopy"
          @click="copyContent"
          class="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-colors"
          :title="isCopied ? 'Copied!' : 'Copy thinking'"
        >
          <svg 
            class="w-4 h-4" 
            :class="{ 'text-green-600': isCopied, 'text-gray-600 dark:text-gray-400': !isCopied }"
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor"
          >
            <path v-if="!isCopied" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </button>
        
        <button
          v-if="shouldShowCollapse"
          @click="toggleExpanded"
          class="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-colors"
          :title="isExpanded ? 'Collapse thinking' : 'Expand thinking'"
        >
          <svg
            class="w-4 h-4 text-gray-600 dark:text-gray-400 transform transition-transform"
            :class="{ 'rotate-180': isExpanded }"
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>
    </div>

    <div 
      class="markdown-body thinking-content"
      :class="{ 
        'max-h-96 overflow-hidden': !isExpanded && shouldShowCollapse,
        'line-clamp-none': isExpanded || !shouldShowCollapse
      }"
      v-html="renderedContent"
    />
    
    <div 
      v-if="shouldShowCollapse && !isExpanded"
      class="mt-2 text-sm text-gray-500 dark:text-gray-400 text-center cursor-pointer hover:text-gray-700 dark:hover:text-gray-300"
      @click="toggleExpanded"
    >
      ... Show more thinking
    </div>
  </div>
</template>

<style lang="less">
@import url('./style.less');

.thinking-content {
  line-height: 1.6;
  
  pre {
    margin: 8px 0;
    background-color: rgba(0, 0, 0, 0.05);
    border-radius: 6px;
    padding: 12px;
    overflow-x: auto;
    
    .dark & {
      background-color: rgba(255, 255, 255, 0.1);
    }
  }
  
  code {
    background-color: rgba(0, 0, 0, 0.05);
    padding: 2px 4px;
    border-radius: 3px;
    font-size: 0.9em;
    
    .dark & {
      background-color: rgba(255, 255, 255, 0.1);
    }
  }
  
  p {
    margin: 8px 0;
  }
  
  ul, ol {
    margin: 8px 0;
    padding-left: 20px;
  }
}
</style>