import axios from 'axios'
import { withBasePath } from '@/lib/runtime'

function getApiBaseUrl() {
  if (process.env.NEXT_PUBLIC_API_BASE_URL) {
    return process.env.NEXT_PUBLIC_API_BASE_URL
  }

  if (typeof window !== 'undefined') {
    return withBasePath('/api/v1').replace('/gptimg/api/v1', '/gptimg-api/api/v1')
  }

  return 'http://127.0.0.1:8080/api/v1'
}

const apiClient = axios.create({
  baseURL: getApiBaseUrl(),
  timeout: 330000,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use(
  (config) => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    return config
  },
  (error) => Promise.reject(error)
)

apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401 && typeof window !== 'undefined') {
      localStorage.removeItem('access_token')
      localStorage.removeItem('user')
      window.location.assign(withBasePath('/login'))
    }
    const errData = error.response?.data
    return Promise.reject(new Error(errData?.message || error.message || 'Request failed'))
  }
)

export default apiClient
