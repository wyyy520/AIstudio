"""
AI Studio Engine - HTTP Server Mode
=====================================

提供持久化的 HTTP 服务，包含健康检查端点。
Launcher 启动 Engine 后通过 GET /health 检查服务状态。
Backend 可通过 HTTP API 分发任务给 Engine。

Usage:
    python server.py --port 8082
    python server.py --port 8082 --host 127.0.0.1

Endpoints:
    GET  /health  - 健康检查，返回 {"status": "ok"}
    POST /task    - 提交任务（JSON body），返回 {"status": "accepted"}
    GET  /env     - 环境信息（Python 版本、依赖状态等）
"""

import argparse
import json
import os
import platform
import sys
import threading
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from typing import Any

# 将 Engine 目录加入 Python 路径
_ENGINE_DIR = Path(__file__).resolve().parent
if str(_ENGINE_DIR) not in sys.path:
    sys.path.insert(0, str(_ENGINE_DIR))

from sdk.logger import log


# -----------------------------------------------------------------------------
# 全局状态
# -----------------------------------------------------------------------------

START_TIME = time.time()
TASK_QUEUE = []          # 任务队列（简单实现）
TASK_LOCK = threading.Lock()  # 保护任务队列的锁


class EngineHandler(BaseHTTPRequestHandler):
    """Engine HTTP 请求处理器"""

    # -------------------------------------------------------------------------
    # GET 请求路由
    # -------------------------------------------------------------------------

    def do_GET(self):
        """处理 GET 请求"""
        if self.path == "/health":
            self._handle_health()
        elif self.path == "/env":
            self._handle_env()
        else:
            self._json_response(404, {"error": "not found", "path": self.path})

    # -------------------------------------------------------------------------
    # POST 请求路由
    # -------------------------------------------------------------------------

    def do_POST(self):
        """处理 POST 请求"""
        if self.path == "/task":
            self._handle_task()
        else:
            self._json_response(404, {"error": "not found", "path": self.path})

    # -------------------------------------------------------------------------
    # 健康检查端点 /health
    # -------------------------------------------------------------------------

    def _handle_health(self):
        """返回服务健康状态"""
        uptime = round(time.time() - START_TIME, 2)
        with TASK_LOCK:
            pending_tasks = len(TASK_QUEUE)

        self._json_response(200, {
            "status": "ok",
            "service": "aistudio-engine",
            "uptime": uptime,
            "pending_tasks": pending_tasks,
            "python_version": sys.version.split()[0],
        })

    # -------------------------------------------------------------------------
    # 环境信息端点 /env
    # -------------------------------------------------------------------------

    def _handle_env(self):
        """返回运行环境信息"""
        env_info = {
            "python_version": sys.version.split()[0],
            "platform": platform.platform(),
            "engine_dir": str(_ENGINE_DIR),
            "executable": sys.executable,
        }

        # 尝试获取更详细的环境状态
        try:
            from runtime.env_detector import get_full_status
            env_info["detection"] = get_full_status()
        except Exception as e:
            env_info["detection_error"] = str(e)

        self._json_response(200, env_info)

    # -------------------------------------------------------------------------
    # 任务提交端点 /task
    # -------------------------------------------------------------------------

    def _handle_task(self):
        """接收任务并加入队列"""
        content_length = int(self.headers.get("Content-Length", 0))
        if content_length == 0:
            self._json_response(400, {"error": "empty request body"})
            return

        body = self.rfile.read(content_length)
        try:
            task = json.loads(body)
        except json.JSONDecodeError as e:
            self._json_response(400, {"error": f"invalid JSON: {e}"})
            return

        # 验证必需字段
        required = ["task_id", "plugin", "action"]
        missing = [k for k in required if k not in task]
        if missing:
            self._json_response(400, {
                "error": f"missing required fields: {missing}"
            })
            return

        # 加入任务队列
        task_id = task["task_id"]
        with TASK_LOCK:
            TASK_QUEUE.append(task)

        log("INFO", f"[engine-server] Task accepted: {task_id}")

        self._json_response(200, {
            "status": "accepted",
            "task_id": task_id,
            "queue_size": len(TASK_QUEUE),
        })

    # -------------------------------------------------------------------------
    # 辅助方法
    # -------------------------------------------------------------------------

    def _json_response(self, code: int, data: dict[str, Any]):
        """发送 JSON 响应"""
        self.send_response(code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(json.dumps(data, ensure_ascii=False).encode("utf-8"))

    def log_message(self, format: str, *args):
        """覆盖默认日志输出，使用 Engine SDK logger"""
        log("INFO", f"[engine-server] {format % args}")


# -----------------------------------------------------------------------------
# 主函数
# -----------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(
        description="AI Studio Engine - HTTP Server"
    )
    parser.add_argument(
        "--port", type=int, default=8082,
        help="HTTP 服务端口（默认 8082）"
    )
    parser.add_argument(
        "--host", type=str, default="127.0.0.1",
        help="监听地址（默认 127.0.0.1）"
    )
    args = parser.parse_args()

    # 打印启动信息
    log("INFO", f"[engine-server] 启动 Engine HTTP 服务")
    log("INFO", f"[engine-server] 监听地址: http://{args.host}:{args.port}")
    log("INFO", f"[engine-server] 健康检查: http://{args.host}:{args.port}/health")
    log("INFO", f"[engine-server] Python: {sys.version.split()[0]}")
    log("INFO", f"[engine-server] 工作目录: {_ENGINE_DIR}")

    # 创建并启动 HTTP 服务器
    server = HTTPServer((args.host, args.port), EngineHandler)

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        log("INFO", "[engine-server] 收到中断信号，正在停止...")
    finally:
        server.shutdown()
        server.server_close()
        log("INFO", "[engine-server] 服务已停止")


if __name__ == "__main__":
    main()
