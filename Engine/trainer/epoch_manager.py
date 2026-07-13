"""
Epoch Manager - manages epoch-level training state and per-epoch results.

Provides:
  - Epoch lifecycle tracking (start/complete/error)
  - Per-epoch metrics recording
  - Training history accumulation
  - ETA estimation
"""

import time
from dataclasses import dataclass, field
from typing import Any, Optional

from sdk.logger import log


@dataclass
class EpochResult:
    epoch: int
    loss: float
    metrics: dict[str, Any] = field(default_factory=dict)
    lr: float = 0.0
    duration_seconds: float = 0.0
    phase: str = "train"


class EpochManager:
    """
    Manages epoch-level training lifecycle.

    Usage:
        manager = EpochManager(total_epochs=100, start_epoch=0)
        for epoch in range(manager.start_epoch, 100):
            manager.start_epoch(epoch)
            result = train_one_epoch(epoch)
            manager.complete_epoch(epoch, result)
            print(f"ETA: {manager.eta}")
    """

    def __init__(self, total_epochs: int, start_epoch: int = 0):
        self.total_epochs = total_epochs
        self.start_epoch = start_epoch
        self.current_epoch: int = start_epoch
        self._epoch_start_time: float = 0.0
        self._epoch_durations: list[float] = []
        self._history: list[EpochResult] = []
        self._best_loss: Optional[float] = None
        self._best_epoch: int = 0
        self._stopped_early: bool = False

    @property
    def eta(self) -> float:
        if not self._epoch_durations:
            return 0.0
        avg_duration = sum(self._epoch_durations) / len(self._epoch_durations)
        remaining = self.total_epochs - self.current_epoch - 1
        return avg_duration * remaining

    @property
    def eta_formatted(self) -> str:
        seconds = self.eta
        if seconds < 60:
            return f"{seconds:.0f}s"
        elif seconds < 3600:
            return f"{seconds / 60:.1f}m"
        else:
            return f"{seconds / 3600:.1f}h"

    def start_epoch(self, epoch: int):
        self.current_epoch = epoch
        self._epoch_start_time = time.time()

    def complete_epoch(self, epoch: int, result: EpochResult):
        duration = time.time() - self._epoch_start_time
        result.duration_seconds = duration
        self._epoch_durations.append(duration)
        self._history.append(result)

        if result.loss is not None:
            if self._best_loss is None or result.loss < self._best_loss:
                self._best_loss = result.loss
                self._best_epoch = epoch

        log("INFO", f"[epoch] Epoch {epoch + 1}/{self.total_epochs} "
                    f"completed in {duration:.1f}s, loss={result.loss:.4f}, "
                    f"ETA={self.eta_formatted}")

    def stop_early(self, epoch: int):
        self._stopped_early = True
        log("INFO", f"[epoch] Training stopped early at epoch {epoch + 1}")

    def get_history(self) -> list[dict]:
        return [
            {
                "epoch": r.epoch,
                "loss": r.loss,
                "metrics": r.metrics,
                "lr": r.lr,
                "duration": r.duration_seconds,
                "phase": r.phase,
            }
            for r in self._history
        ]

    def get_summary(self) -> dict:
        losses = [r.loss for r in self._history if r.loss is not None]
        return {
            "total_epochs": self.total_epochs,
            "completed_epochs": len(self._history),
            "best_loss": self._best_loss,
            "best_epoch": self._best_epoch + 1,
            "avg_epoch_time": sum(self._epoch_durations) / len(self._epoch_durations)
            if self._epoch_durations else 0,
            "stopped_early": self._stopped_early,
            "losses": losses,
        }