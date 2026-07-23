import type { QuickAction } from './types'

export const mockQuickActions: QuickAction[] = [
  {
    id: 'generate-workflow',
    label: '生成工作流',
    icon: 'M6 3h3v6H6V3zm0 12h3v6H6v-6zm9-12h3v6h-3V3zm0 12h3v6h-3v-6zm-9 0V9m3 9v-3m3 3v-3m3 3V9',
    prompt: '帮我创建一个工作流，包含数据输入、处理和输出节点',
  },
  {
    id: 'analyze-error',
    label: '分析错误',
    icon: 'M12 9v2m0 4h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0z',
    prompt: '请分析最近的错误日志，帮我找出问题原因',
  },
  {
    id: 'explain-node',
    label: '解释节点',
    icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2V7zm0 8h2v2h-2v-2z',
    prompt: '解释一下工作流中各个节点的功能和最佳实践',
  },
  {
    id: 'optimize-workflow',
    label: '优化建议',
    icon: 'M13 2L3 14h9l-1 8 10-12h-9l1-8z',
    prompt: '请对我的工作流提出优化建议，提高执行效率',
  },
]
