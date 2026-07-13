<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="app-modal-overlay" @click.self="handleOverlayClick">
        <div class="app-modal" :style="{ width }">
          <div class="app-modal__header">
            <h3 class="app-modal__title">{{ title }}</h3>
            <button class="app-modal__close" @click="close">✕</button>
          </div>
          <div class="app-modal__body">
            <slot />
          </div>
          <div v-if="$slots.footer" class="app-modal__footer">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean
  title: string
  width?: string
  closeOnOverlay?: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

function close() {
  emit('update:visible', false)
}

function handleOverlayClick() {
  if (true) close()
}
</script>

<style scoped>
.app-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
}

.app-modal {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl);
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.app-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.app-modal__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.app-modal__close {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-tag);
  color: var(--text-tertiary);
  transition: all var(--transition-fast);
}

.app-modal__close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.app-modal__body {
  padding: 20px 24px;
  overflow-y: auto;
  flex: 1;
}

.app-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 24px;
  border-top: 1px solid var(--border-subtle);
}

.modal-enter-active { transition: opacity 200ms ease; }
.modal-enter-active .app-modal { transition: transform 250ms cubic-bezier(0.4, 0, 0.2, 1), opacity 250ms ease; }
.modal-leave-active { transition: opacity 150ms ease; }
.modal-leave-active .app-modal { transition: transform 200ms ease, opacity 200ms ease; }
.modal-enter-from { opacity: 0; }
.modal-enter-from .app-modal { transform: scale(0.95); opacity: 0; }
.modal-leave-to { opacity: 0; }
.modal-leave-to .app-modal { transform: scale(0.95); opacity: 0; }
</style>