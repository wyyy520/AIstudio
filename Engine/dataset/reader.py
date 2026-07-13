"""
Dataset Reader - reads datasets from various formats and directories.

Supports:
  - YOLO format (images/ + labels/ + data.yaml)
  - COCO JSON format
  - Image classification (class-named subdirectories)
  - CSV/TSV tabular data
  - Generic file listing
"""

import json
import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

from sdk.logger import log


@dataclass
class DatasetInfo:
    name: str
    path: str
    format: str
    num_samples: int = 0
    num_classes: int = 0
    class_names: list[str] = field(default_factory=list)
    image_size: Optional[tuple] = None
    metadata: dict = field(default_factory=dict)


class DatasetReader:
    """
    Reads and validates datasets from disk.

    Usage:
        reader = DatasetReader()
        info = reader.read("path/to/dataset", format="yolo")
        print(f"Found {info.num_samples} samples across {info.num_classes} classes")
    """

    SUPPORTED_FORMATS = {"yolo", "coco", "classification", "csv", "auto"}

    def read(self, dataset_path: str, format: str = "auto") -> DatasetInfo:
        path = Path(dataset_path)
        if not path.exists():
            raise FileNotFoundError(f"Dataset path not found: {dataset_path}")

        if format == "auto":
            format = self._detect_format(path)

        log("INFO", f"[dataset] Reading dataset: {dataset_path} (format={format})")

        if format == "yolo":
            return self._read_yolo(path)
        elif format == "coco":
            return self._read_coco(path)
        elif format == "classification":
            return self._read_classification(path)
        elif format == "csv":
            return self._read_csv(path)
        else:
            raise ValueError(f"Unsupported format: {format}. "
                             f"Supported: {self.SUPPORTED_FORMATS}")

    def _detect_format(self, path: Path) -> str:
        if (path / "data.yaml").exists() or (path / "dataset.yaml").exists():
            return "yolo"
        if path.is_dir():
            for item in path.iterdir():
                if item.suffix == ".json":
                    try:
                        with open(item, "r") as f:
                            data = json.load(f)
                        if "images" in data and "annotations" in data:
                            return "coco"
                    except (json.JSONDecodeError, IOError):
                        pass
                if item.is_dir() and not item.name.startswith("."):
                    return "classification"
        if path.suffix == ".csv" or path.suffix == ".tsv":
            return "csv"
        return "yolo"

    def _read_yolo(self, path: Path) -> DatasetInfo:
        yaml_path = path / "data.yaml"
        if not yaml_path.exists():
            yaml_path = path / "dataset.yaml"

        classes = []
        nc = 0

        if yaml_path.exists():
            try:
                import yaml
                with open(yaml_path, "r") as f:
                    data = yaml.safe_load(f)
                names = data.get("names", [])
                if isinstance(names, dict):
                    classes = [names[i] for i in sorted(names.keys())]
                elif isinstance(names, list):
                    classes = names
                nc = data.get("nc", len(classes))
            except ImportError:
                content = yaml_path.read_text()
                for line in content.split("\n"):
                    if "names:" in line:
                        break
                classes = self._parse_yolo_names(content)
                nc = len(classes)
            except Exception:
                pass

        train_images = list(path.glob("images/train*/*.*")) + \
                       list(path.glob("train/images/*.*")) + \
                       list(path.glob("images/*.*"))
        val_images = list(path.glob("images/val*/*.*")) + \
                     list(path.glob("val/images/*.*"))

        all_images = train_images + val_images
        if not all_images:
            all_images = list(path.rglob("*.jpg")) + list(path.rglob("*.png")) + \
                         list(path.rglob("*.jpeg")) + list(path.rglob("*.bmp"))

        if not classes:
            label_files = list(path.rglob("labels/*.txt")) + list(path.rglob("*.txt"))
            class_ids = set()
            for lf in label_files[:100]:
                try:
                    for line in lf.read_text().strip().split("\n"):
                        if line.strip():
                            class_ids.add(int(line.split()[0]))
                except (ValueError, IndexError, IOError):
                    pass
            classes = [f"class_{i}" for i in sorted(class_ids)]
            nc = len(classes)

        return DatasetInfo(
            name=path.name,
            path=str(path),
            format="yolo",
            num_samples=len(all_images),
            num_classes=nc,
            class_names=classes,
            metadata={
                "train_images": len(train_images),
                "val_images": len(val_images),
                "yaml_path": str(yaml_path) if yaml_path.exists() else None,
            },
        )

    def _parse_yolo_names(self, yaml_content: str) -> list[str]:
        classes = []
        in_names = False
        for line in yaml_content.split("\n"):
            stripped = line.strip()
            if stripped.startswith("names:"):
                in_names = True
                continue
            if in_names:
                if ":" in stripped and not stripped.startswith("#"):
                    name = stripped.split(":", 1)[1].strip().strip("'").strip('"')
                    if name:
                        classes.append(name)
                elif stripped and not stripped.startswith("-") and not stripped.startswith("#"):
                    if ":" not in stripped:
                        in_names = False
        return classes

    def _read_coco(self, path: Path) -> DatasetInfo:
        json_files = list(path.glob("*.json"))
        if not json_files:
            json_files = list(path.glob("annotations/*.json"))

        annotation_file = None
        for jf in json_files:
            try:
                with open(jf, "r") as f:
                    data = json.load(f)
                if "images" in data and "annotations" in data:
                    annotation_file = jf
                    break
            except (json.JSONDecodeError, IOError):
                pass

        if annotation_file is None:
            raise FileNotFoundError(f"No COCO annotation file found in {path}")

        with open(annotation_file, "r") as f:
            coco_data = json.load(f)

        categories = coco_data.get("categories", [])
        class_names = [c["name"] for c in sorted(categories, key=lambda x: x["id"])]
        images = coco_data.get("images", [])

        return DatasetInfo(
            name=path.name,
            path=str(path),
            format="coco",
            num_samples=len(images),
            num_classes=len(categories),
            class_names=class_names,
            metadata={
                "annotation_file": str(annotation_file),
                "num_annotations": len(coco_data.get("annotations", [])),
            },
        )

    def _read_classification(self, path: Path) -> DatasetInfo:
        class_dirs = [
            d for d in path.iterdir()
            if d.is_dir() and not d.name.startswith(".")
        ]

        class_names = sorted([d.name for d in class_dirs])
        total = 0
        for cls_dir in class_dirs:
            images = list(cls_dir.glob("*.jpg")) + list(cls_dir.glob("*.png")) + \
                     list(cls_dir.glob("*.jpeg")) + list(cls_dir.glob("*.bmp"))
            total += len(images)

        return DatasetInfo(
            name=path.name,
            path=str(path),
            format="classification",
            num_samples=total,
            num_classes=len(class_names),
            class_names=class_names,
            metadata={
                "per_class": {d.name: len(list(d.glob("*.*"))) for d in class_dirs},
            },
        )

    def _read_csv(self, path: Path) -> DatasetInfo:
        import csv
        with open(path, "r", newline="") as f:
            reader = csv.reader(f)
            header = next(reader, [])
            rows = sum(1 for _ in reader)

        return DatasetInfo(
            name=path.stem,
            path=str(path),
            format="csv",
            num_samples=rows,
            num_classes=0,
            class_names=[],
            metadata={
                "columns": header,
                "delimiter": "," if path.suffix == ".csv" else "\t",
            },
        )

    def list_datasets(self, root_dir: str) -> list[DatasetInfo]:
        root = Path(root_dir)
        if not root.exists():
            return []

        datasets = []
        for item in root.iterdir():
            if item.is_dir() and not item.name.startswith("."):
                try:
                    info = self.read(str(item), format="auto")
                    datasets.append(info)
                except Exception:
                    pass
        return datasets