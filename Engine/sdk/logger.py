"""
AI Studio Engine SDK - Logger.

Outputs JSON lines to stdout so the Go backend can parse
progress, results, and errors in real time.
"""

import json
import sys
import time
from typing import Any, Optional


def _emit(event_type: str, data: dict):
    """Write a single JSON line event to stdout."""
    payload = {
        "type": event_type,
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "data": data,
    }
    line = json.dumps(payload, ensure_ascii=False)
    sys.stdout.write(line + "\n")
    sys.stdout.flush()


def progress(epoch: int, total_epochs: int, loss: Optional[float] = None,
             metrics: Optional[dict] = None, step: str = ""):
    """Emit training progress."""
    data = {
        "epoch": epoch,
        "total_epochs": total_epochs,
    }
    if loss is not None:
        data["loss"] = round(loss, 6)
    if metrics:
        data["metrics"] = metrics
    if step:
        data["step"] = step
    _emit("progress", data)


def log(level: str, message: str, source: str = "python"):
    """Emit a log message."""
    _emit("log", {
        "level": level,
        "message": message,
        "source": source,
    })


def result(status: str, model_path: str = "", metrics: Optional[dict] = None,
           error: str = ""):
    """Emit the final result."""
    data: dict[str, Any] = {"status": status}
    if model_path:
        data["model_path"] = model_path
    if metrics:
        data["metrics"] = metrics
    if error:
        data["error"] = error
    _emit("result", data)


def error(message: str):
    """Emit a fatal error."""
    _emit("error", {"message": message})