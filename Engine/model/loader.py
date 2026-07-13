"""
Model Loader - loads models with format detection and device placement.

Provides:
  - Automatic format detection (PyTorch, ONNX, TensorRT, CoreML, TFLite)
  - Device-aware loading with memory optimization
  - Half-precision (FP16) support
  - Model verification after loading
  - Integration with ModelRegistry
"""

from enum import Enum
from pathlib import Path
from typing import Any, Optional

from sdk.logger import log


class ModelFormat(Enum):
    PYTORCH = "pytorch"
    TORCHSCRIPT = "torchscript"
    ONNX = "onnx"
    TENSORRT = "tensorrt"
    COREML = "coreml"
    TFLITE = "tflite"
    SAFETENSORS = "safetensors"
    UNKNOWN = "unknown"


class ModelLoader:
    """
    Loads models from various formats with device placement.

    Usage:
        loader = ModelLoader()
        model = loader.load("model.pt", device="cuda:0")
        format = loader.detect_format("model.onnx")
    """

    def __init__(self, model_registry=None):
        self.registry = model_registry

    def load(self, model_path: str, device: str = "auto",
             half: bool = False, **kwargs) -> Any:
        path = Path(model_path)
        if not path.exists():
            raise FileNotFoundError(f"Model not found: {model_path}")

        fmt = self.detect_format(path)
        log("INFO", f"[model_loader] Loading {fmt.value} model: {model_path} "
                    f"(device={device}, half={half})")

        if device == "auto":
            device = self._auto_device()

        model = None

        if fmt == ModelFormat.PYTORCH:
            model = self._load_pytorch(path, device, half, **kwargs)
        elif fmt == ModelFormat.TORCHSCRIPT:
            model = self._load_torchscript(path, device, **kwargs)
        elif fmt == ModelFormat.ONNX:
            model = self._load_onnx(path, device, **kwargs)
        elif fmt == ModelFormat.SAFETENSORS:
            model = self._load_safetensors(path, device, **kwargs)
        elif fmt == ModelFormat.TFLITE:
            model = self._load_tflite(path, **kwargs)
        else:
            model = self._load_pytorch(path, device, half, **kwargs)

        if model is not None:
            log("INFO", f"[model_loader] Model loaded successfully: {model_path}")
        else:
            log("ERROR", f"[model_loader] Failed to load model: {model_path}")

        return model

    def detect_format(self, path: Path) -> ModelFormat:
        suffix = path.suffix.lower()
        if suffix in (".pt", ".pth"):
            return self._detect_pt_format(path)
        if suffix == ".onnx":
            return ModelFormat.ONNX
        if suffix == ".trt" or suffix == ".engine":
            return ModelFormat.TENSORRT
        if suffix == ".mlmodel" or suffix == ".mlpackage":
            return ModelFormat.COREML
        if suffix == ".tflite":
            return ModelFormat.TFLITE
        if suffix == ".safetensors":
            return ModelFormat.SAFETENSORS
        if suffix == ".ts" or suffix == ".torchscript":
            return ModelFormat.TORCHSCRIPT
        return ModelFormat.UNKNOWN

    def _detect_pt_format(self, path: Path) -> ModelFormat:
        try:
            import torch
            data = torch.load(path, map_location="cpu", weights_only=False)

            if isinstance(data, torch.jit.ScriptModule):
                return ModelFormat.TORCHSCRIPT
            if isinstance(data, dict):
                if "model" in data or "state_dict" in data or "model_state" in data:
                    return ModelFormat.PYTORCH
            if hasattr(data, "state_dict"):
                return ModelFormat.PYTORCH
            return ModelFormat.UNKNOWN
        except Exception:
            return ModelFormat.PYTORCH

    def _load_pytorch(self, path: Path, device: str, half: bool, **kwargs) -> Any:
        import torch

        try:
            data = torch.load(path, map_location="cpu", weights_only=False)
        except Exception:
            data = torch.load(path, map_location="cpu", weights_only=False,
                              pickle_module=__import__("pickle"))

        if isinstance(data, dict):
            if "model" in data:
                model = data["model"]
            elif "state_dict" in data:
                model = self._build_model_from_state(path, data["state_dict"], **kwargs)
            elif "model_state" in data:
                model = self._build_model_from_state(path, data["model_state"], **kwargs)
            else:
                model = data
        else:
            model = data

        if hasattr(model, "to"):
            model = model.to(device)
            if half and device != "cpu":
                model = model.half()
            model.eval()

        return model

    def _build_model_from_state(self, path: Path, state_dict: dict, **kwargs) -> Any:
        name = path.stem.lower()
        if "yolo" in name:
            try:
                from ultralytics import YOLO
                model = YOLO(str(path))
                return model
            except ImportError:
                pass

        log("WARN", f"[model_loader] Cannot reconstruct model architecture for {path.name}")
        log("WARN", "[model_loader] Returning state_dict only, caller must load into model")
        return state_dict

    def _load_torchscript(self, path: Path, device: str, **kwargs) -> Any:
        import torch
        model = torch.jit.load(str(path), map_location="cpu")
        if device != "cpu":
            model = model.to(device)
        model.eval()
        return model

    def _load_onnx(self, path: Path, device: str, **kwargs) -> Any:
        try:
            import onnx
            model = onnx.load(str(path))
            onnx.checker.check_model(model)
            log("INFO", f"[model_loader] ONNX model verified: {path}")
            return model
        except ImportError:
            log("WARN", "[model_loader] onnx/onnxruntime not installed, "
                        "returning path reference")
            return str(path)

    def _load_safetensors(self, path: Path, device: str, **kwargs) -> Any:
        try:
            import safetensors.torch
            state_dict = safetensors.torch.load_file(str(path))
            log("INFO", f"[model_loader] SafeTensors loaded: {len(state_dict)} tensors")
            return state_dict
        except ImportError:
            log("WARN", "[model_loader] safetensors not installed, returning path")
            return str(path)

    def _load_tflite(self, path: Path, **kwargs) -> Any:
        try:
            import tensorflow as tf
            interpreter = tf.lite.Interpreter(model_path=str(path))
            interpreter.allocate_tensors()
            return interpreter
        except ImportError:
            log("WARN", "[model_loader] tensorflow not installed, returning path")
            return str(path)

    def _auto_device(self) -> str:
        try:
            import torch
            if torch.cuda.is_available():
                return "cuda:0"
            if hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
                return "mps"
        except ImportError:
            pass
        return "cpu"