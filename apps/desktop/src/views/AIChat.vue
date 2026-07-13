<template>
  <div class="aichat">
    <div class="aichat__header">
      <span class="aichat__title">AI 助手</span>
      <div class="aichat__actions">
        <button class="aichat__btn" @click="clearChat" title="清空对话">清空</button>
      </div>
    </div>

    <div ref="messagesRef" class="aichat__messages">
      <div v-if="messages.length === 0" class="aichat__welcome">
        <div class="aichat__welcome-icon">🤖</div>
        <h3>你好，我是 AIStudio 助手</h3>
        <p>我可以帮你：</p>
        <ul>
          <li>自然语言描述需求，自动生成工作流</li>
          <li>分析错误，帮你修复问题</li>
          <li>解释节点功能，推荐最佳实践</li>
          <li>分析日志，找出问题原因</li>
        </ul>
      </div>

      <div v-for="(msg, idx) in messages" :key="idx" :class="['aichat__msg', `aichat__msg--${msg.role}`]">
        <div class="aichat__bubble">
          <div v-if="msg.role === 'assistant'" class="aichat__bubble-content" v-html="renderMarkdown(msg.content)" />
          <div v-else class="aichat__bubble-content">{{ msg.content }}</div>
        </div>
      </div>

      <div v-if="streaming" class="aichat__msg aichat__msg--assistant">
        <div class="aichat__bubble">
          <div class="aichat__bubble-content" v-html="renderMarkdown(streamingContent)" />
          <span class="aichat__cursor" />
        </div>
      </div>
    </div>

    <div class="aichat__input-area">
      <textarea
        ref="inputRef"
        v-model="inputText"
        class="aichat__input"
        placeholder="输入你的问题... (Enter 发送, Ctrl+Enter 换行)"
        rows="1"
        @keydown="handleKeydown"
      />
      <button :class="['aichat__send', { 'aichat__send--active': inputText.trim() }]" @click="sendMessage" :disabled="!inputText.trim() || streaming">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { sendChatStream } from '@/api/agent'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

const messages = ref<Message[]>([])
const inputText = ref('')
const streaming = ref(false)
const streamingContent = ref('')
const messagesRef = ref<HTMLElement>()
const inputRef = ref<HTMLTextAreaElement>()

let abortController: AbortController | null = null

function renderMarkdown(text: string) {
  return text
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code class="lang-$1">$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\n/g, '<br/>')
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

async function sendMessage() {
  const text = inputText.value.trim()
  if (!text || streaming.value) return

  messages.value.push({ role: 'user', content: text })
  inputText.value = ''
  streaming.value = true
  streamingContent.value = ''
  scrollToBottom()

  const controller = new AbortController()
  abortController = controller

  await sendChatStream(
    { message: text },
    {
      signal: controller.signal,
      onChunk: (chunk) => {
        streamingContent.value = chunk
        scrollToBottom()
      },
      onDone: () => {
        messages.value.push({ role: 'assistant', content: streamingContent.value })
        streaming.value = false
        streamingContent.value = ''
        abortController = null
        scrollToBottom()
      },
      onError: (e) => {
        messages.value.push({ role: 'assistant', content: `抱歉，请求失败：${e.message || '未知错误'}` })
        streaming.value = false
        streamingContent.value = ''
        abortController = null
        scrollToBottom()
      },
    }
  )
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.ctrlKey && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

function clearChat() {
  messages.value = []
}

onMounted(() => {
  inputRef.value?.focus()
})
</script>

<style scoped>
.aichat {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.aichat__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 48px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-secondary);
  flex-shrink: 0;
}

.aichat__title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.aichat__btn {
  font-size: 13px;
  color: var(--text-tertiary);
  padding: 4px 12px;
  border-radius: var(--radius-xs);
  transition: all var(--transition-fast);
}

.aichat__btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.aichat__messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.aichat__welcome {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  text-align: center;
  color: var(--text-secondary);
}

.aichat__welcome-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.aichat__welcome h3 {
  font-size: 18px;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.aichat__welcome ul {
  list-style: none;
  padding: 0;
  text-align: left;
  margin-top: 8px;
}

.aichat__welcome li {
  padding: 4px 0;
  font-size: 13px;
}

.aichat__welcome li::before {
  content: '• ';
  color: var(--primary);
}

.aichat__msg {
  display: flex;
}

.aichat__msg--user {
  justify-content: flex-end;
}

.aichat__msg--assistant {
  justify-content: flex-start;
}

.aichat__bubble {
  max-width: 75%;
  padding: 10px 16px;
  border-radius: var(--radius-md);
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
}

.aichat__msg--user .aichat__bubble {
  background: var(--primary);
  color: #fff;
  border-top-right-radius: 4px;
}

.aichat__msg--assistant .aichat__bubble {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border-top-left-radius: 4px;
}

.aichat__bubble-content :deep(pre) {
  background: var(--bg-primary);
  padding: 12px;
  border-radius: var(--radius-xs);
  overflow-x: auto;
  margin: 8px 0;
  font-family: var(--font-mono);
  font-size: 13px;
}

.aichat__bubble-content :deep(code) {
  font-family: var(--font-mono);
  font-size: 13px;
  background: var(--bg-hover);
  padding: 2px 6px;
  border-radius: 4px;
}

.aichat__bubble-content :deep(pre code) {
  background: none;
  padding: 0;
}

.aichat__cursor {
  display: inline-block;
  width: 2px;
  height: 16px;
  background: var(--primary);
  margin-left: 2px;
  animation: pulse 1s ease-in-out infinite;
  vertical-align: text-bottom;
}

.aichat__input-area {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-secondary);
  flex-shrink: 0;
}

.aichat__input {
  flex: 1;
  min-height: 40px;
  max-height: 200px;
  padding: 10px 14px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xs);
  color: var(--text-primary);
  font-size: 14px;
  resize: none;
  outline: none;
  transition: border-color var(--transition-fast);
  line-height: 1.5;
}

.aichat__input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.aichat__input::placeholder {
  color: var(--text-tertiary);
}

.aichat__send {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.aichat__send--active {
  background: var(--primary);
  color: #fff;
}

.aichat__send--active:hover {
  background: var(--primary-hover);
}

.aichat__send:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>