'use client'

import Link from 'next/link'
import { useChatStore } from '@/lib/store/chatStore'
import { useAuthStore } from '@/lib/store/authStore'
import { withBasePath } from '@/lib/runtime'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'

interface SidebarProps {
  isOpen: boolean
  onClose: () => void
}

export default function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { sessions, currentSession, selectSession, createSession, deleteSession, isLoadingSessions } = useChatStore()
  const { user, clearAuth } = useAuthStore()
  const handleLogout = () => { clearAuth(); window.location.assign(withBasePath('/login')) }
  const handleSelectSession = (sessionId: string | number) => {
    selectSession(String(sessionId))
    onClose()
  }
  const handleNewChatClick = async () => {
    await createSession('New Chat')
    onClose()
  }

  return (
    <>
      <div
        className={`fixed inset-0 z-30 bg-black/60 backdrop-blur-sm transition-opacity duration-200 md:hidden ${
          isOpen ? 'opacity-100' : 'pointer-events-none opacity-0'
        }`}
        onClick={onClose}
      />

      <aside
        className={`fixed inset-y-0 left-0 z-40 flex w-[280px] flex-col border-r border-border bg-card transition-transform duration-200 ease-out md:static md:w-56 md:translate-x-0 xl:w-64 ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="flex items-center justify-between p-4 md:hidden">
          <div>
            <div className="text-sm font-semibold text-foreground">Conversations</div>
            <div className="text-xs text-muted-foreground">{user?.credits ?? 0} credits left</div>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </Button>
        </div>

        <div className="p-3">
          <Button onClick={handleNewChatClick} className="w-full gap-2" size="sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            New Chat
          </Button>
        </div>

        <Separator />

        <ScrollArea className="flex-1">
          <div className="p-2">
            {isLoadingSessions ? (
              <div className="py-8 text-center text-sm text-muted-foreground">Loading...</div>
            ) : sessions.length === 0 ? (
              <div className="py-8 text-center text-sm text-muted-foreground">No conversations yet</div>
            ) : (
              sessions.map((session) => (
                <div
                  key={session.id}
                  className={`group relative mb-1 cursor-pointer rounded-lg px-3 py-2.5 transition-colors ${
                    currentSession?.id === session.id
                      ? 'bg-accent text-accent-foreground'
                      : 'text-foreground hover:bg-accent/50'
                  }`}
                  onClick={() => handleSelectSession(session.id)}
                >
                  <div className="truncate pr-6 text-sm font-medium">{session.title || 'Untitled'}</div>
                  <div className="mt-0.5 text-xs text-muted-foreground">
                    {new Date(session.updated_at).toLocaleDateString()}
                  </div>
                  <button
                    className="absolute right-2 top-2.5 rounded p-0.5 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
                    onClick={(e) => {
                      e.stopPropagation()
                      if (confirm('Delete this conversation?')) deleteSession(session.id)
                    }}
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <line x1="18" y1="6" x2="6" y2="18" />
                      <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                  </button>
                </div>
              ))
            )}
          </div>
        </ScrollArea>

        <Separator />

        <div className="p-3">
          <div className="mb-2 flex items-center justify-between px-1">
            <span className="truncate text-sm text-foreground">{user?.username}</span>
            <span className="text-xs text-muted-foreground">{user?.credits} cr</span>
          </div>
          <div className="flex gap-1.5">
            <Button variant="ghost" size="sm" className="flex-1" asChild>
              <Link href="/settings" prefetch={false} onClick={onClose}>Settings</Link>
            </Button>
            <Button onClick={handleLogout} variant="ghost" size="sm" className="flex-1 text-muted-foreground">
              Logout
            </Button>
          </div>
        </div>
      </aside>
    </>
  )
}
