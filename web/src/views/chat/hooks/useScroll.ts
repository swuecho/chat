import type { Ref } from 'vue'
import { nextTick, ref, onUnmounted, watch } from 'vue'

type ScrollElement = HTMLDivElement | null

interface ScrollReturn {
  scrollRef: Ref<ScrollElement>
  scrollToBottom: () => Promise<void>
  scrollToTop: () => Promise<void>
  scrollToBottomIfAtBottom: () => Promise<void>
  smoothScrollToBottomIfAtBottom: () => Promise<void>
}

export function useScroll(): ScrollReturn {
  const scrollRef = ref<ScrollElement>(null)
  
  // State tracking for scroll behavior
  let isAutoScrolling = false
  let manualScrollTimeout: number | null = null
  let userHasManuallyScrolled = false
  let currentAnimation: number | null = null

  // Detect manual scrolling
  const handleScroll = () => {
    if (isAutoScrolling) return // Ignore scroll events during auto-scroll
    
    // Clear existing timeout
    if (manualScrollTimeout) {
      clearTimeout(manualScrollTimeout)
    }
    
    // Mark as manually scrolled
    userHasManuallyScrolled = true
    
    // Cancel any ongoing auto-scroll animation
    if (currentAnimation) {
      cancelAnimationFrame(currentAnimation)
      currentAnimation = null
    }
    
    // Reset manual scroll flag after user stops scrolling
    manualScrollTimeout = window.setTimeout(() => {
      userHasManuallyScrolled = false
    }, 2000) // 2 seconds of no scrolling
  }

  const scrollToBottom = async () => {
    await nextTick()
    if (scrollRef.value) {
      isAutoScrolling = true
      scrollRef.value.scrollTop = scrollRef.value.scrollHeight
      // Reset auto-scroll flag after a brief delay
      setTimeout(() => { isAutoScrolling = false }, 50)
    }
  }

  const scrollToTop = async () => {
    await nextTick()
    if (scrollRef.value) {
      isAutoScrolling = true
      scrollRef.value.scrollTop = 0
      setTimeout(() => { isAutoScrolling = false }, 50)
    }
  }

  const scrollToBottomIfAtBottom = async () => {
    await nextTick()
    if (scrollRef.value && !userHasManuallyScrolled) {
      const element = scrollRef.value
      const threshold = Math.max(400, element.clientHeight * 0.25) // Dynamic threshold: 400px minimum or 25% of viewport
      const distanceToBottom = element.scrollHeight - element.scrollTop - element.clientHeight
      if (distanceToBottom <= threshold) {
        isAutoScrolling = true
        scrollRef.value.scrollTop = element.scrollHeight
        setTimeout(() => { isAutoScrolling = false }, 50)
      }
    }
  }

  const smoothScrollToBottomIfAtBottom = async () => {
    await nextTick()
    if (scrollRef.value && !userHasManuallyScrolled) {
      const element = scrollRef.value
      const threshold = Math.max(200, element.clientHeight * 0.1) // Smaller threshold: 200px minimum or 10% of viewport
      const distanceToBottom = element.scrollHeight - element.scrollTop - element.clientHeight
      
      if (distanceToBottom <= threshold) {
        // Cancel any existing animation to prevent conflicts
        if (currentAnimation) {
          cancelAnimationFrame(currentAnimation)
          currentAnimation = null
        }
        
        // Simple instant scroll to bottom without animation
        isAutoScrolling = true
        element.scrollTop = element.scrollHeight
        setTimeout(() => { isAutoScrolling = false }, 50)
      }
    }
  }

  // Setup event listener when scrollRef becomes available
  watch(scrollRef, (newElement, oldElement) => {
    // Remove listener from old element
    if (oldElement) {
      oldElement.removeEventListener('scroll', handleScroll)
    }
    
    // Add listener to new element
    if (newElement) {
      newElement.addEventListener('scroll', handleScroll, { passive: true })
    }
  }, { immediate: true })

  onUnmounted(() => {
    if (scrollRef.value) {
      scrollRef.value.removeEventListener('scroll', handleScroll)
    }
    if (manualScrollTimeout) {
      clearTimeout(manualScrollTimeout)
    }
    if (currentAnimation) {
      cancelAnimationFrame(currentAnimation)
    }
  })

  return {
    scrollRef,
    scrollToBottom,
    scrollToTop,
    scrollToBottomIfAtBottom,
    smoothScrollToBottomIfAtBottom,
  }
}
