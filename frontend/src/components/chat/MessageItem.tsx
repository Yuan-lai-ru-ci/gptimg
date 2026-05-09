'use client'

import { memo, useState } from 'react'
import type { ImageRecord } from '@/types/image'
import { useChatStore } from '@/lib/store/chatStore'

interface MessageItemProps {
  message: ImageRecord
}

interface ParsedPptMeta {
  slideNumber?: string
  title?: string
  objective?: string
}

function parsePptMeta(prompt: string): ParsedPptMeta | null {
  const isLegacyPpt = prompt.includes('Create one polished 16:9 PPT slide visual')
  const isCurrentPpt = prompt.includes('请生成一张中文科创答辩PPT页面')
  if (!isLegacyPpt && !isCurrentPpt) {
    return null
  }

  const slideNumber = prompt.match(/Slide (\d+) title:/)?.[1]
  const title =
    prompt.match(/Slide \d+ title: ([^\n]+)/)?.[1]?.trim() ||
    prompt.match(/标题是：([^\n]+)/)?.[1]?.trim()
  const objective =
    prompt.match(/Slide objective: ([^\n]+)/)?.[1]?.trim() ||
    prompt.match(/核心内容是：([^\n]+)/)?.[1]?.trim()

  return {
    slideNumber,
    title,
    objective,
  }
}

function getAssetBaseUrl() {
  const configuredApiBase = process.env.NEXT_PUBLIC_API_BASE_URL
  if (configuredApiBase) {
    return configuredApiBase.replace('/api/v1', '')
  }

  if (typeof window !== 'undefined') {
    return '/gptimg-api'
  }

  return 'http://127.0.0.1:8080'
}

function MessageItem({ message }: MessageItemProps) {
  const [showFullImage, setShowFullImage] = useState(false)
  const [isPreparingEdit, setIsPreparingEdit] = useState(false)
  const { useImageAsReference } = useChatStore()
  const pptMeta = parsePptMeta(message.prompt)
  const imageUrl = message.local_path
    ? `${getAssetBaseUrl()}/storage/${message.local_path}`
    : message.image_url

  const handleUseAsReference = async () => {
    try {
      setIsPreparingEdit(true)
      await useImageAsReference({
        imageUrl,
        label: `Editing from: ${message.prompt.slice(0, 48)}`,
      })
    } catch (error: any) {
      alert(error?.message || 'Failed to load this image for editing')
    } finally {
      setIsPreparingEdit(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex gap-3">
        <div className="w-8 h-8 rounded-full bg-blue-600 flex items-center justify-center text-foreground text-sm flex-shrink-0">U</div>
        <div className="min-w-0 flex-1">
          {pptMeta ? (
            <div className="rounded-2xl border border-sky-500/20 bg-sky-500/5 p-4">
              <div className="text-xs uppercase tracking-[0.18em] text-sky-300">
                {pptMeta.slideNumber ? `Slide ${pptMeta.slideNumber}` : 'PPT Slide'}
              </div>
              <div className="mt-2 text-base font-semibold text-foreground sm:text-lg">
                {pptMeta.title || 'Generated slide'}
              </div>
              {pptMeta.objective ? (
                <p className="mt-2 break-words text-sm text-secondary-foreground sm:text-base">
                  {pptMeta.objective}
                </p>
              ) : null}
            </div>
          ) : (
            <p className="break-words text-sm text-foreground sm:text-base">{message.prompt}</p>
          )}
        </div>
      </div>
      <div className="flex gap-3">
        <div className="w-8 h-8 rounded-full bg-emerald-600 flex items-center justify-center text-foreground text-sm flex-shrink-0">AI</div>
        <div className="min-w-0 flex-1">
          {message.status === 'success' ? (
            <div>
              <div className="relative group">
                <img
                  src={imageUrl}
                  alt={message.prompt}
                  loading="lazy"
                  decoding="async"
                  className="w-full max-w-md cursor-pointer rounded-xl transition-opacity hover:opacity-90"
                  onClick={() => setShowFullImage(true)}
                />
                <div className="absolute bottom-2 right-2 flex gap-2 opacity-0 transition-opacity group-hover:opacity-100">
                  <button
                    type="button"
                    className="rounded-lg bg-black/70 px-3 py-1 text-sm text-foreground hover:bg-black/90"
                    onClick={(e) => {
                      e.stopPropagation()
                      void handleUseAsReference()
                    }}
                    disabled={isPreparingEdit}
                  >
                    {isPreparingEdit ? 'Loading...' : 'Edit This'}
                  </button>
                  <a href={imageUrl} download className="bg-black/70 text-foreground px-3 py-1 rounded-lg text-sm hover:bg-black/90" onClick={(e) => e.stopPropagation()}>Download</a>
                </div>
              </div>
              <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                <span>{message.size}</span><span>{message.quality}</span>
                {message.generation_time > 0 && <span>{(message.generation_time / 1000).toFixed(1)}s</span>}
                <span>{message.credits_used} credit{message.credits_used > 1 ? 's' : ''}</span>
              </div>
            </div>
          ) : message.status === 'failed' ? (
            <div className="bg-red-900/20 border border-red-800 rounded-lg p-4">
              <p className="text-red-400">Failed to generate image</p>
              {message.error_message && <p className="text-red-500 text-sm mt-1">{message.error_message}</p>}
            </div>
          ) : <div className="text-muted-foreground">Processing...</div>}
        </div>
      </div>
      {showFullImage && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4" onClick={() => setShowFullImage(false)}>
          <div className="relative max-h-[90vh] w-full max-w-4xl">
            <img src={imageUrl} alt={message.prompt} className="max-w-full max-h-[90vh] object-contain rounded-lg" />
            <button className="absolute right-2 top-2 flex h-8 w-8 items-center justify-center rounded-full bg-black/70 text-foreground hover:bg-black/90" onClick={() => setShowFullImage(false)}>×</button>
          </div>
        </div>
      )}
    </div>
  )
}

export default memo(MessageItem)
