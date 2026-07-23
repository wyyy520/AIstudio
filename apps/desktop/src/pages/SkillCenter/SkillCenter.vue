<template>
  <div class="skill-page">
    <div class="page-header">
      <h1>AI 技能中心</h1>
      <p class="subtitle">工作流规划、智能诊断、参数优化与环境管理</p>
    </div>

    <div class="skill-grid">
      <!-- 技能分类 -->
      <div class="skill-categories">
        <div
          v-for="cat in categories"
          :key="cat.key"
          :class="['category-card', { active: activeCategory === cat.key }]"
          @click="activeCategory = cat.key"
        >
          <span class="category-icon">{{ cat.icon }}</span>
          <div class="category-info">
            <span class="category-name">{{ cat.name }}</span>
            <span class="category-count">{{ cat.count }} 个技能</span>
          </div>
        </div>
      </div>

      <!-- 技能列表 -->
      <div class="skill-main">
        <div class="card">
          <h3>{{ currentCategoryName }}</h3>
          <div class="skill-list">
            <div
              v-for="skill in filteredSkills"
              :key="skill.id"
              :class="['skill-card', { 'skill-card--running': skill.status === 'running' }]"
            >
              <div class="skill-header">
                <span class="skill-icon">{{ skill.icon }}</span>
                <div class="skill-info">
                  <span class="skill-name">{{ skill.name }}</span>
                  <span class="skill-desc">{{ skill.description }}</span>
                </div>
                <span v-if="skill.requiresAI" class="ai-badge">AI</span>
                <span
                  :class="['status-dot', skill.status === 'running' ? 'dot-running' : 'dot-idle']"
                />
              </div>
              <div class="skill-actions">
                <button
                  class="btn btn-primary"
                  :disabled="store.isRunning"
                  @click="runSkill(skill.id)"
                >
                  {{ store.currentSkillId === skill.id ? '执行中...' : '执行' }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 执行结果 -->
        <div v-if="store.results.length > 0" class="card">
          <div class="results-header">
            <h3>执行结果</h3>
            <button class="btn btn-secondary btn-sm" @click="store.clearResults()">清除</button>
          </div>
          <div class="results-list">
            <div
              v-for="(result, idx) in store.results"
              :key="idx"
              class="result-card"
            >
              <div class="result-header">
                <span class="result-skill">{{ skillName(result.skillId) }}</span>
                <span class="result-time">{{ formatTime(result.timestamp) }}</span>
              </div>
              <p class="result-content">{{ result.content }}</p>
              <div v-if="result.suggestions && result.suggestions.length > 0" class="result-suggestions">
                <h4>建议操作：</h4>
                <ul>
                  <li v-for="(s, si) in result.suggestions" :key="si">{{ s }}</li>
                </ul>
              </div>
              <div v-if="result.errors && result.errors.length > 0" class="result-errors">
                <h4>发现的问题：</h4>
                <ul>
                  <li v-for="(e, ei) in result.errors" :key="ei">{{ e }}</li>
                </ul>
              </div>
              <div v-if="result.nodes" class="result-nodes">
                <h4>推荐节点：</h4>
                <div class="node-tags">
                  <span v-for="n in result.nodes" :key="n.type" class="node-tag">
                    {{ n.type }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useSkillCenterStore } from '@/stores/skillCenter'

const store = useSkillCenterStore()
const activeCategory = ref<string>('planner')

const categories = computed(() => [
  { key: 'planner', name: '工作流规划', icon: '📋', count: store.plannerSkills.length },
  { key: 'diagnose', name: '智能诊断', icon: '🔍', count: store.diagnoseSkills.length },
  { key: 'explain', name: '知识解释', icon: '💡', count: store.explainSkills.length },
  { key: 'optimize', name: '优化建议', icon: '⚡', count: store.optimizeSkills.length },
  { key: 'environment', name: '环境管理', icon: '🌍', count: store.environmentSkills.length },
  { key: 'automation', name: '自动化', icon: '🤖', count: store.automationSkills.length },
])

const currentCategoryName = computed(() => {
  const cat = categories.value.find(c => c.key === activeCategory.value)
  return cat?.name || '全部技能'
})

const filteredSkills = computed(() => {
  const map: Record<string, any[]> = {
    planner: store.plannerSkills,
    diagnose: store.diagnoseSkills,
    explain: store.explainSkills,
    optimize: store.optimizeSkills,
    environment: store.environmentSkills,
    automation: store.automationSkills,
  }
  return map[activeCategory.value] || []
})

function skillName(id: string) {
  const skill = store.skills.find(s => s.id === id)
  return skill?.name || id
}

function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString()
}

