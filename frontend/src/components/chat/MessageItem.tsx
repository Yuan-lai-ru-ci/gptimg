'use client'

import { useState } from 'react'
import type { ImageRecord } from '@/types/image'
import { Button } from '@/components/ui/button'

interface MessageItemProps {
  message: ImageRecord
}

export default function MessageItem({ message }: MessageItemProps) {
  const [showFullImage, setShowFullImage] = useState(false)
  const apiBase = process.env.NEXT_PUBLIC_API_BASE_URL?.replace('/api/v1', '') || 'http://localhost:8080'

  const imageUrl = message.local_path
    ? `${apiBase}/storage/${message.local_path}`
    : message.image_url

  return (
    <div className="space-y-4">
      <div className="flex gap-3">
        <div className="w-8 h-8 rounded-full bg-blue-600 flex items-center justify-center text-white text-sm flex-shrink-0">
          U
        </div>
        <div className="flex-1">
          <p className="text-foreground">{message.prompt}</p>
        </div>
      </div>

      <div className="flex gap-3">
        <div className="w-8 h-8 rounded-full bg-emerald-600 flex items-center justify-center text-white text-sm flex-shrink-0">
          AI
        </div>
        <div className="flex-1">
          {message.status === 'success' ? (
            <div>
              {message.revised_prompt && message.revised_prompt !== message.prompt && (
                <p className="text-muted-foreground text-sm mb-3 italic">
                  {message.revised_prompt}
                </p>
              )}

              <div className="relative group">
                <img
                  src={imageUrl}
                  alt={message.prompt}
                  className="rounded-xl max-w-md w-full cursor-pointer hover:opacity-90 transition-opacity"
                  onClick={() => setShowFullImage(true)}
                />
                <div className="absolute bottom-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-2">
                  <Button size="sm" variant="secondary" asChild>
                    <a
                      href={imageUrl}
                      download
                      onClick={(e) => e.stopPropagation()}
                    >
                      Download
                    </a>
                  </Button>
                </div>
              </div>

              <div className="flex gap-4 mt-2 text-xs text-muted-foreground">
                <span>{message.size}</span>
                <span>{message.quality}</span>
                {message.generation_time > 0 && (
                  <span>{(message.generation_time / 1000).toFixed(1)}s</span>
                )}
                <span>{message.credits_used} credit{message.credits_used > 1 ? 's' : ''}</span>
              </div>
            </div>
          ) : message.status === 'failed' ? (
            <div className="bg-destructive/10 border border-destructive/50 rounded-lg p-4">
              <p className="text-destructive">Failed to generate image</p>
              {message.error_message && (
                <p className="text-destructive/80 text-sm mt-1">{message.error_message}</p>
              )}
            </div>
          ) : (
            <div className="text-muted-foreground">Processing...</div>
          )}
        </div>
      </div>

      {showFullImage && (
        <div
          className="fixed inset-0 bg-black/80 z-50 flex items-center justify-center p-4"
          onClick={() => setShowFullImage(false)}
        >
          <div className="relative max-w-4xl max-h-[90vh]">
            <img
              src={imageUrl}
              alt={message.prompt}
              className="max-w-full max-h-[90vh] object-contain rounded-lg"
            />
            <Button
              variant="secondary"
              size="icon"
              className="absolute top-2 right-2"
              onClick={() => setShowFullImage(false)}
            >
              ×
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
