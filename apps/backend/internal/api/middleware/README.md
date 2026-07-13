# Middleware 层

生产级 HTTP 中间件组件，基于 Gin 框架构建。

## 组件列表

| 中间件 | 文件 | 说明 |
|--------|------|------|
| Auth | `auth.go` | JWT Bearer Token 认证 |
| CORS | `cors.go` | 跨域资源共享（支持 localhost + Tauri） |
| Logger | `logger.go` | 结构化 JSON 请求日志 |
| Recovery | `recovery.go` | Panic 恢复（防止服务器崩溃） |
| RateLimit | `rate_limit.go` | IP 限流（Token Bucket 算法） |

## 执行顺序

```
Request
  ↓
[1] Recovery      — 捕获所有 panic，返回 500
  ↓
[2] Logger        — 记录请求方法、路径、状态码、耗时、IP
  ↓
[3] CORS          — 设置跨域头，处理 OPTIONS 预检
  ↓
[4] RateLimit     — 限流（默认 100 req/min/IP）
  ↓
[5] Auth          — 验证 JWT Token（公开路径跳过）
  ↓
Handler
```

## 统一注册

所有中间件通过 `middleware.Apply()` 统一注册：

```go
import "github.com/aistudio/backend/internal/api/middleware"

cfg := middleware.DefaultConfig()
cfg.JWTSecret = os.Getenv("JWT_SECRET")
cfg.Development = os.Getenv("APP_ENV") == "development"

// 自定义 RateLimit
cfg.RateLimit = &middleware.RateLimitConfig{
    Rate:  100.0 / 60.0, // 100 req/min
    Burst: 100,
}

// 自定义 CORS
cfg.CORS = &middleware.CORSConfig{
    AllowedOrigins: []string{"http://localhost:5173", "tauri://localhost"},
}

middleware.Apply(router, cfg)
```

## 认证机制

### JWT 实现

- 算法: HS256 (HMAC-SHA256)
- 无外部依赖（使用 Go 标准库 crypto/hmac）
- Token 格式: `base64URL(header).base64URL(payload).base64URL(signature)`

### 公开路径

以下路径不需要认证：

| 路径 | 说明 |
|------|------|
| `/api/health` | 健康检查 |
| `/api/auth/login` | 用户登录 |

其他所有 `/api/*` 路径均需要 Bearer Token。

### 登录示例

```bash
curl -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 响应
{
  "code": 0,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

### 认证请求

```bash
curl http://localhost:8081/api/projects \
  -H "Authorization: Bearer <token>"
```

## CORS 配置

默认允许的来源：

- `http://localhost:5173` — Vite 开发服务器
- `http://localhost:3000` — 备选开发端口
- `http://localhost:8080` — 备选开发端口
- `tauri://localhost` — Tauri 桌面应用
- `https://tauri.localhost` — Tauri 生产环境
- `capacitor://localhost` — Capacitor 移动应用
- `http://localhost` — 通用 localhost

## 限流机制

### Token Bucket 算法

- 每个 IP 独立计数
- 默认: 100 个请求/分钟
- 超出返回 `429 Too Many Requests`
- OPTIONS 预检请求不计入限流
- 过期 bucket 自动清理（每 5 分钟）

### 响应头

限流时返回:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
```

## 结构化日志

Logger 输出 JSON 格式日志到 stdout：

```json
{
  "timestamp": "2024-01-01T12:00:00.123456Z",
  "level": "INFO",
  "method": "GET",
  "path": "/api/health",
  "status": 200,
  "duration_ms": "1.23",
  "ip": "127.0.0.1",
  "user_agent": "curl/8.0",
  "user_id": "",
  "source": "api"
}
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `JWT_SECRET` | JWT 签名密钥 | `aistudio-default-secret-change-in-production` |
| `APP_ENV` | 运行环境 (`development` / `production`) | 空（production） |

## 测试

```bash
cd Backend
go test ./internal/api/middleware/ -v
```

测试覆盖率：

| 测试 | 说明 |
|------|------|
| TestAuth_* | 认证: 公开路径、缺失Token、无效Token、过期Token、格式错误 |
| TestCORS_* | CORS: 允许来源、预检请求、拒绝来源 |
| TestLogger_* | 日志: 结构化 JSON 输出格式 |
| TestRecovery_* | 恢复: Panic 捕获、开发模式详情 |
| TestRateLimit_* | 限流: 正常请求、超限拒绝、预检跳过 |
| TestJWT_* | JWT: 生成、验证、过期、签名错误、格式错误 |
| TestFullStack_* | 全栈集成: 认证失败 + CORS、认证成功 + CORS |