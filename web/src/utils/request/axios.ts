import axios, { type AxiosResponse } from 'axios'
import { useAuthStore } from '@/store'
import { logger } from '@/utils/logger'

const service = axios.create({
  baseURL: "/api",
  withCredentials: true, // Include httpOnly cookies (for refresh token)
})

service.interceptors.request.use(
  async (config) => {
    const authStore = useAuthStore()

    // Skip token refresh for auth endpoints
    if (config.url?.includes('/auth/')) {
      return config
    }

    // Wait for auth initialization to complete before making API calls
    if (authStore.isInitializing) {
      logger.debug('Waiting for auth initialization to complete', 'Axios', { url: config.url })
      // Wait for initialization to complete
      while (authStore.isInitializing) {
        await new Promise(resolve => setTimeout(resolve, 50))
      }
      logger.debug('Auth initialization completed', 'Axios', { url: config.url })
    }

    // Check if token needs refresh
    if (authStore.needsRefresh && !authStore.isRefreshing) {
      try {
        await authStore.refreshToken()
      } catch (error) {
        // Refresh failed - redirect to login or handle appropriately
        console.error('Token refresh failed:', error)
      }
    }

    // Add access token to Authorization header
    const token = authStore.getToken
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error) => {
    return Promise.reject(error.response)
  },
)

service.interceptors.response.use(
  (response: AxiosResponse): AxiosResponse => {
    if (response.status === 200 || response.status === 201 || response.status === 204)
      return response
    throw new Error(response.status.toString())
  },
  async (error) => {
    const authStore = useAuthStore()

    logger.logApiError(error.config?.method || 'unknown', error.config?.url || 'unknown', error, error.response?.status)

    // Handle 401 errors with automatic token refresh
    if (error.response?.status === 401 && !error.config?.url?.includes('/auth/')) {
      logger.debug('Handling 401 error, attempting token refresh', 'Axios')
      try {
        await authStore.refreshToken()
        // Retry the original request with new token
        const token = authStore.getToken
        if (token) {
          logger.debug('Retrying request with new token', 'Axios')
          error.config.headers.Authorization = `Bearer ${token}`
          return service.request(error.config)
        }
      } catch (refreshError) {
        // Refresh failed - user needs to login again
        logger.warn('Token refresh failed, clearing auth state', 'Axios', refreshError)
        authStore.removeToken()
        authStore.removeExpiresIn()
        // Don't redirect - the login modal will appear automatically when authStore.isValid becomes false
      }
    }

    return Promise.reject(error)
  },
)

export default service
