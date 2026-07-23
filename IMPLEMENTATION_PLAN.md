# EngStudio 项目实施计划

> 基于 EngStudio.md 设计文档，按优先级分阶段实施

---

## 第一阶段：Compiler 核心模块（最高优先级）

### 1.1 Graph Optimizer（图优化器）
- [ ] 死节点消除（Dead Node Elimination）
- [ ] 无效边清理（Invalid Edge Removal）
- [ ] 重复边去除（Duplicate Edge Removal）
- [ ] 相同类型节点融合（Node Fusion）
- [ ] 不可达节点移除（Unreachable Node Removal）
- [ ] 环路检测与处理（Cycle Detection）

### 1.2 EWIR Builder（工程中间表示构建器）
- [ ] ui.json 生成（编辑器状态）
- [ ] workflow.ir.json 生成（工程中间表示）
- [ ] 数据流与控制流分离
- [ ] 依赖关系提取

### 1.3 Execution Plan Builder（执行计划构建器）
- [ ] 拓扑排序（Kahn 算法）
- [ ] 依赖分析
- [ ] 执行步骤生成
- [ ] Domain 分发

### 1.4 Plugin Manifest Generator（插件清单生成器）
- [ ] plugin_manifest.json 生成
- [ ] 插件依赖分析

---

## 第二阶段：Generator 与 Template Engine

### 2.1 Template Engine（模板引擎）
- [ ] Go Template 模板引擎封装
- [ ] 模板变量替换
- [ ] 模板路径管理
- [ ] 模板扩展机制

### 2.2 Generator 系统
- [ ] Generator 统一接口定义
- [ ] Python Generator
- [ ] MATLAB Generator
- [ ] STM32 Generator
- [ ] ANSYS Generator
- [ ] Generator Registry（注册表）

### 2.3 工程模板
- [ ] Python 工程模板
- [ ] MATLAB 工程模板
- [ ] STM32 工程模板
- [ ] ANSYS 工程模板

---

## 第三阶段：Runtime 执行引擎

### 3.1 Runtime 核心
- [ ] Runtime 统一接口
- [ ] Executor Registry
- [ ] Python Executor
- [ ] MATLAB Executor
- [ ] STM32 Executor
- [ ] 执行状态管理
- [ ] 执行日志采集

### 3.2 多 Runtime 支持
- [ ] Python Runtime
- [ ] MATLAB Runtime
- [ ] STM32 Runtime
- [ ] ANSYS Runtime

---

## 第四阶段：前端页面完善

### 4.1 缺失页面开发
- [ ] Compiler 页面（编译面板）
- [ ] Generator 页面（工程生成器面板）
- [ ] Runtime 页面（运行时执行监控）
- [ ] Diagnose Center 页面（诊断中心）
- [ ] Skill Center 页面（AI 技能中心）
- [ ] Log Center 页面（日志中心）

### 4.2 工作流节点完善
- [ ] 87 个节点定义
- [ ] 节点属性面板
- [ ] 节点端口配置
- [ ] 节点分类与搜索

---

## 第五阶段：AI 与智能辅助

### 5.1 Skill Center
- [ ] Workflow Planner
- [ ] Explain Skill
- [ ] Diagnose Skill
- [ ] Optimize Skill
- [ ] Auto Connect
- [ ] Environment Skill
- [ ] Generate Workflow

### 5.2 RAG 工程知识库
- [ ] 知识库索引
- [ ] 向量数据库集成
- [ ] 检索增强生成

---

## 第六阶段：插件系统完善

### 6.1 插件 SDK
- [ ] Plugin SDK
- [ ] Node SDK
- [ ] Generator SDK
- [ ] Runtime SDK
- [ ] Skill SDK
- [ ] Provider SDK
- [ ] Template SDK

### 6.2 插件市场
- [ ] 插件安装/卸载
- [ ] 插件版本管理
- [ ] 插件依赖管理

---

## 第七阶段：测试与优化

### 7.1 单元测试
- [ ] Compiler 测试
- [ ] Generator 测试
- [ ] Runtime 测试
- [ ] Workflow 测试

### 7.2 集成测试
- [ ] 端到端测试
- [ ] 性能测试
- [ ] 压力测试

---

## 当前优先级

**立即开始：第一阶段 - Compiler 核心模块**

这是整个系统的核心，所有后续模块都依赖 Compiler 的输出。