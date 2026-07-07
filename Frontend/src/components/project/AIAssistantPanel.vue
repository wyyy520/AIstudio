<template>
  <div class="ai-assistant-panel">
    <div class="panel-header">
      <div class="header-title">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
        </svg>
        <span>AI Assistant</span>
      </div>
    </div>

    <div class="panel-content">
      <div class="info-section">
        <div class="info-item">
          <span class="info-label">Current Project</span>
          <span class="info-value">{{ currentProjectName }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Current Workflow</span>
          <span class="info-value">{{ currentWorkflowName }}</span>
        </div>
      </div>

      <div v-if="currentError" class="error-section">
        <div class="error-header">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 8v4M12 16h.01" />
          </svg>
          <span>检测到问题</span>
        </div>
        <div class="error-message">{{ currentError }}</div>
      </div>

      <div class="suggestion-section">
        <div class="suggestion-header">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83" />
          </svg>
          <span>AI 建议</span>
        </div>
        <div class="suggestion-content">
          <p>{{ suggestion }}</p>
        </div>
        <button v-if="currentError" class="fix-btn" @click="$emit('apply-fix')">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
            <polyline points="20 6 9 17 4 12" />
          </svg>
          Apply Fix
        </button>
      </div>

      <div class="chat-section">
        <div class="chat-messages">
          <div
            v-for="msg in messages"
            :key="msg.id"
            class="chat-message"
            :class="`message-${msg.role}`"
          >
            <div class="message-avatar">
              <svg v-if="msg.role === 'assistant'" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M12 2L2 7l10 5 10-5-10-5zm0 22L2 17l10-5 10 5-10 5z" />
              </svg>
              <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M12 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8z" />
              </svg>
            </div>
            <div class="message-content">{{ msg.content }}</div>
          </div>
        </div>

        <div class="chat-input">
          <input
            v-model="inputText"
            type="text"
            placeholder="向 AI 提问..."
            @keydown.enter="sendMessage"
          />
          <button class="send-btn" @click="sendMessage">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5">
              <line x1="22" y1="2" x2="11" y2="13" />
              <polygon points="22 2 15 22 11 13 2 9 22 2" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useProjectStore } from '@/store/project'

const store = useProjectStore()

defineEmits<{
  'apply-fix': []
}>()

const inputText = ref('')

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
}

const messages = ref<ChatMessage[]>([
  {
    id: '1',
    role: 'assistant',
    content: '你好！我是 AI 助手，可以帮助你管理项目和解决开发中的问题。',
  },
])

const currentProjectName = computed(() => store.currentProject?.name || '未选择')
const currentWorkflowName = computed(() => {
  if (!store.currentProject) return '无'
  const running = store.currentProject.workflows.find(w => w.status === 'running')
  return running?.name || '无'
})

const currentError = computed(() => {
  if (!store.currentProject) return ''
  if (store.currentProject.environment.gpuStatus === 'error') {
    return 'CUDA 版本不匹配，可能导致训练失败'
  }
  const outdated = store.currentProject.environment.dependencies.find(d => d.status === 'outdated')
  if (outdated) {
    return `${outdated.name} 需要更新到最新版本`
  }
  return ''
})

const suggestion = computed(() => {
  if (currentError.value) {
    if (store.currentProject?.environment.gpuStatus === 'error') {
      return '建议更新 CUDA 驱动到与 PyTorch 兼容的版本，或者重新安装 PyTorch。'
    }
    return '建议运行环境修复工具来解决依赖问题。'
  }
  return '项目状态良好，可以继续进行开发工作。'
})

function sendMessage() {
  if (!inputText.value.trim()) return

  messages.value.push({
    id: Date.now().toString(),
    role: 'user',
    content: inputText.value,
  })

  inputText.value = ''

  setTimeout(() => {
    messages.value.push({
      id: (Date.now() + 1).toString(),
      role: 'assistant',
      content: '收到你的问题，让我帮你分析一下...',
    })
  }, 500)
}
</script>

<style scoped>
.ai-assistant-panel {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
  border-left: 1px solid var(--border-subtle);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border-subtle);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--primary);
}

.panel-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.info-section {
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border-subtle);
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: var(--spacing-2);
}

.info-item:last-child {
  margin-bottom: 0;
}

.info-label {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.info-value {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
}

.error-section {
  margin: var(--spacing-3) var(--spacing-4);
  padding: var(--spacing-3);
  background: var(--error-bg);
  border-radius: var(--radius-md);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.error-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  color: var(--error);
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  margin-bottom: var(--spacing-2);
}

.error-message {
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

.suggestion-section {
  padding: var(--spacing-3) var(--spacing-4);
  border-bottom: 1px solid var(--border-subtle);
}

.suggestion-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  color: var(--warning);
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  margin-bottom: var(--spacing-2);
}

.suggestion-content {
  font-size: var(--text-caption);
  color: var(--text-secondary);
  line-height: var(--leading-caption);
  margin-bottom: var(--spacing-3);
}

.fix-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 32px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-md);
  background: var(--primary);
  border: none;
  color: white;
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.fix-btn:hover {
  background: var(--primary-hover);
}

.chat-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-3) var(--spacing-4);
}

.chat-message {
  display: flex;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-3);
}

.message-avatar {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  flex-shrink: 0;
}

.message-assistant .message-avatar {
  background: var(--primary-bg);
  color: var(--primary);
}

.message-user .message-avatar {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.message-content {
  flex: 1;
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  line-height: var(--leading-body-sm);
  padding: var(--spacing-2) var(--spacing-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
}

.message-user .message-content {
  background: var(--primary-bg);
}

.chat-input {
  display: flex;
  gap: var(--spacing-2);
  padding: var(--spacing-3) var(--spacing-4);
  border-top: 1px solid var(--border-subtle);
}

.chat-input input {
  flex: 1;
  height: 36px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: var(--text-body-sm);
}

.chat-input input:focus {
  border-color: var(--primary);
  outline: none;
}

.chat-input input::placeholder {
  color: var(--text-tertiary);
}

.send-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary);
  border: none;
  border-radius: var(--radius-md);
  color: white;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.send-btn:hover {
  background: var(--primary-hover);
}
</style>
