<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="visible" class="history-drawer-overlay" @click.self="$emit('close')">
        <div class="history-drawer">
          <div class="history-drawer-header">
            <h3 class="history-drawer-title">历史记录</h3>
            <button class="history-drawer-close" @click="$emit('close')">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>

          <div class="history-drawer-search">
            <svg class="history-search-icon" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <input
              class="history-search-input"
              v-model="searchQuery"
              placeholder="搜索对话..."
            />
          </div>

          <div class="history-drawer-list">
            <div v-if="filteredConversations.length === 0" class="history-empty">
              暂无对话记录
            </div>
            <template v-else>
              <div
                v-for="group in groupedConversations"
                :key="group.label"
                class="history-group"
              >
                <div class="history-group-label">{{ group.label }}</div>
                <button
                  v-for="conv in group.items"
                  :key="conv.id"
                  class="history-item"
                  :class="{ active: conv.id === selectedId }"
                  @click="$emit('select', conv.id)"
                >
                  <div class="history-item-content">
                    <div class="history-item-title">
                      <svg v-if="conv.isFavorite" class="history-item-star" viewBox="0 0 24 24" width="12" height="12" fill="currentColor" stroke="currentColor" stroke-width="1.5">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                      </svg>
                      {{ conv.title }}
                    </div>
                    <div class="history-item-meta">
                      <span>{{ conv.messageCount }} 条消息</span>
                      <span>·</span>
                      <span>{{ conv.model }}</span>
                    </div>
                  </div>
                  <div class="history-item-actions">
                    <button
                      class="history-item-action"
                      :class="{ favorited: conv.isFavorite }"
                      @click.stop="$emit('favorite', conv.id)"
                      title="收藏"
                    >
                      <svg viewBox="0 0 24 24" width="14" height="14" :fill="conv.isFavorite ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                      </svg>
                    </button>
                    <button
                      class="history-item-action history-item-action--danger"
                      @click.stop="$emit('delete', conv.id)"
                      title="删除"
                    >
                      <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <polyline points="3 6 5 6 21 6" />
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                      </svg>
                    </button>
                  </div>
                </button>
              </div>
            </template>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Conversation } from '../../types'

const props = defineProps<{
  conversations: Conversation[]
  visible: boolean
  selectedId?: string
}>()

defineEmits<{
  close: []
  select: [id: string]
  delete: [id: string]
  favorite: [id: string]
}>()

const searchQuery = ref('')

const filteredConversations = computed(() => {
  if (!searchQuery.value) return props.conversations
  const q = searchQuery.value.toLowerCase()
  return props.conversations.filter(c => c.title.toLowerCase().includes(q))
})

const groupedConversations = computed(() => {
  const now = Date.now()
  const day = 86400000
  const groups: { label: string; items: Conversation[] }[] = []

  const today: Conversation[] = []
  const yesterday: Conversation[] = []
  const older: Conversation[] = []

  for (const conv of filteredConversations.value) {
    const diff = now - conv.updatedAt
    if (diff < day) today.push(conv)
    else if (diff < day * 2) yesterday.push(conv)
    else older.push(conv)
  }

  if (today.length) groups.push({ label: '今天', items: today })
  if (yesterday.length) groups.push({ label: '昨天', items: yesterday })
  if (older.length) groups.push({ label: '更早', items: older })

  return groups
})
</script>

<style scoped>
.history-drawer-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 900;
}

.history-drawer {
  position: fixed;
  top: 0;
  left: var(--sidebar-width);
  bottom: var(--statusbar-height);
  width: 320px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-subtle);
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
}

.history-drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3) var(--spacing-3) var(--spacing-2);
  flex-shrink: 0;
}

.history-drawer-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.history-drawer-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.history-drawer-close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.history-drawer-search {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin: 0 var(--spacing-3) var(--spacing-2);
  padding: 0 var(--spacing-2);
  height: 32px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  transition: border-color var(--transition-fast);
}

.history-drawer-search:focus-within {
  border-color: var(--primary);
}

.history-search-icon {
  flex-shrink: 0;
  color: var(--text-tertiary);
}

.history-search-input {
  flex: 1;
  height: 100%;
  background: transparent;
  color: var(--text-primary);
  font-size: var(--text-body-sm);
  font-family: var(--font-family-sans);
  outline: none;
}

.history-search-input::placeholder {
  color: var(--text-disabled);
}

.history-drawer-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 var(--spacing-2);
}

.history-empty {
  padding: var(--spacing-8) var(--spacing-4);
  text-align: center;
  color: var(--text-disabled);
  font-size: var(--text-body-sm);
}

.history-group {
  margin-bottom: var(--spacing-3);
}

.history-group-label {
  padding: var(--spacing-1) var(--spacing-2);
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  font-weight: var(--font-semibold);
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: var(--spacing-2);
  border: none;
  background: transparent;
  text-align: left;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background var(--transition-fast);
  font-family: var(--font-family-sans);
}

.history-item:hover {
  background: var(--bg-hover);
}

.history-item.active {
  background: var(--bg-active);
}

.history-item-content {
  flex: 1;
  min-width: 0;
}

.history-item-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.history-item-star {
  color: var(--warning);
  flex-shrink: 0;
}

.history-item-meta {
  display: flex;
  gap: var(--spacing-1);
  margin-top: 2px;
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.history-item-actions {
  display: flex;
  gap: var(--spacing-1);
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.history-item:hover .history-item-actions {
  opacity: 1;
}

.history-item-action {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.history-item-action:hover {
  background: var(--bg-active);
  color: var(--text-primary);
}

.history-item-action.favorited {
  color: var(--warning);
}

.history-item-action--danger:hover {
  color: var(--error);
}

/* Drawer transition */
.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 200ms ease;
}
.drawer-enter-active .history-drawer,
.drawer-leave-active .history-drawer {
  transition: transform 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}
.drawer-enter-from .history-drawer {
  transform: translateX(-100%);
}
.drawer-leave-to .history-drawer {
  transform: translateX(-100%);
}
</style>