async function runSkill(id: string) {
  await store.runSkill(id)
}
</script>

<style scoped>
.skill-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header h1 { font-size: 22px; font-weight: 700; color: var(--text-primary); margin: 0; }
.subtitle { font-size: 13px; color: var(--text-secondary); margin: 4px 0 0; }

.skill-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 16px;
  min-height: 0;
}

/* Categories */
.skill-categories {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.category-card {
  display: flex; align-items: center; gap: 10px;
  padding: 12px; border-radius: var(--radius-xs);
  background: var(--bg-secondary); border: 2px solid transparent;
  cursor: pointer; transition: all var(--transition-fast);
}
.category-card:hover { border-color: var(--border-hover); }
.category-card.active { border-color: var(--primary); background: var(--bg-active); }
.category-icon { font-size: 22px; }
.category-name { font-size: 13px; font-weight: 600; display: block; color: var(--text-primary); }
.category-count { font-size: 11px; color: var(--text-tertiary); }

/* Skills */
.skill-main { overflow-y: auto; display: flex; flex-direction: column; gap: 12px; }

.card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 16px;
}
.card h3 { font-size: 14px; font-weight: 600; color: var(--text-primary); margin: 0 0 12px; }

.skill-list { display: flex; flex-direction: column; gap: 8px; }

.skill-card {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px; border-radius: var(--radius-xs);
  background: var(--bg-primary); border: 1px solid var(--border-subtle);
  transition: all var(--transition-fast);
}
.skill-card:hover { border-color: var(--border-hover); }
.skill-card--running { border-color: var(--primary); background: var(--bg-active); }
.skill-header { display: flex; align-items: center; gap: 10px; flex: 1; }
.skill-icon { font-size: 24px; }
.skill-info { flex: 1; }
.skill-name { font-size: 14px; font-weight: 600; display: block; }
.skill-desc { font-size: 12px; color: var(--text-tertiary); }
.ai-badge {
  font-size: 10px; padding: 2px 6px; border-radius: 8px;
  background: #e3f2fd; color: #1565c0; font-weight: 600;
}
.status-dot { width: 8px; height: 8px; border-radius: 50%; }
.dot-running { background: #ffbd2e; animation: pulse 1s infinite; }
.dot-idle { background: #27c93f; }

@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }

/* Buttons */
.btn {
  padding: 8px 18px; border-radius: var(--radius-xs);
  font-size: 13px; font-weight: 500; cursor: pointer;
  transition: all var(--transition-fast); border: none;
}
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover:not(:disabled) { opacity: 0.9; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: var(--bg-hover); color: var(--text-primary); }
.btn-sm { padding: 4px 10px; font-size: 12px; }

/* Results */
.results-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px;
}
.results-header h3 { margin: 0; }

.results-list { display: flex; flex-direction: column; gap: 10px; }

.result-card {
  padding: 14px; border-radius: var(--radius-xs);
  background: var(--bg-primary); border: 1px solid var(--border-subtle);
}
.result-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;
}
.result-skill { font-weight: 600; font-size: 13px; color: var(--primary); }
.result-time { font-size: 11px; color: var(--text-tertiary); }
.result-content { font-size: 13px; color: var(--text-primary); line-height: 1.5; margin: 8px 0; }

.result-suggestions, .result-errors {
  margin-top: 8px; padding: 10px; border-radius: var(--radius-xs);
}
.result-suggestions { background: #e8f5e9; }
.result-errors { background: #fbe9e7; }
.result-suggestions h4, .result-errors h4, .result-nodes h4 {
  font-size: 12px; font-weight: 600; margin: 0 0 6px;
}
.result-suggestions ul, .result-errors ul {
  margin: 0; padding-left: 16px; font-size: 12px;
}
.result-suggestions li, .result-errors li { margin-bottom: 3px; }

.result-nodes { margin-top: 8px; }
.node-tags { display: flex; gap: 6px; flex-wrap: wrap; }
.node-tag {
  padding: 3px 10px; border-radius: 12px; font-size: 11px;
  background: var(--bg-active); color: var(--primary); font-weight: 500;
}
</style>
