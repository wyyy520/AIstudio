"""
Checkpoint Manager - save/load/resume training checkpoints.

Provides:
  - Periodic checkpoint saving
  - Best-model tracking (lowest loss or highest metric)
  - Checkpoint loading for resume training
  - Checkpoint cleanup (keep only N best)
  - Metadata tracking (epoch, metrics, timestamp)
"""

import json
import os
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Optional

from sdk.logger import log


@dataclass
class CheckpointInfo:
    path: str
    epoch: int
    loss: float
    metrics: dict[str, Any] = field(default_factory=dict)
    is_best: bool = False
    timestamp: float = field(default_factory=time.time)
    file_size_mb: float = 0.0


class CheckpointManager:
    """
    Manages training checkpoints: save, load, track best, cleanup.

    Usage:
        manager = CheckpointManager(output_dir="models/", save_best_only=True)
        manager.save(model, optimizer, epoch=10, metrics={"acc": 0.95}, loss=0.1)
        state = manager.load("models/checkpoint_epoch_10.pt")
    """

    def __init__(self, output_dir: str = "Storage/models",
                 save_best_only: bool = True,
                 save_period: int = 10,
                 max_keep: int = 5):
        self.output_dir = Path(output_dir)
        self.save_best_only = save_best_only
        self.save_period = save_period
        self.max_keep = max_keep
        self._checkpoints: list[CheckpointInfo] = []
        self._best_loss: float = float("inf")
        self._best_model_path: str = ""

        self.output_dir.mkdir(parents=True, exist_ok=True)
        self._load_index()

    def save(self, model: Any, optimizer: Optional[Any] = None,
             epoch: int = 0, metrics: dict[str, Any] = None,
             loss: float = 0.0, is_best: bool = False) -> Optional[str]:
        metrics = metrics or {}

        should_save = False
        if is_best:
            should_save = True
        elif not self.save_best_only and epoch % self.save_period == 0:
            should_save = True
        elif self.save_best_only and loss < self._best_loss:
            should_save = True
            is_best = True

        if not should_save:
            return None

        if is_best and loss < self._best_loss:
            self._best_loss = loss

        checkpoint_name = f"checkpoint_epoch_{epoch:04d}.pt"
        if is_best:
            checkpoint_name = f"best_epoch_{epoch:04d}.pt"

        checkpoint_path = self.output_dir / checkpoint_name

        try:
            import torch
            state = {
                "epoch": epoch,
                "model_state": model.state_dict() if hasattr(model, "state_dict") else None,
                "loss": loss,
                "metrics": metrics,
                "timestamp": time.time(),
            }
            if optimizer and hasattr(optimizer, "state_dict"):
                state["optimizer_state"] = optimizer.state_dict()

            torch.save(state, checkpoint_path)
        except ImportError:
            state = {
                "epoch": epoch,
                "loss": loss,
                "metrics": metrics,
                "timestamp": time.time(),
            }
            with open(checkpoint_path, "w") as f:
                json.dump(state, f, indent=2)

        file_size = checkpoint_path.stat().st_size / (1024 * 1024) if checkpoint_path.exists() else 0

        info = CheckpointInfo(
            path=str(checkpoint_path),
            epoch=epoch,
            loss=loss,
            metrics=metrics,
            is_best=is_best,
            file_size_mb=file_size,
        )
        self._checkpoints.append(info)

        if is_best:
            self._best_model_path = str(checkpoint_path)
            latest_path = self.output_dir / "best.pt"
            try:
                import shutil
                shutil.copy2(checkpoint_path, latest_path)
            except Exception:
                pass

        self._save_index()
        self._cleanup()

        log("INFO", f"[checkpoint] Saved: {checkpoint_name} "
                    f"(epoch={epoch}, loss={loss:.4f}, "
                    f"size={file_size:.1f}MB{' [BEST]' if is_best else ''})")

        return str(checkpoint_path)

    def load(self, checkpoint_path: str) -> Optional[dict]:
        path = Path(checkpoint_path)
        if not path.exists():
            log("ERROR", f"[checkpoint] File not found: {checkpoint_path}")
            return None

        try:
            import torch
            state = torch.load(path, map_location="cpu", weights_only=False)
            log("INFO", f"[checkpoint] Loaded: {checkpoint_path} (epoch={state.get('epoch', '?')})")
            return state
        except ImportError:
            with open(path, "r") as f:
                state = json.load(f)
            log("INFO", f"[checkpoint] Loaded JSON: {checkpoint_path}")
            return state
        except Exception as e:
            log("ERROR", f"[checkpoint] Failed to load: {e}")
            return None

    def get_best_model_path(self) -> str:
        return self._best_model_path

    def get_summary(self) -> dict:
        return {
            "output_dir": str(self.output_dir),
            "num_checkpoints": len(self._checkpoints),
            "best_model_path": self._best_model_path,
            "best_loss": self._best_loss,
            "checkpoints": [
                {
                    "path": c.path,
                    "epoch": c.epoch,
                    "loss": c.loss,
                    "is_best": c.is_best,
                    "file_size_mb": c.file_size_mb,
                }
                for c in self._checkpoints[-10:]
            ],
        }

    def _cleanup(self):
        if len(self._checkpoints) <= self.max_keep:
            return

        self._checkpoints.sort(key=lambda c: c.loss)
        keep = self._checkpoints[:self.max_keep]
        remove = self._checkpoints[self.max_keep:]

        for c in remove:
            try:
                path = Path(c.path)
                if path.exists():
                    path.unlink()
                    log("INFO", f"[checkpoint] Cleaned up: {c.path}")
            except Exception:
                pass

        self._checkpoints = keep

    def _save_index(self):
        index_path = self.output_dir / "checkpoints.json"
        try:
            with open(index_path, "w") as f:
                json.dump([
                    {
                        "path": c.path,
                        "epoch": c.epoch,
                        "loss": c.loss,
                        "is_best": c.is_best,
                        "file_size_mb": c.file_size_mb,
                        "timestamp": c.timestamp,
                    }
                    for c in self._checkpoints
                ], f, indent=2)
        except Exception:
            pass

    def _load_index(self):
        index_path = self.output_dir / "checkpoints.json"
        if not index_path.exists():
            return
        try:
            with open(index_path, "r") as f:
                data = json.load(f)
            self._checkpoints = [
                CheckpointInfo(**item) for item in data
            ]
        except Exception:
            pass