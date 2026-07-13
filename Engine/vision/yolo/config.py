"""
YOLO Configuration - model definitions, default parameters, and validation.
"""

from dataclasses import dataclass, field
from typing import Optional

# Supported YOLO model variants
YOLO_MODELS = {
    "yolov8n": {"size": "nano", "params_m": 3.2, "mAP": 37.3},
    "yolov8s": {"size": "small", "params_m": 11.2, "mAP": 44.9},
    "yolov8m": {"size": "medium", "params_m": 25.9, "mAP": 50.2},
    "yolov8l": {"size": "large", "params_m": 43.7, "mAP": 52.9},
    "yolov8x": {"size": "xlarge", "params_m": 68.2, "mAP": 53.9},
}

# Default training configuration
DEFAULT_TRAIN_CONFIG = {
    "model": "yolov8n.pt",
    "epochs": 100,
    "batch": 16,
    "img_size": 640,
    "device": "cuda",
    "workers": 8,
    "lr0": 0.01,
    "lrf": 0.01,
    "momentum": 0.937,
    "weight_decay": 0.0005,
    "warmup_epochs": 3.0,
    "warmup_momentum": 0.8,
    "warmup_bias_lr": 0.1,
    "patience": 50,
    "save_period": -1,
    "cache": False,
    "resume": False,
    "pretrained": True,
    "optimizer": "auto",
    "verbose": True,
    "seed": 0,
    "single_cls": False,
    "rect": False,
    "cos_lr": False,
    "close_mosaic": 10,
    "amp": True,
    "fraction": 1.0,
    "overlap_mask": True,
    "mask_ratio": 4,
    "dropout": 0.0,
    "val": True,
    "plots": False,
    "project": "Storage/models",
    "name": "train",
}


@dataclass
class YOLOTrainConfig:
    """YOLO training configuration."""
    model: str = "yolov8n.pt"
    dataset: str = ""
    epochs: int = 100
    batch: int = 16
    img_size: int = 640
    device: str = "cuda"
    workers: int = 8
    lr0: float = 0.01
    lrf: float = 0.01
    momentum: float = 0.937
    weight_decay: float = 0.0005
    warmup_epochs: float = 3.0
    warmup_momentum: float = 0.8
    warmup_bias_lr: float = 0.1
    patience: int = 50
    save_period: int = -1
    cache: bool = False
    resume: bool = False
    pretrained: bool = True
    optimizer: str = "auto"
    seed: int = 0
    single_cls: bool = False
    rect: bool = False
    cos_lr: bool = False
    close_mosaic: int = 10
    amp: bool = True
    fraction: float = 1.0
    dropout: float = 0.0
    val: bool = True
    plots: bool = False
    project: str = "Storage/models"
    name: str = "train"
    output_dir: str = "Storage/models/"

    extra: dict = field(default_factory=dict)

    def to_ultralytics_args(self) -> dict:
        """Convert to the argument dict that ultralytics YOLO expects."""
        skip = {"extra", "output_dir", "dataset"}
        args = {}
        for key, value in self.__dict__.items():
            if key in skip or key.startswith("_"):
                continue
            args[key] = value
        args["data"] = self.dataset
        args["project"] = self.output_dir or self.project
        return args

    @classmethod
    def from_dict(cls, params: dict) -> "YOLOTrainConfig":
        """Create config from a flat dictionary (e.g., from task.json params)."""
        valid_keys = {f.name for f in cls.__dataclass_fields__.values()}
        filtered = {k: v for k, v in params.items() if k in valid_keys}
        extra = {k: v for k, v in params.items() if k not in valid_keys}
        filtered["extra"] = extra
        return cls(**filtered)


@dataclass
class YOLOPredictConfig:
    """YOLO inference configuration."""
    model_path: str = ""
    source: str = ""
    conf: float = 0.25
    iou: float = 0.7
    img_size: int = 640
    device: str = "cuda"
    save: bool = True
    save_txt: bool = False
    save_conf: bool = False
    save_crop: bool = False
    show: bool = False
    project: str = "Storage/predictions"
    name: str = "predict"

    extra: dict = field(default_factory=dict)

    def to_ultralytics_args(self) -> dict:
        skip = {"extra", "model_path", "source"}
        args = {}
        for key, value in self.__dict__.items():
            if key in skip or key.startswith("_"):
                continue
            args[key] = value
        return args

    @classmethod
    def from_dict(cls, params: dict) -> "YOLOPredictConfig":
        valid_keys = {f.name for f in cls.__dataclass_fields__.values()}
        filtered = {k: v for k, v in params.items() if k in valid_keys}
        extra = {k: v for k, v in params.items() if k not in valid_keys}
        filtered["extra"] = extra
        return cls(**filtered)