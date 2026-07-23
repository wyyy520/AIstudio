import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type GeneratorStatus =
  | 'idle'
  | 'planning'
  | 'loading_template'
  | 'generating'
  | 'completed'
  | 'failed'

export interface GeneratedProject {
  target: string
  name: string
  outputDir: string
  entryPoints: string[]
  fileCount: number
  estimatedSize: string
}

export interface GeneratedFileEntry {
  path: string
  size: number
  type: 'code' | 'config' | 'resource' | 'template' | 'other'
  language?: string
}

export interface GeneratorLog {
  timestamp: string
  level: 'info' | 'warning' | 'error'
  message: string
  file?: string
}

export const useGeneratorStore = defineStore('generator', () => {
  const status = ref<GeneratorStatus>('idle')
  const progress = ref(0)
  const error = ref<string | null>(null)
  const logs = ref<GeneratorLog[]>([])
  const generatedFiles = ref<GeneratedFileEntry[]>([])
  const currentTarget = ref<string | null>(null)
  const projectInfo = ref<GeneratedProject | null>(null)
  const availableTargets = ref([
    { id: 'python', name: 'Python', icon: '🐍', description: 'AI/ML 工程', version: '3.8+' },
    { id: 'matlab', name: 'MATLAB', icon: '📊', description: '仿真与控制工程', version: 'R2020b+' },
    { id: 'stm32', name: 'STM32', icon: '🔌', description: '嵌入式工程', version: 'CubeMX' },
    { id: 'ansys', name: 'ANSYS', icon: '🔧', description: '有限元仿真工程', version: '2023+' },
    { id: 'cpp', name: 'C++', icon: '⚡', description: '高性能工程', version: 'C++17+' },
    { id: 'java', name: 'Java', icon: '☕', description: '企业级工程', version: '17+' },
    { id: 'ros2', name: 'ROS2', icon: '🤖', description: '机器人工程', version: 'Humble+' },
    { id: 'unity', name: 'Unity', icon: '🎮', description: '游戏/仿真工程', version: '2022 LTS+' },
    { id: 'docker', name: 'Docker', icon: '🐳', description: '容器化部署', version: '24+' },
  ])

  const isGenerating = computed(() => {
    return !['idle', 'completed', 'failed'].includes(status.value)
  })

  const generatedCodeFileCount = computed(() =>
    generatedFiles.value.filter(f => f.type === 'code').length
  )

  const generatedConfigFileCount = computed(() =>
    generatedFiles.value.filter(f => f.type === 'config').length
  )

  function addLog(level: GeneratorLog['level'], message: string, file?: string) {
    logs.value.push({ timestamp: new Date().toISOString(), level, message, file })
  }

  async function generate(target: string, projectId: string) {
    status.value = 'planning'
    progress.value = 0
    error.value = null
    logs.value = []
    generatedFiles.value = []
    currentTarget.value = target
    projectInfo.value = null

    addLog('info', `开始生成 ${target} 工程...`)

    try {
      await simulateDelay(400)
      progress.value = 10
      addLog('info', '读取 Execution Plan...')

      await simulateDelay(300)
      progress.value = 20
      addLog('info', `定位 ${target} Template...`)
      status.value = 'loading_template'

      await simulateDelay(500)
      progress.value = 35
      addLog('info', '模板加载成功，准备渲染')
      status.value = 'generating'

      // Simulate file generation
      const fileList = generateFileList(target)
      for (let i = 0; i < fileList.length; i++) {
        await simulateDelay(100 + Math.random() * 200)
        const file = fileList[i]
        generatedFiles.value.push(file)
        addLog('info', `生成文件: ${file.path}`, file.path)
        progress.value = 35 + Math.floor((i + 1) / fileList.length * 55)
      }

      await simulateDelay(300)
      progress.value = 95
      addLog('info', '创建项目目录结构...')
      addLog('info', '写入工程配置文件...')

      await simulateDelay(200)
      progress.value = 100
      status.value = 'completed'

      projectInfo.value = {
        target,
        name: `generated-${target}-project`,
        outputDir: `./generated/${target}/`,
        entryPoints: getEntryPoints(target),
        fileCount: fileList.length,
        estimatedSize: `${(fileList.length * 2.5).toFixed(1)} KB`,
      }

      addLog('info', `✅ ${target} 工程生成完成！共 ${fileList.length} 个文件`)
    } catch (e) {
      status.value = 'failed'
      error.value = e instanceof Error ? e.message : '生成失败'
      addLog('error', `❌ 生成失败: ${error.value}`)
    }
  }

  function reset() {
    status.value = 'idle'
    progress.value = 0
    error.value = null
    logs.value = []
    generatedFiles.value = []
    currentTarget.value = null
    projectInfo.value = null
  }

  return {
    status, progress, error, logs, generatedFiles,
    currentTarget, projectInfo, availableTargets,
    isGenerating, generatedCodeFileCount, generatedConfigFileCount,
    generate, reset, addLog,
  }
})

