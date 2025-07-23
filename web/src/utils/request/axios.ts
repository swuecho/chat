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
    
    // Handle 401 errors with automatic token refresh
    if (error.response?.status === 401 && !error.config?.url?.includes('/auth/')) {
      try {
        await authStore.refreshToken()
        // Retry the original request with new token
        const token = authStore.getToken()
        if (token) {
          error.config.headers.Authorization = `Bearer ${token}`
          return service.request(error.config)
        }
      } catch (refreshError) {
        // Refresh failed - user needs to login again
        authStore.removeToken()
        authStore.removeExpiresIn()
        // Redirect to login or show login modal
        window.location.href = '/login'
      }
    }
    
    return Promise.reject(error)
  },
)

export default service
