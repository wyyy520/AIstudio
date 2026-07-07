<template>
  <div
    class="input-box"
    :class="{ 'is-focused': focused, 'is-disabled': disabled }"
    @dragover.prevent="dragOver = true"
    @dragleave="dragOver = false"
    @drop.prevent="handleDrop"
  >
    <div v-if="attachments.length > 0" class="input-box-attachments">
      <div v-for="file in attachments" :key="file.id" class="attachment-chip">
        <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <polyline points="14 2 14 8 20 8" />
        </svg>
        <span class="attachment-chip-name">{{ file.name }}</span>
        <button class="attachment-chip-remove" @click="removeAttachment(file.id)">
          <svg viewBox="0 0 24 24" width="10" height="10" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
    </div>

    <div v-if="dragOver" class="input-box-drop-overlay">
      <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
        <polyline points="17 8 12 3 7 8" />
        <line x1="12" y1="3" x2="12" y2="15" />
      </svg>
      <span>拖放文件到这里</span>
    </div>

    <div class="input-box-inner">
      <button class="input-box-attach" @click="triggerFileInput" title="上传文件">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
        </svg>
      </button>

      <textarea
        ref="textareaRef"
        class="input-box-textarea selectable"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        rows="1"
        @input="handleInput"
        @keydown="handleKeydown"
        @focus="focused = true"
        @blur="focused = false"
      />

      <button
        v-if="modelValue || attachments.length > 0"
        class="input-box-send"
        :disabled="disabled"
        @click="handleSend"
        title="发送"
      >
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="22" y1="2" x2="11" y2="13" />
          <polygon points="22 2 15 22 11 13 2 9 22 2" />
        </svg>
      </button>

      <button
        v-else-if="disabled"
        class="input-box-stop"
        @click="$emit('stop')"
        title="停止生成"
      >
        <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
          <rect x="6" y="6" width="12" height="12" rx="2" />
        </svg>
      </button>
    </div>

    <input
      ref="fileInputRef"
      type="file"
      multiple
      hidden
      @change="handleFileSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import type { Attachment } from '../../types'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
}>(), {
  placeholder: '输入消息... (Enter 发送, Shift+Enter 换行)',
  disabled: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  send: [content: string, attachments: Attachment[]]
  stop: []
  upload: [files: FileList]
}>()

const textareaRef = ref<HTMLTextAreaElement>()
const fileInputRef = ref<HTMLInputElement>()
const focused = ref(false)
const dragOver = ref(false)
const attachments = ref<Attachment[]>([])

function handleInput(e: Event) {
  const target = e.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
  autoResize()
}

function autoResize() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 200) + 'px'
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

function handleSend() {
  if (!props.modelValue.trim() && attachments.value.length === 0) return
  emit('send', props.modelValue, [...attachments.value])
  emit('update:modelValue', '')
  attachments.value = []
  nextTick(() => {
    if (textareaRef.value) {
      textareaRef.value.style.height = 'auto'
    }
  })
}

function triggerFileInput() {
  fileInputRef.value?.click()
}

function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files) {
    addFiles(input.files)
    input.value = ''
  }
}

function handleDrop(e: DragEvent) {
  dragOver.value = false
  if (e.dataTransfer?.files) {
    addFiles(e.dataTransfer.files)
  }
}

function addFiles(files: FileList) {
  for (const file of Array.from(files)) {
    const isImage = file.type.startsWith('image/')
    attachments.value.push({
      id: `att-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      type: isImage ? 'image' : 'file',
      name: file.name,
      size: file.size,
      mimeType: file.type,
    })
  }
}

function removeAttachment(id: string) {
  attachments.value = attachments.value.filter(a => a.id !== id)
}
</script>

<style scoped>
.input-box {
  position: relative;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.input-box.is-focused {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.input-box.is-disabled {
  opacity: 0.5;
}

.input-box-attachments {
  display: flex;
  gap: var(--spacing-1);
  padding: var(--spacing-2) var(--spacing-3) 0;
  flex-wrap: wrap;
}

.attachment-chip {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 24px;
  padding: 0 var(--spacing-2);
  background: var(--bg-active);
  border-radius: var(--radius-sm);
  font-size: var(--text-caption);
  color: var(--text-secondary);
}

.attachment-chip-name {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attachment-chip-remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  border-radius: 50%;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.attachment-chip-remove:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.input-box-drop-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  background: rgba(139, 92, 246, 0.1);
  border: 2px dashed var(--primary);
  border-radius: var(--radius-lg);
  color: var(--primary);
  font-size: var(--text-body-sm);
  z-index: 10;
}

.input-box-inner {
  display: flex;
  align-items: flex-end;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-3);
}

.input-box-attach,
.input-box-send,
.input-box-stop {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  flex-shrink: 0;
  transition: all var(--transition-fast);
}

.input-box-attach {
  background: transparent;
  color: var(--text-tertiary);
}

.input-box-attach:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.input-box-send {
  background: var(--primary);
  color: white;
}

.input-box-send:hover:not(:disabled) {
  background: var(--primary-hover);
}

.input-box-send:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.input-box-stop {
  background: var(--error);
  color: white;
}

.input-box-stop:hover {
  opacity: 0.9;
}

.input-box-textarea {
  flex: 1;
  min-height: 24px;
  max-height: 200px;
  padding: var(--spacing-1) 0;
  background: transparent;
  color: var(--text-primary);
  font-size: var(--text-body);
  font-family: var(--font-family-sans);
  line-height: 1.5;
  resize: none;
  outline: none;
}

.input-box-textarea::placeholder {
  color: var(--text-tertiary);
}
</style>
