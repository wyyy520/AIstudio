<template>
  <div class="announcement-card">
    <div class="card-header">
      <h3 class="card-title">最新公告</h3>
      <span class="version-tag">v0.1.0</span>
    </div>
    <div v-if="announcement" class="announcement-content">
      <ul class="update-list">
        <li v-for="(item, idx) in announcement.updates" :key="idx" class="update-item">
          <span class="update-bullet">•</span>
          <span class="update-text">{{ item }}</span>
        </li>
      </ul>
    </div>
    <div v-else class="welcome-content">
      <p class="welcome-text">欢迎使用 AIStudio 👋</p>
      <p class="welcome-hint">智能工作流开发平台，通过可视化拖拽编排 AI 能力节点，实现从数据处理、模型推理到自动化部署的全流程开发。</p>
    </div>
    <div class="card-footer">
      <a href="#" class="card-link">查看全部更新日志 →</a>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Announcement {
  version: string
  updates: string[]
}

const announcement: Announcement | null = {
  version: 'v0.1.0',
  updates: [
    '🎨 完成基础 UI 框架和设计系统',
    '⚡ 支持响应式布局，适配不同窗口',
    '🔧 新增核心公共组件（按钮/输入框/开关等）',
  ],
}
</script>

<style scoped>
.announcement-card {
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-4);
}

.card-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  line-height: var(--leading-h3);
}

.version-tag {
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  background: var(--primary-bg);
  color: var(--primary);
  font-size: var(--text-caption);
  font-weight: var(--font-regular);
  line-height: var(--leading-caption);
}

/* ===== 公告内容 ===== */
.announcement-content {
  flex: 1;
}

.update-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.update-item {
  display: flex;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-2);
  font-size: var(--text-body-sm);
  line-height: var(--leading-body-sm);
  color: var(--text-secondary);
}

.update-bullet {
  flex-shrink: 0;
  color: var(--primary);
}

/* ===== 欢迎内容 ===== */
.welcome-content {
  flex: 1;
}

.welcome-text {
  font-size: var(--text-body);
  color: var(--text-primary);
  margin-bottom: var(--spacing-2);
}

.welcome-hint {
  font-size: var(--text-body-sm);
  line-height: var(--leading-body-sm);
  color: var(--text-secondary);
}

/* ===== 页脚 ===== */
.card-footer {
  margin-top: var(--spacing-4);
  padding-top: var(--spacing-4);
  border-top: 1px solid var(--border-subtle);
}

.card-link {
  color: var(--primary);
  font-size: var(--text-body-sm);
  text-decoration: none;
}

.card-link:hover {
  color: var(--primary-hover);
  text-decoration: underline;
}
</style>