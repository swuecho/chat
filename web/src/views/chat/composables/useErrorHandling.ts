import { ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { t } from '@/locales'
import { useAuthStore } from '@/store'
import * as notificationManager from '@/utils/notificationManager'

interface AppError {
  code?: number | string
  message: string
  details?: any
  timestamp: Date
  context?: string
}

interface ErrorState {
  hasError: boolean
  currentError: AppError | null
  errorHistory: AppError[]
}

export function useErrorHandling() {
  const nui_msg = useMessage()
  
  const errorState = ref<ErrorState>({
    hasError: false,
    currentError: null,
    errorHistory: []
  })

  const hasRecentErrors = computed(() => {
    const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000)
    return errorState.value.errorHistory.some(error => error.timestamp > fiveMinutesAgo)
  })

  const errorCount = computed(() => errorState.value.errorHistory.length)

  function getErrorTitle(errorType: string, errorCode: string | number): string {
    switch (errorType) {
      case 'network':
        return 'Connection Problem'
      case 'server':
        return errorCode >= 500 ? 'Server Error' : 'Request Failed'
      case 'auth':
        return 'Authentication Required'
      case 'timeout':
        return 'Request Timeout'
      default:
        return 'Error'
    }
  }

  function logError(error: Partial<AppError>, context?: string): void {
    const appError: AppError = {
      code: error.code,
      message: error.message || 'Unknown error occurred',
      details: error.details,
      timestamp: new Date(),
      context: context || 'general'
    }

    errorState.value.errorHistory.push(appError)
    errorState.value.currentError = appError
    errorState.value.hasError = true

    // Limit error history to prevent memory leaks
    if (errorState.value.errorHistory.length > 50) {
      errorState.value.errorHistory.shift()
    }

    console.error(`[${context}] Error:`, appError)
  }

  function handleApiError(error: any, context: string = 'api'): void {
    let errorMessage = 'An unexpected error occurred'
    let errorCode: string | number = 'UNKNOWN'
    let errorType: 'network' | 'auth' | 'server' | 'client' | 'timeout' | 'unknown' = 'unknown'
    let action: { text: string; onClick: () => void } | undefined

    if (error?.response) {
      // HTTP error response
      errorCode = error.response.status
      errorMessage = error.response.data?.message || `HTTP ${error.response.status}`
      
      if (error.response.status === 401) {
        errorMessage = t('error.unauthorized') || 'Session expired. Please login again.'
        errorType = 'auth'
        action = {
          text: 'Login',
          onClick: () => {
            const authStore = useAuthStore()
            authStore.removeToken()
            authStore.removeExpiresIn()
          }
        }
      } else if (error.response.status === 403) {
        errorMessage = t('error.forbidden') || 'Access denied. You don\'t have permission for this action.'
        errorType = 'auth'
      } else if (error.response.status === 404) {
        errorMessage = t('error.notFound') || 'The requested resource was not found.'
        errorType = 'client'
      } else if (error.response.status === 429) {
        errorMessage = 'Too many requests. Please wait a moment before trying again.'
        errorType = 'client'
        action = {
          text: 'Retry',
          onClick: () => window.location.reload()
        }
      } else if (error.response.status >= 500) {
        errorMessage = t('error.serverError') || 'Server error. Our team has been notified and is working on a fix.'
        errorType = 'server'
        action = {
          text: 'Retry',
          onClick: () => window.location.reload()
        }
      } else {
        errorType = 'client'
      }
    } else if (error?.message) {
      // Network or other errors
      errorMessage = error.message
      
      if (error.message.includes('timeout') || error.message.includes('TIMEOUT')) {
        errorMessage = t('error.timeout') || 'Request timed out. Please check your connection and try again.'
        errorType = 'timeout'
        action = {
          text: 'Retry',
          onClick: () => window.location.reload()
        }
      } else if (error.message.includes('network') || error.message.includes('Network Error') || error.code === 'ECONNABORTED') {
        errorMessage = t('error.network') || 'Network connection error. Please check your internet connection.'
        errorType = 'network'
        action = {
          text: 'Retry',
          onClick: () => window.location.reload()
        }
      }
    } else if (error?.code === 'ERR_CANCELED') {
      errorMessage = 'Request was cancelled.'
      errorType = 'client'
      return // Don't show notification for cancelled requests
    }

    logError({
      code: errorCode,
      message: errorMessage,
      details: error
    }, context)

    // Use enhanced notifications for better visual hierarchy
    if (errorType === 'server' || errorType === 'network') {
      notificationManager.showEnhancedErrorNotification(
        getErrorTitle(errorType, errorCode),
        errorMessage,
        { persistent: true, action }
      )
    } else if (errorType === 'auth') {
      notificationManager.showEnhancedWarningNotification(
        'Authentication Required',
        errorMessage,
        { persistent: true, action }
      )
    } else {
      notificationManager.showEnhancedErrorNotification(
        'Error',
        errorMessage,
        { duration: 5000, action }
      )
    }
  }

  function handleStreamError(responseText: string, context: string = 'stream'): void {
    try {
      const errorJson = JSON.parse(responseText)
      const errorMessage = t(`error.${errorJson.code}`) || errorJson.message || 'Stream error occurred'
      
      logError({
        code: errorJson.code,
        message: errorMessage,
        details: errorJson.details
      }, context)

      // Handle specific stream errors with better messages
      let action: { text: string; onClick: () => void } | undefined
      if (errorJson.code === 'MODEL_006' || errorJson.code === 'INTN_004') {
        action = {
          text: 'Retry',
          onClick: () => window.location.reload()
        }
      }

      notificationManager.showEnhancedErrorNotification(
        'Stream Error', 
        errorMessage, 
        { duration: 5000, action }
      )
    } catch (parseError) {
      logError({
        message: 'Failed to parse error response',
        details: { responseText, parseError }
      }, context)

      notificationManager.showEnhancedErrorNotification(
        'Connection Error',
        'Connection interrupted. Please check your connection and try again.',
        { 
          persistent: true, 
          action: {
            text: 'Retry',
            onClick: () => window.location.reload()
          }
        }
      )
    }
  }

  function showErrorNotification(message: string, duration: number = 5000, action?: { text: string; onClick: () => void }): void {
    notificationManager.showErrorNotification(message, { duration, action })
  }

  function showWarningNotification(message: string, duration: number = 3000, action?: { text: string; onClick: () => void }): void {
    notificationManager.showWarningNotification(message, { duration, action })
  }

  function showSuccessNotification(message: string, duration: number = 3000): void {
    notificationManager.showSuccessNotification(message, { duration })
  }

  function showInfoNotification(message: string, duration: number = 3000): void {
    notificationManager.showInfoNotification(message, { duration })
  }

  function showPersistentErrorNotification(message: string, action?: { text: string; onClick: () => void }): void {
    notificationManager.showPersistentNotification(message, 'error', action)
  }

  function clearError(): void {
    errorState.value.hasError = false
    errorState.value.currentError = null
  }

  function clearErrorHistory(): void {
    errorState.value.errorHistory = []
    clearError()
  }

  function retryOperation<T>(
    operation: () => Promise<T>,
    maxRetries: number = 3,
    delay: number = 1000
  ): Promise<T> {
    return new Promise<T>(async (resolve, reject) => {
      for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
          const result = await operation()
          resolve(result)
          return
        } catch (error) {
          if (attempt === maxRetries) {
            handleApiError(error, 'retry-operation')
            reject(error)
            return
          }

          // Show retry notification
          if (attempt === 1) {
            showWarningNotification(`Retrying... (${attempt}/${maxRetries})`, 2000)
          }

          // Wait before retrying
          await new Promise(resolve => setTimeout(resolve, delay * attempt))
        }
      }
    })
  }

  function showNetworkStatusNotification(): void {
    if (!navigator.onLine) {
      showPersistentErrorNotification('You are offline. Please check your internet connection.', {
        text: 'Retry',
        onClick: () => window.location.reload()
      })
    }
  }

  function clearAllNotifications(): void {
    notificationManager.clearAllNotifications()
  }

  return {
    errorState: computed(() => errorState.value),
    hasRecentErrors,
    errorCount,
    logError,
    handleApiError,
    handleStreamError,
    showErrorNotification,
    showWarningNotification,
    showSuccessNotification,
    showInfoNotification,
    showPersistentErrorNotification,
    clearError,
    clearErrorHistory,
    retryOperation,
    showNetworkStatusNotification,
    clearAllNotifications
  }
}