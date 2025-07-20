import axios, { type AxiosResponse } from 'axios'
import { useAuthStore } from '@/store'

// For local development: set to true to bypass proxy and connect directly to backend
const USE_DIRECT_API_IN_DEV = true

const service = axios.create({
  baseURL: USE_DIRECT_API_IN_DEV && window.location.hostname === 'localhost'
    ? "http://localhost:8080/api"
    : "/api"
})

service.interceptors.request.use(
  (config) => {
    const token = useAuthStore().getToken()

    // clear token if expired
    const expiresIn = useAuthStore().getExpiresIn()
    if (expiresIn && expiresIn < Date.now() / 1000)
      useAuthStore().removeToken()

    if (token)
      config.headers.Authorization = `Bearer ${token}`
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
  (error) => {
    return Promise.reject(error)
  },
)

export default service
