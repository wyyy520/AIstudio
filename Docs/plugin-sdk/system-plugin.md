# 系统插件开发指南

## 1. 支持的系统插件

| 插件 | 目录 | 功能 |
|------|------|------|
| Python | Plugins/System/Python/ | 执行 Python 脚本 |
| Git | Plugins/System/Git/ | Git 操作 |
| CUDA | Plugins/System/CUDA/ | GPU 状态查询 |
| Download | Plugins/System/Download/ | 文件下载 |
| Terminal | Plugins/System/Terminal/ | 终端命令执行 |
| Environment | Plugins/System/Environment/ | 环境变量管理 |

## 2. Python 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | script | text | Python 脚本内容 |
| 输入 | args | json | 脚本参数 |
| 输入 | file_path | file | 脚本文件路径（与 script 二选一） |
| 输出 | stdout | text | 标准输出 |
| 输出 | stderr | text | 标准错误 |
| 输出 | exit_code | number | 退出码 |
| 输出 | result | json | 解析后的结果 |

### 示例代码

```python
import subprocess
import sys
import json
import tempfile
import os

class PythonPlugin:
    
    def setup(self, config):
        self.python_path = config.get("python_path", "python")
        self.timeout = config.get("timeout", 300)
    
    def execute(self, inputs, config):
        script = inputs.get("script", "")
        file_path = inputs.get("file_path", "")
        args = inputs.get("args", {})
        
        # 写入临时脚本文件
        if file_path:
            script_path = file_path
        else:
            with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
                f.write(script)
                script_path = f.name
        
        # 构造参数
        cmd = [self.python_path, script_path]
        for k, v in args.items():
            cmd.extend([f"--{k}", str(v)])
        
        # 执行
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=self.timeout
            )
            
            # 尝试解析 stdout 为 JSON
            result_json = None
            try:
                result_json = json.loads(result.stdout)
            except (json.JSONDecodeError, TypeError):
                pass
            
            return {
                "stdout": result.stdout,
                "stderr": result.stderr,
                "exit_code": result.returncode,
                "result": result_json
            }
        except subprocess.TimeoutExpired:
            return {
                "stdout": "",
                "stderr": f"Script timed out after {self.timeout}s",
                "exit_code": -1,
                "result": None
            }
        finally:
            if not file_path and os.path.exists(script_path):
                os.unlink(script_path)
```

## 3. Git 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | command | text | Git 命令（clone/commit/push/pull/status） |
| 输入 | repo_path | file | 仓库路径 |
| 输入 | message | text | commit 信息 |
| 输出 | output | text | 命令输出 |
| 输出 | exit_code | number | 退出码 |

### 示例代码

```python
import subprocess
import os

class GitPlugin:
    
    def execute(self, inputs, config):
        command = inputs.get("command", "")
        repo_path = inputs.get("repo_path", ".")
        message = inputs.get("message", "")
        
        if command == "clone":
            url = inputs.get("url", "")
            cmd = ["git", "clone", url, repo_path]
        elif command == "commit":
            cmd = ["git", "-C", repo_path, "commit", "-m", message]
        elif command == "push":
            cmd = ["git", "-C", repo_path, "push"]
        elif command == "status":
            cmd = ["git", "-C", repo_path, "status", "--porcelain"]
        else:
            cmd = ["git", "-C", repo_path, command]
        
        result = subprocess.run(cmd, capture_output=True, text=True)
        
        return {
            "output": result.stdout or result.stderr,
            "exit_code": result.returncode
        }
```

## 4. CUDA 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输出 | status | json | GPU 状态信息 |
| 输出 | memory | json | 显存使用情况 |
| 输出 | devices | json | 可用设备列表 |

### 示例代码

```python
import torch

class CUDAPlugin:
    
    def execute(self, inputs, config):
        if not torch.cuda.is_available():
            return {
                "status": {"available": False, "device": "cpu"},
                "memory": {},
                "devices": []
            }
        
        devices = []
        for i in range(torch.cuda.device_count()):
            props = torch.cuda.get_device_properties(i)
            memory_allocated = torch.cuda.memory_allocated(i)
            memory_reserved = torch.cuda.memory_reserved(i)
            
            devices.append({
                "id": i,
                "name": props.name,
                "total_memory_mb": props.total_memory / (1024 * 1024),
                "allocated_mb": memory_allocated / (1024 * 1024),
                "reserved_mb": memory_reserved / (1024 * 1024),
                "free_mb": (props.total_memory - memory_reserved) / (1024 * 1024)
            })
        
        return {
            "status": {"available": True, "device_count": len(devices)},
            "memory": {"devices": devices},
            "devices": [d["name"] for d in devices]
        }
```

## 5. Download 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | url | text | 下载地址 |
| 输入 | output_path | file | 保存路径 |
| 输出 | file_path | file | 下载后的文件路径 |
| 输出 | size_bytes | number | 文件大小 |

### 示例代码

```python
import requests
import os

class DownloadPlugin:
    
    def setup(self, config):
        self.timeout = config.get("timeout", 60)
    
    def execute(self, inputs, config):
        url = inputs.get("url", "")
        output_path = inputs.get("output_path", "./downloads/")
        
        os.makedirs(output_path, exist_ok=True)
        filename = os.path.basename(url.split("?")[0])
        full_path = os.path.join(output_path, filename)
        
        response = requests.get(url, timeout=self.timeout, stream=True)
        response.raise_for_status()
        
        with open(full_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        
        size = os.path.getsize(full_path)
        
        return {
            "file_path": full_path,
            "size_bytes": size,
            "filename": filename
        }
```

## 6. Environment 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | action | text | 操作（get/set/list/clear） |
| 输入 | key | text | 变量名 |
| 输入 | value | text | 变量值 |
| 输出 | result | json | 操作结果 |

### 示例代码

```python
import os

class EnvironmentPlugin:
    
    def execute(self, inputs, config):
        action = inputs.get("action", "list")
        key = inputs.get("key", "")
        
        if action == "get":
            value = os.environ.get(key, "")
            return {"result": {"key": key, "value": value}}
        
        elif action == "set":
            value = inputs.get("value", "")
            os.environ[key] = value
            return {"result": {"status": "set", "key": key}}
        
        elif action == "list":
            env_vars = {k: v for k, v in os.environ.items() 
                       if not k.startswith("_")}
            return {"result": {"variables": env_vars}}
        
        elif action == "clear":
            if key:
                os.environ.pop(key, None)
            return {"result": {"status": "cleared"}}
        
        return {"result": {}}
```
