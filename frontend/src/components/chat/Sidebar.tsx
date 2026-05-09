'use client'

import { useChatStore } from '@/lib/store/chatStore'
import { useAuthStore } from '@/lib/store/authStore'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'

export default function Sidebar() {
  const { sessions, currentSession, selectSession, createSession, deleteSession, isLoadingSessions } =
    useChatStore()
  const { user, clearAuth } = useAuthStore()

  const handleNewChat = async () => {
    await createSession('New Chat')
  }

  const handleLogout = () => {
    clearAuth()
    window.location.href = '/login'
  }

  return (
    <div className="w-64 bg-card border-r border-border flex flex-col">
      <div className="p-4">
        <Button onClick={handleNewChat} className="w-full" size="sm">
          + New Chat
        </Button>
      </div>

      <Separator />

      <ScrollArea className="flex-1 px-2 py-2">
        {isLoadingSessions ? (
          <div className="text-center text-muted-foreground py-4">Loading...</div>
        ) : sessions.length === 0 ? (
          <div className="text-center text-muted-foreground py-4">No conversations yet</div>
        ) : (
          sessions.map((session) => (
            <div
              key={session.id}
              className={`p-3 rounded-lg cursor-pointer mb-1 group relative transition-colors ${
                currentSession?.id === session.id
                  ? 'bg-accent'
                  : 'hover:bg-accent'
              }`}
              onClick={() => selectSession(session.id)}
            >
              <div className="text-sm text-foreground truncate pr-8">
                {session.title || 'Untitled'}
              </div>
              <div className="text-xs text-muted-foreground mt-1">
                {new Date(session.updated_at).toLocaleDateString()}
              </div>
              <button
                className="absolute right-2 top-3 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity"
                onClick={(e) => {
                  e.stopPropagation()
                  if (confirm('Delete this conversation?')) {
                    deleteSession(session.id)
                  }
                }}
              >
                ×
              </button>
            </div>
          ))
        )}
      </ScrollArea>

      <Separator />

      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="text-sm text-foreground">{user?.username}</div>
          <div className="text-sm text-muted-foreground">{user?.credits} credits</div>
        </div>
        <div className="flex gap-2">
          <Button variant="ghost" size="sm" className="flex-1" asChild>
            <a href="/settings">Settings</a>
          </Button>
          <Button onClick={handleLogout} variant="ghost" size="sm" className="flex-1">
            Logout
          </Button>
        </div>
      </div>
    </div>
  )
}
