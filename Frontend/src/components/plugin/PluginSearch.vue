<template>
  <div class="plugin-search">
    <svg
      class="plugin-search-icon"
      viewBox="0 0 24 24"
      width="16"
      height="16"
      fill="none"
      stroke="currentColor"
      stroke-width="1.5"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <circle cx="11" cy="11" r="8" />
      <path d="M21 21l-4.35-4.35" />
    </svg>
    <input
      class="plugin-search-input"
      type="text"
      :value="modelValue"
      :placeholder="placeholder"
      @input="handleInput"
      @keydown.escape="handleClear"
    />
    <button
      v-if="modelValue"
      class="plugin-search-clear"
      type="button"
      @click="handleClear"
      title="Clear"
    >
      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
        <circle cx="12" cy="12" r="10" />
        <path d="m15 9-6 6" />
        <path d="m9 9 6 6" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
interface Props {
  modelValue: string
  placeholder?: string
}

withDefaults(defineProps<Props>(), {
  placeholder: 'Search plugins...',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  search: [query: string]
  clear: []
}>()

function handleInput(e: Event) {
  const value = (e.target as HTMLInputElement).value
  emit('update:modelValue', value)
  emit('search', value)
}

function handleClear() {
  emit('update:modelValue', '')
  emit('search', '')
  emit('clear')
}
</script>

<style scoped>
.plugin-search {
  display: flex;
  align-items: center;
  height: 36px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  padding: 0 12px;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.plugin-search:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.plugin-search-icon {
  flex-shrink: 0;
  color: var(--text-tertiary);
  margin-right: 8px;
}

.plugin-search-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  line-height: var(--leading-body-sm);
}

.plugin-search-input::placeholder {
  color: var(--text-tertiary);
}

.plugin-search-clear {
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

.plugin-search-clear:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}
</style>