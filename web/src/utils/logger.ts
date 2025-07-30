export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

export interface LogEntry {
  level: LogLevel
  message: string
  timestamp: string
  context?: string
  data?: any
}

class Logger {
  private level: LogLevel
  private isProduction: boolean
  private logs: LogEntry[] = []

  constructor() {
    this.level = this.getLogLevelFromEnv()
    this.isProduction = process.env.NODE_ENV === 'production'
  }

  private getLogLevelFromEnv(): LogLevel {
    const envLevel = process.env.VUE_APP_LOG_LEVEL
    switch (envLevel) {
      case 'debug': return LogLevel.DEBUG
      case 'info': return LogLevel.INFO
      case 'warn': return LogLevel.WARN
      case 'error': return LogLevel.ERROR
      default: return this.isProduction ? LogLevel.WARN : LogLevel.DEBUG
    }
  }

  private shouldLog(level: LogLevel): boolean {
    return level >= this.level
  }

  private createLogEntry(level: LogLevel, message: string, context?: string, data?: any): LogEntry {
    return {
      level,
      message,
      timestamp: new Date().toISOString(),
      context,
      data,
    }
  }

  private formatMessage(entry: LogEntry): string {
    const levelName = LogLevel[entry.level]
    const contextStr = entry.context ? `[${entry.context}] ` : ''
    const dataStr = entry.data ? ` ${JSON.stringify(entry.data)}` : ''
    return `${entry.timestamp} [${levelName}] ${contextStr}${entry.message}${dataStr}`
  }

  private log(level: LogLevel, message: string, context?: string, data?: any): void {
    if (!this.shouldLog(level)) {
      return
    }

    const entry = this.createLogEntry(level, message, context, data)
    this.logs.push(entry)

    // Only log to console in development or for errors/warnings
    if (!this.isProduction || level >= LogLevel.ERROR) {
      const formattedMessage = this.formatMessage(entry)
      
      switch (level) {
        case LogLevel.DEBUG:
          console.debug(formattedMessage)
          break
        case LogLevel.INFO:
          console.info(formattedMessage)
          break
        case LogLevel.WARN:
          console.warn(formattedMessage)
          break
        case LogLevel.ERROR:
          console.error(formattedMessage)
          break
      }
    }
  }

  // Public logging methods
  debug(message: string, context?: string, data?: any): void {
    this.log(LogLevel.DEBUG, message, context, data)
  }

  info(message: string, context?: string, data?: any): void {
    this.log(LogLevel.INFO, message, context, data)
  }

  warn(message: string, context?: string, data?: any): void {
    this.log(LogLevel.WARN, message, context, data)
  }

  error(message: string, context?: string, data?: any): void {
    this.log(LogLevel.ERROR, message, context, data)
  }

  // Specialized logging methods for common scenarios
  logApiCall(method: string, url: string, status?: number, duration?: number): void {
    this.debug(`API ${method} ${url}`, 'API', { method, url, status, duration })
  }

  logApiError(method: string, url: string, error: any, status?: number): void {
    this.error(`API Error ${method} ${url}`, 'API', { method, url, error, status })
  }

  logStoreAction(action: string, store: string, data?: any): void {
    this.debug(`Store action: ${action}`, store, data)
  }

  logPerformance(metric: string, value: number, unit: string = 'ms'): void {
    this.debug(`Performance: ${metric} = ${value}${unit}`, 'Performance', { metric, value, unit })
  }

  logUserAction(action: string, details?: any): void {
    this.info(`User action: ${action}`, 'User', details)
  }

  // Get logs for debugging
  getLogs(level?: LogLevel): LogEntry[] {
    if (level !== undefined) {
      return this.logs.filter(log => log.level >= level)
    }
    return [...this.logs]
  }

  // Clear logs
  clearLogs(): void {
    this.logs = []
  }

  // Set log level dynamically
  setLevel(level: LogLevel): void {
    this.level = level
  }

  // Export logs for debugging
  exportLogs(): string {
    return JSON.stringify(this.logs, null, 2)
  }
}

// Export singleton instance
export const logger = new Logger()

// Export convenience functions for direct use
export const debug = (message: string, context?: string, data?: any) => logger.debug(message, context, data)
export const info = (message: string, context?: string, data?: any) => logger.info(message, context, data)
export const warn = (message: string, context?: string, data?: any) => logger.warn(message, context, data)
export const error = (message: string, context?: string, data?: any) => logger.error(message, context, data)

// Default export
export default logger