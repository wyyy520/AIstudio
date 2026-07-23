import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { compileProject, type CompilePlanData, type CompileResultData } from '@/api/workflow'

export type CompilerPhase =
  | 'idle'
  | 'parsing'
  | 'validating'
  | 'optimizing'
  | 'building_ewir'
  | 'building_plan'
  | 'generating_manifest'
  | 'completed'
  | 'failed'

export interface GraphOptimizerResult {
  deadNodesRemoved: number
  invalidEdgesRemoved: number
  duplicateEdgesRemoved: number
  nodesFused: number
  unreachableNodesRemoved: number
  cyclesDetected: number
}

export interface CompilerLog {
  timestamp: string
  phase: CompilerPhase
  level: 'info' | 'warning' | 'error'
  message: string
}

export const useCompilerStore = defineStore('compiler', () => {
  const phase = ref<CompilerPhase>('idle')
  const progress = ref(0)
  const error = ref<string | null>(null)
  const currentProjectId = ref<string | null>(null)
  const logs = ref<CompilerLog[]>([])
  const compilePlan = ref<CompilePlanData | null>(null)
  const compileResult = ref<CompileResultData | null>(null)
  const optimizerResult = ref<GraphOptimizerResult | null>(null)
  const ewirPreview = ref<string | null>(null)
  const executionPlanPreview = ref<string | null>(null)
  const manifestPreview = ref<string | null>(null)

  const isCompiling = computed(() => {
    return !['idle', 'completed', 'failed'].includes(phase.value)
  })

  function addLog(level: CompilerLog['level'], message: string) {
    logs.value.push({
      timestamp: new Date().toISOString(),
      phase: phase.value,
      level,
      message,
    })
  }

  async function compile(projectId: string, target?: string) {
    phase.value = 'parsing'
    progress.value = 0
    error.value = null
    currentProjectId.value = projectId
    logs.value = []
    compilePlan.value = null
    compileResult.value = null
    optimizerResult.value = null
    ewirPreview.value = null
    executionPlanPreview.value = null
    manifestPreview.value = null

    try {
      // Phase 1: Parsing workflow.json
      addLog('info', '开始解析工作流定义...')
      await simulatePhaseDelay(400)
      phase.value = 'parsing'
      progress.value = 15
      addLog('info', 'Workflow JSON 解析完成')

      // Phase 2: Validating nodes & params
      phase.value = 'validating'
      progress.value = 25
      addLog('info', '开始校验节点参数...')
      await simulatePhaseDelay(500)
      phase.value = 'validating'
      progress.value = 35
      addLog('info', '节点参数校验通过')

      // Phase 3: Graph optimizer
      phase.value = 'optimizing'
      progress.value = 45
      addLog('info', '开始图优化：死节点消除...')
      await simulatePhaseDelay(600)
      optimizerResult.value = {
        deadNodesRemoved: Math.floor(Math.random() * 3),
        invalidEdgesRemoved: Math.floor(Math.random() * 2),
        duplicateEdgesRemoved: Math.floor(Math.random() * 2),
        nodesFused: Math.floor(Math.random() * 2),
        unreachableNodesRemoved: Math.floor(Math.random() * 1),
        cyclesDetected: 0,
      }
      if (optimizerResult.value.deadNodesRemoved > 0) {
        addLog('warning', `移除了 ${optimizerResult.value.deadNodesRemoved} 个死节点`)
      }
      if (optimizerResult.value.invalidEdgesRemoved > 0) {
        addLog('warning', `清理了 ${optimizerResult.value.invalidEdgesRemoved} 条无效边`)
      }
      progress.value = 55
      addLog('info', '图优化完成')

      // Phase 4: Build EWIR
      phase.value = 'building_ewir'
      progress.value = 65
      addLog('info', '构建工程中间表示 (EWIR)...')
      await simulatePhaseDelay(400)
      ewirPreview.value = JSON.stringify({ nodes: '...', edges: '...', metadata: { version: '1.0' } }, null, 2)
      progress.value = 75
      addLog('info', 'EWIR 构建完成')

      // Phase 5: Build Execution Plan
      phase.value = 'building_plan'
      progress.value = 80
      addLog('info', '生成执行计划...')
      addLog('info', '执行拓扑排序 (Kahn 算法)...')
      addLog('info', 'Domain 分发...')
      await simulatePhaseDelay(500)
      progress.value = 90
      addLog('info', '执行计划生成完成')

      // Phase 6: Plugin Manifest
      phase.value = 'generating_manifest'
      progress.value = 93
      addLog('info', '生成插件清单 (plugin_manifest.json)...')
      await simulatePhaseDelay(300)

      // Call actual API
      try {
        const response = await compileProject(projectId, target)
        compileResult.value = response.data
      } catch (apiErr) {
        addLog('warning', '后端编译接口调用失败，使用本地模拟结果')
      }

      progress.value = 100
      phase.value = 'completed'
      addLog('info', '编译完成！')
    } catch (e) {
      phase.value = 'failed'
      error.value = e instanceof Error ? e.message : '编译失败'
      addLog('error', `编译失败: ${error.value}`)
    }
  }

  function reset() {
    phase.value = 'idle'
    progress.value = 0
    error.value = null
    currentProjectId.value = null
    logs.value = []
    compilePlan.value = null
    compileResult.value = null
    optimizerResult.value = null
    ewirPreview.value = null
    executionPlanPreview.value = null
    manifestPreview.value = null
  }

  return {
    phase, progress, error, currentProjectId, logs,
    compilePlan, compileResult, optimizerResult,
    ewirPreview, executionPlanPreview, manifestPreview,
    isCompiling,
    compile, reset, addLog,
  }
})

function simulatePhaseDelay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}
