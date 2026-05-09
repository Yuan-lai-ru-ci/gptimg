'use client'

import { useEffect } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/store/authStore'
import SettingsPanel from '@/components/settings/SettingsPanel'

export default function SettingsPage() {
  const router = useRouter()
  const { isAuthenticated, isLoading, loadAuth } = useAuthStore()
  useEffect(() => { loadAuth() }, [loadAuth])
  useEffect(() => { if (!isLoading && !isAuthenticated) router.push('/login') }, [isAuthenticated, isLoading, router])
  if (isLoading || !isAuthenticated) return <div className="min-h-screen bg-background flex items-center justify-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div></div>
  return <div className="min-h-screen bg-background"><nav className="border-b border-border px-4 py-3 sm:p-4"><div className="mx-auto flex max-w-4xl items-center justify-between gap-3"><h1 className="text-foreground font-bold">GPT Image</h1><Link href="/chat" className="text-right text-sm text-muted-foreground hover:text-foreground">Back to Chat</Link></div></nav><SettingsPanel /></div>
}
