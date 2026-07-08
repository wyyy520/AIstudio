export type WebSocketEventType = 'task_status' | 'task_log' | 'task_progress' | 'task_error' | 'task_complete'

export interface WebSocketEvent {
  type: WebSocketEventType
  taskId: string
  data: {
    status?: string
    progress?: number
    message?: string
    level?: string
    step?: string
    error?: string
    result?: unknown
    timestamp: string
  }
}

export type WebSocketCallback = (event: WebSocketEvent) => void

class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = 2000
  private maxReconnectDelay = 30000
  private callbacks: Set<WebSocketCallback> = new Set()
  private taskCallbacks: Map<string, Set<WebSocketCallback>> = new Map()
  private isConnected = false
  private shouldReconnect = true

  constructor() {
    const baseUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8081'
    this.url = `${baseUrl}/api/ws`
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return

    this.shouldReconnect = true
    try {
      this.ws = new WebSocket(this.url)
      this.ws.onopen = this.handleOpen.bind(this)
      this.ws.onmessage = this.handleMessage.bind(this)
      this.ws.onclose = this.handleClose.bind(this)
      this.ws.onerror = this.handleError.bind(this)
    } catch (err) {
      console.error('[ws] connection error:', err)
      this.scheduleReconnect()
    }
  }

  disconnect(): void {
    this.shouldReconnect = false
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.ws?.close()
    this.ws = null
    this.isConnected = false
  }

  subscribe(callback: WebSocketCallback): () => void {
    this.callbacks.add(callback)
    return () => this.callbacks.delete(callback)
  }

  subscribeTask(taskId: string, callback: WebSocketCallback): () => void {
    let callbacks = this.taskCallbacks.get(taskId)
    if (!callbacks) {
      callbacks = new Set()
      this.taskCallbacks.set(taskId, callbacks)
    }
    callbacks.add(callback)
    return () => {
      callbacks?.delete(callback)
      if (callbacks?.size === 0) {
        this.taskCallbacks.delete(taskId)
      }
    }
  }

  get connected(): boolean {
    return this.isConnected
  }

  private handleOpen(): void {
    this.isConnected = true
    this.reconnectDelay = 2000
    console.log('[ws] connected')
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const parsed: WebSocketEvent = JSON.parse(event.data)
      this.dispatch(parsed)
    } catch {
      console.warn('[ws] failed to parse message:', event.data)
    }
  }

  private handleClose(): void {
    this.isConnected = false
    console.log('[ws] disconnected')
    if (this.shouldReconnect) {
      this.scheduleReconnect()
    }
  }

  private handleError(err: Event): void {
    console.error('[ws] error:', err)
  }

  private dispatch(event: WebSocketEvent): void {
    for (const cb of this.callbacks) {
      try { cb(event) } catch { /* ignore */ }
    }

    const taskCallbacks = this.taskCallbacks.get(event.taskId)
    if (taskCallbacks) {
      for (const cb of taskCallbacks) {
        try { cb(event) } catch { /* ignore */ }
      }
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.reconnectDelay = Math.min(this.reconnectDelay * 1.5, this.maxReconnectDelay)
      console.log(`[ws] reconnecting in ${this.reconnectDelay}ms...`)
      this.connect()
    }, this.reconnectDelay)
  }
}

export const wsClient = new WebSocketClient()