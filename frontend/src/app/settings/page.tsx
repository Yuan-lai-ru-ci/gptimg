'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/store/authStore'
import SettingsPanel from '@/components/settings/SettingsPanel'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

export default function SettingsPage() {
  const router = useRouter()
  const { isAuthenticated, isLoading, loadAuth } = useAuthStore()

  useEffect(() => {
    loadAuth()
  }, [loadAuth])

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login')
    }
  }, [isAuthenticated, isLoading, router])

  if (isLoading || !isAuthenticated) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-foreground"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <nav className="border-b border-border p-4">
        <div className="max-w-4xl mx-auto flex justify-between items-center">
          <h1 className="text-foreground font-bold">GPT Image</h1>
          <Button variant="ghost" size="sm" asChild>
            <a href="/chat">Back to Chat</a>
          </Button>
        </div>
      </nav>
      <SettingsPanel />
    </div>
  )
}
