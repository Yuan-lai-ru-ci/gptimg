'use client'

import { useEffect, useRef } from 'react'
import { useChatStore } from '@/lib/store/chatStore'
import MessageItem from './MessageItem'
import Loading from '../common/Loading'

export default function MessageList() {
  const { messages, isLoadingMessages, isGenerating, currentSession, isPlanningPpt, pptProgress, pptPlan } = useChatStore()
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  if (!currentSession) {
    return (
      <div className="flex h-full items-center justify-center px-6 text-muted-foreground">
        <div className="text-center">
          <h2 className="mb-2 text-xl font-bold text-foreground sm:text-2xl">GPT Image Generator</h2>
          <p className="text-sm sm:text-base">Start a new conversation to generate images</p>
        </div>
      </div>
    )
  }

  if (isLoadingMessages) {
    return (
      <div className="h-full flex items-center justify-center">
        <Loading label="Loading conversation..." />
      </div>
    )
  }

  return (
    <div className="h-full min-h-0 overflow-y-auto px-3 py-4 sm:p-4">
      <div className="mx-auto max-w-4xl space-y-4 sm:space-y-5">
        {messages.length === 0 && pptPlan && !isPlanningPpt && !isGenerating ? (
          <div className="rounded-2xl border border-border bg-card/60 p-6 text-center">
            <div className="text-lg font-semibold text-foreground">Slide Preview Area</div>
            <p className="mt-2 text-sm text-muted-foreground">
              Edit the plan on the left, then generate slides. The finished visuals will appear here.
            </p>
          </div>
        ) : null}

        {messages.map((message) => (
          <MessageItem key={message.id} message={message} />
        ))}

        {isPlanningPpt && (
          <div className="rounded-2xl border border-sky-500/20 bg-sky-500/5 p-5">
            <Loading label={pptProgress.message || 'Planning your PPT...'} className="py-3" />
            <p className="mt-3 text-center text-xs text-muted-foreground">
              The AI is drafting a science and innovation deck structure based on your brief.
            </p>
          </div>
        )}

        {isGenerating && (
          <div className="flex justify-center py-8">
            <div className="w-full max-w-2xl rounded-2xl border border-emerald-500/20 bg-emerald-500/5 p-5 text-center">
              <Loading
                label={
                  pptPlan
                    ? (pptProgress.message || 'Generating slide visuals...')
                    : 'Generating image...'
                }
              />
              <p className="mt-3 text-xs text-muted-foreground">
                {pptPlan
                  ? 'The slide visuals will appear here when this batch is complete.'
                  : 'The result will appear in this conversation.'}
              </p>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>
    </div>
  )
}
