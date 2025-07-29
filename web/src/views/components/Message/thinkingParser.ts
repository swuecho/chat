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
    let thinkingContentStr = ''
    const answerContent = text.replace(this.config.thinkingTagPattern, (match, content) => {
      thinkingContentStr = content.trim()
      return ''
    })

    const result: ThinkingParseResult = {
      hasThinking: thinkingContentStr.length > 0,
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