import { ref, computed, watch } from 'vue'
import { parseThinkingContent } from './thinkingParser'
import type { 
  ThinkingContent, 
  ThinkingParseResult, 
  ThinkingComposableReturn, 
  UseThinkingContentOptions 
} from './types/thinking'

export function useThinkingContent(
  text: string | undefined, 
  options: UseThinkingContentOptions = {}
): ThinkingComposableReturn {
  const {
    defaultExpanded = true,
    enableCache = true,
    parserConfig = {}
  } = options

  const isExpanded = ref(defaultExpanded)
  const rawText = ref(text || '')
  const parsedResult = ref<ThinkingParseResult | null>(null)

  // Parse content when raw text changes
  const parseContent = () => {
    if (!rawText.value) {
      parsedResult.value = {
        hasThinking: false,
        thinkingContent: { content: '', isExpanded: defaultExpanded },
        answerContent: '',
        rawText: ''
      }
      return
    }

    parsedResult.value = parseThinkingContent(rawText.value)
  }

  // Initial parse
  parseContent()

  // Watch for text changes
  watch(rawText, parseContent, { immediate: true })

  // Computed properties
  const thinkingContent = computed(() => parsedResult.value?.thinkingContent || null)
  const hasThinking = computed(() => parsedResult.value?.hasThinking || false)

  // Methods
  const toggleExpanded = () => {
    isExpanded.value = !isExpanded.value
    if (thinkingContent.value) {
      thinkingContent.value.isExpanded = isExpanded.value
    }
  }

  const setExpanded = (expanded: boolean) => {
    isExpanded.value = expanded
    if (thinkingContent.value) {
      thinkingContent.value.isExpanded = expanded
    }
  }

  const refreshParse = () => {
    parseContent()
  }

  const updateText = (newText: string) => {
    rawText.value = newText
  }

  return {
    thinkingContent,
    hasThinking,
    isExpanded,
    toggleExpanded,
    setExpanded,
    parsedResult,
    refreshParse,
    updateText
  }
}

// Composable for managing multiple thinking contents
export function useMultipleThinkingContent(
  texts: Array<{ id: string; text: string }>,
  options: UseThinkingContentOptions = {}
) {
  const thinkingStates = ref(new Map<string, ThinkingComposableReturn>())

  const getThinkingState = (id: string, text: string) => {
    if (!thinkingStates.value.has(id)) {
      thinkingStates.value.set(id, useThinkingContent(text, options))
    }
    return thinkingStates.value.get(id)!
  }

  const updateText = (id: string, newText: string) => {
    const state = thinkingStates.value.get(id)
    if (state) {
      state.updateText(newText)
    }
  }

  const removeThinkingState = (id: string) => {
    thinkingStates.value.delete(id)
  }

  const clearAllStates = () => {
    thinkingStates.value.clear()
  }

  return {
    getThinkingState,
    updateText,
    removeThinkingState,
    clearAllStates,
    states: thinkingStates
  }
}

// Composable for thinking content statistics
export function useThinkingStats() {
  const totalParsed = ref(0)
  const totalWithThinking = ref(0)
  const averageThinkingLength = ref(0)

  const updateStats = (parsedResult: ThinkingParseResult) => {
    totalParsed.value++
    if (parsedResult.hasThinking) {
      totalWithThinking.value++
      const thinkingLength = parsedResult.thinkingContent.content.length
      averageThinkingLength.value = 
        (averageThinkingLength.value * (totalWithThinking.value - 1) + thinkingLength) / totalWithThinking.value
    }
  }

  const getStats = computed(() => ({
    totalParsed: totalParsed.value,
    totalWithThinking: totalWithThinking.value,
    thinkingRate: totalParsed.value > 0 ? (totalWithThinking.value / totalParsed.value) * 100 : 0,
    averageThinkingLength: Math.round(averageThinkingLength.value)
  }))

  const resetStats = () => {
    totalParsed.value = 0
    totalWithThinking.value = 0
    averageThinkingLength.value = 0
  }

  return {
    updateStats,
    getStats,
    resetStats
  }
}