import type { ImageRecord } from './image'

export interface Session {
  id: string
  user_id: number
  title: string
  last_message_at: string
  message_count: number
  created_at: string
  updated_at: string
}

export interface Message {
  id: number
  type: 'user' | 'assistant'
  content: string
  image?: ImageRecord
  timestamp: string
}

export interface CreateSessionRequest {
  title?: string
}
