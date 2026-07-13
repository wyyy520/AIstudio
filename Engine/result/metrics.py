"""
Metrics Calculator - computes and aggregates training/evaluation metrics.

Provides:
  - Classification metrics (accuracy, precision, recall, F1)
  - Detection metrics (mAP, IoU, precision-recall)
  - Regression metrics (MSE, MAE, R2)
  - Per-class metrics breakdown
  - Confusion matrix computation
  - Metric history smoothing
"""

import math
from collections import defaultdict
from typing import Any, Optional

from sdk.logger import log


def compute_classification_metrics(y_true: list, y_pred: list,
                                   labels: list = None) -> dict:
    labels = labels or sorted(set(y_true) | set(y_pred))
    n = len(labels)

    conf_matrix = [[0] * n for _ in range(n)]
    label_to_idx = {l: i for i, l in enumerate(labels)}

    for t, p in zip(y_true, y_pred):
        if t in label_to_idx and p in label_to_idx:
            conf_matrix[label_to_idx[t]][label_to_idx[p]] += 1

    per_class = {}
    total_tp = 0
    total_fp = 0
    total_fn = 0

    for i, label in enumerate(labels):
        tp = conf_matrix[i][i]
        fp = sum(conf_matrix[j][i] for j in range(n)) - tp
        fn = sum(conf_matrix[i][j] for j in range(n)) - tp

        total_tp += tp
        total_fp += fp
        total_fn += fn

        precision = tp / (tp + fp) if (tp + fp) > 0 else 0.0
        recall = tp / (tp + fn) if (tp + fn) > 0 else 0.0
        f1 = 2 * precision * recall / (precision + recall) if (precision + recall) > 0 else 0.0

        per_class[label] = {
            "precision": round(precision, 4),
            "recall": round(recall, 4),
            "f1": round(f1, 4),
            "support": tp + fn,
        }

    accuracy = total_tp / len(y_true) if len(y_true) > 0 else 0.0
    macro_precision = sum(p["precision"] for p in per_class.values()) / n if n > 0 else 0.0
    macro_recall = sum(p["recall"] for p in per_class.values()) / n if n > 0 else 0.0
    macro_f1 = sum(p["f1"] for p in per_class.values()) / n if n > 0 else 0.0

    return {
        "accuracy": round(accuracy, 4),
        "macro_precision": round(macro_precision, 4),
        "macro_recall": round(macro_recall, 4),
        "macro_f1": round(macro_f1, 4),
        "per_class": per_class,
        "confusion_matrix": conf_matrix,
        "num_classes": n,
        "total_samples": len(y_true),
    }


class MetricsCalculator:
    """
    Computes and aggregates various ML metrics.

    Usage:
        calc = MetricsCalculator()
        calc.update(loss=0.5, accuracy=0.85, epoch=1)
        calc.update(loss=0.3, accuracy=0.90, epoch=2)
        summary = calc.summary()
    """

    def __init__(self):
        self._history: list[dict] = []
        self._current: dict = {}

    def update(self, **metrics):
        self._current = metrics
        self._history.append(metrics)

    def compute_iou(self, box1: list, box2: list) -> float:
        x1 = max(box1[0], box2[0])
        y1 = max(box1[1], box2[1])
        x2 = min(box1[2], box2[2])
        y2 = min(box1[3], box2[3])

        inter_area = max(0, x2 - x1) * max(0, y2 - y1)
        area1 = (box1[2] - box1[0]) * (box1[3] - box1[1])
        area2 = (box2[2] - box2[0]) * (box2[3] - box2[1])

        union = area1 + area2 - inter_area
        return inter_area / union if union > 0 else 0.0

    def compute_ap(self, recalls: list, precisions: list) -> float:
        recalls = [0.0] + sorted(recalls) + [1.0]
        precisions = [0.0] + sorted(precisions) + [0.0]

        for i in range(len(precisions) - 2, -1, -1):
            precisions[i] = max(precisions[i], precisions[i + 1])

        ap = 0.0
        for i in range(1, len(recalls)):
            ap += (recalls[i] - recalls[i - 1]) * precisions[i]
        return ap

    def compute_detection_metrics(self, predictions: list, ground_truth: list,
                                  iou_threshold: float = 0.5,
                                  num_classes: int = 1) -> dict:
        results = {}
        for cls_id in range(num_classes):
            cls_preds = [p for p in predictions if p.get("class_id") == cls_id]
            cls_gts = [g for g in ground_truth if g.get("class_id") == cls_id]

            matched = set()
            tp = 0
            for pred in cls_preds:
                best_iou = 0.0
                best_gt = None
                for idx, gt in enumerate(cls_gts):
                    if idx in matched:
                        continue
                    iou = self.compute_iou(
                        pred.get("bbox", [0, 0, 0, 0]),
                        gt.get("bbox", [0, 0, 0, 0]),
                    )
                    if iou > best_iou:
                        best_iou = iou
                        best_gt = idx

                if best_iou >= iou_threshold and best_gt is not None:
                    tp += 1
                    matched.add(best_gt)

            fp = len(cls_preds) - tp
            fn = len(cls_gts) - tp

            precision = tp / (tp + fp) if (tp + fp) > 0 else 0.0
            recall = tp / (tp + fn) if (tp + fn) > 0 else 0.0
            f1 = 2 * precision * recall / (precision + recall) if (precision + recall) > 0 else 0.0

            results[f"class_{cls_id}"] = {
                "precision": round(precision, 4),
                "recall": round(recall, 4),
                "f1": round(f1, 4),
                "tp": tp,
                "fp": fp,
                "fn": fn,
            }

        return results

    def compute_regression_metrics(self, y_true: list, y_pred: list) -> dict:
        n = len(y_true)
        if n == 0:
            return {}

        mse = sum((t - p) ** 2 for t, p in zip(y_true, y_pred)) / n
        mae = sum(abs(t - p) for t, p in zip(y_true, y_pred)) / n
        rmse = math.sqrt(mse)

        mean_true = sum(y_true) / n
        ss_res = sum((t - p) ** 2 for t, p in zip(y_true, y_pred))
        ss_tot = sum((t - mean_true) ** 2 for t in y_true)
        r2 = 1 - ss_res / ss_tot if ss_tot != 0 else 0.0

        return {
            "mse": round(mse, 4),
            "mae": round(mae, 4),
            "rmse": round(rmse, 4),
            "r2": round(r2, 4),
        }

    def summary(self) -> dict:
        if not self._history:
            return {}

        all_keys = set()
        for entry in self._history:
            all_keys.update(entry.keys())

        summary = {}
        for key in sorted(all_keys):
            values = [e[key] for e in self._history if key in e
                      and isinstance(e[key], (int, float))]
            if values:
                summary[key] = {
                    "min": round(min(values), 4),
                    "max": round(max(values), 4),
                    "mean": round(sum(values) / len(values), 4),
                    "last": round(values[-1], 4),
                    "best": round(min(values), 4) if "loss" in key.lower() else round(max(values), 4),
                }

        return {
            "num_steps": len(self._history),
            "metrics": summary,
        }

    def smooth_history(self, metrics_key: str = "loss",
                       window: int = 5) -> list[float]:
        values = [e[metrics_key] for e in self._history
                  if metrics_key in e and isinstance(e[metrics_key], (int, float))]
        if not values:
            return []

        smoothed = []
        for i in range(len(values)):
            start = max(0, i - window + 1)
            window_vals = values[start:i + 1]
            smoothed.append(round(sum(window_vals) / len(window_vals), 4))

        return smoothed

    def reset(self):
        self._history = []
        self._current = {}