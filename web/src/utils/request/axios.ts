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

    // Check if token is expired before making request
    if (!authStore.isValid) {
      logger.debug('Token is expired or invalid, attempting refresh', 'Axios', { url: config.url })
      try {
        await authStore.refreshToken()
        // Check again after refresh
        if (!authStore.isValid) {
          logger.warn('Token still invalid after refresh attempt', 'Axios')
          return Promise.reject(new Error('Authentication required'))
        }
      } catch (error) {
        logger.error('Token refresh failed in request interceptor', 'Axios', error)
        return Promise.reject(new Error('Authentication required'))
      }
    } 
    // Check if token needs refresh (expires within 5 minutes)
    else if (authStore.needsRefresh && !authStore.isRefreshing) {
      logger.debug('Token needs refresh, refreshing proactively', 'Axios', { url: config.url })
      try {
        await authStore.refreshToken()
      } catch (error) {
        logger.error('Proactive token refresh failed', 'Axios', error)
        // Continue with existing token if proactive refresh fails
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
      
      // Prevent infinite retry loops
      if (error.config._retryCount >= 1) {
        logger.warn('Already retried once, clearing auth state', 'Axios')
        authStore.removeToken()
        authStore.removeExpiresIn()
        return Promise.reject(new Error('Authentication failed after retry'))
      }
      
      try {
        await authStore.refreshToken()
        // Check if refresh was successful
        if (!authStore.isValid) {
          logger.warn('Token invalid after refresh attempt', 'Axios')
          authStore.removeToken()
          authStore.removeExpiresIn()
          return Promise.reject(new Error('Authentication failed'))
        }
        
        // Retry the original request with new token
        const token = authStore.getToken
        if (token) {
          logger.debug('Retrying request with new token', 'Axios')
          error.config.headers.Authorization = `Bearer ${token}`
          error.config._retryCount = (error.config._retryCount || 0) + 1
          return service.request(error.config)
        }
      } catch (refreshError) {
        // Refresh failed - user needs to login again
        logger.warn('Token refresh failed, clearing auth state', 'Axios', refreshError)
        authStore.removeToken()
        authStore.removeExpiresIn()
        return Promise.reject(new Error('Authentication required'))
      }
    }

    return Promise.reject(error)
  },
)

export default service
