import apiClient from './client'
import type { ImageRecord, GenerateImageRequest } from '@/types/image'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export const imageApi = {
  generate: async (data: GenerateImageRequest): Promise<ApiResponse<ImageRecord>> => {
    const referenceImages = data.reference_images?.length
      ? data.reference_images
      : data.reference_image
        ? [data.reference_image]
        : []

    if (referenceImages.length > 0) {
      const formData = new FormData()
      formData.append('prompt', data.prompt)
      if (data.session_id) formData.append('session_id', data.session_id)
      if (data.size) formData.append('size', data.size)
      if (data.quality) formData.append('quality', data.quality)
      if (data.style) formData.append('style', data.style)
      referenceImages.forEach((file) => formData.append('reference_images', file))

      return apiClient.post('/images/generate', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    }

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
