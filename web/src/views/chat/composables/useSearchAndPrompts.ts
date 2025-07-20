import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { type OnSelect } from 'naive-ui/es/auto-complete/src/interface'
import { useChatStore, usePromptStore } from '@/store'
import { useDebounce, useMemoized } from './usePerformanceOptimizations'

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

  // Memoized search options for better performance
  const searchOptions = useMemoized(
    (searchData: { prompt: string; history: any[]; templates: any[] }): SearchOption[] => {
      const { prompt: searchPrompt, history, templates } = searchData
      
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
      
      if (!searchPrompt.startsWith('/')) {
        return []
      }

      const sessionOptions: SearchOption[] = history
        .filter(filterItemsByTitle)
        .map((session: ChatItem) => ({
          label: `UUID|$|${session.uuid}`,
          value: `UUID|$|${session.uuid}`,
        }))

      const promptOptions: SearchOption[] = templates
        .filter(filterItemsByPrompt)
        .map((item: PromptItem) => ({
          label: item.value,
          value: item.value,
        }))
      
      return [...sessionOptions, ...promptOptions]
    },
    () => ({
      prompt: debouncedPrompt.value,
      history: chatStore.history,
      templates: promptTemplate.value
    })
  )

  const renderOption = (option: { label: string }): string[] => {
    // Check if it's a prompt template
    const promptItem = promptTemplate.value.find((item: PromptItem) => item.value === option.label)
    if (promptItem) {
      return [promptItem.key]
    }
    
    // Check if it's a chat session
    const chatItem = chatStore.history.find((chat: ChatItem) => `UUID|$|${chat.uuid}` === option.label)
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