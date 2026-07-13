"""
AIStudio Integration Demo - Mock YOLO Training

This script simulates the complete execution chain:
  Frontend → Backend → Task Manager → Workflow Engine → Plugin Manager → Python Engine

It demonstrates the full JSON-line protocol for progress, logs, and results.

Usage:
    python demo.py
    python demo.py --task Engine/examples/task_demo.json
"""

import argparse
import json
import sys
import time
from pathlib import Path
from typing import Any


def emit_progress(epoch: int, total_epochs: int, loss: float, accuracy: float = 0):
    """Emit a progress event to stdout (JSON line)."""
    event = {
        "type": "progress",
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "data": {
            "epoch": epoch,
            "total_epochs": total_epochs,
            "loss": round(loss, 4),
            "accuracy": round(accuracy, 4),
            "progress_percent": round(epoch / total_epochs * 100, 1),
        },
    }
    print(json.dumps(event), flush=True)


def emit_log(level: str, message: str):
    """Emit a log event to stdout (JSON line)."""
    event = {
        "type": "log",
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "data": {
            "level": level,
            "message": message,
        },
    }
    print(json.dumps(event), flush=True)


def emit_result(status: str, model_path: str = "", metrics: dict | None = None, error: str = ""):
    """Emit the final result event to stdout (JSON line)."""
    event = {
        "type": "result",
        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "data": {
            "status": status,
            "model_path": model_path,
            "metrics": metrics or {},
            "error": error,
        },
    }
    print(json.dumps(event), flush=True)


def mock_train(task: dict[str, Any]):
    """
    Simulate a YOLO training process.

    The full chain:
    1. Frontend clicks "Run" on a YOLO workflow
    2. Backend creates a Task via Task Manager
    3. Task Manager dispatches to workflow.TaskHandler
    4. Workflow Engine executes "YOLO Train" node
    5. Plugin Manager resolves "YOLO Train" → Plugins/Vision/YOLO/train.py
    6. Python Engine (this script) executes the mock training
    7. Progress, logs, and results are streamed back to Go via stdout
    8. Go forwards events to Frontend via WebSocket
    """
    task_id = task.get("task_id", "unknown")
    params = task.get("params", {})
    epochs = params.get("epochs", 5)
    dataset = params.get("dataset", "unknown")
    model = params.get("model", "yolov8n.pt")
    output_dir = params.get("output_dir", "Storage/models/")

    emit_log("INFO", "=" * 60)
    emit_log("INFO", "AIStudio Integration Demo - YOLO Mock Training")
    emit_log("INFO", "=" * 60)
    emit_log("INFO", f"Task ID:    {task_id}")
    emit_log("INFO", f"Plugin:     {task.get('plugin')}")
    emit_log("INFO", f"Action:     {task.get('action')}")
    emit_log("INFO", f"Dataset:    {dataset}")
    emit_log("INFO", f"Model:      {model}")
    emit_log("INFO", f"Epochs:     {epochs}")
    emit_log("INFO", f"Output:     {output_dir}")
    emit_log("INFO", "-" * 60)

    # Step 1: Environment Check
    emit_log("INFO", "[1/5] Checking environment...")
    time.sleep(0.5)
    emit_log("INFO", "  Python: OK (3.11.0)")
    emit_log("INFO", "  PyTorch: 2.0.1 (CPU)")
    emit_log("INFO", "  Ultralytics: 8.0.100")
    emit_log("INFO", "  CUDA: Not available (using CPU)")
    time.sleep(0.3)

    # Step 2: Load Dataset
    emit_log("INFO", "[2/5] Loading dataset...")
    time.sleep(0.5)
    emit_log("INFO", f"  Dataset: {dataset}")
    emit_log("INFO", "  Images: 128 (train), 32 (val)")
    emit_log("INFO", "  Classes: 80 (COCO)")
    time.sleep(0.3)

    # Step 3: Build Model
    emit_log("INFO", "[3/5] Building model...")
    time.sleep(0.5)
    emit_log("INFO", f"  Model: {model}")
    emit_log("INFO", "  Parameters: 3,011,238")
    emit_log("INFO", "  Layers: 225")
    time.sleep(0.3)

    # Step 4: Training Loop
    emit_log("INFO", "[4/5] Training...")
    emit_log("INFO", f"  Epochs: {epochs}")
    emit_log("INFO", f"  Batch size: {params.get('batch', 8)}")
    emit_log("INFO", f"  Image size: {params.get('img_size', 640)}")
    emit_log("INFO", "")

    for epoch in range(1, epochs + 1):
        time.sleep(0.8)  # Simulate training time

        # Simulate decreasing loss
        loss = 2.5 / (epoch + 0.5) + 0.1
        accuracy = 0.3 + epoch * 0.12

        emit_progress(epoch, epochs, loss, accuracy)

        if epoch == 1:
            emit_log("INFO", f"  Epoch {epoch}/{epochs}: loss={loss:.4f}, acc={accuracy:.2%} (warming up)")
        elif epoch == epochs:
            emit_log("INFO", f"  Epoch {epoch}/{epochs}: loss={loss:.4f}, acc={accuracy:.2%} (final)")
        else:
            emit_log("INFO", f"  Epoch {epoch}/{epochs}: loss={loss:.4f}, acc={accuracy:.2%}")

    emit_log("INFO", "")

    # Step 5: Save Model
    emit_log("INFO", "[5/5] Saving model...")
    time.sleep(0.5)
    model_path = f"{output_dir}/best.pt"
    emit_log("INFO", f"  Model saved: {model_path}")
    emit_log("INFO", f"  Size: 12.3 MB")
    time.sleep(0.3)

    # Final Result
    metrics = {
        "mAP50": 0.72 + epochs * 0.02,
        "mAP50-95": 0.45 + epochs * 0.02,
        "precision": 0.78 + epochs * 0.01,
        "recall": 0.68 + epochs * 0.02,
        "train_loss": round(2.5 / (epochs + 0.5) + 0.1, 4),
        "val_loss": round(3.0 / (epochs + 0.5) + 0.15, 4),
    }

    emit_log("INFO", "=" * 60)
    emit_log("INFO", "Training Complete")
    emit_log("INFO", f"  mAP@50:    {metrics['mAP50']:.3f}")
    emit_log("INFO", f"  mAP@50-95: {metrics['mAP50-95']:.3f}")
    emit_log("INFO", f"  Precision: {metrics['precision']:.3f}")
    emit_log("INFO", f"  Recall:    {metrics['recall']:.3f}")
    emit_log("INFO", "=" * 60)

    emit_result("success", model_path, metrics)


def main():
    parser = argparse.ArgumentParser(description="AIStudio Integration Demo")
    parser.add_argument("--task", type=str, default="Engine/examples/task_demo.json",
                        help="Path to task.json file")
    args = parser.parse_args()

    task_path = Path(args.task)
    if not task_path.exists():
        # Try relative to project root
        project_root = Path(__file__).resolve().parent.parent
        task_path = project_root / args.task
        if not task_path.exists():
            emit_log("ERROR", f"Task file not found: {args.task}")
            emit_result("failed", error=f"Task file not found: {args.task}")
            sys.exit(1)

    with open(task_path, "r", encoding="utf-8") as f:
        task = json.load(f)

    try:
        mock_train(task)
    except Exception as e:
        emit_log("ERROR", f"Demo failed: {e}")
        import traceback
        traceback.print_exc(file=sys.stderr)
        emit_result("failed", error=str(e))
        sys.exit(1)


if __name__ == "__main__":
    main()