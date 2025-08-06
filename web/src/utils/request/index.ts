import type { AxiosProgressEvent, AxiosResponse, GenericAbortSignal } from 'axios'
import request from './axios'
import { useAuthStore } from '@/store'

export interface HttpOption {
  url: string
  // rome-ignore lint/suspicious/noExplicitAny: <explanation>
  data?: any
  method?: string
  // rome-ignore lint/suspicious/noExplicitAny: <explanation>
  headers?: any
  onDownloadProgress?: (progressEvent: AxiosProgressEvent) => void
  signal?: GenericAbortSignal
  beforeRequest?: () => void
  afterRequest?: () => void
}

export interface Response<T> {
  data: T
  message: string | null
  status: string
}

function http<T>(
  { url, data, method, headers, onDownloadProgress, signal, beforeRequest, afterRequest }: HttpOption,
) {
  const successHandler = (res: AxiosResponse<Response<T>>) => {
    const authStore = useAuthStore()

    if (res.data.status === 'Success' || typeof res.data === 'string')
      return res.data

    if (res.data.status === 'Unauthorized') {
      authStore.removeToken()
      window.location.reload()
    }

    return Promise.reject(res.data)
  }

  const failHandler = (error: any) => {
    afterRequest?.()
    
    // Enhanced error handling with more detailed error information
    let errorMessage = 'An unexpected error occurred'
    let errorCode = 'UNKNOWN_ERROR'
    
    if (error?.response?.data) {
      errorMessage = error.response.data.message || errorMessage
      errorCode = error.response.data.code || errorCode
    } else if (error?.message) {
      errorMessage = error.message
    } else if (typeof error === 'string') {
      errorMessage = error
    }
    
    // Create enhanced error object with proper typing
    interface EnhancedError extends Error {
      code?: string | number
      status?: number
      originalError?: any
    }
    
    const enhancedError = new Error(errorMessage) as EnhancedError
    enhancedError.name = errorCode
    enhancedError.code = errorCode
    enhancedError.status = error?.response?.status || 0
    enhancedError.originalError = error
    
    throw enhancedError
  }

  beforeRequest?.()

  method = method || 'GET'

  const params = Object.assign(typeof data === 'function' ? data() : data ?? {}, {})

  return method === 'GET'
    ? request.get(url, { params, signal, onDownloadProgress }).then(successHandler, failHandler)
    : request.post(url, params, { headers, signal, onDownloadProgress }).then(successHandler, failHandler)
}

export function get<T>(
  { url, data, method = 'GET', onDownloadProgress, signal, beforeRequest, afterRequest }: HttpOption,
): Promise<Response<T>> {
  return http<T>({
    url,
    method,
    data,
    onDownloadProgress,
    signal,
    beforeRequest,
    afterRequest,
  })
}

export function post<T>(
  { url, data, method = 'POST', headers, onDownloadProgress, signal, beforeRequest, afterRequest }: HttpOption,
): Promise<Response<T>> {
  return http<T>({
    url,
    method,
    data,
    headers,
    onDownloadProgress,
    signal,
    beforeRequest,
    afterRequest,
  })
}

export default post
