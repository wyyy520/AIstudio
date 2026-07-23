import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface Skill {
  id: string
  name: string
  description: string
  icon: string
  category: 'planner' | 'diagnose' | 'explain' | 'optimize' | 'environment' | 'automation'
  status: 'available' | 'running' | 'error'
  requiresAI: boolean
}

export interface SkillResult {
  skillId: string
  timestamp: string
  content: string
  suggestions?: string[]
  nodes?: any[]
  errors?: string[]
}

export const useSkillCenterStore = defineStore('skillCenter', () => {
  const skills = ref<Skill[]>([
    { id: 'workflow-planner', name: 'Workflow Planner', description: '根据自然语言描述自动生成工作流', icon: '📋', category: 'planner', status: 'available', requiresAI: true },
    { id: 'workflow-explain', name: 'Explain Skill', description: '解释当前工作流的逻辑和数据流向', icon: '💡', category: 'explain', status: 'available', requiresAI: true },
    { id: 'workflow-diagnose', name: 'Diagnose Skill', description: '诊断工作流中的问题和潜在错误', icon: '🔍', category: 'diagnose', status: 'available', requiresAI: true },
    { id: 'workflow-optimize', name: 'Optimize Skill', description: '分析并优化工作流结构，提出改进建议', icon: '⚡', category: 'optimize', status: 'available', requiresAI: true },
    { id: 'auto-connect', name: 'Auto Connect', description: '根据端口类型自动连接兼容节点', icon: '🔗', category: 'automation', status: 'available', requiresAI: false },
    { id: 'environment-skill', name: 'Environment Skill', description: '检测和诊断开发环境配置', icon: '🌍', category: 'environment', status: 'available', requiresAI: false },
    { id: 'generate-workflow', name: 'Generate Workflow', description: '从模板快速生成常用工作流', icon: '🚀', category: 'automation', status: 'available', requiresAI: false },
    { id: 'node-recommend', name: 'Node Recommend', description: '根据当前工作流推荐下一步节点', icon: '🎯', category: 'planner', status: 'available', requiresAI: true },
    { id: 'param-optimize', name: 'Param Optimize', description: '智能优化节点参数配置', icon: '🎛️', category: 'optimize', status: 'available', requiresAI: true },
    { id: 'error-analyze', name: 'Error Analyze', description: '分析运行时错误并提供修复建议', icon: '🛠️', category: 'diagnose', status: 'available', requiresAI: true },
  ])

  const results = ref<SkillResult[]>([])
  const isRunning = ref(false)
  const currentSkillId = ref<string | null>(null)

  const plannerSkills = computed(() => skills.value.filter(s => s.category === 'planner'))
  const diagnoseSkills = computed(() => skills.value.filter(s => s.category === 'diagnose'))
  const explainSkills = computed(() => skills.value.filter(s => s.category === 'explain'))
  const optimizeSkills = computed(() => skills.value.filter(s => s.category === 'optimize'))
  const environmentSkills = computed(() => skills.value.filter(s => s.category === 'environment'))
  const automationSkills = computed(() => skills.value.filter(s => s.category === 'automation'))

  async function runSkill(skillId: string) {
    const skill = skills.value.find(s => s.id === skillId)
    if (!skill) return

    isRunning.value = true
    currentSkillId.value = skillId
    skill.status = 'running'

    try {
      await simulateDelay(800 + Math.random() * 1500)

      const result: SkillResult = {
        skillId,
        timestamp: new Date().toISOString(),
        content: '',
        suggestions: [],
      }

      switch (skillId) {
        case 'workflow-planner':
          result.content = '已根据您的描述生成工作流草案。建议添加 Dataset 节点作为数据源，连接 YOLO 训练节点，然后使用 Export 节点导出模型。'
          result.nodes = [{ type: 'dataset' }, { type: 'yolo' }, { type: 'export' }]
          break
        case 'workflow-explain':
          result.content = '当前工作流包含 3 个阶段：数据加载 → 模型训练 → 结果导出。数据从 Dataset 节点流向 YOLO 训练节点，最终由 Export 节点输出模型文件。'
          break
        case 'workflow-diagnose':
          result.content = '工作流诊断完成。发现以下问题：'
          result.errors = ['节点 LSTM_1 缺少数据集输入', 'Export 节点的输出路径未配置']
          result.suggestions = ['将 Dataset 的输出连接到 LSTM_1 的输入端口', '在 Export 节点中设置 output_dir 参数']
          break
        case 'workflow-optimize':
          result.content = '工作流优化分析完成。检测到 2 个优化机会：'
          result.suggestions = ['YOLO 和 LSTM 节点可以并行执行，节省 ~40% 时间', '数据增强节点可以合并为一个复合增强步骤']
          break
        case 'auto-connect':
          result.content = '已自动连接 5 对兼容端口，1 对类型不兼容需要手动确认。'
          break
        case 'environment-skill':
          result.content = '环境检测完成：Python 3.10 ✅ | CUDA 12.1 ✅ | PyTorch 2.1 ✅ | 所有依赖已就绪。'
          break
        default:
          result.content = `${skill.name} 执行完成。`
      }

      results.value.unshift(result)
      skill.status = 'available'
    } catch (e) {
      skill.status = 'error'
    } finally {
      isRunning.value = false
      currentSkillId.value = null
    }
  }

  function clearResults() {
    results.value = []
  }

  return {
    skills, results, isRunning, currentSkillId,
    plannerSkills, diagnoseSkills, explainSkills,
    optimizeSkills, environmentSkills, automationSkills,
    runSkill, clearResults,
  }
})

function simulateDelay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}
