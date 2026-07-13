# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0-rc.1] - 2026-07-10

### Release Candidate 1

这是 AIStudio 的第一个 Release Candidate 版本，标志着从开发阶段进入发布准备阶段。

### 变更

#### 配置
- 将默认环境从 `development` 切换为 `production`
- 将日志级别从 `debug` 调整为 `info`
- 修复 Backend 可执行文件路径配置（`cmd.exe` → `aistudio-backend.exe`）
- 添加 `production.yaml` 环境配置

#### 后端修复
- 移除 `settings.go` 中的 TODO 占位符，改为内存存储实现
- 移除 `llm_provider.go` 中的 TODO 注释，改为清晰的未实现说明
- 优化 Launcher 依赖检查，支持编译二进制和 `go run` 两种启动模式
- 修复 Launcher 中 Backend 默认启动命令

#### 前端修复
- 移除 `ProjectManagement.vue` 中所有 `console.log` 调试语句
- 为未实现的功能按钮添加用户提示（`alert`）
- 修复 `Dashboard.vue` 中的空函数实现
- 移除 `WorkflowEditor.vue` 中的 `console.log` 调试语句
- 移除 `FlowCanvas.vue` 中的 `console.error` 调试语句
- 移除 `workflow.ts` store 中的 Mock 数据回退逻辑
- 修复 `WorkflowValidator` 中的 Mock CUDA 检查，改为运行时自动检测提示

#### 工程化
- 更新 `CHANGELOG.md`，记录 RC 版本变更
- 生成完整的 Release Checklist

### 已知问题
- 设置页面尚未实现持久化存储（当前为内存存储）
- Claude 和 Gemini 的流式对话尚未实现
- 部分功能按钮（模型训练、部署、数据集管理等）尚未实现，将在后续版本中完成

## [Unreleased]

### Added
- 初始化项目架构