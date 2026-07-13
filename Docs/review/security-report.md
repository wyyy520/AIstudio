# AIStudio 安全审查报告

> **审查日期**: 2026-07-10  
> **审查范围**: SQL注入、路径遍历、命令执行、Token管理、API Key存储、密码存储、权限控制、敏感信息泄漏

---

## 一、安全评分总览

| 维度 | 评分 | 风险等级 |
|------|------|---------|
| 认证机制 | ⭐⭐⭐ | 🟡 MEDIUM |
| 授权控制 | ⭐⭐⭐⭐ | 🟢 LOW |
| 数据存储 | ⭐⭐ | 🟠 HIGH |
| 输入验证 | ⭐⭐⭐ | 🟡 MEDIUM |
| 通信安全 | ⭐⭐⭐ | 🟡 MEDIUM |
| 配置安全 | ⭐ | 🔴 CRITICAL |

---

## 二、🔴 CRITICAL 安全问题

### 2.1 JWT Secret 硬编码默认值

**文件**: [token.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/auth/token.go) L19, [middleware.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/middleware/middleware.go) L28, [service.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/service/service.go) L30

```go
secret = "aistudio-default-secret-change-in-production"
```

**风险**: 
- 默认密钥在代码中硬编码，所有部署实例使用相同密钥
- 攻击者可以伪造任意 JWT Token
- 可以伪造管理员权限

**修复**: 
1. 生产环境强制从环境变量读取
2. 启动时检查是否为默认值，如果是则警告/阻止启动
3. 支持密钥轮换

### 2.2 API Key 存储在数据库明文

**文件**: [database/models/api_key.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/database/models/api_key.go), [auth/apikey.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/auth/apikey.go)

**风险**: 数据库泄漏时所有 API Key 暴露。

**修复**: 
1. 使用 bcrypt 或 argon2 哈希存储 API Key
2. 仅存储 Key 前缀用于显示（如 `sk-****abcd`）
3. 支持 Key 轮换和撤销

### 2.3 密码存储安全性不足

**文件**: [auth/user.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/auth/user.go)

**风险**: 需要确认密码是否使用 bcrypt 存储。检查 `golang.org/x/crypto` 的使用。

### 2.4 插件安装路径遍历

**文件**: [plugin/installer.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/installer.go), [plugin/cloner.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/cloner.go)

**风险**: 插件安装时未验证路径是否在预期目录内，允许 `../` 路径遍历。

**修复**: 
1. 使用 `filepath.Clean` 和 `filepath.Abs` 规范化路径
2. 验证路径以插件目录为前缀
3. 禁止符号链接

---

## 三、🟠 HIGH 安全问题

### 3.1 Python 子进程命令注入

**文件**: [python.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/engine/python.go) L85

```go
cmd := exec.CommandContext(ctx, r.pythonPath, runnerScript, "--task", taskPath)
```

**风险**: 
- 如果 `r.pythonPath` 被用户配置修改，可以执行任意命令
- `runnerScript` 路径拼接可能被利用

**修复**: 
1. 验证 pythonPath 是可执行文件路径
2. 禁止在路径中使用 shell 元字符
3. 使用 `exec.Command` 而非 `exec.Command("sh", "-c", ...)`

### 3.2 缺乏 CSRF 防护

**文件**: [middleware.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/middleware/middleware.go)

**风险**: 所有 API 端点（除 login/register）依赖 JWT 认证，但没有 CSRF Token。如果用户在登录状态下访问恶意网站，攻击者可利用用户的认证状态发起请求。

**修复**: 
1. 添加 CSRF Token 中间件
2. 或使用 SameSite=Strict Cookie

### 3.3 API 速率限制配置缺失

**文件**: [rate_limit.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/middleware/rate_limit.go)

**风险**: 需要确认速率限制是否真正生效。如果配置不当，暴力破解攻击可以持续进行。

### 3.4 敏感信息在日志中泄漏

**文件**: 全局检查

**风险**: 
- 错误信息可能包含 SQL 查询、文件路径、堆栈信息
- 开发模式返回详细错误到客户端

**修复**: 
1. 生产环境禁用详细错误信息
2. 日志中过滤密码、Token、API Key 等敏感字段

---

## 四、🟡 MEDIUM 安全问题

### 4.1 文件上传无限制

**文件**: [plugin/installer.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/plugin/installer.go)

**风险**: 插件安装时从 URL 下载文件，没有限制文件大小和类型。

**修复**: 
1. 限制文件大小（如 100MB）
2. 验证文件类型
3. 沙箱执行插件代码

### 4.2 CORS 配置过于宽松

**文件**: [cors.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/api/middleware/cors.go) L16-L26

**风险**: 允许了多个 localhost 源，但在生产环境可能过于宽松。

**修复**: 生产环境限制为具体域名，不使用通配符。

### 4.3 Token 过期时间不合理

**文件**: [service.go](file:///d:/AIstudio-master/AIstudio-master/Backend/internal/service/service.go) L25

```go
tokenMgr := auth.NewTokenManager(24*60*60*1e9, 7*24*60*60*1e9)
```

**风险**: Access Token 有效期 24 小时，Refresh Token 有效期 7 天。如果 Token 被窃取，攻击者有较长的攻击窗口。

**建议**: 
1. Access Token 缩短到 15 分钟
2. Refresh Token 使用 HTTP-Only Cookie
3. 实现 Token 轮换

---

## 五、安全修复计划

### 5.1 立即修复（P0）

| 问题 | 修复方案 | 工作量 |
|------|---------|-------|
| JWT 硬编码密钥 | 启动时检查并强制环境变量 | 1h |
| API Key 明文存储 | 使用 bcrypt 哈希 | 2h |
| 密码存储验证 | 确认 bcrypt 使用 | 1h |
| 路径遍历验证 | 添加路径规范化检查 | 2h |

### 5.2 短期修复（P1）

| 问题 | 修复方案 | 工作量 |
|------|---------|-------|
| 命令注入防护 | 验证 pythonPath 可执行文件 | 2h |
| CSRF 防护 | 添加 CSRF 中间件 | 4h |
| 速率限制生效 | 配置合理的速率限制 | 2h |
| 日志过滤敏感信息 | 实现日志脱敏 | 3h |

### 5.3 长期修复（P2）

| 问题 | 修复方案 | 工作量 |
|------|---------|-------|
| 文件上传限制 | 实现文件类型/大小验证 | 2h |
| Token 过期时间 | 缩短 Token 有效期 | 1h |
| CORS 生产配置 | 按环境配置 CORS | 1h |