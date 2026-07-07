<template>
  <span :class="tagClasses">
    <span class="tag-text">
      <slot />
    </span>
    <button
      v-if="closable"
      class="tag-close"
      type="button"
      @click="handleClose"
      title="移除"
    >
      <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M18 6 6 18" />
        <path d="m6 6 12 12" />
      </svg>
    </button>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type TagSize = 'default' | 'small'
type TagColor = 'default' | 'primary' | 'success' | 'warning' | 'error' | 'info'

interface Props {
  size?: TagSize
  color?: TagColor
  closable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'default',
  color: 'default',
  closable: false,
})

const emit = defineEmits<{
  close: []
}>()

const tagClasses = computed(() => {
  return [
    'app-tag',
    `app-tag--${props.size}`,
    `app-tag--${props.color}`,
    {
      'is-closable': props.closable,
    },
  ]
})

function handleClose(): void {
  emit('close')
}
</script>

<style scoped>
.app-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border-radius: var(--radius-sm);
  font-family: var(--font-family-sans);
  font-weight: var(--font-regular);
  white-space: nowrap;
  flex-shrink: 0;
}

/* ===== 尺寸 ===== */
.app-tag--default {
  height: 24px;
  padding: 0 10px;
  font-size: 12px;
  line-height: 24px;
}

.app-tag--small {
  height: 20px;
  padding: 0 8px;
  font-size: 11px;
  line-height: 20px;
}

/* ===== 颜色 ===== */
.app-tag--default {
  background: var(--bg-active);
  color: var(--text-secondary);
}

.app-tag--primary {
  background: var(--primary-bg);
  color: var(--primary);
  border: 1px solid rgba(139, 92, 246, 0.2);
}

.app-tag--success {
  background: var(--success-bg);
  color: var(--success);
  border: 1px solid rgba(34, 197, 94, 0.2);
}

.app-tag--warning {
  background: var(--warning-bg);
  color: var(--warning);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.app-tag--error {
  background: var(--error-bg);
  color: var(--error);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.app-tag--info {
  background: var(--info-bg);
  color: var(--info);
  border: 1px solid rgba(59, 130, 246, 0.2);
}

/* ===== 关闭按钮 ===== */
.tag-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border: none;
  background: transparent;
  color: inherit;
  opacity: 0.6;
  cursor: pointer;
  border-radius: 3px;
  padding: 0;
  transition: opacity var(--transition-fast), background var(--transition-fast);
}

.tag-close:hover {
  opacity: 1;
  background: rgba(255, 255, 255, 0.1);
}

.tag-text {
  line-height: 1;
}
</style>