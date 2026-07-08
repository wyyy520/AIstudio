"""
AI Studio Engine - Python Runner.

Unified entry point for the Go backend to execute AI tasks.

Usage:
    python runner.py --task task.json

The runner reads a task.json, dispatches to the appropriate plugin
handler, and outputs JSON lines to stdout for the Go backend to
parse progress, logs, and results in real time.

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

import argparse
import json
import sys
import traceback
from pathlib import Path
from typing import Any

# Ensure the Engine directory is on the Python path
_ENGINE_DIR = Path(__file__).resolve().parent
if str(_ENGINE_DIR) not in sys.path:
    sys.path.insert(0, str(_ENGINE_DIR))

from sdk.logger import log, error as emit_error


# Plugin registry: maps plugin name to handler module
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
        emit_error(f"Task file not found: {task_path}")
        sys.exit(1)

    with open(path, "r", encoding="utf-8") as f:
        task = json.load(f)

    required = ["task_id", "plugin", "action"]
    missing = [k for k in required if k not in task]
    if missing:
        emit_error(f"Missing required fields in task.json: {missing}")
        sys.exit(1)

    return task


def dispatch(plugin: str, action: str, params: dict[str, Any]):
    """Dispatch to the appropriate plugin handler."""
    plugin_config = PLUGIN_REGISTRY.get(plugin)
    if not plugin_config:
        emit_error(f"Unknown plugin: {plugin}. Supported: {list(PLUGIN_REGISTRY.keys())}")
        sys.exit(1)

    handler_module = plugin_config.get(action)
    if not handler_module:
        emit_error(f"Unknown action '{action}' for plugin '{plugin}'. "
                   f"Supported: {list(plugin_config.keys())}")
        sys.exit(1)

    log("INFO", f"Dispatching: plugin={plugin}, action={action}, "
                f"handler={handler_module}")

    # Import the handler module
    try:
        mod = __import__(handler_module, fromlist=["run_train", "run_predict"])
    except ImportError as e:
        emit_error(f"Failed to import handler {handler_module}: {e}")
        sys.exit(1)

    # Build config from params
    if action == "train":
        from vision.yolo.config import YOLOTrainConfig
        config = YOLOTrainConfig.from_dict(params)
        run_func = getattr(mod, "run_train", None)
    elif action == "predict":
        from vision.yolo.config import YOLOPredictConfig
        config = YOLOPredictConfig.from_dict(params)
        run_func = getattr(mod, "run_predict", None)
    else:
        emit_error(f"No config mapping for action: {action}")
        sys.exit(1)

    if run_func is None:
        emit_error(f"Handler {handler_module} has no run function for action '{action}'")
        sys.exit(1)

    # Execute
    run_func(config)


def main():
    parser = argparse.ArgumentParser(
        description="AI Studio Engine - Python Runner"
    )
    parser.add_argument(
        "--task", type=str, required=True,
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
        sys.exit(1)


if __name__ == "__main__":
    main()