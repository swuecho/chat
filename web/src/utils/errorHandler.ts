import { logger } from './logger'
import { useAuthStore } from '@/store'

export interface ApiError {
  status: number
  message: string
  code?: string
  details?: any
}

export interface ErrorHandlerOptions {
  logError?: boolean
  showToast?: boolean
  redirectOnError?: boolean
  retryCount?: number
  retryDelay?: number
}

class ErrorHandler {
  private defaultOptions: ErrorHandlerOptions = {
    logError: true,
    showToast: true,
    redirectOnError: true,
    retryCount: 0,
    retryDelay: 1000,
  }

  private isNetworkError(error: any): boolean {
    return (
      !navigator.onLine ||
      error.message === 'Network Error' ||
      error.code === 'ECONNABORTED' ||
      error.code === 'ETIMEDOUT'
    )
  }

  private isAuthError(error: any): boolean {
    return error.status === 401 || error.status === 403
  }

  private isServerError(error: any): boolean {
    return error.status >= 500 && error.status < 600
  }

  private isClientError(error: any): boolean {
    return error.status >= 400 && error.status < 500
  }

  private extractErrorMessage(error: any): string {
    if (error.response?.data?.message) {
      return error.response.data.message
    }
    if (error.message) {
      return error.message
    }
    if (typeof error === 'string') {
      return error
    }
    return 'An unknown error occurred'
  }

  private logApiError(method: string, url: string, error: any, options: ErrorHandlerOptions): void {
    if (!options.logError) return

    const apiError: ApiError = {
      status: error.response?.status || 0,
      message: this.extractErrorMessage(error),
      code: error.code,
      details: error.response?.data,
    }

    logger.logApiError(method, url, apiError, apiError.status)
  }

  private async handleAuthError(error: any): Promise<void> {
    const authStore = useAuthStore()
    
    try {
      logger.debug('Attempting token refresh for auth error', 'ErrorHandler')
      await authStore.refreshToken()
    } catch (refreshError) {
      logger.error('Token refresh failed, clearing auth state', 'ErrorHandler', refreshError)
      authStore.removeToken()
      authStore.removeExpiresIn()
      
      // Don't redirect immediately - let the auth store handle the UI state change
      // The login modal will appear automatically when authStore.isValid becomes false
    }
  }

  private async retryRequest(
    requestFn: () => Promise<any>,
    retryCount: number,
    retryDelay: number
  ): Promise<any> {
    let lastError: any

    for (let attempt = 1; attempt <= retryCount; attempt++) {
      try {
        return await requestFn()
      } catch (error) {
        lastError = error
        
        if (attempt === retryCount) {
          throw error
        }

        if (this.isAuthError(error)) {
          // Don't retry auth errors
          throw error
        }

        logger.debug(`Retrying request (attempt ${attempt + 1}/${retryCount})`, 'ErrorHandler', { error })
        
        // Exponential backoff
        const delay = retryDelay * Math.pow(2, attempt - 1)
        await new Promise(resolve => setTimeout(resolve, delay))
      }
    }

    throw lastError
  }

  async handleApiRequest<T>(
    requestFn: () => Promise<T>,
    method: string,
    url: string,
    options: ErrorHandlerOptions = {}
  ): Promise<T> {
    const finalOptions = { ...this.defaultOptions, ...options }
    const startTime = Date.now()

    try {
      if (finalOptions.retryCount > 0) {
        const result = await this.retryRequest(requestFn, finalOptions.retryCount, finalOptions.retryDelay)
        const duration = Date.now() - startTime
        logger.logApiCall(method, url, 200, duration)
        return result
      } else {
        const result = await requestFn()
        const duration = Date.now() - startTime
        logger.logApiCall(method, url, 200, duration)
        return result
      }
    } catch (error: any) {
      this.logApiError(method, url, error, finalOptions)

      // Handle network errors
      if (this.isNetworkError(error)) {
        logger.warn('Network error detected', 'ErrorHandler', { error })
        throw {
          status: 0,
          message: 'Network error. Please check your internet connection.',
          originalError: error,
        }
      }

      // Handle authentication errors
      if (this.isAuthError(error)) {
        await this.handleAuthError(error)
        throw {
          status: error.status,
          message: 'Authentication failed. Please login again.',
          originalError: error,
        }
      }

      // Handle server errors
      if (this.isServerError(error)) {
        logger.error('Server error occurred', 'ErrorHandler', error)
        throw {
          status: error.status,
          message: 'Server error. Please try again later.',
          originalError: error,
        }
      }

      // Handle client errors
      if (this.isClientError(error)) {
        logger.warn('Client error occurred', 'ErrorHandler', error)
        throw {
          status: error.status,
          message: this.extractErrorMessage(error),
          originalError: error,
        }
      }

      // Handle unknown errors
      logger.error('Unknown error occurred', 'ErrorHandler', error)
      throw {
        status: 0,
        message: 'An unexpected error occurred.',
        originalError: error,
      }
    }
  }

  // Convenience method for GET requests
  async get<T>(url: string, options?: ErrorHandlerOptions): Promise<T> {
    return this.handleApiRequest(
      () => fetch(url, { method: 'GET' }),
      'GET',
      url,
      options
    )
  }

  // Convenience method for POST requests
  async post<T>(url: string, data?: any, options?: ErrorHandlerOptions): Promise<T> {
    return this.handleApiRequest(
      () => fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      }),
      'POST',
      url,
      options
    )
  }

  // Convenience method for PUT requests
  async put<T>(url: string, data?: any, options?: ErrorHandlerOptions): Promise<T> {
    return this.handleApiRequest(
      () => fetch(url, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      }),
      'PUT',
      url,
      options
    )
  }

  // Convenience method for DELETE requests
  async delete<T>(url: string, options?: ErrorHandlerOptions): Promise<T> {
    return this.handleApiRequest(
      () => fetch(url, { method: 'DELETE' }),
      'DELETE',
      url,
      options
    )
  }

  // Global error handler for unhandled promise rejections
  setupGlobalErrorHandler(): void {
    window.addEventListener('unhandledrejection', (event) => {
      logger.error('Unhandled promise rejection', 'GlobalErrorHandler', event.reason)
      
      // Prevent default behavior (logging to console)
      event.preventDefault()
      
      // You could also show a user-friendly error message here
      if (event.reason instanceof Error) {
        console.error('An unexpected error occurred:', event.reason.message)
      }
    })

    window.addEventListener('error', (event) => {
      logger.error('Global error occurred', 'GlobalErrorHandler', {
        message: event.message,
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        error: event.error,
      })
    })
  }
}

// Export singleton instance
export const errorHandler = new ErrorHandler()

// Export convenience functions
export const handleApiRequest = <T>(
  requestFn: () => Promise<T>,
  method: string,
  url: string,
  options?: ErrorHandlerOptions
) => errorHandler.handleApiRequest(requestFn, method, url, options)

// Default export
export default errorHandler