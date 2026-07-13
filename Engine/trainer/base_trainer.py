"""
Base Trainer - unified training interface for all model types.

Provides:
  - Common training configuration (TrainingConfig)
  - Training state tracking (TrainingState)
  - Abstract BaseTrainer that all trainers should extend
  - Integration with EpochManager, CheckpointManager, EarlyStopping
"""

import time
import traceback
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any, Optional

from sdk.logger import log, progress, result, error
from trainer.epoch_manager import EpochManager, EpochResult
from trainer.checkpoint import CheckpointManager
from trainer.early_stop import EarlyStopping, EarlyStopConfig


@dataclass
class TrainingConfig:
    model_name: str = "model"
    output_dir: str = "Storage/models"
    epochs: int = 100
    batch_size: int = 16
    learning_rate: float = 0.001
    device: str = "auto"
    seed: int = 42
    num_workers: int = 4
    save_best_only: bool = True
    save_period: int = 10
    resume: bool = False
    checkpoint_path: Optional[str] = None
    validation: bool = True
    val_interval: int = 1
    log_interval: int = 10
    early_stop: Optional[EarlyStopConfig] = None
    extra: dict = field(default_factory=dict)


class TrainingState(Enum):
    IDLE = "idle"
    PREPARING = "preparing"
    TRAINING = "training"
    VALIDATING = "validating"
    PAUSED = "paused"
    COMPLETED = "completed"
    FAILED = "failed"
    STOPPED = "stopped"


class BaseTrainer(ABC):
    """
    Abstract base class for all trainers.

    Subclasses must implement:
      - _build_model()
      - _train_epoch()
      - _validate_epoch()

    Usage:
        class MyTrainer(BaseTrainer):
            def _build_model(self):
                ...

            def _train_epoch(self, epoch):
                ...
                return EpochResult(...)

            def _validate_epoch(self, epoch):
                ...
                return EpochResult(...)

        trainer = MyTrainer(TrainingConfig(epochs=100))
        result = trainer.train()
    """

    def __init__(self, config: TrainingConfig):
        self.config = config
        self.state = TrainingState.IDLE
        self.model: Any = None
        self.optimizer: Any = None
        self.scheduler: Any = None

        self.epoch_manager = EpochManager(
            total_epochs=config.epochs,
            start_epoch=0,
        )

        self.checkpoint_manager = CheckpointManager(
            output_dir=config.output_dir,
            save_best_only=config.save_best_only,
            save_period=config.save_period,
        )

        self.early_stop: Optional[EarlyStopping] = None
        if config.early_stop:
            self.early_stop = EarlyStopping(config.early_stop)

        self._start_time: float = 0.0
        self._history: list[EpochResult] = []
        self._best_metric: Optional[float] = None
        self._best_epoch: int = 0

    def train(self) -> dict:
        self._start_time = time.time()
        self.state = TrainingState.PREPARING

        log("INFO", f"[trainer] Starting training: model={self.config.model_name}, "
                    f"epochs={self.config.epochs}, device={self.config.device}")

        try:
            self._prepare()
            self._build_model()
            self._setup_device()

            if self.config.resume and self.config.checkpoint_path:
                self._load_checkpoint(self.config.checkpoint_path)

            self.state = TrainingState.TRAINING
            log("INFO", f"[trainer] Training started (epochs {self.epoch_manager.start_epoch + 1}-"
                        f"{self.config.epochs})")

            for epoch in range(self.epoch_manager.start_epoch, self.config.epochs):
                if self.state == TrainingState.STOPPED:
                    break

                self.epoch_manager.start_epoch(epoch)

                train_result = self._train_epoch(epoch)
                self.epoch_manager.complete_epoch(epoch, train_result)
                self._history.append(train_result)

                progress(epoch + 1, self.config.epochs,
                         loss=train_result.loss,
                         metrics=train_result.metrics,
                         step=f"train_epoch_{epoch + 1}")

                if self.config.validation and (epoch + 1) % self.config.val_interval == 0:
                    self.state = TrainingState.VALIDATING
                    val_result = self._validate_epoch(epoch)
                    self.state = TrainingState.TRAINING

                    if val_result is not None:
                        self._history.append(val_result)
                        progress(epoch + 1, self.config.epochs,
                                 loss=val_result.loss,
                                 metrics=val_result.metrics,
                                 step=f"val_epoch_{epoch + 1}")

                    current_metric = val_result.loss if val_result else None
                    is_best = False
                    if current_metric is not None:
                        if self._best_metric is None or current_metric < self._best_metric:
                            self._best_metric = current_metric
                            self._best_epoch = epoch
                            is_best = True

                    self.checkpoint_manager.save(
                        model=self.model,
                        optimizer=self.optimizer,
                        epoch=epoch + 1,
                        metrics=val_result.metrics if val_result else {},
                        loss=val_result.loss if val_result else 0.0,
                        is_best=is_best,
                    )

                if self.early_stop:
                    metric = train_result.loss
                    if self.early_stop.step(metric):
                        log("INFO", f"[trainer] Early stopping triggered at epoch {epoch + 1}")
                        self.epoch_manager.stop_early(epoch)
                        break

            self.state = TrainingState.COMPLETED
            elapsed = time.time() - self._start_time
            log("INFO", f"[trainer] Training completed: {elapsed:.1f}s, "
                        f"best_epoch={self._best_epoch + 1}, "
                        f"best_loss={self._best_metric}")

            output = self.checkpoint_manager.get_summary()
            output["history"] = [
                {"epoch": r.epoch, "loss": r.loss, "metrics": r.metrics}
                for r in self._history
            ]
            output["elapsed_seconds"] = round(elapsed, 2)
            output["best_epoch"] = self._best_epoch + 1
            output["best_metric"] = self._best_metric

            result("success", model_path=output.get("best_model_path", ""),
                   metrics=output)
            return output

        except Exception as e:
            self.state = TrainingState.FAILED
            log("ERROR", f"[trainer] Training failed: {e}")
            traceback.print_exc()
            error(str(e))
            raise

    def stop(self):
        self.state = TrainingState.STOPPED
        log("INFO", "[trainer] Training stopped by user")

    def _prepare(self):
        output_dir = Path(self.config.output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)
        log("INFO", f"[trainer] Output directory: {output_dir}")

    def _setup_device(self):
        if self.config.device == "auto":
            try:
                import torch
                if torch.cuda.is_available():
                    self.config.device = "cuda:0"
                elif hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
                    self.config.device = "mps"
                else:
                    self.config.device = "cpu"
            except ImportError:
                self.config.device = "cpu"
        log("INFO", f"[trainer] Using device: {self.config.device}")

    def _load_checkpoint(self, checkpoint_path: str):
        state = self.checkpoint_manager.load(checkpoint_path)
        if state is None:
            log("WARN", f"[trainer] Checkpoint not found: {checkpoint_path}")
            return

        if self.model and state.get("model_state"):
            self.model.load_state_dict(state["model_state"])
        if self.optimizer and state.get("optimizer_state"):
            self.optimizer.load_state_dict(state["optimizer_state"])

        self.epoch_manager.start_epoch = state.get("epoch", 0)
        log("INFO", f"[trainer] Resumed from epoch {self.epoch_manager.start_epoch + 1}")

    @abstractmethod
    def _build_model(self):
        pass

    @abstractmethod
    def _train_epoch(self, epoch: int) -> EpochResult:
        pass

    @abstractmethod
    def _validate_epoch(self, epoch: int) -> Optional[EpochResult]:
        pass