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
      const threshold = Math.max(400, element.clientHeight * 0.25) // Dynamic threshold: 400px minimum or 25% of viewport
      const distanceToBottom = element.scrollHeight - element.scrollTop - element.clientHeight
      
      if (distanceToBottom <= threshold) {
        // Calculate the distance to scroll
        const targetScrollTop = element.scrollHeight - element.clientHeight
        const currentScrollTop = element.scrollTop
        const distance = targetScrollTop - currentScrollTop
        
        if (distance > 0) {
          // Cancel any existing animation
          if (currentAnimation) {
            cancelAnimationFrame(currentAnimation)
          }
          
          // Human reading speed: approximately 200-300 words per minute
          // Assuming average 5 characters per word, that's ~17-25 chars per second
          // For smooth scrolling, we'll use a slower pace: ~300 pixels per second
          const scrollSpeed = 300 // pixels per second
          const duration = Math.max(distance / scrollSpeed * 1000, 50) // minimum 50ms
          
          // Use smooth scrolling with easing
          const startTime = performance.now()
          const startScrollTop = currentScrollTop
          isAutoScrolling = true
          
          const animateScroll = (currentTime: number) => {
            // Check if user has manually scrolled during animation
            if (userHasManuallyScrolled) {
              currentAnimation = null
              isAutoScrolling = false
              return
            }
            
            const elapsed = currentTime - startTime
            const progress = Math.min(elapsed / duration, 1)
            
            // Ease-out cubic function for natural feeling
            const easeOut = 1 - Math.pow(1 - progress, 3)
            
            element.scrollTop = startScrollTop + (distance * easeOut)
            
            if (progress < 1) {
              currentAnimation = requestAnimationFrame(animateScroll)
            } else {
              currentAnimation = null
              isAutoScrolling = false
            }
          }
          
          currentAnimation = requestAnimationFrame(animateScroll)
        }
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
