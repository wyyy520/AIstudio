"""
AI Studio Engine - Python Runner.

Unified entry point to execute AI tasks.

Usage:
    python runner.py --task task.json
    python runner.py --env-check

The runner reads a task.json, dispatches to the appropriate plugin
handler, and outputs JSON lines to stdout for consumption.

Task JSON format:
{
    "task_id": "uuid",
    "plugin": "yolo",
    "action": "train",
    "params": {
        "dataset": "/path/to/dataset",
        "model": "yolov8n.pt",
        "epochs": 100,
        "batch": 16,
        "img_size": 640,
        "device": "cuda",
        "output_dir": "Storage/models/"
    }
}
"""

from __future__ import annotations

import argparse
import json
import sys
import traceback
from pathlib import Path
from typing import Any

_ENGINE_DIR = Path(__file__).resolve().parent
if str(_ENGINE_DIR) not in sys.path:
    sys.path.insert(0, str(_ENGINE_DIR))

from sdk.logger import log, error as emit_error


PLUGIN_REGISTRY = {
    "yolo": {
        "train": "vision.yolo.train",
        "predict": "vision.yolo.predict",
    },
}


def load_task(task_path: str) -> dict[str, Any]:
    """Load and validate the task.json file."""
    path = Path(task_path)
    if not path.exists():
        raise RuntimeError(f"Task file not found: {task_path}")

    with open(path, "r", encoding="utf-8") as f:
        task = json.load(f)

    required = ["task_id", "plugin", "action"]
    missing = [k for k in required if k not in task]
    if missing:
        raise RuntimeError(f"Missing required fields in task.json: {missing}")

    return task


def dispatch(plugin: str, action: str, params: dict[str, Any]):
    """Dispatch to the appropriate plugin handler."""
    plugin_config = PLUGIN_REGISTRY.get(plugin)
    if not plugin_config:
        raise RuntimeError(f"Unknown plugin: {plugin}. Supported: {list(PLUGIN_REGISTRY.keys())}")

    handler_module = plugin_config.get(action)
    if not handler_module:
        raise RuntimeError(f"Unknown action '{action}' for plugin '{plugin}'. "
                           f"Supported: {list(plugin_config.keys())}")

    log("INFO", f"Dispatching: plugin={plugin}, action={action}, "
                f"handler={handler_module}")

    try:
        mod = __import__(handler_module, fromlist=["run_train", "run_predict"])
    except ImportError as e:
        raise RuntimeError(f"Failed to import handler {handler_module}: {e}")

    if action == "train":
        from vision.yolo.config import YOLOTrainConfig
        config = YOLOTrainConfig.from_dict(params)
        run_func = getattr(mod, "run_train", None)
    elif action == "predict":
        from vision.yolo.config import YOLOPredictConfig
        config = YOLOPredictConfig.from_dict(params)
        run_func = getattr(mod, "run_predict", None)
    else:
        raise RuntimeError(f"No config mapping for action: {action}")

    if run_func is None:
        raise RuntimeError(f"Handler {handler_module} has no run function for action '{action}'")

    run_func(config)


def main():
    parser = argparse.ArgumentParser(
        description="AI Studio Engine - Python Runner"
    )
    parser.add_argument(
        "--task", type=str, default=None,
        help="Path to task.json file"
    )
    parser.add_argument(
        "--env-check", action="store_true",
        help="Run environment detection and exit"
    )
    args = parser.parse_args()

    if args.env_check:
        from runtime.env_detector import get_full_status
        status = get_full_status()
        json.dump(status, sys.stdout, indent=2, ensure_ascii=False)
        sys.stdout.write("\n")
        return

    if not args.task:
        parser.print_help()
        raise RuntimeError("No task file specified. Use --task <path>")

    task = load_task(args.task)

    task_id = task["task_id"]
    plugin = task["plugin"]
    action = task["action"]
    params = task.get("params", {})

    log("INFO", f"Task started: task_id={task_id}, plugin={plugin}, "
                f"action={action}")

    try:
        dispatch(plugin, action, params)
    except Exception as e:
        log("ERROR", f"Task execution failed: {e}")
        traceback.print_exc(file=sys.stderr)
        emit_error(str(e))
        raise RuntimeError(str(e))


if __name__ == "__main__":
    main()
