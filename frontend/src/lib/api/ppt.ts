import apiClient from './client'
import type { GeneratePPTRequest, PPTGeneratedSlide, PPTPlan, PlanPPTDocumentRequest, PlanPPTRequest } from '@/types/ppt'

interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export const pptApi = {
  plan: async (data: PlanPPTRequest): Promise<ApiResponse<PPTPlan>> => {
    return apiClient.post('/ppt/plan', data)
  },

  planDocument: async (data: PlanPPTDocumentRequest): Promise<ApiResponse<PPTPlan>> => {
    const formData = new FormData()
    formData.append('document', data.document)
    if (data.requirement) formData.append('requirement', data.requirement)
    if (data.slide_count) formData.append('slide_count', String(data.slide_count))

    return apiClient.post('/ppt/plan-document', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  generate: async (
    data: GeneratePPTRequest
  ): Promise<ApiResponse<{ deck_title: string; slides: PPTGeneratedSlide[] }>> => {
    if (data.reference_images?.length) {
      const formData = new FormData()
      formData.append('session_id', data.session_id)
      if (data.generation_mode) formData.append('generation_mode', data.generation_mode)
      formData.append('deck_title', data.deck_title)
      formData.append('deck_goal', data.deck_goal)
      formData.append('visual_direction', data.visual_direction)
      formData.append('master_style_description', data.master_style_description)
      formData.append('consistency_rules', JSON.stringify(data.consistency_rules))
      formData.append('slides', JSON.stringify(data.slides))
      if (data.size) formData.append('size', data.size)
      if (data.quality) formData.append('quality', data.quality)
      if (data.style) formData.append('style', data.style)
      data.reference_images.forEach((file) => formData.append('reference_images', file))

      return apiClient.post('/ppt/generate', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
    }

    return apiClient.post('/ppt/generate', data)
  },
}
