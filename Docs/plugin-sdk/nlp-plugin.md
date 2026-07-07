# NLP 插件开发指南

## 1. 支持的 NLP 插件类型

| 插件 | 目录 | 功能 |
|------|------|------|
| Transformer | Plugins/NLP/Transformer/ | 文本分类、NER、情感分析 |
| LLM | Plugins/NLP/LLM/ | 大语言模型对话 |

## 2. Transformer 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | text | text | 输入文本 |
| 输入 | task | text | 任务类型（classify/ner/qa） |
| 输出 | result | json | 处理结果 |

### 示例代码

```python
from transformers import pipeline

class TransformerPlugin:
    
    def setup(self, config):
        model_name = config.get("model", "bert-base-chinese")
        task = config.get("task", "text-classification")
        self.pipe = pipeline(task, model=model_name)
    
    def execute(self, inputs, config):
        text = inputs["text"]
        task = inputs.get("task", "classification")
        
        if task == "classification":
            result = self.pipe(text)
            return {"result": {"label": result[0]["label"], "score": result[0]["score"]}}
        
        elif task == "ner":
            result = self.pipe(text)
            entities = [
                {"entity": r["entity"], "word": r["word"], "start": r["start"], "end": r["end"], "score": r["score"]}
                for r in result
            ]
            return {"result": {"entities": entities}}
        
        elif task == "qa":
            question = inputs.get("question", "")
            context = inputs["text"]
            result = self.pipe(question=question, context=context)
            return {"result": {"answer": result["answer"], "score": result["score"]}}
        
        return {"result": {}}
```

## 3. LLM 插件

### 端口定义

| 方向 | 端口名 | 类型 | 说明 |
|------|--------|------|------|
| 输入 | prompt | text | 用户提示词 |
| 输入 | context | text | 上下文（可选） |
| 输入 | max_tokens | number | 最大生成长度 |
| 输出 | response | text | 模型回复 |
| 输出 | usage | json | Token 用量 |

### 示例代码

```python
from transformers import AutoModelForCausalLM, AutoTokenizer
import torch
from typing import Generator

class LLMPlugin:
    
    def setup(self, config):
        model_name = config.get("model", "Qwen/Qwen2-7B")
        device = "cuda" if torch.cuda.is_available() else "cpu"
        
        self.tokenizer = AutoTokenizer.from_pretrained(model_name, trust_remote_code=True)
        self.model = AutoModelForCausalLM.from_pretrained(
            model_name,
            torch_dtype=torch.float16 if device == "cuda" else torch.float32,
            device_map=device,
            trust_remote_code=True
        )
        self.device = device
    
    def execute(self, inputs, config):
        prompt = inputs["prompt"]
        context = inputs.get("context", "")
        max_tokens = inputs.get("max_tokens", 512)
        
        full_prompt = f"{context}\n{prompt}" if context else prompt
        
        inputs_ids = self.tokenizer(full_prompt, return_tensors="pt").to(self.device)
        output = self.model.generate(
            **inputs_ids,
            max_new_tokens=max_tokens,
            temperature=config.get("temperature", 0.7),
            top_p=config.get("top_p", 0.9),
            do_sample=True
        )
        
        response = self.tokenizer.decode(output[0][inputs_ids["input_ids"].shape[1]:], skip_special_tokens=True)
        
        return {
            "response": response,
            "usage": {
                "prompt_tokens": inputs_ids["input_ids"].shape[1],
                "completion_tokens": output.shape[1] - inputs_ids["input_ids"].shape[1]
            }
        }
    
    def execute_stream(self, inputs, config) -> Generator[str, None, None]:
        """流式输出"""
        prompt = inputs["prompt"]
        context = inputs.get("context", "")
        full_prompt = f"{context}\n{prompt}" if context else prompt
        
        inputs_ids = self.tokenizer(full_prompt, return_tensors="pt").to(self.device)
        
        from transformers import TextIteratorStreamer
        from threading import Thread
        
        streamer = TextIteratorStreamer(
            self.tokenizer,
            skip_prompt=True,
            skip_special_tokens=True
        )
        
        generation_kwargs = {
            **inputs_ids,
            "max_new_tokens": inputs.get("max_tokens", 512),
            "streamer": streamer,
            "temperature": config.get("temperature", 0.7),
            "do_sample": True
        }
        
        thread = Thread(target=self.model.generate, kwargs=generation_kwargs)
        thread.start()
        
        for text in streamer:
            yield text
```

## 4. 配置 Schema

```json
{
  "config_schema": {
    "model": {
      "type": "string",
      "default": "Qwen/Qwen2-7B",
      "description": "模型名称（HuggingFace）"
    },
    "temperature": {
      "type": "number",
      "default": 0.7,
      "min": 0,
      "max": 2,
      "step": 0.1,
      "description": "温度参数（越高越随机）"
    },
    "top_p": {
      "type": "number",
      "default": 0.9,
      "min": 0,
      "max": 1,
      "step": 0.05,
      "description": "Top-P 采样阈值"
    },
    "max_tokens": {
      "type": "number",
      "default": 512,
      "min": 1,
      "max": 8192,
      "description": "最大生成 Token 数"
    }
  }
}
```

## 5. 本地模型 vs API 模型

| 方式 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| 本地模型 | 数据不出本地、无延迟 | 需要 GPU、加载慢 | 隐私要求高、离线场景 |
| API 调用 | 无需 GPU、模型更新快 | 数据外传、有延迟 | 快速验证、无 GPU |
| 本地 API | 速度快、可复用 | 需要部署服务 | 多用户共享场景 |

### API 调用方式示例

```python
import requests

class LLMPlugin:
    
    def setup(self, config):
        self.api_url = config.get("api_url", "http://localhost:8000/v1")
        self.api_key = config.get("api_key", "")
    
    def execute(self, inputs, config):
        response = requests.post(
            f"{self.api_url}/chat/completions",
            headers={"Authorization": f"Bearer {self.api_key}"},
            json={
                "model": config.get("model", "gpt-3.5-turbo"),
                "messages": [{"role": "user", "content": inputs["prompt"]}],
                "max_tokens": inputs.get("max_tokens", 512),
                "temperature": config.get("temperature", 0.7)
            }
        )
        data = response.json()
        return {
            "response": data["choices"][0]["message"]["content"],
            "usage": data["usage"]
        }
```
