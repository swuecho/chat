/**
 * Execution History Service
 * Manages execution history, persistence, and analytics
 */

import { reactive, computed } from 'vue'
import type { ExecutionResult } from './codeRunner'

export interface ExecutionHistoryEntry {
  id: string
  artifactId: string
  timestamp: string
  code: string
  language: string
  results: ExecutionResult[]
  executionTime: number
  success: boolean
  tags: string[]
  notes?: string
}

export interface ExecutionStats {
  totalExecutions: number
  totalTime: number
  successRate: number
  languageBreakdown: Record<string, number>
  averageExecutionTime: number
  recentActivity: ExecutionHistoryEntry[]
}

class ExecutionHistoryService {
  private static instance: ExecutionHistoryService
  private history: ExecutionHistoryEntry[] = reactive([])
  private maxHistorySize = 1000
  private storageKey = 'code-runner-history'

  private constructor() {
    this.loadFromStorage()
  }

  static getInstance(): ExecutionHistoryService {
    if (!ExecutionHistoryService.instance) {
      ExecutionHistoryService.instance = new ExecutionHistoryService()
    }
    return ExecutionHistoryService.instance
  }

  /**
   * Add an execution to history
   */
  addExecution(entry: Omit<ExecutionHistoryEntry, 'id' | 'timestamp'>): string {
    const id = this.generateId()
    const timestamp = new Date().toISOString()
    
    const historyEntry: ExecutionHistoryEntry = {
      id,
      timestamp,
      ...entry
    }

    this.history.unshift(historyEntry)
    
    // Keep history size manageable
    if (this.history.length > this.maxHistorySize) {
      this.history.splice(this.maxHistorySize)
    }

    this.saveToStorage()
    return id
  }

  /**
   * Get execution history
   */
  getHistory(limit?: number): ExecutionHistoryEntry[] {
    return limit ? this.history.slice(0, limit) : [...this.history]
  }

  /**
   * Get history for specific artifact
   */
  getArtifactHistory(artifactId: string, limit?: number): ExecutionHistoryEntry[] {
    const artifactHistory = this.history.filter(entry => entry.artifactId === artifactId)
    return limit ? artifactHistory.slice(0, limit) : artifactHistory
  }

  /**
   * Get execution by ID
   */
  getExecution(id: string): ExecutionHistoryEntry | undefined {
    return this.history.find(entry => entry.id === id)
  }

  /**
   * Search history
   */
  searchHistory(query: string, filters?: {
    language?: string
    success?: boolean
    tags?: string[]
    dateFrom?: Date
    dateTo?: Date
  }): ExecutionHistoryEntry[] {
    let results = this.history

    // Text search
    if (query.trim()) {
      const searchTerm = query.toLowerCase()
      results = results.filter(entry => 
        entry.code.toLowerCase().includes(searchTerm) ||
        entry.notes?.toLowerCase().includes(searchTerm) ||
        entry.tags.some(tag => tag.toLowerCase().includes(searchTerm))
      )
    }

    // Apply filters
    if (filters) {
      if (filters.language) {
        results = results.filter(entry => entry.language === filters.language)
      }
      
      if (filters.success !== undefined) {
        results = results.filter(entry => entry.success === filters.success)
      }
      
      if (filters.tags && filters.tags.length > 0) {
        results = results.filter(entry => 
          filters.tags!.some(tag => entry.tags.includes(tag))
        )
      }
      
      if (filters.dateFrom) {
        results = results.filter(entry => 
          new Date(entry.timestamp) >= filters.dateFrom!
        )
      }
      
      if (filters.dateTo) {
        results = results.filter(entry => 
          new Date(entry.timestamp) <= filters.dateTo!
        )
      }
    }

    return results
  }

  /**
   * Get execution statistics
   */
  getStats(): ExecutionStats {
    const totalExecutions = this.history.length
    const totalTime = this.history.reduce((sum, entry) => sum + entry.executionTime, 0)
    const successfulExecutions = this.history.filter(entry => entry.success).length
    const successRate = totalExecutions > 0 ? successfulExecutions / totalExecutions : 0
    
    // Language breakdown
    const languageBreakdown: Record<string, number> = {}
    this.history.forEach(entry => {
      languageBreakdown[entry.language] = (languageBreakdown[entry.language] || 0) + 1
    })
    
    const averageExecutionTime = totalExecutions > 0 ? totalTime / totalExecutions : 0
    
    // Recent activity (last 10 executions)
    const recentActivity = this.history.slice(0, 10)
    
    return {
      totalExecutions,
      totalTime,
      successRate,
      languageBreakdown,
      averageExecutionTime,
      recentActivity
    }
  }

  /**
   * Update execution notes
   */
  updateNotes(id: string, notes: string): boolean {
    const entry = this.history.find(entry => entry.id === id)
    if (entry) {
      entry.notes = notes
      this.saveToStorage()
      return true
    }
    return false
  }

  /**
   * Add tags to execution
   */
  addTags(id: string, tags: string[]): boolean {
    const entry = this.history.find(entry => entry.id === id)
    if (entry) {
      const newTags = tags.filter(tag => !entry.tags.includes(tag))
      entry.tags.push(...newTags)
      this.saveToStorage()
      return true
    }
    return false
  }

  /**
   * Remove tags from execution
   */
  removeTags(id: string, tags: string[]): boolean {
    const entry = this.history.find(entry => entry.id === id)
    if (entry) {
      entry.tags = entry.tags.filter(tag => !tags.includes(tag))
      this.saveToStorage()
      return true
    }
    return false
  }

