'use client'

import { useEffect, useState } from 'react'
import { useChatStore } from '@/lib/store/chatStore'
import { useAuthStore } from '@/lib/store/authStore'
import Sidebar from './Sidebar'
import MessageList from './MessageList'
import InputBox from './InputBox'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

export default function ChatContainer() {
  const { loadSessions, currentSession } = useChatStore()
  const { user } = useAuthStore()
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <Sidebar isOpen={isSidebarOpen} onClose={() => setIsSidebarOpen(false)} />

      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <header className="flex items-center gap-3 border-b border-border px-4 py-3 md:hidden">
          <Button variant="ghost" size="icon" onClick={() => setIsSidebarOpen(true)} aria-label="Open menu">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          </Button>
          <div className="min-w-0 flex-1 text-center">
            <div className="truncate text-sm font-medium text-foreground">
              {currentSession?.title || 'GPT Image'}
            </div>
          </div>
          <div className="flex h-7 w-7 items-center justify-center rounded-full bg-accent text-xs font-medium text-accent-foreground">
            {user?.username?.slice(0, 2)?.toUpperCase()}
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-hidden">
          <MessageList />
        </div>

        <Separator />
        <InputBox />
      </div>
    </div>
  )
}
