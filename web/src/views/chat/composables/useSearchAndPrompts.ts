import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { type OnSelect } from 'naive-ui/es/auto-complete/src/interface'
import { useChatStore, usePromptStore } from '@/store'
import { useDebounce } from './usePerformanceOptimizations'

interface PromptItem {
  key: string
  value: string
}

interface ChatItem {
  uuid: string
  title: string
}

interface SearchOption {
  label: string
  value: string
}

export function useSearchAndPrompts() {
  const prompt = ref<string>('')
  const chatStore = useChatStore()
  const promptStore = usePromptStore()
  
  // Get reactive store refs without explicit typing to avoid type issues
  const storeRefs = storeToRefs(promptStore)
  const promptTemplate = computed(() => storeRefs.promptList?.value || [])
  
  // Debounce search input for better performance
  const debouncedPrompt = useDebounce(prompt, 300)

  // Search options computed directly - much simpler!
  const searchOptions = computed((): SearchOption[] => {
    let searchPrompt = debouncedPrompt.value
    
    // Ensure searchPrompt is a string
    if (typeof searchPrompt !== 'string') {
      console.warn('debouncedPrompt.value is not a string:', typeof searchPrompt, searchPrompt)
      searchPrompt = String(searchPrompt || '')
    }
    
    if (!searchPrompt.startsWith('/')) {
      return []
    }
    
    const filterItemsByPrompt = (item: PromptItem): boolean => {
      const lowerCaseKey = item.key.toLowerCase()
      const lowerCasePrompt = searchPrompt.substring(1).toLowerCase()
      return lowerCaseKey.includes(lowerCasePrompt)
    }
    
    const filterItemsByTitle = (item: ChatItem): boolean => {
      const lowerCaseTitle = item.title.toLowerCase()
      const lowerCasePrompt = searchPrompt.substring(1).toLowerCase()
      return lowerCaseTitle.includes(lowerCasePrompt)
    }

    // Get all sessions from workspace history
    const allSessions: ChatItem[] = []
    for (const sessions of Object.values(chatStore.workspaceHistory)) {
      allSessions.push(...sessions)
    }
    
    const sessionOptions: SearchOption[] = allSessions
      .filter(filterItemsByTitle)
      .map((session: ChatItem) => ({
        label: `UUID|$|${session.uuid}`,
        value: `UUID|$|${session.uuid}`,
      }))

    const promptOptions: SearchOption[] = promptTemplate.value
      .filter(filterItemsByPrompt)
      .map((item: PromptItem) => ({
        label: item.value,
        value: item.value,
      }))
    
    return [...sessionOptions, ...promptOptions]
  })

  const renderOption = (option: { label: string }): string[] => {
    // Check if it's a prompt template
    const promptItem = promptTemplate.value.find((item: PromptItem) => item.value === option.label)
    if (promptItem) {
      return [promptItem.key]
    }
    
    // Check if it's a chat session across all workspace histories
    let chatItem = null
    for (const sessions of Object.values(chatStore.workspaceHistory)) {
      chatItem = sessions.find((chat: ChatItem) => `UUID|$|${chat.uuid}` === option.label)
      if (chatItem) break
    }
    if (chatItem) {
      return [chatItem.title]
    }
    
    return []
  }

  const handleSelectAutoComplete: OnSelect = function (v: string | number) {
    if (typeof v === 'string' && v.startsWith('UUID|$|')) {
      chatStore.setActive(v.split('|$|')[1])
    }
  }

  const handleUsePrompt = (_: string, value: string): void => {
    prompt.value = value
  }

  return {
    prompt,
    searchOptions,
    renderOption,
    handleSelectAutoComplete,
    handleUsePrompt
  }
}