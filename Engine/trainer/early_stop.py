"""
Early Stopping - stops training when validation metric stops improving.

Supports:
  - Patience-based stopping (stop after N epochs without improvement)
  - Minimum delta for improvement
  - Mode: minimize (loss) or maximize (accuracy)
  - Best metric tracking
  - Counter reset on improvement
"""

import time
from dataclasses import dataclass, field
from typing import Optional

from sdk.logger import log


@dataclass
class EarlyStopConfig:
    patience: int = 10
    min_delta: float = 0.0
    mode: str = "min"  # "min" for loss, "max" for accuracy/mAP
    baseline: Optional[float] = None


class EarlyStopping:
    """
    Early stopping monitor.

    Usage:
        stopper = EarlyStopping(EarlyStopConfig(patience=10, mode="min"))
        for epoch in range(epochs):
            loss = train_one_epoch()
            if stopper.step(loss):
                print(f"Stopped at epoch {epoch}")
                break
        print(f"Best: {stopper.best_metric} at counter {stopper.best_counter}")
    """

    def __init__(self, config: EarlyStopConfig):
        self.config = config
        self.patience = config.patience
        self.min_delta = config.min_delta
        self.mode = config.mode
        self.baseline = config.baseline
        self.counter = 0
        self.best_metric: Optional[float] = None
        self.best_counter: int = 0
        self._start_time = time.time()
        self._history: list[float] = []

    def step(self, metric: float) -> bool:
        self._history.append(metric)

        if self.best_metric is None:
            self.best_metric = metric
            self.best_counter = self.counter
            self.counter = 0
            return False

        if self._is_improvement(metric):
            self.best_metric = metric
            self.best_counter = self.counter
            self.counter = 0
        else:
            self.counter += 1

        if self.counter >= self.patience:
            log("INFO", f"[early_stop] Stopping: no improvement for {self.patience} epochs "
                        f"(best={self.best_metric:.4f} at step {self.best_counter})")
            return True

        return False

    def _is_improvement(self, metric: float) -> bool:
        if self.best_metric is None:
            return True
        if self.mode == "min":
            return metric < self.best_metric - self.min_delta
        else:
            return metric > self.best_metric + self.min_delta

    def reset(self):
        self.counter = 0
        self.best_metric = None
        self.best_counter = 0
        self._history = []

    def get_state(self) -> dict:
        return {
            "counter": self.counter,
            "best_metric": self.best_metric,
            "best_counter": self.best_counter,
            "patience": self.patience,
            "mode": self.mode,
            "history": self._history,
        }