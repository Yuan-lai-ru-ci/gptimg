export interface User {
  id: number
  username: string
  email: string
  avatar_url?: string
  credits: number
  quota_limit: number
  role: string
  status: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}
