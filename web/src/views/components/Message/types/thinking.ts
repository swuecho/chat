export interface ThinkingContent {
  id?: string
  content: string
  isExpanded?: boolean
  createdAt?: Date
  updatedAt?: Date
}

export interface ThinkingParseResult {
  hasThinking: boolean
  thinkingContent: ThinkingContent
  answerContent: string
  rawText: string
}

export interface ThinkingRenderOptions {
  enableMarkdown?: boolean
  enableCollapsible?: boolean
  defaultExpanded?: boolean
  showBorder?: boolean
  borderColor?: string
  maxLines?: number
  enableCopy?: boolean
}

export interface ThinkingCacheEntry {
  rawText: string
  parsedResult: ThinkingParseResult
  timestamp: number
}

export interface ThinkingParserConfig {
  cacheSize?: number
  cacheTTL?: number
  enableLogging?: boolean
  thinkingTagPattern?: RegExp
}

export interface ThinkingComposableReturn {
  thinkingContent: ThinkingContent | null
  hasThinking: boolean
  isExpanded: boolean
  toggleExpanded: () => void
  setExpanded: (expanded: boolean) => void
  parsedResult: ThinkingParseResult | null
  refreshParse: () => void
  updateText: (newText: string) => void
}

export interface ThinkingRendererProps {
  content: ThinkingContent
  options?: ThinkingRenderOptions
  class?: string
  onToggle?: (expanded: boolean) => void
}

export interface ThinkingRendererEmits {
  toggle: [expanded: boolean]
  copy: [content: string]
}

export interface UseThinkingContentOptions {
  defaultExpanded?: boolean
  enableCache?: boolean
  parserConfig?: ThinkingParserConfig
}