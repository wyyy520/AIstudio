# 逻辑控制插件开发指南

## 1. 支持的逻辑插件

| 插件 | 目录 | 功能 |
|------|------|------|
| If | Plugins/Logic/If/ | 条件判断 |
| Else | Plugins/Logic/Else/ | 条件分支（与 If 配合） |
| Switch | Plugins/Logic/Switch/ | 多路分支 |
| Loop | Plugins/Logic/Loop/ | 循环执行 |
| Retry | Plugins/Logic/Retry/ | 失败重试 |

## 2. If 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | input | json | 任意输入数据 |
| 输出 | true | json | 条件为真时的输出 |
| 输出 | false | json | 条件为假时的输出 |

### 示例代码

```python
import ast
import operator

class IfPlugin:
    
    def setup(self, config):
        self.condition = config.get("condition", "")
    
    def execute(self, inputs, config):
        input_data = inputs.get("input", {})
        condition = config.get("condition", self.condition)
        
        # 安全的条件表达式求值
        result = self._evaluate_condition(condition, input_data)
        
        if result:
            return {"true": input_data}
        else:
            return {"false": input_data}
    
    def _evaluate_condition(self, condition, data):
        """
        支持的条件表达式:
        - len(detections.boxes) > 0
        - score > 0.8
        - status == "success"
        - count >= 5 and confidence > 0.9
        """
        # 构造安全求值上下文
        safe_ops = {
            "__builtins__": {},
            "len": len,
            "abs": abs,
            "min": min,
            "max": max,
            "sum": sum,
            "str": str,
            "int": int,
            "float": float,
        }
        
        # 将 data 的 key 作为变量注入
        context = {**safe_ops, **self._flatten_dict(data)}
        
        try:
            return bool(eval(condition, context))
        except Exception:
            return False
    
    def _flatten_dict(self, d, prefix=""):
        """将嵌套 dict 展平为点分隔变量"""
        result = {}
        for k, v in d.items():
            key = f"{prefix}{k}" if not prefix else f"{prefix}.{k}"
            if isinstance(v, dict):
                result.update(self._flatten_dict(v, key))
            else:
                result[k] = v
        return result
```

### 配置 Schema

```json
{
  "config_schema": {
    "condition": {
      "type": "text",
      "default": "len(boxes) > 0",
      "description": "条件表达式（Python 语法）"
    }
  }
}
```

## 3. Switch 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | input | json | 输入数据 |
| 输出 | case_0 | json | 匹配 case 0 |
| 输出 | case_1 | json | 匹配 case 1 |
| 输出 | case_2 | json | 匹配 case 2 |
| 输出 | default | json | 无匹配时输出 |

### 示例代码

```python
class SwitchPlugin:
    
    def setup(self, config):
        self.key = config.get("key", "")
        self.cases = config.get("cases", [])
    
    def execute(self, inputs, config):
        input_data = inputs.get("input", {})
        key = config.get("key", self.key)
        cases = config.get("cases", self.cases)
        
        # 获取比较值
        value = self._get_nested(input_data, key)
        
        result = {}
        matched = False
        for i, case in enumerate(cases):
            if value == case:
                result[f"case_{i}"] = input_data
                matched = True
                break
        
        if not matched:
            result["default"] = input_data
        
        return result
    
    def _get_nested(self, data, key):
        """支持点分隔的嵌套 key，如 detections.count"""
        keys = key.split(".")
        value = data
        for k in keys:
            if isinstance(value, dict):
                value = value.get(k)
            else:
                return None
        return value
```

## 4. Loop 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | input | json | 输入数据（list 时自动迭代） |
| 输入 | count | number | 循环次数（固定次数时使用） |
| 输出 | output | json | 每次循环的输出 |
| 输出 | completed | json | 循环完成后的汇总 |

### 示例代码

```python
class LoopPlugin:
    
    def setup(self, config):
        self.mode = config.get("mode", "iterate")  # iterate / count / while
    
    def execute(self, inputs, config):
        mode = config.get("mode", self.mode)
        
        if mode == "iterate":
            return self._iterate_mode(inputs, config)
        elif mode == "count":
            return self._count_mode(inputs, config)
        elif mode == "while":
            return self._while_mode(inputs, config)
    
    def _iterate_mode(self, inputs, config):
        """迭代列表中的每个元素"""
        items = inputs.get("input", [])
        if not isinstance(items, list):
            items = [items]
        
        outputs = []
        for i, item in enumerate(items):
            outputs.append({
                "index": i,
                "item": item
            })
        
        return {
            "output": outputs,  # 每次循环输出一个元素（引擎会逐个触发下游）
            "completed": {"total": len(items), "results": outputs}
        }
    
    def _count_mode(self, inputs, config):
        """固定次数循环"""
        count = int(inputs.get("count", config.get("count", 3)))
        
        outputs = []
        for i in range(count):
            outputs.append({"index": i, "item": inputs.get("input")})
        
        return {
            "output": outputs,
            "completed": {"total": count}
        }
```

## 5. Retry 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | input | json | 输入数据 |
| 输出 | output | json | 成功时的输出 |
| 输出 | failed | json | 重试耗尽后的输出 |

### 示例代码

```python
import time

class RetryPlugin:
    
    def setup(self, config):
        self.max_retries = config.get("max_retries", 3)
        self.delay = config.get("delay", 1.0)
        self.backoff = config.get("backoff", 2.0)  # 指数退避倍数
    
    def execute(self, inputs, config):
        max_retries = config.get("max_retries", self.max_retries)
        delay = config.get("delay", self.delay)
        backoff = config.get("backoff", self.backoff)
        
        # Retry 插件本身不执行逻辑，而是包装下游节点的执行
        # 引擎会在下游节点失败时自动重试
        return {
            "output": inputs.get("input"),
            "_retry_config": {
                "max_retries": max_retries,
                "delay": delay,
                "backoff": backoff
            }
        }
```

## 6. 工作流中的逻辑控制示意

```
[图像输入] → [YOLO检测] → [If: len(boxes)>0]
                              ├── true  → [保存结果]
                              └── false → [日志记录]
```

```
[数据集] → [Loop: iterate] → [预处理] → [推理] → [汇总结果]
```
