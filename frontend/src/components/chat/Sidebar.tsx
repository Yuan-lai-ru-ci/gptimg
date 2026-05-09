'use client'

import Link from 'next/link'
import { useChatStore } from '@/lib/store/chatStore'
import { useAuthStore } from '@/lib/store/authStore'
import { withBasePath } from '@/lib/runtime'
import { Button } from '@/components/ui/button'

interface SidebarProps {
  isOpen: boolean
  onClose: () => void
}

export default function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { sessions, currentSession, selectSession, createSession, deleteSession, isLoadingSessions } = useChatStore()
  const { user, clearAuth } = useAuthStore()
  const handleNewChat = async () => { await createSession('New Chat') }
  const handleLogout = () => { clearAuth(); window.location.assign(withBasePath('/login')) }
  const handleSelectSession = (sessionId: string | number) => {
    selectSession(String(sessionId))
    onClose()
  }
  const handleNewChatClick = async () => {
    await handleNewChat()
    onClose()
  }

  return (
    <>
      <div
        className={`fixed inset-0 z-30 bg-black/60 transition-opacity md:hidden ${
          isOpen ? 'opacity-100' : 'pointer-events-none opacity-0'
        }`}
        onClick={onClose}
      />

      <aside
        className={`fixed inset-y-0 left-0 z-40 flex w-[85vw] max-w-72 flex-col border-r border-border bg-card transition-transform md:static md:w-56 md:max-w-none md:translate-x-0 xl:w-64 ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="flex items-center justify-between border-b border-border p-4 md:hidden">
          <div>
            <div className="text-sm font-semibold text-foreground">Conversations</div>
            <div className="text-xs text-muted-foreground">{user?.credits ?? 0} credits left</div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            aria-label="Close conversations"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <div className="border-b border-border p-3 xl:p-4">
          <Button onClick={handleNewChatClick} className="w-full" size="sm">
            + New Chat
          </Button>
        </div>

        <div className="flex-1 overflow-y-auto p-2">
          {isLoadingSessions ? (
            <div className="py-4 text-center text-muted-foreground">Loading...</div>
          ) : sessions.length === 0 ? (
            <div className="py-4 text-center text-muted-foreground">No conversations yet</div>
          ) : (
            sessions.map((session) => (
              <div
                key={session.id}
                className={`group relative mb-1.5 rounded-lg p-2.5 xl:p-3 ${
                  currentSession?.id === session.id ? 'bg-accent' : 'hover:bg-accent'
                } cursor-pointer`}
                onClick={() => handleSelectSession(session.id)}
              >
                <div className="truncate pr-8 text-sm text-foreground">{session.title || 'Untitled'}</div>
                <div className="mt-1 text-xs text-muted-foreground">
                  {new Date(session.updated_at).toLocaleDateString()}
                </div>
                <button
                  className="absolute right-2 top-3 text-muted-foreground transition-opacity hover:text-red-400 md:opacity-0 md:group-hover:opacity-100"
                  onClick={(e) => {
                    e.stopPropagation()
                    if (confirm('Delete this conversation?')) deleteSession(session.id)
                  }}
                >
                  ×
                </button>
              </div>
            ))
          )}
        </div>

        <div className="border-t border-border p-3 xl:p-4">
          <div className="mb-3 flex items-center justify-between gap-3">
            <div className="min-w-0 text-sm text-secondary-foreground">{user?.username}</div>
            <div className="shrink-0 text-sm text-muted-foreground">{user?.credits} credits</div>
          </div>
          <div className="flex gap-1.5 xl:gap-2">
            <Link
              href="/settings"
              prefetch={false}
              onClick={onClose}
              className="flex-1 rounded-lg py-1.5 text-center text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            >
              Settings
            </Link>
            <Button onClick={handleLogout} variant="ghost" size="sm" className="flex-1">
              Logout
            </Button>
          </div>
        </div>
      </aside>
    </>
  )
}
