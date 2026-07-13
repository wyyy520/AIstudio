export type ProjectType = 'detection' | 'classification' | 'segmentation' | 'timeseries' | 'custom'

export type ProjectStatus = 'active' | 'idle' | 'running' | 'error' | 'archived'

export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'

export type PluginFramework = 'pytorch' | 'tensorflow' | 'onnx' | 'tensorrt' | 'auto'

export interface ProjectEnvironment {
  pythonVersion: string
  cudaVersion: string
  pytorchVersion: string
  gpuStatus: 'ready' | 'warning' | 'error'
  dependencies: Dependency[]
}

export interface Dependency {
  name: string
  version: string
  status: 'installed' | 'outdated' | 'missing'
}

export interface ProjectWorkflow {
  id: string
  name: string
  version: string
  nodeCount: number
  updatedAt: string
  status: TaskStatus
}

export interface ProjectDataset {
  id: string
  name: string
  format: string
  size: string
  imageCount: number
  classCount: number
  status: TaskStatus
}

export interface ProjectModel {
  id: string
  name: string
  version: string
  framework: PluginFramework
  size: string
  source: string
  trainedAt: string
  accuracy: number
}

export interface ProjectExperiment {
  id: string
  modelId: string
  modelName: string
  epoch: number
  loss: number
  accuracy: number
  gpu: string
  duration: string
  status: TaskStatus
}

export interface Project {
  id: string
  name: string
  type: ProjectType
  status: ProjectStatus
  createdAt: string
  updatedAt: string
  description: string
  template: string
  framework: PluginFramework
  plugins: string[]
  workflows: ProjectWorkflow[]
  datasets: ProjectDataset[]
  models: ProjectModel[]
  experiments: ProjectExperiment[]
  environment: ProjectEnvironment
  outputs: string[]
  logs: string[]
}

export interface ProjectTemplate {
  id: string
  name: string
  description: string
  type: ProjectType
  framework: PluginFramework
  plugins: string[]
}
