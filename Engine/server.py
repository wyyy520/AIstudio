"""
AI Studio Engine - HTTP Server Mode

Thin HTTP wrapper around runner.py.

Usage:
    python server.py --port 8082
    python server.py --port 8082 --host 0.0.0.0

Environment variables:
    ENGINE_HOST - Bind address (default 0.0.0.0)
    ENGINE_PORT - HTTP port (default 8082)

Endpoints:
    GET  /health      - Health check: {"status": "ok"}
    GET  /env         - Environment info (Python, PyTorch)
    POST /task        - Submit task (JSON body), dispatches via runner
"""

from __future__ import annotations

import argparse
import json
import os
import platform
import sys
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from typing import Any, Optional

_ENGINE_DIR = Path(__file__).resolve().parent
if str(_ENGINE_DIR) not in sys.path:
    sys.path.insert(0, str(_ENGINE_DIR))

from sdk.logger import log

START_TIME = time.time()


class EngineHandler(BaseHTTPRequestHandler):

    def do_GET(self):
        if self.path == "/health":
            self._handle_health()
        elif self.path == "/env":
            self._handle_env()
        else:
            self._json_response(404, {"error": "not found", "path": self.path})

    def do_POST(self):
        if self.path == "/task":
            self._handle_task()
        else:
            self._json_response(404, {"error": "not found", "path": self.path})

    def _handle_health(self):
        uptime = round(time.time() - START_TIME, 2)
        self._json_response(200, {
            "status": "ok",
            "service": "aistudio-engine",
            "uptime": uptime,
            "python_version": sys.version.split()[0],
        })

    def _handle_env(self):
        env_info = {
            "python_version": sys.version.split()[0],
            "platform": platform.platform(),
            "engine_dir": str(_ENGINE_DIR),
            "executable": sys.executable,
        }
        try:
            from runtime.env_detector import get_full_status
            env_info["detection"] = get_full_status()
        except Exception as e:
            env_info["detection_error"] = str(e)
        self._json_response(200, env_info)

    def _handle_task(self):
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

        required = ["task_id", "plugin", "action"]
        missing = [k for k in required if k not in task]
        if missing:
            self._json_response(400, {
                "error": f"missing required fields: {missing}"
            })
            return

        task_id = task["task_id"]
        log("INFO", f"[engine-server] Task accepted: {task_id}")

        from runner import dispatch
        try:
            dispatch(task["plugin"], task["action"], task.get("params", {}))
            self._json_response(200, {
                "status": "completed",
                "task_id": task_id,
            })
        except Exception as e:
            self._json_response(500, {
                "status": "failed",
                "task_id": task_id,
                "error": str(e),
            })

    def log_message(self, format: str, *args):
        log("INFO", f"[engine-server] {format % args}")


def main():
    default_host = os.environ.get("ENGINE_HOST", "0.0.0.0")
    default_port = int(os.environ.get("ENGINE_PORT", "8082"))

    parser = argparse.ArgumentParser(
        description="AI Studio Engine - HTTP Server"
    )
    parser.add_argument(
        "--port", type=int, default=default_port,
        help=f"HTTP port (default {default_port}, env: ENGINE_PORT)"
    )
    parser.add_argument(
        "--host", type=str, default=default_host,
        help=f"Bind address (default {default_host}, env: ENGINE_HOST)"
    )
    args = parser.parse_args()

    log("INFO", f"[engine-server] Starting Engine HTTP server on "
                f"http://{args.host}:{args.port}")
    log("INFO", f"[engine-server] Python: {sys.version.split()[0]}")

    server = HTTPServer((args.host, args.port), EngineHandler)

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        log("INFO", "[engine-server] Stopping...")
    finally:
        server.shutdown()
        server.server_close()
        log("INFO", "[engine-server] Stopped")


if __name__ == "__main__":
    main()