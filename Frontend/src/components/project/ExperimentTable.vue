<template>
  <div class="experiment-table">
    <div class="table-header">
      <h2 class="section-title">Experiments</h2>
      <div class="table-actions">
        <button class="action-btn" @click="$emit('compare')">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M18 20V10M12 20V4M6 20v-6" />
          </svg>
          对比实验
        </button>
      </div>
    </div>

    <div v-if="experiments.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M9 3h6v11l-3 3-3-3V3z" />
      </svg>
      <span>暂无实验记录</span>
      <span class="empty-hint">训练模型后将记录实验数据</span>
    </div>

    <div v-else class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th class="sortable" @click="toggleSort('id')">
              Experiment ID
              <span v-if="sortKey === 'id'" class="sort-icon">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th>Model</th>
            <th class="sortable" @click="toggleSort('epoch')">
              Epoch
              <span v-if="sortKey === 'epoch'" class="sort-icon">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="sortable" @click="toggleSort('loss')">
              Loss
              <span v-if="sortKey === 'loss'" class="sort-icon">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th class="sortable" @click="toggleSort('accuracy')">
              Accuracy
              <span v-if="sortKey === 'accuracy'" class="sort-icon">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
            </th>
            <th>GPU</th>
            <th>Time</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="exp in sortedExperiments"
            :key="exp.id"
            :class="{ 'row-active': selectedId === exp.id }"
            @click="selectedId = exp.id"
          >
            <td class="cell-id">{{ exp.id }}</td>
            <td>{{ exp.modelName }}</td>
            <td>{{ exp.epoch }}</td>
            <td>{{ exp.loss.toFixed(4) }}</td>
            <td class="cell-accuracy">{{ (exp.accuracy * 100).toFixed(1) }}%</td>
            <td>{{ exp.gpu }}</td>
            <td>{{ exp.duration }}</td>
            <td>
              <span class="status-badge" :class="`status-${exp.status}`">{{ statusLabel(exp.status) }}</span>
            </td>
            <td>
              <button class="table-btn" @click.stop="$emit('detail', exp)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5">
                  <circle cx="12" cy="12" r="10" />
                  <path d="M12 16v-4M12 8h.01" />
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { ProjectExperiment } from '@/types/project'

const props = defineProps<{
  experiments: ProjectExperiment[]
}>()

defineEmits<{
  compare: []
  detail: [experiment: ProjectExperiment]
}>()

const sortKey = ref<'id' | 'epoch' | 'loss' | 'accuracy'>('id')
const sortDir = ref<'asc' | 'desc'>('desc')
const selectedId = ref<string>('')

function toggleSort(key: 'id' | 'epoch' | 'loss' | 'accuracy') {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'desc'
  }
}

const sortedExperiments = computed(() => {
  return [...props.experiments].sort((a, b) => {
    const aVal = a[sortKey.value]
    const bVal = b[sortKey.value]
    if (typeof aVal === 'string') {
      return sortDir.value === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal)
    }
    return sortDir.value === 'asc' ? (aVal as number) - (bVal as number) : (bVal as number) - (aVal as number)
  })
})

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '等待中',
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
  }
  return map[status] || status
}
</script>

<style scoped>
.experiment-table {
  padding: var(--spacing-6);
  height: 100%;
  overflow-y: auto;
}

.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-4);
}

.section-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.table-actions {
  display: flex;
  gap: var(--spacing-2);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
  height: 32px;
  padding: 0 var(--spacing-3);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-strong);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-3);
  height: 300px;
  color: var(--text-tertiary);
}

.empty-state span {
  font-size: var(--text-body);
}

.empty-hint {
  font-size: var(--text-caption);
  color: var(--text-disabled);
}

.table-container {
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: var(--spacing-3) var(--spacing-4);
  text-align: left;
  font-size: var(--text-caption);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.data-table th.sortable {
  cursor: pointer;
  user-select: none;
}

.data-table th.sortable:hover {
  color: var(--text-primary);
}

.sort-icon {
  margin-left: var(--spacing-1);
  color: var(--primary);
}

.data-table td {
  padding: var(--spacing-3) var(--spacing-4);
  font-size: var(--text-body-sm);
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-subtle);
}

.data-table tr:last-child td {
  border-bottom: none;
}

.data-table tr:hover {
  background: var(--bg-hover);
}

.data-table tr.row-active {
  background: var(--primary-bg);
}

.cell-id {
  font-family: var(--font-family-mono);
  font-size: var(--text-caption);
  color: var(--text-tertiary);
}

.cell-accuracy {
  color: var(--success);
  font-weight: var(--font-semibold);
}

.status-badge {
  display: inline-flex;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-size: var(--text-caption);
}

.status-badge.status-completed {
  background: var(--success-bg);
  color: var(--success);
}

.status-badge.status-running {
  background: var(--info-bg);
  color: var(--info);
}

.status-badge.status-failed {
  background: var(--error-bg);
  color: var(--error);
}

.status-badge.status-pending {
  background: var(--warning-bg);
  color: var(--warning);
}

.table-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  background: transparent;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.table-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
</style>
