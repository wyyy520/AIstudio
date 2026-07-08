"""
YOLO Training Module.

Supports YOLOv8 training via ultralytics.
Outputs training progress as JSON lines to stdout for the Go backend.
"""

import os
import sys
import traceback
from pathlib import Path

from vision.yolo.config import YOLOTrainConfig, YOLO_MODELS
from sdk.logger import progress, log, result, error


def run_train(config: YOLOTrainConfig):
    """
    Run YOLO training with the given configuration.

    Args:
        config: YOLOTrainConfig with dataset, model, epochs, etc.
    """
    log("INFO", f"Starting YOLO training: model={config.model}, "
                f"dataset={config.dataset}, epochs={config.epochs}, "
                f"device={config.device}, batch={config.batch}")

    # Validate dataset
    if not config.dataset:
        error("Dataset path is required")
        sys.exit(1)

    dataset_path = Path(config.dataset)
    if not dataset_path.exists():
        error(f"Dataset path does not exist: {config.dataset}")
        sys.exit(1)

    # Check for data.yaml in dataset directory
    data_yaml = dataset_path / "data.yaml"
    if not data_yaml.exists():
        data_yaml = dataset_path / "dataset.yaml"
    if not data_yaml.exists():
        error(f"No data.yaml or dataset.yaml found in {config.dataset}")
        sys.exit(1)

    config.dataset = str(data_yaml)

    # Create output directory
    output_dir = Path(config.output_dir) if config.output_dir else Path("Storage/models")
    output_dir.mkdir(parents=True, exist_ok=True)
    log("INFO", f"Output directory: {output_dir}")

    # Check device availability
    device = config.device
    if device == "cuda":
        try:
            import torch
            if not torch.cuda.is_available():
                log("WARN", "CUDA not available, falling back to CPU")
                device = "cpu"
                config.device = "cpu"
        except ImportError:
            log("WARN", "PyTorch not installed, falling back to CPU")
            device = "cpu"
            config.device = "cpu"

    log("INFO", f"Using device: {device}")

    # Import ultralytics
    try:
        from ultralytics import YOLO
    except ImportError as e:
        error(f"ultralytics not installed: {e}")
        error("Run: pip install ultralytics")
        sys.exit(1)

    # Load model
    model_path = config.model
    if config.pretrained and not os.path.exists(model_path):
        # Download pretrained model
        log("INFO", f"Downloading pretrained model: {model_path}")
        model = YOLO(model_path)
    else:
        model = YOLO(model_path)

    log("INFO", f"Model loaded: {model_path}")

    # Build training arguments
    train_args = config.to_ultralytics_args()

    # Add custom callback for progress reporting
    def on_train_epoch_end(trainer):
        """Callback after each training epoch."""
        epoch = trainer.epoch + 1
        total = trainer.epochs
        metrics = trainer.metrics

        loss = None
        if hasattr(trainer, "loss_items") and trainer.loss_items is not None:
            loss_items = trainer.loss_items
            if hasattr(loss_items, "tolist"):
                loss_items = loss_items.tolist()
            if isinstance(loss_items, (list, tuple)) and len(loss_items) > 0:
                loss = float(loss_items[0])

        # Extract key metrics
        metric_dict = {}
        if metrics:
            for k, v in metrics.items():
                if isinstance(v, (int, float)):
                    metric_dict[k] = round(float(v), 4)
                elif hasattr(v, "tolist"):
                    metric_dict[k] = v.tolist()
                elif hasattr(v, "item"):
                    metric_dict[k] = float(v.item())

        progress(epoch, total, loss=loss, metrics=metric_dict,
                 step=f"Epoch {epoch}/{total}")

    try:
        # Register callback
        model.add_callback("on_train_epoch_end", on_train_epoch_end)

        # Start training
        log("INFO", "Training started...")
        train_results = model.train(**train_args)

        # Process results
        final_metrics = {}
        if hasattr(train_results, "results_dict"):
            final_metrics = {
                k: round(float(v), 4) if isinstance(v, (int, float)) else v
                for k, v in train_results.results_dict.items()
                if isinstance(v, (int, float))
            }

        # Find the saved model
        best_pt_path = ""
        run_dir = output_dir / config.name
        if run_dir.exists():
            weights_dir = run_dir / "weights"
            if weights_dir.exists():
                best_pt = weights_dir / "best.pt"
                if best_pt.exists():
                    best_pt_path = str(best_pt.absolute())
                    log("INFO", f"Best model saved: {best_pt_path}")

        if not best_pt_path:
            # Search recursively
            for pt_file in output_dir.rglob("best.pt"):
                best_pt_path = str(pt_file.absolute())
                break

        log("INFO", "Training completed successfully")
        result("success", model_path=best_pt_path, metrics=final_metrics)

    except Exception as e:
        log("ERROR", f"Training failed: {e}")
        traceback.print_exc(file=sys.stderr)
        error(str(e))
        sys.exit(1)


def main():
    """Entry point for standalone training."""
    import argparse
    import json

    parser = argparse.ArgumentParser(description="YOLO Training")
    parser.add_argument("--task", type=str, required=True,
                        help="Path to task.json")
    args = parser.parse_args()

    with open(args.task, "r", encoding="utf-8") as f:
        task_data = json.load(f)

    params = task_data.get("params", {})
    config = YOLOTrainConfig.from_dict(params)
    run_train(config)


if __name__ == "__main__":
    main()