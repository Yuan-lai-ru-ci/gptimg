export interface ImageRecord {
  id: number
  user_id: number
  session_id: string
  prompt: string
  revised_prompt?: string
  image_url: string
  local_path: string
  size: string
  quality: string
  style?: string
  model: string
  credits_used: number
  generation_time: number
  status: string
  error_message?: string
  created_at: string
}

export interface GenerateImageRequest {
  prompt: string
  session_id?: string
  size?: string
  quality?: string
  style?: string
  reference_image?: File | null
  reference_images?: File[]
}

export interface ImageOptions {
  size: '1024x1024' | '1792x1024' | '1536x1024' | '1024x1536'
  quality: 'standard' | 'hd'
  style?: 'vivid' | 'natural'
}
