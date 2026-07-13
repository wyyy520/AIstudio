# 项目管理接口

## GET /projects

获取项目列表。

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 搜索关键词 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页条数 |

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 5,
    "items": [
      {
        "id": "proj_001",
        "name": "智能交通检测",
        "description": "基于视觉的车辆检测与追踪",
        "path": "/storage/projects/proj_001",
        "workflow_count": 3,
        "created_at": "2026-07-01T10:00:00Z",
        "updated_at": "2026-07-07T14:00:00Z"
      }
    ]
  }
}
```

---

## POST /projects

创建项目。

### 请求体

```json
{
  "name": "智能交通检测",
  "description": "基于视觉的车辆检测与追踪",
  "template": "blank"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 项目名称 |
| description | string | 否 | 项目描述 |
| template | string | 否 | 模板（blank / vision / nlp / simulation） |

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "proj_002",
    "name": "智能交通检测",
    "path": "/storage/projects/proj_002",
    "created_at": "2026-07-07T14:52:00Z"
  }
}
```

---

## GET /projects/:id

获取项目详情。

### 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "proj_001",
    "name": "智能交通检测",
    "description": "基于视觉的车辆检测与追踪",
    "path": "/storage/projects/proj_001",
    "settings": {
      "python_env": "python3.11",
      "cuda_device": 0,
      "work_dir": "/storage/projects/proj_001/workspace"
    },
    "stats": {
      "workflow_count": 3,
      "dataset_count": 5,
      "model_count": 2,
      "storage_used_mb": 1024
    },
    "created_at": "2026-07-01T10:00:00Z",
    "updated_at": "2026-07-07T14:00:00Z"
  }
}
```

---

## PUT /projects/:id

更新项目信息。

### 请求体

```json
{
  "name": "智能交通检测 v2",
  "description": "更新后的描述",
  "settings": {
    "cuda_device": 1
  }
}
```

---

## DELETE /projects/:id

删除项目（同时删除关联的工作流和数据）。

### 响应

```json
{
  "code": 0,
  "message": "success"
}
```
