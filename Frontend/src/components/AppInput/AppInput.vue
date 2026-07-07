<template>
  <div :class="wrapperClasses">
    <div class="input-inner">
      <div v-if="$slots.prefix || prefixIcon" class="input-prefix">
        <svg
          v-if="prefixIcon"
          class="input-icon"
          viewBox="0 0 24 24"
          width="16"
          height="16"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path :d="prefixIcon" />
        </svg>
        <slot v-else name="prefix" />
      </div>

      <input
        v-if="type !== 'textarea'"
        :ref="inputRef"
        :type="inputType"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :autocomplete="autocomplete"
        :maxlength="maxlength"
        :autofocus="autofocus"
        @input="handleInput"
        @change="handleChange"
        @focus="handleFocus"
        @blur="handleBlur"
        :class="inputClasses"
      />

      <textarea
        v-else
        :ref="inputRef"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :autocomplete="autocomplete"
        :maxlength="maxlength"
        :rows="rows"
        @input="handleInput"
        @change="handleChange"
        @focus="handleFocus"
        @blur="handleBlur"
        :class="inputClasses"
      ></textarea>

      <div v-if="$slots.suffix || suffixIcon || clearable" class="input-suffix">
        <slot name="suffix" />
        <svg
          v-if="suffixIcon && !clearable"
          class="input-icon"
          viewBox="0 0 24 24"
          width="16"
          height="16"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path :d="suffixIcon" />
        </svg>
        <button
          v-if="clearable && modelValue"
          class="input-clear-btn"
          type="button"
          @click="clearValue"
          title="清除"
        >
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="10" />
            <path d="m15 9-6 6" />
            <path d="m9 9 6 6" />
          </svg>
        </button>
        <button
          v-if="type === 'password'"
          class="input-toggle-btn"
          type="button"
          @click="togglePassword"
          :title="showPassword ? '隐藏密码' : '显示密码'"
        >
          <svg
            v-if="showPassword"
            viewBox="0 0 24 24"
            width="16"
            height="16"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
            <circle cx="12" cy="12" r="3" />
          </svg>
          <svg
            v-else
            viewBox="0 0 24 24"
            width="16"
            height="16"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
            <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
            <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.62 9.62 0 0 0 4.29-.89" />
            <path d="M4 4l16 16" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, type PropType } from 'vue'

type InputSize = 'large' | 'medium' | 'small'

interface Props {
  modelValue: string | number
  type?: 'text' | 'password' | 'email' | 'number' | 'tel' | 'url' | 'textarea'
  size?: InputSize
  placeholder?: string
  disabled?: boolean
  readonly?: boolean
  error?: boolean
  clearable?: boolean
  prefixIcon?: string
  suffixIcon?: string
  autocomplete?: string
  maxlength?: number
  autofocus?: boolean
  rows?: number
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  type: 'text',
  size: 'medium',
  placeholder: '',
  disabled: false,
  readonly: false,
  error: false,
  clearable: false,
  rows: 4,
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  input: [value: string]
  change: [value: string]
  focus: []
  blur: []
  clear: []
}>()

const inputRef = ref<HTMLInputElement | HTMLTextAreaElement>()
const showPassword = ref(false)

const inputType = computed(() => {
  if (props.type === 'password' && !showPassword.value) {
    return 'password'
  }
  return 'text'
})

const wrapperClasses = computed(() => {
  return [
    'app-input-wrapper',
    `app-input--${props.size}`,
    {
      'is-disabled': props.disabled,
      'is-readonly': props.readonly,
      'is-error': props.error,
    },
  ]
})

const inputClasses = computed(() => {
  return [
    'app-input',
    {
      'has-prefix': props.prefixIcon || !!$slots.prefix,
      'has-suffix': props.suffixIcon || props.clearable || props.type === 'password' || !!$slots.suffix,
    },
  ]
})

function handleInput(e: Event): void {
  const target = e.target as HTMLInputElement | HTMLTextAreaElement
  emit('update:modelValue', target.value)
  emit('input', target.value)
}

function handleChange(e: Event): void {
  const target = e.target as HTMLInputElement | HTMLTextAreaElement
  emit('change', target.value)
}

function handleFocus(): void {
  emit('focus')
}

function handleBlur(): void {
  emit('blur')
}

function clearValue(): void {
  emit('update:modelValue', '')
  emit('input', '')
  emit('clear')
  inputRef.value?.focus()
}

function togglePassword(): void {
  showPassword.value = !showPassword.value
}
</script>

<style scoped>
.app-input-wrapper {
  width: 100%;
  position: relative;
}

.input-inner {
  display: flex;
  align-items: center;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.input-inner:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.app-input-wrapper.is-error .input-inner {
  border-color: var(--error);
  box-shadow: 0 0 0 3px var(--error-bg);
}

.app-input-wrapper.is-disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* ===== 尺寸 ===== */
.app-input--large .input-inner {
  height: 44px;
  padding: 12px 16px;
  border-radius: 10px;
}

.app-input--medium .input-inner {
  height: 36px;
  padding: 10px 12px;
  border-radius: 8px;
}

.app-input--small .input-inner {
  height: 28px;
  padding: 8px 10px;
  border-radius: 6px;
}

/* ===== input / textarea ===== */
.app-input {
  flex: 1;
  width: 100%;
  background: transparent;
  color: var(--text-primary);
  font-family: var(--font-family-sans);
  font-size: var(--text-body);
  line-height: var(--leading-body);
}

.app-input::placeholder {
  color: var(--text-tertiary);
}

textarea.app-input {
  height: auto;
  min-height: 80px;
  resize: vertical;
  line-height: 1.5;
}

/* ===== 前缀/后缀 ===== */
.input-prefix,
.input-suffix {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.app-input--large .input-prefix,
.app-input--large .input-suffix {
  gap: 8px;
}

.app-input--small .input-prefix,
.app-input--small .input-suffix {
  gap: 4px;
}

.input-prefix {
  margin-right: var(--spacing-2);
}

.input-suffix {
  margin-left: var(--spacing-2);
}

.app-input--large .input-prefix {
  margin-right: 10px;
}

.app-input--large .input-suffix {
  margin-left: 10px;
}

.app-input--small .input-prefix {
  margin-right: 4px;
}

.app-input--small .input-suffix {
  margin-left: 4px;
}

.input-icon {
  flex-shrink: 0;
  opacity: 0.7;
}

.input-clear-btn,
.input-toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  border-radius: 4px;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.input-clear-btn:hover,
.input-toggle-btn:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}
</style>