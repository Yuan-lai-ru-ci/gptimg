import { create } from 'zustand'
import type { Session } from '@/types/chat'
import type { GenerateImageRequest, ImageRecord } from '@/types/image'
import type { PPTPlan, PPTWorkflowMode } from '@/types/ppt'
import { sessionApi } from '@/lib/api/session'
import { imageApi } from '@/lib/api/image'
import { pptApi } from '@/lib/api/ppt'

interface PptProgress {
  stage: 'idle' | 'planning' | 'planned' | 'generating'
  message: string
  totalSlides: number
}

const sleep = (ms: number) => new Promise((resolve) => window.setTimeout(resolve, ms))

const splitUserTextIntoSlides = (content: string): string[] => {
  const trimmed = content.trim()
  if (!trimmed) return []

  const markerPattern = /(?:^|\n)\s*(?=(?:第\s*[一二三四五六七八九十\d]+\s*[页頁]|P(?:age)?\s*\d+|Slide\s*\d+)[:：\.\s-]*)/i
  const markerParts = trimmed
    .split(markerPattern)
    .map((part) => part.trim())
    .filter(Boolean)

  if (markerParts.length > 1) {
    return markerParts
  }

  const blockParts = trimmed
    .split(/\n\s*\n+/)
    .map((part) => part.trim())
    .filter(Boolean)

  return blockParts.length > 1 ? blockParts : [trimmed]
}

const cleanupSlideTitle = (value: string) =>
  value
    .replace(/^(?:第\s*[一二三四五六七八九十\d]+\s*[页頁]|P(?:age)?\s*\d+|Slide\s*\d+)[:：\.\s-]*/i, '')
    .trim() || value.trim()

interface ChatStore {
  sessions: Session[]
  currentSession: Session | null
  messages: ImageRecord[]
  composerMode: 'image' | 'ppt'
  pptWorkflowMode: PPTWorkflowMode
  pptPlan: PPTPlan | null
  pptProgress: PptProgress
  isGenerating: boolean
  isPlanningPpt: boolean
  isLoadingSessions: boolean
  isLoadingMessages: boolean
  referenceImage: File | null
  referenceImages: File[]
  referenceImageLabel: string
  pptReferenceImages: File[]

  loadSessions: () => Promise<void>
  createSession: (title?: string) => Promise<Session>
  selectSession: (sessionId: string) => Promise<void>
  deleteSession: (sessionId: string) => Promise<void>
  generateImage: (prompt: string, options?: Omit<GenerateImageRequest, 'prompt' | 'session_id'>) => Promise<ImageRecord>
  planPpt: (requirement: string, options?: { slideCount?: number; mode?: PPTWorkflowMode }) => Promise<PPTPlan>
  planPptFromStyleText: (content: string) => Promise<PPTPlan>
  planPptFromImageSet: (styleDescription: string, imageCount: number) => Promise<PPTPlan>
  planPptFromDocument: (document: File, requirement?: string, options?: { slideCount?: number }) => Promise<PPTPlan>
  updatePptPlan: (plan: PPTPlan) => void
  clearPptPlan: () => void
  generatePptDeck: (options?: { size?: string; quality?: string; style?: string }) => Promise<ImageRecord[]>
  generatePptDirect: (prompt: string) => Promise<ImageRecord[]>
  setComposerMode: (mode: 'image' | 'ppt') => void
  setPptWorkflowMode: (mode: PPTWorkflowMode) => void
  setReferenceImage: (file: File | null, label?: string) => void
  setReferenceImages: (files: File[], label?: string) => void
  clearReferenceImage: () => void
  useImageAsReference: (payload: { imageUrl: string; label?: string }) => Promise<void>
  setGenerating: (isGenerating: boolean) => void
}

