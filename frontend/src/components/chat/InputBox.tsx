'use client'

import { useEffect, useRef, useState, KeyboardEvent } from 'react'
import { useChatStore } from '@/lib/store/chatStore'
import { useAuthStore } from '@/lib/store/authStore'
import type { PPTWorkflowMode } from '@/types/ppt'

const pptModeOptions: Array<{
  id: PPTWorkflowMode
  title: string
  hint: string
}> = [
  {
    id: 'style_text',
    title: '① 风格图 + 文字',
    hint: '1张图定风格，文字决定每页内容',
  },
  {
    id: 'image_content_style',
    title: '② 多图内容 + 风格',
    hint: '第N张图对应第N页，文字只管统一风格',
  },
  {
    id: 'free_mix',
    title: '③ 自由美化',
    hint: '图片和文字一起交给AI整理补全',
  },
]

export default function InputBox() {
  const [prompt, setPrompt] = useState('')
  const [showOptions, setShowOptions] = useState(false)
  const [size, setSize] = useState('1024x1024')
  const [quality, setQuality] = useState('standard')
  const [style, setStyle] = useState('')
  const [referencePreviewUrls, setReferencePreviewUrls] = useState<string[]>([])
  const [isMobile, setIsMobile] = useState(false)
  const [isDraggingImages, setIsDraggingImages] = useState(false)
  const [documentFile, setDocumentFile] = useState<File | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const documentInputRef = useRef<HTMLInputElement>(null)

  const {
    generateImage,
    isGenerating,
    composerMode,
    setComposerMode,
    pptWorkflowMode,
    setPptWorkflowMode,
    generatePptDirect,
    isPlanningPpt,
    pptProgress,
    referenceImage,
    referenceImages,
    referenceImageLabel,
    setReferenceImages,
    clearReferenceImage,
  } = useChatStore()
  const { user, refreshUser } = useAuthStore()

  useEffect(() => {
    if (referenceImages.length === 0) {
      setReferencePreviewUrls([])
      return
    }

    const nextUrls = referenceImages.map((file) => URL.createObjectURL(file))
    setReferencePreviewUrls(nextUrls)

    return () => nextUrls.forEach((url) => URL.revokeObjectURL(url))
  }, [referenceImages])

  useEffect(() => {
    if (typeof window === 'undefined') return

    const media = window.matchMedia('(max-width: 767px)')
    const updateViewport = () => setIsMobile(media.matches)

    updateViewport()
    media.addEventListener('change', updateViewport)
    return () => media.removeEventListener('change', updateViewport)
  }, [])

  useEffect(() => {
    if (isMobile && composerMode === 'ppt') {
      setComposerMode('image')
    }
  }, [composerMode, isMobile, setComposerMode])

  useEffect(() => {
    if (composerMode === 'ppt') {
      setSize('1792x1024')
      setQuality('hd')
    }
  }, [composerMode])

  const handleSubmit = async () => {
    if ((!prompt.trim() && !documentFile) || isGenerating || isPlanningPpt) return

    const currentPrompt = prompt
    setPrompt('')

    try {
      if (composerMode === 'ppt') {
        if (referenceImages.length === 0) {
          alert('PPT模式需要先上传参考图')
          setPrompt(currentPrompt)
          return
        }
        await generatePptDirect(currentPrompt)
      } else {
        await generateImage(currentPrompt, {
          size,
          quality,
          style: style || undefined,
          reference_image: referenceImage,
          reference_images: referenceImages,
        })
        if (fileInputRef.current) {
          fileInputRef.current.value = ''
        }
        await refreshUser()
      }
    } catch (error: any) {
      alert(error?.message || (composerMode === 'ppt' ? 'Failed to plan PPT' : 'Failed to generate image'))
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

  const attachReferenceImages = (files: File[], append = false) => {
    if (files.length === 0) {
      if (!append) {
        setReferenceImages([])
      }
      return
    }

    const imageFiles = files.filter((file) => file.type.startsWith('image/'))
    if (imageFiles.length !== files.length) {
      alert('Please choose image files only')
      return
    }

    const nextImages = append ? [...referenceImages, ...imageFiles] : imageFiles
    setReferenceImages(nextImages, `${nextImages.length} uploaded image${nextImages.length > 1 ? 's' : ''}`)
  }

  const handleReferenceImageChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    attachReferenceImages(Array.from(event.target.files || []), true)
  }

  const handleDocumentChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] || null
    if (!file) {
      setDocumentFile(null)
      return
    }

    const lowerName = file.name.toLowerCase()
    const isSupported = ['.txt', '.md', '.markdown', '.csv', '.docx', '.pptx'].some((ext) => lowerName.endsWith(ext))
    if (!isSupported) {
      alert('Please upload txt, md, csv, docx, or pptx')
      event.target.value = ''
      return
    }

    setDocumentFile(file)
    setComposerMode('ppt')
  }

  const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setIsDraggingImages(false)
    attachReferenceImages(Array.from(event.dataTransfer.files || []), true)
  }

  const handlePaste = (event: React.ClipboardEvent<HTMLTextAreaElement>) => {
    const clipboardItems = Array.from(event.clipboardData.items || [])
    const imageFiles = clipboardItems
      .filter((item) => item.kind === 'file' && item.type.startsWith('image/'))
      .map((item, index) => {
        const file = item.getAsFile()
        if (!file) return null
        const extension = file.type.split('/')[1] || 'png'
        return new File([file], file.name || `clipboard-image-${index + 1}.${extension}`, { type: file.type })
      })
      .filter((file): file is File => Boolean(file))

    if (imageFiles.length > 0) {
      event.preventDefault()
      attachReferenceImages(imageFiles, true)
    }
  }

  return (
    <div className="mx-auto w-full max-w-4xl px-3 py-2 sm:px-4 sm:py-3">
      <div className="mb-2 flex flex-wrap items-center gap-2">
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setComposerMode('image')}
            className={`rounded-full px-3 py-1 text-xs ${composerMode === 'image' ? 'bg-white text-black' : 'bg-card text-muted-foreground'}`}
          >
            Image
          </button>
          {!isMobile && (
            <button
              type="button"
              onClick={() => setComposerMode('ppt')}
              className={`rounded-full px-3 py-1 text-xs ${composerMode === 'ppt' ? 'bg-white text-black' : 'bg-card text-muted-foreground'}`}
            >
              PPT Mode
            </button>
          )}
        </div>
        {!isMobile && (
          <a
            href="https://aippt.wps.cn/aippt/convert-ppt/home"
            target="_blank"
            rel="noreferrer"
            className="ml-1 inline-flex items-center gap-1.5 rounded-full bg-card px-3 py-1 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            title="WPS AIPPT"
          >
            <img src="/gptimg/wps-aippt-logo.png" alt="" className="h-4 w-4 rounded-sm" />
            <span>WPS AIPPT</span>
          </a>
        )}
      </div>

      {composerMode === 'ppt' && !isMobile && (
        <div className="mb-2 grid gap-2 md:grid-cols-3">
          {pptModeOptions.map((option) => (
            <button
              key={option.id}
              type="button"
              onClick={() => setPptWorkflowMode(option.id)}
              className={`rounded-xl border px-3 py-2 text-left transition-colors ${
                pptWorkflowMode === option.id
                  ? 'border-white/70 bg-white text-black'
                  : 'border-border bg-card text-secondary-foreground hover:border-white/30 hover:bg-accent'
              }`}
            >
              <div className="text-xs font-semibold">{option.title}</div>
              <div className={`mt-1 text-[11px] ${pptWorkflowMode === option.id ? 'text-black/60' : 'text-muted-foreground'}`}>
                {option.hint}
              </div>
            </button>
          ))}
        </div>
      )}

      {showOptions && (
        <div className="mb-2 rounded-lg border border-border bg-card p-3">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3 sm:gap-4">
            <div>
              <label className="block text-xs text-muted-foreground mb-1">Size</label>
              <select
                value={size}
                onChange={(e) => setSize(e.target.value)}
                className="w-full rounded border border-border bg-background px-2 py-2 text-sm text-foreground"
              >
                <option value="1024x1024">1024×1024</option>
                <option value="1792x1024">1792×1024</option>
                <option value="1536x1024">1536×1024</option>
                <option value="1024x1536">1024×1536</option>
              </select>
            </div>
            <div>
              <label className="block text-xs text-muted-foreground mb-1">Quality</label>
              <select
                value={quality}
                onChange={(e) => setQuality(e.target.value)}
                className="w-full rounded border border-border bg-background px-2 py-2 text-sm text-foreground"
              >
                <option value="standard">Standard / Medium (1 credit)</option>
                <option value="hd">HD / High (2 credits)</option>
              </select>
            </div>
            <div>
              <label className="block text-xs text-muted-foreground mb-1">Style</label>
              <select
                value={style}
                onChange={(e) => setStyle(e.target.value)}
                className="w-full rounded border border-border bg-background px-2 py-2 text-sm text-foreground"
              >
                <option value="">Default</option>
                <option value="vivid">Vivid</option>
                <option value="natural">Natural</option>
              </select>
            </div>
          </div>
        </div>
      )}

      <div
        className={`relative rounded-2xl border bg-card transition-colors ${
          isDraggingImages ? 'border-sky-400 bg-sky-500/10' : 'border-border'
        }`}
        onDragEnter={(event) => {
          event.preventDefault()
          setIsDraggingImages(true)
        }}
        onDragOver={(event) => event.preventDefault()}
        onDragLeave={(event) => {
          if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
            setIsDraggingImages(false)
          }
        }}
        onDrop={handleDrop}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          multiple
          className="hidden"
          onChange={handleReferenceImageChange}
        />
        <input
          ref={documentInputRef}
          type="file"
          accept=".txt,.md,.markdown,.csv,.docx,.pptx"
          className="hidden"
          onChange={handleDocumentChange}
        />

        {composerMode === 'ppt' && documentFile && (
          <div className="border-b border-border px-3 py-2">
            <div className="flex items-center gap-3 rounded-xl border border-sky-500/20 bg-sky-500/10 px-3 py-2">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-sky-400/15 text-xs font-semibold text-sky-200">
                DOC
              </div>
              <div className="min-w-0 flex-1">
                <div className="truncate text-sm text-foreground">{documentFile.name}</div>
                <div className="text-xs text-sky-200/80">DeepSeek will split this document into PPT pages first</div>
              </div>
              <button
                type="button"
                onClick={() => {
                  setDocumentFile(null)
                  if (documentInputRef.current) {
                    documentInputRef.current.value = ''
                  }
                }}
                className="rounded-lg px-2 py-1 text-xs text-secondary-foreground transition-colors hover:bg-accent hover:text-foreground"
              >
                Remove
              </button>
            </div>
          </div>
        )}

        {referenceImages.length > 0 && (
          <div className="border-b border-border px-3 py-2">
            <div className="flex items-center gap-3">
              <div className="grid max-w-[118px] grid-cols-4 gap-1">
                {referencePreviewUrls.slice(0, 6).map((url, index) => (
                  <img
                    key={url}
                    src={url}
                    alt={`Reference ${index + 1}`}
                    className="h-9 w-9 rounded-md object-cover"
                  />
                ))}
                {referenceImages.length > 6 && (
                  <div className="flex h-9 w-9 items-center justify-center rounded-md bg-accent text-xs text-secondary-foreground">
                    +{referenceImages.length - 6}
                  </div>
                )}
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-sm text-foreground">
                  {composerMode === 'ppt'
                    ? `${referenceImages.length} PPT reference image${referenceImages.length > 1 ? 's' : ''} attached`
                    : `${referenceImages.length} reference image${referenceImages.length > 1 ? 's' : ''} attached`}
                </div>
                <div className="truncate text-xs text-muted-foreground">
                  {referenceImageLabel || referenceImages.map((file) => file.name).join(', ')}
                </div>
              </div>
              <button
                type="button"
                onClick={() => {
                  clearReferenceImage()
                  if (fileInputRef.current) {
                    fileInputRef.current.value = ''
                  }
                }}
                className="rounded-lg px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              >
                Remove
              </button>
            </div>
          </div>
        )}

        <textarea
          ref={textareaRef}
          value={prompt}
          onChange={(e) => {
            setPrompt(e.target.value)
            handleTextareaInput()
          }}
          onKeyDown={handleKeyDown}
          onPaste={handlePaste}
          placeholder={
            isDraggingImages
              ? 'Drop images here to attach them...'
              : composerMode === 'ppt'
                ? documentFile
                  ? 'Optional: add style, audience, slide count, or key emphasis for this document...'
                  : pptWorkflowMode === 'style_text'
                    ? 'Paste page text here. You can write Page 1 / Page 2, or paste long content and let DeepSeek split it...'
                    : pptWorkflowMode === 'image_content_style'
                      ? 'Describe the unified PPT style, e.g. clean green academic style, technology defense style...'
                      : 'Paste images and text. AI will freely organize, beautify, and complete the PPT...'
                : 'Describe the image you want to generate...'
          }
          className="min-h-[86px] w-full resize-none rounded-2xl bg-transparent px-4 py-3 pr-4 text-foreground placeholder-gray-500 focus:outline-none sm:min-h-0 sm:pr-24"
          rows={1}
          disabled={isGenerating}
        />

        <div className="absolute bottom-2 right-2 flex items-center gap-2">
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            className={`rounded-lg p-1.5 transition-colors ${
              referenceImages.length > 0
                ? 'bg-white text-black hover:bg-gray-200'
                : 'text-muted-foreground hover:bg-accent hover:text-foreground'
            }`}
            title="Reference image"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <path d="M21 15l-5-5L5 21" />
            </svg>
          </button>
          {false && composerMode === 'ppt' && !isMobile && (
            <button
              type="button"
              onClick={() => documentInputRef.current?.click()}
              className={`rounded-lg p-1.5 transition-colors ${
                documentFile
                  ? 'bg-sky-200 text-black hover:bg-sky-100'
                  : 'text-muted-foreground hover:bg-accent hover:text-foreground'
              }`}
              title="Generate PPT from document"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <path d="M14 2v6h6" />
                <path d="M8 13h8" />
                <path d="M8 17h5" />
              </svg>
            </button>
          )}
          <button
            onClick={() => setShowOptions(!showOptions)}
            className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            title="Options"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
            </svg>
          </button>

          <button
            onClick={handleSubmit}
            disabled={(!prompt.trim() && !documentFile) || isGenerating || isPlanningPpt}
            className="rounded-lg bg-white p-1.5 text-black transition-colors hover:bg-gray-200 disabled:cursor-not-allowed disabled:opacity-30"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="22" y1="2" x2="11" y2="13" />
              <polygon points="22 2 15 22 11 13 2 9 22 2" />
            </svg>
          </button>
        </div>
      </div>

      <div className="mt-1.5 flex flex-col gap-1 px-1 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-2">
        <span>
          Credits: {user?.credits ?? 0}
        </span>
        <span>
          {composerMode === 'ppt' && !isMobile
            ? (isPlanningPpt
                ? 'Planning PPT...'
                : documentFile
                  ? 'Document PPT is hidden as a legacy path'
                  : pptProgress.message ||
                    (pptWorkflowMode === 'style_text'
                      ? 'Mode 1: first image is style only; text becomes slide content'
                      : pptWorkflowMode === 'image_content_style'
                        ? 'Mode 2: each uploaded image maps to one slide; text controls shared style'
                        : 'Mode 3: images and text are freely organized and beautified'))
            : referenceImages.length > 0
              ? `Prompt + ${referenceImages.length} image${referenceImages.length > 1 ? 's' : ''} will be sent together`
              : 'Shift+Enter for new line'}
        </span>
      </div>
    </div>
  )
}
