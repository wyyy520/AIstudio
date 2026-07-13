# AI Studio Engine - Trainer Module
#
# Provides:
#   - base_trainer: Unified training interface
#   - epoch_manager: Epoch management and scheduling
#   - checkpoint: Checkpoint save/load/resume
#   - early_stop: Early stopping with configurable patience

from trainer.base_trainer import BaseTrainer, TrainingConfig, TrainingState
from trainer.epoch_manager import EpochManager, EpochResult
from trainer.checkpoint import CheckpointManager, CheckpointInfo
from trainer.early_stop import EarlyStopping, EarlyStopConfig

__all__ = [
    "BaseTrainer", "TrainingConfig", "TrainingState",
    "EpochManager", "EpochResult",
    "CheckpointManager", "CheckpointInfo",
    "EarlyStopping", "EarlyStopConfig",
]