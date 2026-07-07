# AI Studio

AI Studio - 智能工作流开发平台

## 项目结构

| 目录 | 技术栈 | 说明 |
|------|--------|------|
| Frontend/ | Vue3 + TypeScript + Tauri | 前端界面 |
| Backend/ | Go | 后端服务 |
| Engine/ | Python | AI 执行引擎 |
| Plugins/ | - | 插件系统 |
| Runtime/ | - | 运行时文件（可删除） |
| Storage/ | - | 用户数据 |
| Docs/ | - | 项目文档 |
| Scripts/ | - | 构建/部署脚本 |

## 快速开始

```bash
# 前端
cd Frontend && npm install && npm run dev

# 后端
cd Backend && go run cmd/main.go

# 引擎
cd Engine && pip install -r requirements.txt
```

## License

MIT
