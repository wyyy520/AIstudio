# API 接口总览

## 1. 基础信息

| 项目 | 值 |
|------|-----|
| Base URL | `http://localhost:8080/api/v1` |
| 协议 | HTTP / WebSocket |
| 数据格式 | JSON |
| 认证方式 | Bearer Token |

## 2. 通用响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 0=成功，非0=错误码 |
| message | string | 状态描述 |
| data | object | 响应数据 |

## 3. 错误码定义

| 错误码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未认证 |
| 1003 | 权限不足 |
| 2001 | 工作流不存在 |
| 2002 | 工作流校验失败 |
| 3001 | 插件不存在 |
| 3002 | 插件执行失败 |
| 4001 | 任务不存在 |
| 4002 | 任务超时 |
| 5001 | 引擎连接失败 |
| 5002 | 模型加载失败 |

## 4. 接口列表

### 认证模块

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /auth/login | 用户登录 |
| POST | /auth/logout | 用户登出 |
| GET | /auth/profile | 获取当前用户 |

### 项目管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /projects | 项目列表 |
| POST | /projects | 创建项目 |
| GET | /projects/:id | 项目详情 |
| PUT | /projects/:id | 更新项目 |
| DELETE | /projects/:id | 删除项目 |

### 工作流

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /workflows | 工作流列表 |
| POST | /workflows | 创建工作流 |
| GET | /workflows/:id | 工作流详情 |
| PUT | /workflows/:id | 更新工作流 |
| DELETE | /workflows/:id | 删除工作流 |
| POST | /workflows/:id/run | 运行工作流 |
| GET | /workflows/:id/status | 运行状态 |
| POST | /workflows/:id/stop | 停止运行 |
| GET | /workflows/:id/nodes/:nodeId/output | 节点输出 |

### 插件管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /plugins | 插件列表 |
| GET | /plugins/:name | 插件详情 |
| POST | /plugins/install | 安装插件 |
| DELETE | /plugins/:name | 卸载插件 |
| GET | /plugins/:name/config-schema | 插件配置 Schema |

### 任务

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /tasks | 任务列表 |
| GET | /tasks/:id | 任务详情 |
| GET | /tasks/:id/logs | 任务日志 |
| POST | /tasks/:id/cancel | 取消任务 |

### Agent

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /agent/chat | 发送对话 |
| GET | /agent/sessions | 对话列表 |
| POST | /agent/sessions | 创建对话 |
| DELETE | /agent/sessions/:id | 删除对话 |

### WebSocket

| 路径 | 说明 |
|------|------|
| ws://localhost:8080/ws/task/:taskId | 任务实时状态推送 |
| ws://localhost:8080/ws/agent/:sessionId | Agent 流式对话 |