  /**
   * Delete execution from history
   */
  deleteExecution(id: string): boolean {
    const index = this.history.findIndex(entry => entry.id === id)
    if (index >= 0) {
      this.history.splice(index, 1)
      this.saveToStorage()
      return true
    }
    return false
  }

  /**
   * Clear all history
   */
  clearHistory(): void {
    this.history.splice(0)
    this.saveToStorage()
  }

  /**
   * Export history as JSON
   */
  exportHistory(): string {
    return JSON.stringify(this.history, null, 2)
  }

  /**
   * Import history from JSON
   */
  importHistory(jsonData: string): boolean {
    try {
      const importedHistory = JSON.parse(jsonData) as ExecutionHistoryEntry[]
      
      // Validate imported data
      if (!Array.isArray(importedHistory)) {
        throw new Error('Invalid format: expected array')
      }
      
      // Validate each entry
      importedHistory.forEach((entry, index) => {
        if (!entry.id || !entry.timestamp || !entry.code || !entry.language) {
          throw new Error(`Invalid entry at index ${index}: missing required fields`)
        }
      })
      
      // Merge with existing history, avoiding duplicates
      const existingIds = new Set(this.history.map(entry => entry.id))
      const newEntries = importedHistory.filter(entry => !existingIds.has(entry.id))
      
      this.history.unshift(...newEntries)
      
      // Sort by timestamp
      this.history.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
      
      // Trim to max size
      if (this.history.length > this.maxHistorySize) {
        this.history.splice(this.maxHistorySize)
      }
      
      this.saveToStorage()
      return true
    } catch (error) {
      console.error('Failed to import history:', error)
      return false
    }
  }

  /**
   * Get popular code snippets
   */
  getPopularSnippets(language?: string, limit = 10): Array<{
    code: string
    language: string
    count: number
    lastUsed: string
  }> {
    let entries = this.history
    
    if (language) {
      entries = entries.filter(entry => entry.language === language)
    }
    
    // Group by similar code (first 100 characters)
    const codeGroups: Record<string, {
      code: string
      language: string
      count: number
      lastUsed: string
    }> = {}
    
    entries.forEach(entry => {
      const key = entry.code.substring(0, 100).trim()
      if (key.length > 10) { // Ignore very short snippets
        if (!codeGroups[key]) {
          codeGroups[key] = {
            code: entry.code,
            language: entry.language,
            count: 0,
            lastUsed: entry.timestamp
          }
        }
        codeGroups[key].count++
        if (entry.timestamp > codeGroups[key].lastUsed) {
          codeGroups[key].lastUsed = entry.timestamp
        }
      }
    })
    
    return Object.values(codeGroups)
      .sort((a, b) => b.count - a.count)
      .slice(0, limit)
  }

  /**
   * Get performance trends
   */
  getPerformanceTrends(days = 30): Array<{
    date: string
    executions: number
    averageTime: number
    successRate: number
  }> {
    const cutoffDate = new Date()
    cutoffDate.setDate(cutoffDate.getDate() - days)
    
    const recentEntries = this.history.filter(entry => 
      new Date(entry.timestamp) >= cutoffDate
    )
    
    // Group by date
    const dailyData: Record<string, {
      executions: number
      totalTime: number
      successes: number
    }> = {}
    
    recentEntries.forEach(entry => {
      const date = new Date(entry.timestamp).toISOString().split('T')[0]
      if (!dailyData[date]) {
        dailyData[date] = { executions: 0, totalTime: 0, successes: 0 }
      }
      dailyData[date].executions++
      dailyData[date].totalTime += entry.executionTime
      if (entry.success) {
        dailyData[date].successes++
      }
    })
    
    return Object.entries(dailyData)
      .map(([date, data]) => ({
        date,
        executions: data.executions,
        averageTime: data.totalTime / data.executions,
        successRate: data.successes / data.executions
      }))
      .sort((a, b) => a.date.localeCompare(b.date))
  }

  private generateId(): string {
    return `exec_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private saveToStorage(): void {
    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.history))
    } catch (error) {
      console.warn('Failed to save execution history to localStorage:', error)
    }
  }

  private loadFromStorage(): void {
    try {
      const stored = localStorage.getItem(this.storageKey)
      if (stored) {
        const parsedHistory = JSON.parse(stored) as ExecutionHistoryEntry[]
        this.history.splice(0, this.history.length, ...parsedHistory)
      }
    } catch (error) {
      console.warn('Failed to load execution history from localStorage:', error)
    }
  }
}

// Export singleton instance
export const executionHistory = ExecutionHistoryService.getInstance()

// Export composable for Vue components
export function useExecutionHistory() {
  return {
    history: computed(() => executionHistory.getHistory()),
    stats: computed(() => executionHistory.getStats()),
    addExecution: executionHistory.addExecution.bind(executionHistory),
    getArtifactHistory: executionHistory.getArtifactHistory.bind(executionHistory),
    searchHistory: executionHistory.searchHistory.bind(executionHistory),
    updateNotes: executionHistory.updateNotes.bind(executionHistory),
    addTags: executionHistory.addTags.bind(executionHistory),
    removeTags: executionHistory.removeTags.bind(executionHistory),
    deleteExecution: executionHistory.deleteExecution.bind(executionHistory),
    clearHistory: executionHistory.clearHistory.bind(executionHistory),
    exportHistory: executionHistory.exportHistory.bind(executionHistory),
    importHistory: executionHistory.importHistory.bind(executionHistory),
    getPopularSnippets: executionHistory.getPopularSnippets.bind(executionHistory),
    getPerformanceTrends: executionHistory.getPerformanceTrends.bind(executionHistory)
  }
}