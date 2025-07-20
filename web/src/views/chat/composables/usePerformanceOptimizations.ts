import { ref, computed, watch, type Ref } from 'vue'

/**
 * Debounce utility for search input
 */
export function useDebounce<T>(value: Ref<T> | T, delay: number) {
  // Handle both refs and raw values
  const isRef = value && typeof value === 'object' && '__v_isRef' in value
  const initialValue = isRef ? (value as Ref<T>).value : value as T
  const debouncedValue = ref<T>(initialValue)
  
  let timeoutId: NodeJS.Timeout
  
  if (isRef) {
    // If it's a ref, watch the ref directly
    watch(value as Ref<T>, (newValue) => {
      clearTimeout(timeoutId)
      timeoutId = setTimeout(() => {
        debouncedValue.value = newValue
      }, delay)
    }, { immediate: true })
  } else {
    // If it's a raw value, watch it as a getter
    watch(() => value, (newValue) => {
      clearTimeout(timeoutId)
      timeoutId = setTimeout(() => {
        debouncedValue.value = newValue as T
      }, delay)
    }, { immediate: true })
  }
  
  return debouncedValue
}

/**
 * Virtual scrolling helper for large lists
 */
export function useVirtualList<T>(
  items: T[],
  itemHeight: number,
  containerHeight: number
) {
  const scrollTop = ref(0)
  
  const visibleItems = computed(() => {
    const startIndex = Math.floor(scrollTop.value / itemHeight)
    const endIndex = Math.min(
      startIndex + Math.ceil(containerHeight / itemHeight) + 1,
      items.length
    )
    
    return {
      startIndex,
      endIndex,
      items: items.slice(startIndex, endIndex),
      offsetY: startIndex * itemHeight,
      totalHeight: items.length * itemHeight
    }
  })
  
  const handleScroll = (event: Event) => {
    const target = event.target as HTMLElement
    scrollTop.value = target.scrollTop
  }
  
  return {
    visibleItems,
    handleScroll
  }
}

/**
 * Throttle utility for high-frequency events
 */
export function useThrottle<T extends (...args: any[]) => any>(
  fn: T,
  delay: number
): T {
  let lastCall = 0
  
  return ((...args: Parameters<T>) => {
    const now = Date.now()
    if (now - lastCall >= delay) {
      lastCall = now
      return fn(...args)
    }
  }) as T
}