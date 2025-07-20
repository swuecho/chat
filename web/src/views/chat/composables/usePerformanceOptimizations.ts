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
      // Handle circular references by using a simpler key strategy
      if (error instanceof TypeError && error.message.includes('circular')) {
        // Create a simple hash based on object properties to avoid circular refs
        key = createSimpleKey(dep)
      } else {
        throw error
      }
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
 * Creates a simple key for objects that may contain circular references
 */
function createSimpleKey(obj: any): string {
  if (obj === null || obj === undefined) return 'null'
  if (typeof obj !== 'object') return String(obj)
  
  const keys: string[] = []
  const visited = new WeakSet()
  
  function serialize(value: any, depth = 0): string {
    if (depth > 5) return '[max-depth]' // Prevent infinite recursion
    if (value === null || value === undefined) return 'null'
    if (typeof value !== 'object') return String(value)
    if (visited.has(value)) return '[circular]'
    
    visited.add(value)
    
    if (Array.isArray(value)) {
      return `[${value.length}]`
    }
    
    const sortedKeys = Object.keys(value).sort()
    return sortedKeys.slice(0, 10).map(key => `${key}:${serialize(value[key], depth + 1)}`).join(',')
  }
  
  return serialize(obj)
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