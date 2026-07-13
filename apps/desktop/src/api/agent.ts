import request from './request'
import type { AgentResponse, ToolCall, PlanItem } from '@/pages/AIChat/types'
import { useUserStore } from '@/stores/user'

export interface ChatRequest {
  message: string
  projectId?: string
  userId?: string
  context?: Record<string, unknown>
}

export interface ChatSendOptions {
  signal?: AbortSignal
  onChunk?: (text: string) => void
  onStatus?: (status: string) => void
  onToolCall?: (call: ToolCall) => void
  onPlan?: (plan: PlanItem[]) => void
  onDone?: (response: AgentResponse) => void
  onError?: (error: Error) => void
}

export async function sendChat(data: ChatRequest) {
  return request.post('/agent/chat', data)
}

export async function sendChatStream(data: ChatRequest, options: ChatSendOptions): Promise<void> {
  const { onChunk, onStatus, onToolCall, onPlan, onDone, onError } = options
  const baseURL = import.meta.env.VITE_API_BASE_URL || ''
  const token = useUserStore().token

  try {
    onStatus?.('thinking')

    const response = await fetch(`${baseURL}/agent/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(data),
      signal: options.signal,
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    if (options.signal?.aborted) return

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('ReadableStream not supported')
    }

    const decoder = new TextDecoder()
    let buffer = ''
    let accumulatedText = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      let currentEvent = ''
      let currentData = ''

      for (const line of lines) {
        if (line.startsWith('event: ')) {
          currentEvent = line.slice(7).trim()
        } else if (line.startsWith('data: ')) {
          currentData = line.slice(6).trim()
        } else if (line === '') {
          if (currentEvent && currentData) {
            processSSEEvent(currentEvent, currentData)
          }
          currentEvent = ''
          currentData = ''
        }
      }
    }

    if (buffer.trim()) {
      const lastLines = buffer.split('\n')
      let currentEvent = ''
      let currentData = ''
      for (const line of lastLines) {
        if (line.startsWith('event: ')) {
          currentEvent = line.slice(7).trim()
        } else if (line.startsWith('data: ')) {
          currentData = line.slice(6).trim()
        }
      }
      if (currentEvent && currentData) {
        processSSEEvent(currentEvent, currentData)
      }
    }
  } catch (err: any) {
    if (err.name === 'AbortError') return
    onStatus?.('error')
    onError?.(err instanceof Error ? err : new Error(String(err)))
  }

  function processSSEEvent(event: string, dataStr: string) {
    try {
      const parsed = JSON.parse(dataStr)
      switch (event) {
        case 'token':
          accumulatedText += parsed.text || ''
          onChunk?.(accumulatedText)
          break
        case 'action':
          onStatus?.(parsed.type || 'thinking')
          break
        case 'done':
          onStatus?.('finished')
          onDone?.({
            reply: accumulatedText,
            summary: accumulatedText,
            goal: '',
            explanation: '',
            plan: [],
            steps: [],
            status: 'completed',
          })
          break
        case 'error':
          onStatus?.('error')
          onError?.(new Error(parsed.message || parsed.code || 'Unknown error'))
          break
      }
    } catch {
      // Ignore malformed SSE data
    }
  }
}

export function planOnly(data: { message: string }) {
  return request.post('/agent/generate-workflow', data)
}
