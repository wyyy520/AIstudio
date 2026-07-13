<template>
  <button
    :class="switchClasses"
    type="button"
    role="switch"
    :aria-checked="modelValue"
    :disabled="disabled"
    @click="toggle"
  >
    <span class="switch-dot">
      <svg
        v-if="loading"
        class="switch-loading"
        viewBox="0 0 24 24"
        width="10"
        height="10"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <path d="M21 12a9 9 0 1 1-6.219-8.56" />
      </svg>
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type SwitchSize = 'default' | 'small'

interface Props {
  modelValue: boolean
  size?: SwitchSize
  disabled?: boolean
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  size: 'default',
  disabled: false,
  loading: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  change: [value: boolean]
}>()

const switchClasses = computed(() => {
  return [
    'app-switch',
    `app-switch--${props.size}`,
    {
      'is-active': props.modelValue,
      'is-disabled': props.disabled,
      'is-loading': props.loading,
    },
  ]
})

function toggle(): void {
  if (props.disabled || props.loading) return
  const newValue = !props.modelValue
  emit('update:modelValue', newValue)
  emit('change', newValue)
}
</script>

<style scoped>
.app-switch {
  position: relative;
  display: inline-flex;
  align-items: center;
  border: none;
  outline: none;
  cursor: pointer;
  padding: 0;
  background: var(--bg-hover);
  border-radius: 999px;
  transition: background var(--transition-fast);
  flex-shrink: 0;
}

.app-switch.is-active {
  background: var(--primary);
}

.app-switch.is-disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* ===== 尺寸 ===== */
.app-switch--default {
  width: 40px;
  height: 22px;
}

.app-switch--small {
  width: 32px;
  height: 18px;
}

/* ===== 圆点 ===== */
.switch-dot {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border-radius: 50%;
  transition: transform var(--transition-fast), background var(--transition-fast);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.app-switch--default .switch-dot {
  width: 18px;
  height: 18px;
  left: 2px;
}

.app-switch--default.is-active .switch-dot {
  transform: translateX(18px);
}

.app-switch--small .switch-dot {
  width: 14px;
  height: 14px;
  left: 2px;
}

.app-switch--small.is-active .switch-dot {
  transform: translateX(14px);
}

/* ===== 加载动画 ===== */
.switch-loading {
  animation: dot-spin 1s linear infinite;
}

@keyframes dot-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>