function simulateDelay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

function generateFileList(target: string): GeneratedFileEntry[] {
  const templates: Record<string, GeneratedFileEntry[]> = {
    python: [
      { path: 'main.py', size: 1024, type: 'code', language: 'python' },
      { path: 'runtime.py', size: 2048, type: 'code', language: 'python' },
      { path: 'registry.py', size: 512, type: 'code', language: 'python' },
      { path: 'executors/yolo_executor.py', size: 4096, type: 'code', language: 'python' },
      { path: 'executors/dataset_executor.py', size: 2048, type: 'code', language: 'python' },
      { path: 'requirements.txt', size: 256, type: 'config' },
      { path: 'pyproject.toml', size: 512, type: 'config' },
      { path: 'execution_plan.json', size: 2048, type: 'config' },
      { path: '.gitignore', size: 128, type: 'config' },
      { path: 'README.md', size: 512, type: 'resource' },
    ],
    matlab: [
      { path: 'main.m', size: 1024, type: 'code', language: 'matlab' },
      { path: 'runtime.m', size: 2048, type: 'code', language: 'matlab' },
      { path: 'executors/pid_executor.m', size: 3072, type: 'code', language: 'matlab' },
      { path: 'executors/simulink_executor.m', size: 4096, type: 'code', language: 'matlab' },
      { path: 'execution_plan.json', size: 1536, type: 'config' },
      { path: 'startup.m', size: 256, type: 'config' },
      { path: 'README.md', size: 512, type: 'resource' },
    ],
    stm32: [
      { path: 'Core/Src/main.c', size: 3072, type: 'code', language: 'c' },
      { path: 'Core/Inc/main.h', size: 512, type: 'code', language: 'c' },
      { path: 'Drivers/CMSIS/system_stm32.c', size: 4096, type: 'code', language: 'c' },
      { path: 'executors/gpio_executor.c', size: 2048, type: 'code', language: 'c' },
      { path: 'project.ioc', size: 8192, type: 'config' },
      { path: 'Makefile', size: 1024, type: 'config' },
      { path: 'execution_plan.json', size: 1536, type: 'config' },
      { path: 'README.md', size: 512, type: 'resource' },
    ],
    ansys: [
      { path: 'workbench.wbpj', size: 2048, type: 'config' },
      { path: 'mechanical.dat', size: 4096, type: 'code', language: 'apdl' },
      { path: 'journal.wbjn', size: 1024, type: 'code', language: 'python' },
      { path: 'mesh/mesh.cdb', size: 8192, type: 'resource' },
      { path: 'execution_plan.json', size: 1024, type: 'config' },
      { path: 'README.md', size: 512, type: 'resource' },
    ],
    cpp: [
      { path: 'src/main.cpp', size: 2048, type: 'code', language: 'cpp' },
      { path: 'include/project.h', size: 1024, type: 'code', language: 'cpp' },
      { path: 'src/executor.cpp', size: 4096, type: 'code', language: 'cpp' },
      { path: 'CMakeLists.txt', size: 1024, type: 'config' },
      { path: 'execution_plan.json', size: 1536, type: 'config' },
      { path: 'README.md', size: 512, type: 'resource' },
    ],
  }

  return templates[target] || [
    { path: 'main.go', size: 2048, type: 'code', language: 'go' },
    { path: 'config.yaml', size: 512, type: 'config' },
    { path: 'execution_plan.json', size: 1024, type: 'config' },
    { path: 'README.md', size: 512, type: 'resource' },
  ]
}

function getEntryPoints(target: string): string[] {
  const entries: Record<string, string[]> = {
    python: ['main.py'],
    matlab: ['main.m'],
    stm32: ['Core/Src/main.c'],
    ansys: ['workbench.wbpj'],
    cpp: ['build/project'],
  }
  return entries[target] || ['main']
}
