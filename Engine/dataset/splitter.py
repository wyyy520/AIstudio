"""
Dataset Splitter - automatically splits datasets into train/val/test sets.

Supports:
  - Stratified splitting (preserves class distribution)
  - Random splitting with fixed seed
  - Custom ratios (train/val/test)
  - YOLO-format output (symlinks or copies)
  - Classification-format output
"""

import os
import random
import shutil
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

from sdk.logger import log


@dataclass
class SplitConfig:
    train_ratio: float = 0.8
    val_ratio: float = 0.1
    test_ratio: float = 0.1
    seed: int = 42
    stratified: bool = True
    method: str = "copy"  # "copy" or "symlink"
    output_dir: Optional[str] = None


class DatasetSplitter:
    """
    Splits datasets into train/val/test sets.

    Usage:
        splitter = DatasetSplitter()
        config = SplitConfig(train_ratio=0.8, val_ratio=0.1, test_ratio=0.1)
        result = splitter.split("path/to/dataset", config)
        # Result: {train: [...], val: [...], test: [...]}
    """

    def split(self, dataset_path: str, config: SplitConfig = None) -> dict:
        if config is None:
            config = SplitConfig()

        total = config.train_ratio + config.val_ratio + config.test_ratio
        if abs(total - 1.0) > 0.001:
            raise ValueError(
                f"Split ratios must sum to 1.0, got {total} "
                f"(train={config.train_ratio}, val={config.val_ratio}, "
                f"test={config.test_ratio})"
            )

        random.seed(config.seed)
        path = Path(dataset_path)

        log("INFO", f"[splitter] Splitting dataset: {dataset_path} "
                    f"(train={config.train_ratio}, val={config.val_ratio}, "
                    f"test={config.test_ratio}, seed={config.seed})")

        if (path / "data.yaml").exists() or (path / "dataset.yaml").exists():
            result = self._split_yolo(path, config)
        elif self._is_classification(path):
            result = self._split_classification(path, config)
        else:
            result = self._split_generic(path, config)

        if config.output_dir:
            self._write_split(result, config)

        return result

    def _is_classification(self, path: Path) -> bool:
        for item in path.iterdir():
            if item.is_dir() and not item.name.startswith("."):
                has_images = bool(
                    list(item.glob("*.jpg")) + list(item.glob("*.png")) +
                    list(item.glob("*.jpeg"))
                )
                if has_images:
                    return True
        return False

    def _split_yolo(self, path: Path, config: SplitConfig) -> dict:
        images = []
        for ext in ("*.jpg", "*.png", "*.jpeg", "*.bmp"):
            images.extend(path.rglob(f"images/**/{ext}"))
            images.extend(path.rglob(f"train/images/**/{ext}"))
            images.extend(path.rglob(f"val/images/**/{ext}"))

        if not images:
            images = list(path.glob("*.jpg")) + list(path.glob("*.png"))

        images = sorted(set(images))
        random.shuffle(images)

        return self._partition(images, config)

    def _split_classification(self, path: Path, config: SplitConfig) -> dict:
        class_dirs = [
            d for d in path.iterdir()
            if d.is_dir() and not d.name.startswith(".")
        ]

        train_all, val_all, test_all = [], [], []

        for cls_dir in class_dirs:
            images = []
            for ext in ("*.jpg", "*.png", "*.jpeg", "*.bmp"):
                images.extend(cls_dir.glob(ext))
            images = sorted(images)
            random.shuffle(images)

            if config.stratified:
                t, v, te = self._partition(images, config)
                train_all.extend(t)
                val_all.extend(v)
                test_all.extend(te)
            else:
                t, v, te = self._partition(images, config)
                train_all.extend(t)
                val_all.extend(v)
                test_all.extend(te)

        return {
            "train": train_all,
            "val": val_all,
            "test": test_all,
        }

    def _split_generic(self, path: Path, config: SplitConfig) -> dict:
        files = sorted([
            f for f in path.iterdir()
            if f.is_file() and not f.name.startswith(".")
        ])
        random.shuffle(files)
        return self._partition(files, config)

    def _partition(self, items: list, config: SplitConfig) -> dict:
        n = len(items)
        n_train = int(n * config.train_ratio)
        n_val = int(n * config.val_ratio)

        train = items[:n_train]
        val = items[n_train:n_train + n_val]
        test = items[n_train + n_val:]

        log("INFO", f"[splitter] Partitioned: train={len(train)}, "
                    f"val={len(val)}, test={len(test)} (total={n})")

        return {"train": train, "val": val, "test": test}

    def _write_split(self, split_result: dict, config: SplitConfig):
        output_dir = Path(config.output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)

        for split_name, items in split_result.items():
            split_dir = output_dir / split_name
            split_dir.mkdir(parents=True, exist_ok=True)

            for item in items:
                dest = split_dir / item.name
                if config.method == "symlink":
                    if not dest.exists():
                        os.symlink(item.resolve(), dest)
                else:
                    shutil.copy2(item, dest)

        log("INFO", f"[splitter] Split written to: {output_dir}")