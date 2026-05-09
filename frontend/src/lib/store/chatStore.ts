import { create } from 'zustand'
import type { Session } from '@/types/chat'
import type { ImageRecord } from '@/types/image'
import { sessionApi } from '@/lib/api/session'
import { imageApi } from '@/lib/api/image'

interface ChatStore {
  sessions: Session[]
  currentSession: Session | null
  messages: ImageRecord[]
  isGenerating: boolean
  isLoadingSessions: boolean
  isLoadingMessages: boolean

  loadSessions: () => Promise<void>
  createSession: (title?: string) => Promise<Session>
  selectSession: (sessionId: string) => Promise<void>
  deleteSession: (sessionId: string) => Promise<void>
  generateImage: (prompt: string, options?: any) => Promise<ImageRecord>
  setGenerating: (isGenerating: boolean) => void
}

export const useChatStore = create<ChatStore>((set, get) => ({
  sessions: [],
  currentSession: null,
  messages: [],
  isGenerating: false,
  isLoadingSessions: false,
  isLoadingMessages: false,

  loadSessions: async () => {
    set({ isLoadingSessions: true })
    try {
      const res = await sessionApi.getList()
      set({ sessions: res.data || [] })
    } catch (error) {
      console.error('Failed to load sessions:', error)
    } finally {
      set({ isLoadingSessions: false })
    }
  },

  createSession: async (title?: string) => {
    const res = await sessionApi.create({ title })
    const session = res.data
    set((state) => ({
      sessions: [session, ...state.sessions],
      currentSession: session,
      messages: [],
    }))
    return session
  },

  selectSession: async (sessionId: string) => {
    set({ isLoadingMessages: true })
    try {
      const sessionRes = await sessionApi.getById(sessionId)
      const messagesRes = await sessionApi.getMessages(sessionId)
      set({ currentSession: sessionRes.data, messages: messagesRes.data || [] })
    } catch (error) {
      console.error('Failed to load session:', error)
    } finally {
      set({ isLoadingMessages: false })
    }
  },

  deleteSession: async (sessionId: string) => {
    await sessionApi.delete(sessionId)
    set((state) => ({
      sessions: state.sessions.filter((s) => s.id !== sessionId),
      currentSession: state.currentSession?.id === sessionId ? null : state.currentSession,
      messages: state.currentSession?.id === sessionId ? [] : state.messages,
    }))
  },

  generateImage: async (prompt: string, options?: any) => {
    const { currentSession } = get()

    let sessionId = currentSession?.id
    if (!sessionId) {
      const newSession = await get().createSession()
      sessionId = newSession.id
    }

    set({ isGenerating: true })
    try {
      const res = await imageApi.generate({
        prompt,
        session_id: sessionId,
        ...options,
      })
      const record = res.data

      set((state) => ({
        messages: [...state.messages, record],
      }))

      return record
    } finally {
      set({ isGenerating: false })
    }
  },

  setGenerating: (isGenerating: boolean) => {
    set({ isGenerating })
  },
}))
