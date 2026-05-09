'use client'

import { useMemo } from 'react'
import { useChatStore } from '@/lib/store/chatStore'
import { Button } from '@/components/ui/button'

interface PptPlanPanelProps {
  docked?: boolean
}

export default function PptPlanPanel({ docked = false }: PptPlanPanelProps) {
  const { pptPlan, updatePptPlan, clearPptPlan, generatePptDeck, isGenerating, pptProgress } = useChatStore()

  const slideCount = useMemo(() => pptPlan?.slides.length ?? 0, [pptPlan])
  const modeLabel = useMemo(() => {
    if (pptPlan?.generation_mode === 'style_text') return '模式① 风格图 + 文字'
    if (pptPlan?.generation_mode === 'image_content_style') return '模式② 多图内容 + 风格'
    if (pptPlan?.generation_mode === 'free_mix') return '模式③ 自由美化'
    return 'PPT Mode'
  }, [pptPlan?.generation_mode])

  if (!pptPlan) return null

  return (
    <div className={docked ? 'h-full min-h-0 p-3 xl:p-4' : 'mx-auto max-w-5xl p-4'}>
      <div className={`rounded-2xl border border-border bg-card p-3 xl:p-4 ${docked ? 'h-full min-h-0 overflow-y-auto' : ''}`}>
        <div className="mb-3 flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
          <div>
            <div className="text-lg font-semibold text-foreground">PPT Plan</div>
            <div className="mt-1 text-sm text-muted-foreground">
              {modeLabel} · {slideCount} slides planned. Review and edit before batch generation.
            </div>
            {pptProgress.message ? (
              <div className="mt-2 text-xs text-sky-300">{pptProgress.message}</div>
            ) : null}
          </div>
          <div className="flex gap-2">
            <Button size="sm" variant="ghost" onClick={clearPptPlan}>
              Clear
            </Button>
            <Button size="sm" onClick={() => void generatePptDeck()} disabled={isGenerating}>
              {isGenerating ? 'Generating...' : 'Generate Slides'}
            </Button>
          </div>
        </div>

        <div className="mb-3 grid gap-3 md:grid-cols-2">
          <label className="block">
            <span className="mb-2 block text-sm text-secondary-foreground">Deck Title</span>
            <input
              value={pptPlan.deck_title}
              onChange={(e) => updatePptPlan({ ...pptPlan, deck_title: e.target.value })}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground"
            />
          </label>
          <label className="block">
            <span className="mb-2 block text-sm text-secondary-foreground">Deck Goal</span>
            <input
              value={pptPlan.deck_goal}
              onChange={(e) => updatePptPlan({ ...pptPlan, deck_goal: e.target.value })}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground"
            />
          </label>
        </div>

        <label className="mb-3 block">
          <span className="mb-2 block text-sm text-secondary-foreground">Visual Direction</span>
          <textarea
            value={pptPlan.visual_direction}
            onChange={(e) => updatePptPlan({ ...pptPlan, visual_direction: e.target.value })}
            className="min-h-[72px] w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </label>

        <label className="mb-3 block">
          <span className="mb-2 block text-sm text-secondary-foreground">Master Style Description</span>
          <textarea
            value={pptPlan.master_style_description ?? ''}
            onChange={(e) => updatePptPlan({ ...pptPlan, master_style_description: e.target.value })}
            className="min-h-[88px] w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
          <div className="mt-2 text-xs text-muted-foreground">
            This is the reusable middle-slide design system. Keep palette, material, camera language, overlays,
            spacing rhythm, and editable title/content zones stable here.
          </div>
        </label>

        {pptPlan.consistency_rules.length > 0 ? (
          <div className="mb-3">
            <div className="mb-2 text-sm text-secondary-foreground">Consistency Rules</div>
            <div className="flex flex-wrap gap-2">
              {pptPlan.consistency_rules.map((rule, index) => (
                <span
                  key={`${rule}-${index}`}
                  className="rounded-full border border-sky-500/20 bg-sky-500/10 px-3 py-1 text-xs text-sky-200"
                >
                  {rule}
                </span>
              ))}
            </div>
          </div>
        ) : null}

        <div className="space-y-2.5">
          {pptPlan.slides.map((slide, index) => (
            <div key={index} className="rounded-xl border border-border bg-background p-3">
              <div className="mb-2 text-sm font-medium text-foreground">Slide {slide.slide_number}</div>
              <div className="grid gap-3 md:grid-cols-2">
                <label className="block">
                  <span className="mb-2 block text-xs text-muted-foreground">Title</span>
                  <input
                    value={slide.title}
                    onChange={(e) => {
                      const nextSlides = [...pptPlan.slides]
                      nextSlides[index] = { ...slide, title: e.target.value }
                      updatePptPlan({ ...pptPlan, slides: nextSlides })
                    }}
                    className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground"
                  />
                </label>
                <label className="block">
                  <span className="mb-2 block text-xs text-muted-foreground">Objective</span>
                  <input
                    value={slide.objective}
                    onChange={(e) => {
                      const nextSlides = [...pptPlan.slides]
                      nextSlides[index] = { ...slide, objective: e.target.value }
                      updatePptPlan({ ...pptPlan, slides: nextSlides })
                    }}
                    className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground"
                  />
                </label>
              </div>
              <label className="mt-3 block">
                <span className="mb-2 block text-xs text-muted-foreground">Page Description</span>
                <textarea
                  value={slide.page_description ?? ''}
                  onChange={(e) => {
                    const nextSlides = [...pptPlan.slides]
                    nextSlides[index] = { ...slide, page_description: e.target.value }
                    updatePptPlan({ ...pptPlan, slides: nextSlides })
                  }}
                  className="min-h-[88px] w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground"
                />
              </label>
              <label className="mt-3 block">
                <span className="mb-2 block text-xs text-muted-foreground">Image Prompt</span>
                <textarea
                  value={slide.image_prompt}
                  onChange={(e) => {
                    const nextSlides = [...pptPlan.slides]
                    nextSlides[index] = { ...slide, image_prompt: e.target.value }
                    updatePptPlan({ ...pptPlan, slides: nextSlides })
                  }}
                  className="min-h-[72px] w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground"
                />
              </label>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
