'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/store/authStore'

export default function Home() {
  const router = useRouter()
  const { isAuthenticated, isLoading, loadAuth } = useAuthStore()

  useEffect(() => {
    loadAuth()
  }, [loadAuth])

  useEffect(() => {
    if (!isLoading) {
      if (isAuthenticated) {
        router.push('/chat')
      } else {
        router.push('/login')
      }
    }
  }, [isAuthenticated, isLoading, router])

  return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-foreground"></div>
    </div>
  )
}
