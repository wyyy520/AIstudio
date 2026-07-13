<template>
  <button
    :class="classes"
    :disabled="disabled || loading"
    :type="nativeType"
    @click="handleClick"
  >
    <div class="button-content">
      <svg
        v-if="loading"
        class="button-icon button-icon-loading"
        viewBox="0 0 24 24"
        :width="iconSize"
        :height="iconSize"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M21 12a9 9 0 1 1-6.219-8.56" />
      </svg>
      <svg
        v-else-if="iconLeft"
        class="button-icon"
        :width="iconSize"
        :height="iconSize"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path :d="iconLeft" />
      </svg>
      <span v-if="$slots.default || label" class="button-text">
        {{ label }}
        <slot />
      </span>
      <svg
        v-if="iconRight"
        class="button-icon"
        :width="iconSize"
        :height="iconSize"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path :d="iconRight" />
      </svg>
    </div>
  </button>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue'

type ButtonSize = 'large' | 'medium' | 'small' | 'mini'
type ButtonType = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger' | 'text'

interface Props {
  size?: ButtonSize
  type?: ButtonType
  label?: string
  nativeType?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
  iconLeft?: string
  iconRight?: string
  block?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'medium',
  type: 'secondary',
  nativeType: 'button',
  disabled: false,
  loading: false,
  block: false,
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const classes = computed(() => {
  return [
    'app-button',
    `app-button--${props.size}`,
    `app-button--${props.type}`,
    {
      'is-loading': props.loading,
      'is-disabled': props.disabled,
      'is-block': props.block,
    },
  ]
})

const iconSize = computed(() => {
  const map: Record<ButtonSize, number> = {
    large: 18,
    medium: 16,
    small: 14,
    mini: 12,
  }
  return map[props.size]
})

function handleClick(event: MouseEvent): void {
  if (props.loading || props.disabled) {
    event.preventDefault()
    return
  }
  emit('click', event)
}
</script>

<style scoped>
.app-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  outline: none;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-family-sans);
  font-weight: var(--font-regular);
  position: relative;
  flex-shrink: 0;
  white-space: nowrap;
}

.app-button.is-block {
  display: flex;
  width: 100%;
}

.button-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.button-text {
  line-height: 1;
}

.button-icon {
  flex-shrink: 0;
}

.button-icon-loading {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* ===== 尺寸 ===== */
.app-button--large {
  height: 44px;
  padding: 0 24px;
  border-radius: 12px;
  font-size: 15px;
}

.app-button--medium {
  height: 36px;
  padding: 0 18px;
  border-radius: 10px;
  font-size: 14px;
}

.app-button--small {
  height: 28px;
  padding: 0 12px;
  border-radius: 8px;
  font-size: 13px;
}

.app-button--mini {
  height: 22px;
  padding: 0 8px;
  border-radius: 6px;
  font-size: 12px;
}

/* ===== 类型 ===== */
/* Primary */
.app-button--primary {
  background: var(--primary);
  color: white;
}

.app-button--primary:hover:not(.is-disabled):not(.is-loading) {
  background: var(--primary-hover);
}

.app-button--primary:active:not(.is-disabled):not(.is-loading) {
  background: var(--primary-active);
  transform: scale(0.97);
}

.app-button--primary:disabled,
.app-button--primary.is-disabled,
.app-button--primary.is-loading {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Secondary */
.app-button--secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.app-button--secondary:hover:not(.is-disabled):not(.is-loading) {
  background: var(--bg-hover);
}

.app-button--secondary:active:not(.is-disabled):not(.is-loading) {
  background: var(--bg-active);
  transform: scale(0.97);
}

.app-button--secondary:disabled,
.app-button--secondary.is-disabled,
.app-button--secondary.is-loading {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Outline */
.app-button--outline {
  background: transparent;
  color: var(--text-primary);
  border: 1px solid var(--border-default);
}

.app-button--outline:hover:not(.is-disabled):not(.is-loading) {
  border-color: var(--border-strong);
  background: var(--bg-hover);
}

.app-button--outline:active:not(.is-disabled):not(.is-loading) {
  background: var(--bg-active);
  transform: scale(0.97);
}

.app-button--outline:disabled,
.app-button--outline.is-disabled,
.app-button--outline.is-loading {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Ghost */
.app-button--ghost {
  background: transparent;
  color: var(--text-primary);
}

.app-button--ghost:hover:not(.is-disabled):not(.is-loading) {
  background: var(--bg-hover);
}

.app-button--ghost:active:not(.is-disabled):not(.is-loading) {
  background: var(--bg-active);
  transform: scale(0.97);
}

.app-button--ghost:disabled,
.app-button--ghost.is-disabled,
.app-button--ghost.is-loading {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Danger */
.app-button--danger {
  background: var(--error);
  color: white;
}

.app-button--danger:hover:not(.is-disabled):not(.is-loading) {
  opacity: 0.9;
}

.app-button--danger:active:not(.is-disabled):not(.is-loading) {
  opacity: 1;
  transform: scale(0.97);
}

.app-button--danger:disabled,
.app-button--danger.is-disabled,
.app-button--danger.is-loading {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Text */
.app-button--text {
  background: transparent;
  color: var(--primary);
}

.app-button--text:hover:not(.is-disabled):not(.is-loading) {
  background: var(--primary-bg);
}

.app-button--text:active:not(.is-disabled):not(.is-loading) {
  opacity: 0.8;
  transform: scale(0.97);
}

.app-button--text:disabled,
.app-button--text.is-disabled,
.app-button--text.is-loading {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>