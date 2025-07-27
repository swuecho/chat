import axios, { type AxiosResponse } from 'axios'
import { useAuthStore } from '@/store'

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
      console.log('⏳ Waiting for auth initialization to complete before API call:', config.url)
      // Wait for initialization to complete
      while (authStore.isInitializing) {
        await new Promise(resolve => setTimeout(resolve, 50))
      }
      console.log('✅ Auth initialization completed, proceeding with API call:', config.url)
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
    const token = authStore.getToken()
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

    console.log('Axios response error:', error.response?.status, error.config?.url)

    // Handle 401 errors with automatic token refresh
    if (error.response?.status === 401 && !error.config?.url?.includes('/auth/')) {
      console.log('Handling 401 error, attempting token refresh...')
      try {
        await authStore.refreshToken()
        // Retry the original request with new token
        const token = authStore.getToken()
        if (token) {
          console.log('Retrying request with new token...')
          error.config.headers.Authorization = `Bearer ${token}`
          return service.request(error.config)
        }
      } catch (refreshError) {
        // Refresh failed - user needs to login again
        console.log('Token refresh failed, clearing auth state...')
        authStore.removeToken()
        authStore.removeExpiresIn()
        // Don't redirect - the login modal will appear automatically when authStore.isValid becomes false
        console.error('Token refresh failed, user needs to login again:', refreshError)
      }
    }

    return Promise.reject(error)
  },
)

export default service
