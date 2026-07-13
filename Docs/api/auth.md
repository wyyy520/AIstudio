# 认证接口

## POST /auth/login

用户登录，获取访问 Token。

### 请求参数

```json
{
  "username": "admin",
  "password": "encrypted_password"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码（客户端加密） |

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-07-08T14:00:00Z",
    "user": {
      "id": "u_001",
      "username": "admin",
      "avatar": "/assets/avatar/default.png",
      "role": "admin"
    }
  }
}
```

---

## POST /auth/logout

用户登出，注销当前 Token。

### 请求头

```
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "message": "success"
}
```

---

## GET /auth/profile

获取当前登录用户信息。

### 请求头

```
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "u_001",
    "username": "admin",
    "avatar": "/assets/avatar/default.png",
    "role": "admin",
    "created_at": "2026-07-01T10:00:00Z",
    "settings": {
      "theme": "dark",
      "language": "zh-CN",
      "default_device": "auto"
    }
  }
}
```