export const useChatStore = create<ChatStore>((set, get) => ({
  sessions: [],
  currentSession: null,
  messages: [],
  composerMode: 'image',
  pptWorkflowMode: 'style_text',
  pptPlan: null,
  pptProgress: {
    stage: 'idle',
    message: '',
    totalSlides: 0,
  },
  isGenerating: false,
  isPlanningPpt: false,
  isLoadingSessions: false,
  isLoadingMessages: false,
  referenceImage: null,
  referenceImages: [],
  referenceImageLabel: '',
  pptReferenceImages: [],

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

  generateImage: async (prompt: string, options?: Omit<GenerateImageRequest, 'prompt' | 'session_id'>) => {
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
        referenceImage: null,
        referenceImages: [],
        referenceImageLabel: '',
      }))

      return record
    } finally {
      set({ isGenerating: false })
    }
  },

  planPpt: async (requirement: string, options?: { slideCount?: number; mode?: PPTWorkflowMode }) => {
    const { currentSession, referenceImages } = get()
    const styleReference = referenceImages[0]
    if (!styleReference) {
      throw new Error('模式①需要先上传1张风格参考图')
    }
    let sessionId = currentSession?.id
    if (!sessionId) {
      const newSession = await get().createSession('PPT Deck')
      sessionId = newSession.id
    }

    set({
      isPlanningPpt: true,
      pptPlan: null,
      pptProgress: {
        stage: 'planning',
        message: 'Planning a science and innovation deck structure...',
        totalSlides: options?.slideCount || 0,
      },
    })
    try {
      const res = await pptApi.plan({
        requirement,
        slide_count: options?.slideCount,
        generation_mode: options?.mode || get().pptWorkflowMode,
      })
      const plan = { ...res.data, generation_mode: options?.mode || get().pptWorkflowMode }

      if (sessionId && plan.deck_title) {
        const updated = await sessionApi.update(sessionId, { title: plan.deck_title })
        set((state) => ({
          currentSession: state.currentSession?.id === sessionId ? updated.data : state.currentSession,
          sessions: state.sessions.map((item) => (item.id === sessionId ? updated.data : item)),
          pptPlan: plan,
          pptReferenceImages: [],
          composerMode: 'ppt',
          pptProgress: {
            stage: 'planned',
            message: `Plan ready: ${plan.slides.length} slides.`,
            totalSlides: plan.slides.length,
          },
        }))
      } else {
        set({
          pptPlan: plan,
          pptReferenceImages: [],
          composerMode: 'ppt',
          pptProgress: {
            stage: 'planned',
            message: `Plan ready: ${plan.slides.length} slides.`,
            totalSlides: plan.slides.length,
          },
        })
      }

      return plan
    } finally {
      set({ isPlanningPpt: false })
    }
  },

  planPptFromStyleText: async (content: string) => {
    const { currentSession, referenceImages } = get()
    const styleReference = referenceImages[0]
    if (!styleReference) {
      throw new Error('模式①需要先上传1张风格参考图')
    }
    let sessionId = currentSession?.id
    if (!sessionId) {
      const newSession = await get().createSession('Style PPT')
      sessionId = newSession.id
    }

    const slides = splitUserTextIntoSlides(content).map((slideText, index) => {
      const lines = slideText
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter(Boolean)
      const title = cleanupSlideTitle(lines[0] || `第 ${index + 1} 页`)
      const body = lines.length > 1 ? lines.slice(1).join('\n') : slideText

      return {
        slide_number: index + 1,
        title,
        objective: body,
        layout_notes: '用户文字为当前页唯一内容来源；上传图片只用于风格参考。',
        page_description: `这一页是：内容页\n标题是：${title}\n用户原文内容：\n${body}`,
        image_prompt: `标题是：${title}\n用户原文内容：\n${body}`,
        speaker_notes: '',
      }
    })

    const plan: PPTPlan = {
      deck_title: slides[0]?.title || '风格参考PPT',
      deck_goal: '按用户逐页文字生成PPT，上传图片仅作为统一视觉样板。',
      visual_direction: '完全参考上传图片的整体视觉风格，包括配色、标题位置、版式节奏、字体感觉和装饰元素。',
      master_style_description: '所有页面均以上传的第一张图片作为统一视觉样板；只学习风格，不复制参考图内容。',
      consistency_rules: [
        '上传图片只作为视觉样板，不作为页面内容来源',
        '每页内容只来自用户为该页输入的文字',
        '不同页之间不要串页，不要混用其他页文字',
      ],
      generation_mode: 'style_text',
      slides,
    }

    if (sessionId) {
      const updated = await sessionApi.update(sessionId, { title: plan.deck_title })
      set((state) => ({
        currentSession: state.currentSession?.id === sessionId ? updated.data : state.currentSession,
        sessions: state.sessions.map((item) => (item.id === sessionId ? updated.data : item)),
        pptPlan: plan,
        pptReferenceImages: [styleReference],
        composerMode: 'ppt',
        pptWorkflowMode: 'style_text',
        pptProgress: {
          stage: 'planned',
          message: `Text split ready: ${plan.slides.length} slides.`,
          totalSlides: plan.slides.length,
        },
      }))
    }

    return plan
  },

  planPptFromImageSet: async (styleDescription: string, imageCount: number) => {
    const { currentSession, referenceImages } = get()
    const imageReferences = referenceImages.slice(0, imageCount)
    let sessionId = currentSession?.id
    if (!sessionId) {
      const newSession = await get().createSession('Image PPT')
      sessionId = newSession.id
    }

    const normalizedStyle = styleDescription.trim() || '统一美化为清晰、现代、适合汇报展示的中文PPT风格。'
    const slides = Array.from({ length: imageCount }, (_, index) => ({
      slide_number: index + 1,
      title: `第 ${index + 1} 页`,
      objective: `保留第 ${index + 1} 张上传图片的核心内容，并按统一风格美化成PPT页面。`,
      layout_notes: '上传图片是本页核心内容参考，文字描述只用于统一设计风格。',
      page_description: `这一页是：图片内容美化页\n标题是：第 ${index + 1} 页\n核心内容是：保留第 ${index + 1} 张上传图片的主要画面、结构和关键信息，用统一风格重新排版美化。`,
      image_prompt: `保留第 ${index + 1} 张上传图片的核心内容，按以下统一风格生成一张中文PPT页面：${normalizedStyle}`,
      speaker_notes: '',
    }))
    const plan: PPTPlan = {
      deck_title: '图片内容美化PPT',
      deck_goal: '把上传图片逐页整理、美化为统一风格的PPT图片。',
      visual_direction: normalizedStyle,
      master_style_description: normalizedStyle,
      consistency_rules: [
        '每页只使用对应页号的上传图片作为核心内容',
        '所有页面共享同一套配色、标题位置、装饰语言和信息密度',
        '不要把其他页图片或文字混入当前页',
      ],
      generation_mode: 'image_content_style',
      slides,
    }

    if (sessionId) {
      const updated = await sessionApi.update(sessionId, { title: plan.deck_title })
      set((state) => ({
        currentSession: state.currentSession?.id === sessionId ? updated.data : state.currentSession,
        sessions: state.sessions.map((item) => (item.id === sessionId ? updated.data : item)),
        pptPlan: plan,
        pptReferenceImages: imageReferences,
        composerMode: 'ppt',
        pptWorkflowMode: 'image_content_style',
        pptProgress: {
          stage: 'planned',
          message: `Image plan ready: ${plan.slides.length} slides.`,
          totalSlides: plan.slides.length,
        },
      }))
    }

    return plan
  },

  planPptFromDocument: async (document: File, requirement?: string, options?: { slideCount?: number }) => {
    const { currentSession } = get()
    let sessionId = currentSession?.id
    if (!sessionId) {
      const newSession = await get().createSession('Document PPT')
      sessionId = newSession.id
    }

    set({
      isPlanningPpt: true,
      pptPlan: null,
      pptReferenceImages: [],
      pptProgress: {
        stage: 'planning',
        message: `Reading ${document.name} and planning PPT...`,
        totalSlides: options?.slideCount || 0,
      },
    })
    try {
      const res = await pptApi.planDocument({
        document,
        requirement,
        slide_count: options?.slideCount,
      })
      const plan = res.data

      if (sessionId && plan.deck_title) {
        const updated = await sessionApi.update(sessionId, { title: plan.deck_title })
        set((state) => ({
          currentSession: state.currentSession?.id === sessionId ? updated.data : state.currentSession,
          sessions: state.sessions.map((item) => (item.id === sessionId ? updated.data : item)),
          pptPlan: plan,
          pptReferenceImages: [],
          composerMode: 'ppt',
          pptProgress: {
            stage: 'planned',
            message: `Document plan ready: ${plan.slides.length} slides.`,
            totalSlides: plan.slides.length,
          },
        }))
      } else {
        set({
          pptPlan: plan,
          pptReferenceImages: [],
          composerMode: 'ppt',
          pptProgress: {
            stage: 'planned',
            message: `Document plan ready: ${plan.slides.length} slides.`,
            totalSlides: plan.slides.length,
          },
        })
      }

      return plan
    } finally {
      set({ isPlanningPpt: false })
    }
  },

  updatePptPlan: (plan: PPTPlan) => set({ pptPlan: plan }),

  clearPptPlan: () =>
    set({
      pptPlan: null,
      pptProgress: {
        stage: 'idle',
        message: '',
        totalSlides: 0,
      },
    }),

  generatePptDeck: async (options?: { size?: string; quality?: string; style?: string }) => {
    const { currentSession, pptPlan } = get()
    if (!currentSession || !pptPlan) {
      throw new Error('PPT plan not ready')
    }

    const finalSize = options?.size || '1792x1024'
    const finalQuality = options?.quality || 'hd'
    const sessionId = currentSession.id
    const initialIds = new Set(get().messages.map((message) => message.id))
    const expectedSlides = pptPlan.slides.length
    const stateSnapshot = get()
    const referenceImagesSnapshot =
      stateSnapshot.pptReferenceImages.length > 0
        ? [...stateSnapshot.pptReferenceImages]
        : [...stateSnapshot.referenceImages]

    set({
      isGenerating: true,
      referenceImage: null,
      referenceImages: [],
      referenceImageLabel: '',
      pptProgress: {
        stage: 'generating',
        message: `Generating 0/${expectedSlides} slide visuals...`,
        totalSlides: expectedSlides,
      },
    })
    try {
      await pptApi.generate({
        session_id: sessionId,
        deck_title: pptPlan.deck_title,
        deck_goal: pptPlan.deck_goal,
        visual_direction: pptPlan.visual_direction,
        master_style_description: pptPlan.master_style_description,
        consistency_rules: pptPlan.consistency_rules,
        slides: pptPlan.slides,
        size: finalSize,
        quality: finalQuality,
        style: options?.style,
        generation_mode: pptPlan.generation_mode || get().pptWorkflowMode,
        reference_images: referenceImagesSnapshot,
      })

      let latestRecords: ImageRecord[] = []
      const startedAt = Date.now()
      const maxWaitMs = Math.max(45 * 60 * 1000, expectedSlides * 7 * 60 * 1000)

      while (Date.now() - startedAt < maxWaitMs) {
        const messagesRes = await sessionApi.getMessages(sessionId, 200, 0)
        const allMessages = messagesRes.data || []
        const generatedRecords = allMessages.filter((message) => !initialIds.has(message.id))
        latestRecords = generatedRecords

        set((state) => {
          if (state.currentSession?.id !== sessionId) {
            return state
          }

          return {
            messages: allMessages,
            pptProgress: {
              stage: 'generating',
              message: `Generating ${Math.min(generatedRecords.length, expectedSlides)}/${expectedSlides} slide visuals...`,
              totalSlides: expectedSlides,
            },
          }
        })

        if (generatedRecords.length >= expectedSlides) {
          break
        }

        await sleep(5000)
      }

      set((state) => ({
        pptProgress: {
          stage: 'planned',
          message: `Generated ${Math.min(latestRecords.length, expectedSlides)}/${expectedSlides} slide visuals.`,
          totalSlides: state.pptPlan?.slides.length || expectedSlides,
        },
      }))

      return latestRecords
    } finally {
      set({ isGenerating: false })
    }
  },

  generatePptDirect: async (prompt: string) => {
    const { currentSession, pptWorkflowMode, referenceImages } = get()
    if (referenceImages.length === 0) {
      throw new Error('PPT模式需要先上传参考图')
    }

    let sessionId = currentSession?.id
    if (!sessionId) {
      const newSession = await get().createSession('PPT Generation')
      sessionId = newSession.id
    }

    const promptBlocks = splitUserTextIntoSlides(prompt)
    const slideCount =
      pptWorkflowMode === 'image_content_style'
        ? referenceImages.length
        : Math.max(promptBlocks.length, 1)
    const slides = Array.from({ length: slideCount }, (_, index) => {
      const userPrompt =
        pptWorkflowMode === 'image_content_style'
          ? prompt
          : promptBlocks[index] || promptBlocks[promptBlocks.length - 1] || prompt
      const title = cleanupSlideTitle(userPrompt.split(/\r?\n/).find(Boolean) || `第 ${index + 1} 页`)

      return {
        slide_number: index + 1,
        title,
        objective: userPrompt,
        layout_notes: '',
        page_description: userPrompt,
        image_prompt: userPrompt,
        speaker_notes: '',
      }
    })

    const referenceImagesSnapshot =
      pptWorkflowMode === 'style_text'
        ? [referenceImages[0]]
        : [...referenceImages]
    const initialIds = new Set(get().messages.map((message) => message.id))

    set({
      isGenerating: true,
      pptPlan: null,
      pptReferenceImages: referenceImagesSnapshot,
      referenceImage: null,
      referenceImages: [],
      referenceImageLabel: '',
      pptProgress: {
        stage: 'generating',
        message: `Generating 0/${slides.length} PPT pages...`,
        totalSlides: slides.length,
      },
    })

    try {
      await pptApi.generate({
        session_id: sessionId,
        generation_mode: pptWorkflowMode,
        deck_title: slides[0]?.title || 'PPT Generation',
        deck_goal: '',
        visual_direction: '',
        master_style_description: '',
        consistency_rules: [],
        slides,
        size: '1792x1024',
        quality: 'hd',
        reference_images: referenceImagesSnapshot,
      })

      let latestRecords: ImageRecord[] = []
      const startedAt = Date.now()
      const maxWaitMs = Math.max(45 * 60 * 1000, slides.length * 7 * 60 * 1000)

      while (Date.now() - startedAt < maxWaitMs) {
        const messagesRes = await sessionApi.getMessages(sessionId, 200, 0)
        const allMessages = messagesRes.data || []
        const generatedRecords = allMessages.filter((message) => !initialIds.has(message.id))
        latestRecords = generatedRecords

        set((state) => {
          if (state.currentSession?.id !== sessionId) {
            return state
          }
          return {
            messages: allMessages,
            pptProgress: {
              stage: 'generating',
              message: `Generating ${Math.min(generatedRecords.length, slides.length)}/${slides.length} PPT pages...`,
              totalSlides: slides.length,
            },
          }
        })

        if (generatedRecords.length >= slides.length) {
          break
        }
        await sleep(5000)
      }

      set({
        pptProgress: {
          stage: 'idle',
          message: '',
          totalSlides: 0,
        },
        pptReferenceImages: [],
      })

      return latestRecords
    } finally {
      set({ isGenerating: false })
    }
  },

  setComposerMode: (mode: 'image' | 'ppt') => set({ composerMode: mode }),

  setPptWorkflowMode: (mode: PPTWorkflowMode) => set({ pptWorkflowMode: mode }),

  setReferenceImage: (file: File | null, label?: string) =>
    set({
      referenceImage: file,
      referenceImages: file ? [file] : [],
      referenceImageLabel: file ? label || file.name : '',
    }),

  setReferenceImages: (files: File[], label?: string) =>
    set({
      referenceImage: files[0] || null,
      referenceImages: files,
      referenceImageLabel: files.length > 0 ? label || `${files.length} reference image${files.length > 1 ? 's' : ''}` : '',
    }),

  clearReferenceImage: () =>
    set({
      referenceImage: null,
      referenceImages: [],
      referenceImageLabel: '',
    }),

  useImageAsReference: async ({ imageUrl, label }) => {
    const response = await fetch(imageUrl)
    if (!response.ok) {
      throw new Error('Failed to load the selected image')
    }

    const blob = await response.blob()
    const inferredType = blob.type || 'image/png'
    const extension = inferredType.split('/')[1] || 'png'
    const file = new File([blob], `chat-reference.${extension}`, { type: inferredType })

    set({
      referenceImage: file,
      referenceImages: [file],
      referenceImageLabel: label || 'Selected from conversation',
      composerMode: 'image',
    })
  },

  setGenerating: (isGenerating: boolean) => {
    set({ isGenerating })
  },
}))
