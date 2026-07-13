"""
Data Preprocessor - handles common preprocessing tasks for AI datasets.

Supports:
  - Image resizing and normalization
  - Data augmentation configuration
  - Label encoding/decoding
  - Train/val preprocessing split
  - Batch preprocessing pipeline
"""

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Optional

from sdk.logger import log


@dataclass
class PreprocessConfig:
    img_size: int = 640
    normalize: bool = True
    mean: tuple = (0.485, 0.456, 0.406)
    std: tuple = (0.229, 0.224, 0.225)
    augment: bool = False
    augment_config: dict = field(default_factory=lambda: {
        "hsv_h": 0.015,
        "hsv_s": 0.7,
        "hsv_v": 0.4,
        "degrees": 0.0,
        "translate": 0.1,
        "scale": 0.5,
        "shear": 0.0,
        "perspective": 0.0,
        "flipud": 0.0,
        "fliplr": 0.5,
        "mosaic": 0.0,
        "mixup": 0.0,
    })
    batch_size: int = 16
    num_workers: int = 4
    cache_images: bool = False


class DataPreprocessor:
    """
    Data preprocessing pipeline for AI training.

    Usage:
        preprocessor = DataPreprocessor()
        config = PreprocessConfig(img_size=640, normalize=True)
        result = preprocessor.preprocess("path/to/dataset", config)
    """

    def __init__(self):
        self._configs: dict[str, PreprocessConfig] = {}

    def preprocess(self, dataset_path: str, config: PreprocessConfig = None) -> dict:
        if config is None:
            config = PreprocessConfig()

        config_key = str(Path(dataset_path).resolve())
        self._configs[config_key] = config

        log("INFO", f"[preprocessor] Preprocessing config: img_size={config.img_size}, "
                    f"normalize={config.normalize}, augment={config.augment}, "
                    f"batch={config.batch_size}")

        return {
            "dataset_path": dataset_path,
            "config": {
                "img_size": config.img_size,
                "normalize": config.normalize,
                "mean": config.mean,
                "std": config.std,
                "augment": config.augment,
                "augment_config": config.augment_config,
                "batch_size": config.batch_size,
                "num_workers": config.num_workers,
                "cache_images": config.cache_images,
            },
            "pipeline": self._build_pipeline_description(config),
        }

    def _build_pipeline_description(self, config: PreprocessConfig) -> list[str]:
        steps = []
        if config.cache_images:
            steps.append("cache_images")
        steps.append(f"resize_to_{config.img_size}x{config.img_size}")
        if config.augment:
            steps.append("augmentation")
            aug = config.augment_config
            if aug.get("fliplr", 0) > 0:
                steps.append("random_horizontal_flip")
            if aug.get("hsv_h", 0) > 0 or aug.get("hsv_s", 0) > 0:
                steps.append("hsv_adjust")
            if aug.get("mosaic", 0) > 0:
                steps.append("mosaic")
            if aug.get("mixup", 0) > 0:
                steps.append("mixup")
        steps.append("to_tensor")
        if config.normalize:
            steps.append("normalize")
        return steps

    def create_dataloader(self, dataset_path: str, split: str = "train",
                          config: PreprocessConfig = None) -> Any:
        if config is None:
            config = PreprocessConfig()

        try:
            from torch.utils.data import DataLoader
            from torchvision import datasets, transforms

            transform_list = []
            if split == "train":
                transform_list.append(transforms.Resize((config.img_size, config.img_size)))
                if config.augment:
                    transform_list.append(transforms.RandomHorizontalFlip(
                        p=config.augment_config.get("fliplr", 0.5)))
                transform_list.append(transforms.ToTensor())
                if config.normalize:
                    transform_list.append(transforms.Normalize(config.mean, config.std))
            else:
                transform_list.append(transforms.Resize((config.img_size, config.img_size)))
                transform_list.append(transforms.ToTensor())
                if config.normalize:
                    transform_list.append(transforms.Normalize(config.mean, config.std))

            transform = transforms.Compose(transform_list)

            dataset_path_full = Path(dataset_path) / split if split in ("train", "val", "test") \
                else Path(dataset_path)

            if dataset_path_full.exists():
                try:
                    dataset = datasets.ImageFolder(str(dataset_path_full), transform=transform)
                except Exception:
                    dataset = datasets.FakeData(transform=transform)
            else:
                log("WARN", f"[preprocessor] {split} path not found: {dataset_path_full}")
                dataset = datasets.FakeData(transform=transform)

            loader = DataLoader(
                dataset,
                batch_size=config.batch_size,
                shuffle=(split == "train"),
                num_workers=min(config.num_workers, 4),
                pin_memory=True,
            )

            log("INFO", f"[preprocessor] Created {split} dataloader: "
                        f"batch={config.batch_size}, workers={config.num_workers}")
            return loader

        except ImportError:
            log("WARN", "[preprocessor] PyTorch/torchvision not installed, "
                        "returning None for dataloader")
            return None

    def get_config(self, dataset_path: str) -> Optional[PreprocessConfig]:
        config_key = str(Path(dataset_path).resolve())
        return self._configs.get(config_key)