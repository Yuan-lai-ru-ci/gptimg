import type { ImageRecord } from './image'

export type PPTWorkflowMode = 'style_text' | 'image_content_style' | 'free_mix'

export interface PPTSlidePlan {
  slide_number: number
  title: string
  objective: string
  layout_notes: string
  page_description: string
  image_prompt: string
  speaker_notes: string
}

export interface PPTPlan {
  deck_title: string
  deck_goal: string
  visual_direction: string
  master_style_description: string
  consistency_rules: string[]
  generation_mode?: PPTWorkflowMode
  slides: PPTSlidePlan[]
}

export interface PlanPPTRequest {
  requirement: string
  slide_count?: number
  generation_mode?: PPTWorkflowMode
}

export interface PlanPPTDocumentRequest {
  document: File
  requirement?: string
  slide_count?: number
}

export interface GeneratePPTRequest {
  session_id: string
  generation_mode?: PPTWorkflowMode
  deck_title: string
  deck_goal: string
  visual_direction: string
  master_style_description: string
  consistency_rules: string[]
  slides: PPTSlidePlan[]
  size?: string
  quality?: string
  style?: string
  reference_images?: File[]
}

export interface PPTGeneratedSlide {
  slide_number: number
  title: string
  record: ImageRecord
}
