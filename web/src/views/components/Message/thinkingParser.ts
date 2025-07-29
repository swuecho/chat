import type { ThinkingParseResult, ThinkingCacheEntry, ThinkingParserConfig } from './types/thinking'

class ThinkingParser {
  private cache: Map<string, ThinkingCacheEntry> = new Map()
  private config: Required<ThinkingParserConfig>

  constructor(config: ThinkingParserConfig = {}) {
    this.config = {
      cacheSize: 100,
      cacheTTL: 5 * 60 * 1000, // 5 minutes
      enableLogging: false,
      thinkingTagPattern: /<think>(.*?)<\/think>/gs,
      ...config
    }
  }

  parseText(text: string): ThinkingParseResult {
    // Check cache first
    const cacheKey = this.generateCacheKey(text)
    const cached = this.getFromCache(cacheKey)
    if (cached) {
      if (this.config.enableLogging) {
        console.log('Cache hit for thinking content')
      }
      return cached
    }

    // Parse thinking content
    const thinkingContents: string[] = []
    let answerContent = text
    
    // Check for complete thinking tags first
    const completeMatch = text.match(this.config.thinkingTagPattern)
    if (completeMatch) {
      // Complete thinking tags found, extract content
      answerContent = text.replace(this.config.thinkingTagPattern, (_, content) => {
        thinkingContents.push(content.trim())
        return ''
      })
    } else {
      // Check for incomplete thinking tags (opening without closing)
      const openingTagMatch = text.match(/<think>/)
      const closingTagMatch = text.match(/<\/think>/)
      
      if (openingTagMatch && !closingTagMatch) {
        // Incomplete: has opening tag but no closing tag
        // Extract content after opening tag as thinking content
        const openingTagIndex = text.indexOf('<think>')
        const content = text.substring(openingTagIndex + 7) // 7 is length of '<think>'
        // Always add content, even if empty - this indicates we're in thinking mode
        thinkingContents.push(content)
        answerContent = text.substring(0, openingTagIndex)
      } else if (!openingTagMatch && closingTagMatch) {
        // Incomplete: has closing tag but no opening tag
        // Treat everything before closing tag as thinking content
        const closingTagIndex = text.indexOf('</think>')
        const content = text.substring(0, closingTagIndex)
        // Always add content, even if empty - this indicates thinking content was present
        thinkingContents.push(content)
        answerContent = ''
      }
      // If both tags are missing or both are present (already handled), no special handling needed
    }

    const thinkingContentStr = thinkingContents.map(content => content.trim()).join('\n\n')
    
    // We have thinking if there's content OR if we found an incomplete opening tag
    const hasThinkingContent = thinkingContents.length > 0
    
    const result: ThinkingParseResult = {
      hasThinking: hasThinkingContent,
      thinkingContent: {
        content: thinkingContentStr,
        isExpanded: true,
        createdAt: new Date(),
        updatedAt: new Date()
      },
      answerContent,
      rawText: text
    }

    // Cache the result
    this.setToCache(cacheKey, result)

    if (this.config.enableLogging) {
      console.log('Parsed thinking content:', {
        hasThinking: result.hasThinking,
        thinkingLength: thinkingContentStr.length,
        answerLength: answerContent.length
      })
    }

    return result
  }

  private generateCacheKey(text: string): string {
    // Simple hash function for cache key
    let hash = 0
    for (let i = 0; i < text.length; i++) {
      const char = text.charCodeAt(i)
      hash = ((hash << 5) - hash) + char
      hash = hash & hash // Convert to 32-bit integer
    }
    return hash.toString(36)
  }

  private getFromCache(key: string): ThinkingParseResult | null {
    const entry = this.cache.get(key)
    if (!entry) return null

    // Check TTL
    const now = Date.now()
    if (now - entry.timestamp > this.config.cacheTTL) {
      this.cache.delete(key)
      return null
    }

    return entry.parsedResult
  }

  private setToCache(key: string, result: ThinkingParseResult): void {
    // Clean up cache if it's too large
    if (this.cache.size >= this.config.cacheSize) {
      this.cleanupCache()
    }

    this.cache.set(key, {
      rawText: result.rawText,
      parsedResult: result,
      timestamp: Date.now()
    })
  }

  private cleanupCache(): void {
    // Remove oldest entries
    const entries = Array.from(this.cache.entries())
    const toRemove = entries.slice(0, Math.floor(this.config.cacheSize * 0.3))
    
    toRemove.forEach(([key]) => {
      this.cache.delete(key)
    })

    if (this.config.enableLogging) {
      console.log(`Cleaned up ${toRemove.length} cache entries`)
    }
  }

  clearCache(): void {
    this.cache.clear()
    if (this.config.enableLogging) {
      console.log('Thinking parser cache cleared')
    }
  }

  getCacheStats(): { size: number; hitRate: number } {
    return {
      size: this.cache.size,
      hitRate: 0 // Could be enhanced with hit tracking
    }
  }
}

// Export singleton instance
export const thinkingParser = new ThinkingParser()

// Export utility functions
export const parseThinkingContent = (text: string): ThinkingParseResult => {
  return thinkingParser.parseText(text)
}

export const clearThinkingCache = (): void => {
  thinkingParser.clearCache()
}

export const getThinkingCacheStats = () => {
  return thinkingParser.getCacheStats()
}

// Backward compatibility - re-export types
export type { ThinkingParseResult, ThinkingCacheEntry, ThinkingParserConfig } from './types/thinking'