import { ref, computed, watch } from 'vue'

/**
 * Debounce utility for search input
 */
export function useDebounce<T>(value: T, delay: number) {
  const debouncedValue = ref<T>(value)
  
  let timeoutId: NodeJS.Timeout
  
  watch(() => value, (newValue) => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => {
      debouncedValue.value = newValue
    }, delay)
  }, { immediate: true })
  
  return debouncedValue
}

/**
 * Memoization for expensive computations
 */
export function useMemoized<T, R>(
  fn: (arg: T) => R,
  dependency: () => T
) {
  const cache = new Map<string, R>()
  
  return computed(() => {
    const dep = dependency()
    let key: string
    
    try {
      key = JSON.stringify(dep)
    } catch (error) {
      // For circular references, disable caching and compute every time
      // This is safer than trying to create complex fallback keys
      console.warn('Memoization disabled due to circular reference in dependency')
      return fn(dep)
    }
    
    if (cache.has(key)) {
      return cache.get(key)!
    }
    
    const result = fn(dep)
    cache.set(key, result)
    
    // Limit cache size to prevent memory leaks
    if (cache.size > 100) {
      const firstKey = cache.keys().next().value
      if (firstKey !== undefined) {
        cache.delete(firstKey)
      }
    }
    
    return result
  })
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