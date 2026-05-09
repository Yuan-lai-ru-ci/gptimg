'use client'

import { useEffect } from 'react'
import { useChatStore } from '@/lib/store/chatStore'
import { useAuthStore } from '@/lib/store/authStore'
import Sidebar from './Sidebar'
import MessageList from './MessageList'
import InputBox from './InputBox'
import { Separator } from '@/components/ui/separator'

export default function ChatContainer() {
  const { loadSessions, currentSession } = useChatStore()
  const { user } = useAuthStore()

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />

      <div className="flex-1 flex flex-col">
        <div className="flex-1 overflow-hidden">
          <MessageList />
        </div>

        <Separator />
        <InputBox />
      </div>
    </div>
  )
}
