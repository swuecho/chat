import { ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { t } from '@/locales'

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

    if (error?.response) {
      // HTTP error response
      errorCode = error.response.status
      errorMessage = error.response.data?.message || `HTTP ${error.response.status}`
      
      if (error.response.status === 401) {
        errorMessage = t('error.unauthorized') || 'Unauthorized access'
      } else if (error.response.status === 403) {
        errorMessage = t('error.forbidden') || 'Access forbidden'
      } else if (error.response.status === 404) {
        errorMessage = t('error.notFound') || 'Resource not found'
      } else if (error.response.status >= 500) {
        errorMessage = t('error.serverError') || 'Server error occurred'
      }
    } else if (error?.message) {
      // Network or other errors
      errorMessage = error.message
      if (error.message.includes('timeout')) {
        errorMessage = t('error.timeout') || 'Request timed out'
      } else if (error.message.includes('network')) {
        errorMessage = t('error.network') || 'Network error'
      }
    }

    logError({
      code: errorCode,
      message: errorMessage,
      details: error
    }, context)

    showErrorNotification(errorMessage)
  }

  function handleStreamError(responseText: string, context: string = 'stream'): void {
    try {
      const errorJson = JSON.parse(responseText)
      const errorMessage = t(`error.${errorJson.code}`) || errorJson.message
      
      logError({
        code: errorJson.code,
        message: errorMessage,
        details: errorJson.details
      }, context)

      showErrorNotification(`${errorJson.code}: ${errorMessage}`)
    } catch (parseError) {
      logError({
        message: 'Failed to parse error response',
        details: { responseText, parseError }
      }, context)

      showErrorNotification('An unexpected error occurred')
    }
  }

  function showErrorNotification(message: string, duration: number = 5000): void {
    nui_msg.error(message, {
      duration,
      closable: true
    })
  }

  function showWarningNotification(message: string, duration: number = 3000): void {
    nui_msg.warning(message, {
      duration,
      closable: true
    })
  }

  function showSuccessNotification(message: string, duration: number = 3000): void {
    nui_msg.success(message, {
      duration,
      closable: true
    })
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

          // Wait before retrying
          await new Promise(resolve => setTimeout(resolve, delay * attempt))
        }
      }
    })
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
    clearError,
    clearErrorHistory,
    retryOperation
  }
}