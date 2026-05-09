'use client'

import { useEffect, useState } from 'react'
import { useChatStore } from '@/lib/store/chatStore'
import { useAuthStore } from '@/lib/store/authStore'
import Sidebar from './Sidebar'
import MessageList from './MessageList'
import InputBox from './InputBox'

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
        <div className="flex items-center justify-between border-b border-border px-4 py-3 md:hidden">
          <button
            type="button"
            onClick={() => setIsSidebarOpen(true)}
            className="rounded-lg border border-border bg-card p-2 text-gray-200 transition-colors hover:text-foreground"
            aria-label="Open conversations"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          </button>
          <div className="min-w-0 px-3 text-center">
            <div className="truncate text-sm font-semibold text-foreground">
              {currentSession?.title || 'GPT Image'}
            </div>
            <div className="truncate text-xs text-muted-foreground">
              {user?.credits ?? 0} credits
            </div>
          </div>
          <div className="w-10 text-right text-xs text-muted-foreground">
            {user?.username?.slice(0, 2)?.toUpperCase()}
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-hidden">
          <MessageList />
        </div>

        <div className="border-t border-border">
          <InputBox />
        </div>
      </div>
    </div>
  )
}
