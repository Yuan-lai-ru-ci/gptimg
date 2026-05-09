'use client'

import { useEffect, useRef } from 'react'
import { useChatStore } from '@/lib/store/chatStore'
import MessageItem from './MessageItem'
import Loading from '../common/Loading'

export default function MessageList() {
  const { messages, isLoadingMessages, isGenerating, currentSession } = useChatStore()
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  if (!currentSession) {
    return (
      <div className="h-full flex items-center justify-center text-muted-foreground">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-2">GPT Image Generator</h2>
          <p>Start a new conversation to generate images</p>
        </div>
      </div>
    )
  }

  if (isLoadingMessages) {
    return (
      <div className="h-full flex items-center justify-center">
        <Loading />
      </div>
    )
  }

  return (
    <div className="h-full overflow-y-auto p-4">
      <div className="max-w-3xl mx-auto space-y-6">
        {messages.map((message) => (
          <MessageItem key={message.id} message={message} />
        ))}

        {isGenerating && (
          <div className="flex justify-center py-8">
            <div className="text-center">
              <Loading />
              <p className="text-muted-foreground mt-4">Generating image...</p>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>
    </div>
  )
}
