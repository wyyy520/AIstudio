# 前端架构

## 1. 技术选型

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue 3 | ^3.4 | UI 框架（Composition API） |
| TypeScript | ^5.3 | 类型安全 |
| Tauri | ^1.5 | 桌面应用封装 |
| Vue Router | ^4.3 | 路由管理 |
| Pinia | ^2.1 | 状态管理 |
| Vite | ^5.0 | 构建工具 |

## 2. 目录结构说明

```
Frontend/
├── public/              # 静态资源
├── src/
│   ├── assets/          # 图片/字体/样式
│   ├── components/      # 全局通用组件
│   ├── pages/           # 页面级组件
│   ├── workflow/        # 工作流编辑器（核心）
│   ├── chat/            # AI 对话界面
│   ├── plugin/          # 插件管理界面
│   ├── project/         # 项目管理界面
│   ├── router/          # 路由配置
│   ├── store/           # Pinia 状态管理
│   ├── api/             # 后端 API 封装
│   ├── utils/           # 工具函数
│   ├── hooks/           # 组合式函数
│   └── types/           # TypeScript 类型定义
└── package.json
```

## 3. 核心模块设计

### 3.1 工作流编辑器 (workflow/)

```
workflow/
├── Canvas.vue          # 画布组件（节点拖拽/连线）
├── NodePanel.vue       # 节点面板（左侧拖拽源）
├── PropertyPanel.vue   # 属性面板（右侧参数配置）
├── Toolbar.vue         # 工具栏（保存/运行/调试）
├── nodes/              # 各类节点组件
│   ├── VisionNode.vue
│   ├── NLPNode.vue
│   ├── LogicNode.vue
│   └── SystemNode.vue
└── engine/             # 前端工作流引擎
    ├── Graph.ts        # 图数据结构
    ├── Executor.ts     # 执行器（预览模式）
    └── Validator.ts    # 节点连线校验
```

### 3.2 状态管理 (store/)

```typescript
// store/workflow.ts - 工作流状态
interface WorkflowState {
  nodes: WorkflowNode[]       // 所有节点
  edges: WorkflowEdge[]       // 所有连线
  selectedNode: string | null  // 当前选中节点
  isRunning: boolean           // 是否运行中
  runtime: NodeRuntime[]       // 运行时状态
}

// store/project.ts - 项目状态
interface ProjectState {
  currentProject: Project | null
  projectList: Project[]
}

// store/chat.ts - 对话状态
interface ChatState {
  messages: Message[]
  isStreaming: boolean
}
```

### 3.3 API 层 (api/)

```typescript
// api/request.ts - 请求封装
// 自动适配 Tauri IPC / HTTP
const request = isTauri ? tauriInvoke : httpFetch

// api/modules/workflow.ts
export const workflowApi = {
  save: (data: WorkflowData) => request('/workflow/save', data),
  run: (id: string) => request('/workflow/run', { id }),
  getStatus: (id: string) => request('/workflow/status', { id }),
}
```

## 4. 页面路由

| 路径 | 页面 | 说明 |
|------|------|------|
| / | 首页/仪表盘 | 项目概览、最近工作流 |
| /workflow/:id | 工作流编辑器 | 核心编辑界面 |
| /chat | AI 对话 | 与 Agent 对话 |
| /plugins | 插件市场 | 浏览/管理插件 |
| /projects | 项目管理 | 项目列表 |
| /settings | 设置 | 全局配置 |

## 5. Tauri 集成

```typescript
// Tauri IPC 调用示例
import { invoke } from '@tauri-apps/api/tauri'

// 调用 Backend 命令
const result = await invoke('run_workflow', { id: 'wf_001' })

// 文件系统操作
import { readTextFile, writeTextFile } from '@tauri-apps/api/fs'
```
