interface ApiConfig {
  baseURL: string
  streamingURL: string
}

/**
 * Get API configuration based on environment
 */
export function getApiConfig(): ApiConfig {
  // Check for explicit configuration first
  const customBackendUrl = (import.meta as any).env?.VITE_BACKEND_URL
  const customStreamingUrl = (import.meta as any).env?.VITE_STREAMING_URL
  
  // If both are explicitly configured, use them
  if (customBackendUrl && customStreamingUrl) {
    return {
      baseURL: customBackendUrl,
      streamingURL: customStreamingUrl
    }
  }
  
  // Environment-based defaults
  const isDevelopment = process.env.NODE_ENV === 'development'
  
  if (isDevelopment) {
    // In development, use direct backend URL for streaming to bypass proxy buffering
    return {
      baseURL: '/api', // Use proxy for regular API calls
      streamingURL: customStreamingUrl || 'http://localhost:8080/api' // Direct connection for streaming
    }
  }
  
  // Production defaults
  return {
    baseURL: customBackendUrl || '/api',
    streamingURL: customStreamingUrl || '/api'
  }
}

/**
 * Get the appropriate URL for streaming endpoints
 */
export function getStreamingUrl(endpoint: string): string {
  const config = getApiConfig()
  return `${config.streamingURL}${endpoint}`
}

/**
 * Get the appropriate URL for regular API endpoints
 */
export function getApiUrl(endpoint: string): string {
  const config = getApiConfig()
  return `${config.baseURL}${endpoint}`
}