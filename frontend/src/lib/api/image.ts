import apiClient from './client'
import type { ImageRecord, GenerateImageRequest } from '@/types/image'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export const imageApi = {
  generate: async (data: GenerateImageRequest): Promise<ApiResponse<ImageRecord>> => {
    return apiClient.post('/images/generate', data)
  },

  getById: async (id: number): Promise<ApiResponse<ImageRecord>> => {
    return apiClient.get(`/images/${id}`)
  },

  delete: async (id: number): Promise<void> => {
    return apiClient.delete(`/images/${id}`)
  },

  getHistory: async (limit = 20, offset = 0): Promise<ApiResponse<ImageRecord[]>> => {
    return apiClient.get('/history', { params: { limit, offset } })
  },
}
