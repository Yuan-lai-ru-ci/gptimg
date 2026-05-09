'use client'

import { useState, useRef, KeyboardEvent } from 'react'
import { useChatStore } from '@/lib/store/chatStore'
import { useAuthStore } from '@/lib/store/authStore'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export default function InputBox() {
  const [prompt, setPrompt] = useState('')
  const [showOptions, setShowOptions] = useState(false)
  const [size, setSize] = useState('1024x1024')
  const [quality, setQuality] = useState('standard')
  const [style, setStyle] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const { generateImage, isGenerating } = useChatStore()
  const { user, refreshUser } = useAuthStore()

  const handleSubmit = async () => {
    if (!prompt.trim() || isGenerating) return

    const currentPrompt = prompt
    setPrompt('')

    try {
      await generateImage(currentPrompt, {
        size,
        quality,
        style: style || undefined,
      })
      await refreshUser()
    } catch (error: any) {
      alert(error?.message || 'Failed to generate image')
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  const handleTextareaInput = () => {
    const textarea = textareaRef.current
    if (textarea) {
      textarea.style.height = 'auto'
      textarea.style.height = Math.min(textarea.scrollHeight, 200) + 'px'
    }
  }

  return (
    <div className="max-w-3xl mx-auto w-full p-4">
      {showOptions && (
        <div className="mb-3 p-3 bg-card rounded-lg border border-border">
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-1.5">
              <label className="text-xs text-muted-foreground">Size</label>
              <Select value={size} onValueChange={setSize}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1024x1024">1024×1024</SelectItem>
                  <SelectItem value="1792x1024">1792×1024</SelectItem>
                  <SelectItem value="1024x1792">1024×1792</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs text-muted-foreground">Quality</label>
              <Select value={quality} onValueChange={setQuality}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="standard">Standard (1 credit)</SelectItem>
                  <SelectItem value="hd">HD (2 credits)</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs text-muted-foreground">Style</label>
              <Select value={style || 'default'} onValueChange={(v) => setStyle(v === 'default' ? '' : v)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="default">Default</SelectItem>
                  <SelectItem value="vivid">Vivid</SelectItem>
                  <SelectItem value="natural">Natural</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
      )}

      <div className="relative bg-card border border-border rounded-2xl">
        <textarea
          ref={textareaRef}
          value={prompt}
          onChange={(e) => {
            setPrompt(e.target.value)
            handleTextareaInput()
          }}
          onKeyDown={handleKeyDown}
          placeholder="Describe the image you want to generate..."
          className="w-full bg-transparent text-foreground placeholder-muted-foreground px-4 py-3 pr-24 resize-none focus:outline-none rounded-2xl"
          rows={1}
          disabled={isGenerating}
        />

        <div className="absolute right-2 bottom-2 flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setShowOptions(!showOptions)}
            title="Options"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
            </svg>
          </Button>

          <Button
            size="icon"
            onClick={handleSubmit}
            disabled={!prompt.trim() || isGenerating}
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="22" y1="2" x2="11" y2="13" />
              <polygon points="22 2 15 22 11 13 2 9 22 2" />
            </svg>
          </Button>
        </div>
      </div>

      <div className="flex justify-between items-center mt-2 px-2">
        <span className="text-xs text-muted-foreground">
          Credits: {user?.credits ?? 0}
        </span>
        <span className="text-xs text-muted-foreground">
          Shift+Enter for new line
        </span>
      </div>
    </div>
  )
}
