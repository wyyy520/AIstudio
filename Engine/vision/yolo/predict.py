"""
YOLO Prediction (Inference) Module.

Runs YOLO inference on images or video and outputs results.
"""

import os
import sys
import traceback
from pathlib import Path

from vision.yolo.config import YOLOPredictConfig
from sdk.logger import progress, log, result, error


def run_predict(config: YOLOPredictConfig):
    """
    Run YOLO inference with the given configuration.

    Args:
        config: YOLOPredictConfig with model_path, source, conf, etc.
    """
    log("INFO", f"Starting YOLO prediction: model={config.model_path}, "
                f"source={config.source}, device={config.device}")

    # Validate model path
    if not config.model_path:
        error("Model path is required for prediction")
        sys.exit(1)

    model_file = Path(config.model_path)
    if not model_file.exists():
        error(f"Model file not found: {config.model_path}")
        sys.exit(1)

    # Validate source
    if not config.source:
        error("Source (image/video path) is required for prediction")
        sys.exit(1)

    source_path = Path(config.source)
    if not source_path.exists():
        error(f"Source path does not exist: {config.source}")
        sys.exit(1)

    # Create output directory
    output_dir = Path(config.project)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Check device
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
        sys.exit(1)

    # Load the trained model
    model = YOLO(config.model_path)
    log("INFO", f"Model loaded: {config.model_path}")

    try:
        # Run inference
        predict_args = config.to_ultralytics_args()
        predict_args["source"] = config.source

        results = model.predict(**predict_args)

        # Process results
        detections = []
        for r in results:
            if hasattr(r, "boxes") and r.boxes is not None:
                boxes = r.boxes
                for i in range(len(boxes)):
                    cls_id = int(boxes.cls[i].item()) if hasattr(boxes.cls[i], "item") else int(boxes.cls[i])
                    conf = float(boxes.conf[i].item()) if hasattr(boxes.conf[i], "item") else float(boxes.conf[i])
                    xyxy = boxes.xyxy[i].tolist() if hasattr(boxes.xyxy[i], "tolist") else list(boxes.xyxy[i])
                    cls_name = model.names.get(cls_id, str(cls_id)) if hasattr(model, "names") else str(cls_id)

                    detections.append({
                        "class_id": cls_id,
                        "class_name": cls_name,
                        "confidence": round(conf, 4),
                        "bbox": [round(float(x), 2) for x in xyxy],
                    })

        summary = {
            "total_detections": len(detections),
            "source": config.source,
            "model": config.model_path,
            "detections": detections[:100],  # Limit for output size
        }

        log("INFO", f"Prediction completed: {len(detections)} detections")
        result("success", model_path=config.model_path, metrics=summary)

    except Exception as e:
        log("ERROR", f"Prediction failed: {e}")
        traceback.print_exc(file=sys.stderr)
        error(str(e))
        sys.exit(1)


def main():
    """Entry point for standalone prediction."""
    import argparse
    import json

    parser = argparse.ArgumentParser(description="YOLO Prediction")
    parser.add_argument("--task", type=str, required=True,
                        help="Path to task.json")
    args = parser.parse_args()

    with open(args.task, "r", encoding="utf-8") as f:
        task_data = json.load(f)

    params = task_data.get("params", {})
    config = YOLOPredictConfig.from_dict(params)
    run_predict(config)


if __name__ == "__main__":
    main()