'use client'

import { memo, useState } from 'react'
import type { ImageRecord } from '@/types/image'
import { useChatStore } from '@/lib/store/chatStore'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

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
  if (!isLegacyPpt && !isCurrentPpt) return null

  const slideNumber = prompt.match(/Slide (\d+) title:/)?.[1]
  const title =
    prompt.match(/Slide \d+ title: ([^\n]+)/)?.[1]?.trim() ||
    prompt.match(/标题是：([^\n]+)/)?.[1]?.trim()
  const objective =
    prompt.match(/Slide objective: ([^\n]+)/)?.[1]?.trim() ||
    prompt.match(/核心内容是：([^\n]+)/)?.[1]?.trim()

  return { slideNumber, title, objective }
}

function getAssetBaseUrl() {
  const configuredApiBase = process.env.NEXT_PUBLIC_API_BASE_URL
  if (configuredApiBase) return configuredApiBase.replace('/api/v1', '')
  if (typeof window !== 'undefined') return '/gptimg-api'
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
      await useImageAsReference({ imageUrl, label: `Editing from: ${message.prompt.slice(0, 48)}` })
    } catch (error: any) {
      alert(error?.message || 'Failed to load this image for editing')
    } finally {
      setIsPreparingEdit(false)
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex gap-3">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-medium text-primary-foreground">U</div>
        <div className="min-w-0 flex-1 pt-1">
          {pptMeta ? (
            <div className="rounded-xl border border-sky-500/20 bg-sky-500/5 p-4">
              <Badge variant="outline" className="mb-2 border-sky-500/30 text-sky-300">
                {pptMeta.slideNumber ? `Slide ${pptMeta.slideNumber}` : 'PPT Slide'}
              </Badge>
              <div className="text-base font-semibold text-foreground">{pptMeta.title || 'Generated slide'}</div>
              {pptMeta.objective && <p className="mt-1.5 text-sm text-muted-foreground">{pptMeta.objective}</p>}
            </div>
          ) : (
            <p className="text-sm text-foreground leading-relaxed">{message.prompt}</p>
          )}
        </div>
      </div>

      <div className="flex gap-3">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-emerald-600 text-xs font-medium text-white">AI</div>
        <div className="min-w-0 flex-1 pt-1">
          {message.status === 'success' ? (
            <div>
              <div className="group relative overflow-hidden rounded-xl border border-border">
                <img
                  src={imageUrl}
                  alt={message.prompt}
                  loading="lazy"
                  decoding="async"
                  className="w-full max-w-lg cursor-pointer transition-transform duration-200 hover:scale-[1.02]"
                  onClick={() => setShowFullImage(true)}
                />
                <div className="absolute inset-x-0 bottom-0 flex justify-end gap-2 bg-gradient-to-t from-black/60 to-transparent p-3 opacity-0 transition-opacity group-hover:opacity-100">
                  <Button
                    size="sm"
                    variant="secondary"
                    className="h-7 text-xs"
                    onClick={(e) => { e.stopPropagation(); void handleUseAsReference() }}
                    disabled={isPreparingEdit}
                  >
                    {isPreparingEdit ? 'Loading...' : 'Edit'}
                  </Button>
                  <Button size="sm" variant="secondary" className="h-7 text-xs" asChild>
                    <a href={imageUrl} download onClick={(e) => e.stopPropagation()}>Download</a>
                  </Button>
                </div>
              </div>
              <div className="mt-2 flex flex-wrap gap-2">
                <Badge variant="secondary" className="text-[10px]">{message.size}</Badge>
                <Badge variant="secondary" className="text-[10px]">{message.quality}</Badge>
                {message.generation_time > 0 && <Badge variant="secondary" className="text-[10px]">{(message.generation_time / 1000).toFixed(1)}s</Badge>}
                <Badge variant="secondary" className="text-[10px]">{message.credits_used} cr</Badge>
              </div>
            </div>
          ) : message.status === 'failed' ? (
            <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
              <p className="text-sm font-medium text-destructive">Failed to generate image</p>
              {message.error_message && <p className="mt-1 text-xs text-destructive/80">{message.error_message}</p>}
            </div>
          ) : (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Processing...
            </div>
          )}
        </div>
      </div>

      {showFullImage && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/90 p-4 backdrop-blur-sm" onClick={() => setShowFullImage(false)}>
          <div className="relative max-h-[90vh] w-full max-w-5xl">
            <img src={imageUrl} alt={message.prompt} className="mx-auto max-h-[90vh] max-w-full rounded-lg object-contain" />
            <Button
              variant="secondary"
              size="icon"
              className="absolute right-2 top-2"
              onClick={() => setShowFullImage(false)}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

export default memo(MessageItem)
