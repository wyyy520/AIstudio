"""
Format Converter - converts between common dataset annotation formats.

Supports:
  - YOLO <-> COCO
  - YOLO <-> VOC XML
  - Classification folder <-> CSV
  - Generic label mapping
"""

import json
import os
import csv
import xml.etree.ElementTree as ET
from pathlib import Path
from typing import Optional

from sdk.logger import log

supported_formats = ["yolo", "coco", "voc", "csv", "classification"]


class FormatConverter:
    """
    Converts datasets between common annotation formats.

    Usage:
        converter = FormatConverter()
        converter.convert("dataset/", "yolo", "coco", output="dataset_coco/")
    """

    def convert(self, input_path: str, from_format: str, to_format: str,
                output_path: Optional[str] = None,
                class_names: Optional[list[str]] = None) -> str:
        if from_format not in supported_formats:
            raise ValueError(f"Unsupported source format: {from_format}")
        if to_format not in supported_formats:
            raise ValueError(f"Unsupported target format: {to_format}")

        if output_path is None:
            output_path = f"{input_path}_{to_format}"

        output_dir = Path(output_path)
        output_dir.mkdir(parents=True, exist_ok=True)

        log("INFO", f"[converter] Converting: {from_format} -> {to_format}")
        log("INFO", f"[converter] Input: {input_path}, Output: {output_path}")

        if from_format == "yolo" and to_format == "coco":
            result = self._yolo_to_coco(input_path, output_dir, class_names)
        elif from_format == "coco" and to_format == "yolo":
            result = self._coco_to_yolo(input_path, output_dir)
        elif from_format == "yolo" and to_format == "voc":
            result = self._yolo_to_voc(input_path, output_dir, class_names)
        elif from_format == "voc" and to_format == "yolo":
            result = self._voc_to_yolo(input_path, output_dir)
        elif from_format == "classification" and to_format == "csv":
            result = self._classification_to_csv(input_path, output_dir)
        elif from_format == "csv" and to_format == "classification":
            result = self._csv_to_classification(input_path, output_dir)
        else:
            raise ValueError(
                f"Conversion {from_format} -> {to_format} not yet supported"
            )

        return result

    def _yolo_to_coco(self, input_path: str, output_dir: Path,
                      class_names: Optional[list[str]] = None) -> str:
        input_dir = Path(input_path)
        images = []
        annotations = []
        image_id = 0
        annotation_id = 0

        image_exts = (".jpg", ".png", ".jpeg", ".bmp")
        image_files = []
        for ext in image_exts:
            image_files.extend(input_dir.rglob(f"**/*{ext}"))

        class_map = {}
        if class_names:
            class_map = {name: i for i, name in enumerate(class_names)}

        for img_path in sorted(image_files):
            image_id += 1
            img_name = img_path.name

            try:
                from PIL import Image
                with Image.open(img_path) as img:
                    width, height = img.size
            except ImportError:
                width, height = 640, 480

            images.append({
                "id": image_id,
                "file_name": img_name,
                "width": width,
                "height": height,
            })

            label_path = img_path.with_suffix(".txt")
            labels_dir = input_dir / "labels"
            if not label_path.exists():
                label_path = labels_dir / img_path.relative_to(
                    img_path.parents[0]).with_suffix(".txt")

            if label_path.exists():
                for line in label_path.read_text().strip().split("\n"):
                    if not line.strip():
                        continue
                    parts = line.strip().split()
                    if len(parts) >= 5:
                        cls_id = int(parts[0])
                        x_center = float(parts[1])
                        y_center = float(parts[2])
                        bbox_w = float(parts[3])
                        bbox_h = float(parts[4])

                        x = (x_center - bbox_w / 2) * width
                        y = (y_center - bbox_h / 2) * height
                        w = bbox_w * width
                        h = bbox_h * height

                        annotation_id += 1
                        annotations.append({
                            "id": annotation_id,
                            "image_id": image_id,
                            "category_id": cls_id + 1,
                            "bbox": [round(x, 2), round(y, 2),
                                     round(w, 2), round(h, 2)],
                            "area": round(w * h, 2),
                            "iscrowd": 0,
                        })

        if not class_names:
            class_ids = sorted(set(a["category_id"] for a in annotations))
            class_names = [f"class_{i}" for i in range(len(class_ids))]

        categories = [
            {"id": i + 1, "name": name, "supercategory": "none"}
            for i, name in enumerate(class_names)
        ]

        coco_data = {
            "images": images,
            "annotations": annotations,
            "categories": categories,
        }

        output_file = output_dir / "annotations.json"
        with open(output_file, "w") as f:
            json.dump(coco_data, f, indent=2)

        log("INFO", f"[converter] COCO annotations written: {output_file}")
        return str(output_file)

    def _coco_to_yolo(self, input_path: str, output_dir: Path) -> str:
        json_files = list(Path(input_path).glob("*.json"))
        if not json_files:
            json_files = list(Path(input_path).glob("annotations/*.json"))

        if not json_files:
            raise FileNotFoundError(f"No COCO JSON found in {input_path}")

        with open(json_files[0], "r") as f:
            coco_data = json.load(f)

        images_dir = output_dir / "images"
        labels_dir = output_dir / "labels"
        images_dir.mkdir(parents=True, exist_ok=True)
        labels_dir.mkdir(parents=True, exist_ok=True)

        image_map = {img["id"]: img for img in coco_data["images"]}

        for ann in coco_data.get("annotations", []):
            img = image_map.get(ann["image_id"])
            if img is None:
                continue

            img_name = Path(img["file_name"]).stem
            label_file = labels_dir / f"{img_name}.txt"

            bbox = ann["bbox"]
            x, y, w, h = bbox[0], bbox[1], bbox[2], bbox[3]
            img_w, img_h = img.get("width", 640), img.get("height", 480)

            x_center = (x + w / 2) / img_w
            y_center = (y + h / 2) / img_h
            norm_w = w / img_w
            norm_h = h / img_h
            cls_id = ann["category_id"] - 1

            with open(label_file, "a") as f:
                f.write(f"{cls_id} {x_center:.6f} {y_center:.6f} "
                        f"{norm_w:.6f} {norm_h:.6f}\n")

        log("INFO", f"[converter] YOLO labels written: {labels_dir}")
        return str(output_dir)

    def _yolo_to_voc(self, input_path: str, output_dir: Path,
                     class_names: Optional[list[str]] = None) -> str:
        annotations_dir = output_dir / "Annotations"
        annotations_dir.mkdir(parents=True, exist_ok=True)

        input_dir = Path(input_path)
        image_files = []
        for ext in (".jpg", ".png", ".jpeg", ".bmp"):
            image_files.extend(input_dir.rglob(f"**/*{ext}"))

        for img_path in sorted(image_files):
            try:
                from PIL import Image
                with Image.open(img_path) as img:
                    width, height = img.size
            except ImportError:
                width, height = 640, 480

            root = ET.Element("annotation")
            ET.SubElement(root, "filename").text = img_path.name
            size = ET.SubElement(root, "size")
            ET.SubElement(size, "width").text = str(width)
            ET.SubElement(size, "height").text = str(height)
            ET.SubElement(size, "depth").text = "3"

            label_path = img_path.with_suffix(".txt")
            labels_dir = input_dir / "labels"
            if not label_path.exists():
                label_path = labels_dir / img_path.relative_to(
                    img_path.parents[0]).with_suffix(".txt")

            if label_path.exists():
                for line in label_path.read_text().strip().split("\n"):
                    if not line.strip():
                        continue
                    parts = line.strip().split()
                    if len(parts) >= 5:
                        cls_id = int(parts[0])
                        x_center = float(parts[1]) * width
                        y_center = float(parts[2]) * height
                        bbox_w = float(parts[3]) * width
                        bbox_h = float(parts[4]) * height

                        obj = ET.SubElement(root, "object")
                        cls_name = class_names[cls_id] if class_names and cls_id < len(class_names) \
                            else f"class_{cls_id}"
                        ET.SubElement(obj, "name").text = cls_name
                        bndbox = ET.SubElement(obj, "bndbox")
                        ET.SubElement(bndbox, "xmin").text = str(int(x_center - bbox_w / 2))
                        ET.SubElement(bndbox, "ymin").text = str(int(y_center - bbox_h / 2))
                        ET.SubElement(bndbox, "xmax").text = str(int(x_center + bbox_w / 2))
                        ET.SubElement(bndbox, "ymax").text = str(int(y_center + bbox_h / 2))

            tree = ET.ElementTree(root)
            xml_path = annotations_dir / f"{img_path.stem}.xml"
            tree.write(xml_path, encoding="utf-8", xml_declaration=True)

        log("INFO", f"[converter] VOC annotations written: {annotations_dir}")
        return str(output_dir)

    def _voc_to_yolo(self, input_path: str, output_dir: Path) -> str:
        labels_dir = output_dir / "labels"
        labels_dir.mkdir(parents=True, exist_ok=True)

        xml_files = list(Path(input_path).glob("*.xml"))
        if not xml_files:
            xml_files = list(Path(input_path).glob("Annotations/*.xml"))

        class_names = []
        for xml_path in xml_files:
            tree = ET.parse(xml_path)
            root = tree.getroot()

            size = root.find("size")
            width = int(size.find("width").text) if size is not None else 640
            height = int(size.find("height").text) if size is not None else 480

            label_lines = []
            for obj in root.findall("object"):
                name = obj.find("name").text
                if name not in class_names:
                    class_names.append(name)
                cls_id = class_names.index(name)

                bndbox = obj.find("bndbox")
                xmin = float(bndbox.find("xmin").text)
                ymin = float(bndbox.find("ymin").text)
                xmax = float(bndbox.find("xmax").text)
                ymax = float(bndbox.find("ymax").text)

                x_center = (xmin + xmax) / 2 / width
                y_center = (ymin + ymax) / 2 / height
                norm_w = (xmax - xmin) / width
                norm_h = (ymax - ymin) / height

                label_lines.append(
                    f"{cls_id} {x_center:.6f} {y_center:.6f} "
                    f"{norm_w:.6f} {norm_h:.6f}"
                )

            label_file = labels_dir / f"{xml_path.stem}.txt"
            label_file.write_text("\n".join(label_lines))

        log("INFO", f"[converter] YOLO labels written: {labels_dir}")
        return str(output_dir)

    def _classification_to_csv(self, input_path: str, output_dir: Path) -> str:
        csv_path = output_dir / "labels.csv"
        rows = []

        input_dir = Path(input_path)
        for cls_dir in sorted(input_dir.iterdir()):
            if not cls_dir.is_dir() or cls_dir.name.startswith("."):
                continue
            for img in cls_dir.glob("*.*"):
                if img.suffix.lower() in (".jpg", ".png", ".jpeg", ".bmp"):
                    rows.append({"image": img.name, "label": cls_dir.name,
                                 "path": str(img)})

        with open(csv_path, "w", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=["image", "label", "path"])
            writer.writeheader()
            writer.writerows(rows)

        log("INFO", f"[converter] CSV written: {csv_path} ({len(rows)} rows)")
        return str(csv_path)

    def _csv_to_classification(self, input_path: str, output_dir: Path) -> str:
        csv_path = Path(input_path)
        if not csv_path.exists():
            csv_path = list(Path(input_path).glob("*.csv"))[0]

        with open(csv_path, "r", newline="") as f:
            reader = csv.DictReader(f)
            for row in reader:
                label = row.get("label", "unknown")
                cls_dir = output_dir / label
                cls_dir.mkdir(parents=True, exist_ok=True)

                src = row.get("path", "")
                if src and Path(src).exists():
                    dest = cls_dir / Path(src).name
                    if not dest.exists():
                        os.symlink(Path(src).resolve(), dest)

        log("INFO", f"[converter] Classification folders written: {output_dir}")
        return str(output_dir)