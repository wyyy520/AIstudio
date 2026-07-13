"""
Base Inference - unified inference interface for all models.

Provides:
  - Common inference configuration
  - Single/batch inference methods
  - Device-aware inference
  - Result formatting
  - Warmup support
"""

import time
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Optional

from sdk.logger import log


@dataclass
class InferenceConfig:
    model_path: str = ""
    device: str = "auto"
    confidence_threshold: float = 0.25
    iou_threshold: float = 0.45
    max_detections: int = 300
    img_size: int = 640
    batch_size: int = 1
    half_precision: bool = False
    warmup: bool = True
    warmup_iterations: int = 3
    extra: dict = field(default_factory=dict)


@dataclass
class InferenceResult:
    predictions: list[dict[str, Any]] = field(default_factory=list)
    inference_time_ms: float = 0.0
    preprocessing_time_ms: float = 0.0
    postprocessing_time_ms: float = 0.0
    total_time_ms: float = 0.0
    model_name: str = ""
    image_size: tuple = (0, 0)
    count: int = 0


class BaseInference(ABC):
    """
    Abstract base class for model inference.

    Usage:
        class MyInference(BaseInference):
            def _load_model(self):
                self.model = load_model(self.config.model_path)

            def _preprocess(self, input_data):
                return preprocessed_input

            def _infer(self, preprocessed):
                return self.model(preprocessed)

            def _postprocess(self, raw_output, original_input):
                return formatted_predictions

        inference = MyInference(InferenceConfig(model_path="model.pt"))
        result = inference.predict("image.jpg")
        results = inference.predict_batch(["img1.jpg", "img2.jpg"])
    """

    def __init__(self, config: InferenceConfig):
        self.config = config
        self.model: Any = None
        self._loaded = False
        self._warmup_done = False

    def load(self):
        if self._loaded:
            return

        if self.config.device == "auto":
            self._auto_detect_device()

        log("INFO", f"[inference] Loading model: {self.config.model_path} "
                    f"(device={self.config.device})")

        self._load_model()
        self._loaded = True

        if self.config.warmup:
            self._do_warmup()

    def _auto_detect_device(self):
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

    def _do_warmup(self):
        log("INFO", f"[inference] Warming up ({self.config.warmup_iterations} iterations)...")
        try:
            import numpy as np
            dummy = np.zeros((self.config.img_size, self.config.img_size, 3), dtype=np.uint8)
            for _ in range(self.config.warmup_iterations):
                self.predict(dummy)
        except Exception as e:
            log("WARN", f"[inference] Warmup failed: {e}")
        self._warmup_done = True

    def predict(self, input_data: Any) -> InferenceResult:
        if not self._loaded:
            self.load()

        t0 = time.time()

        t1 = time.time()
        preprocessed = self._preprocess(input_data)
        t2 = time.time()

        raw_output = self._infer(preprocessed)
        t3 = time.time()

        predictions = self._postprocess(raw_output, input_data)
        t4 = time.time()

        return InferenceResult(
            predictions=predictions,
            inference_time_ms=round((t3 - t2) * 1000, 2),
            preprocessing_time_ms=round((t2 - t1) * 1000, 2),
            postprocessing_time_ms=round((t4 - t3) * 1000, 2),
            total_time_ms=round((t4 - t0) * 1000, 2),
            model_name=self.config.model_path,
            count=len(predictions),
        )

    def predict_batch(self, inputs: list[Any]) -> list[InferenceResult]:
        if not self._loaded:
            self.load()
        results = []
        for input_data in inputs:
            results.append(self.predict(input_data))
        return results

    def unload(self):
        if self.model is not None:
            try:
                if hasattr(self.model, "cpu"):
                    self.model.cpu()
                del self.model
            except Exception:
                pass
            self.model = None
        self._loaded = False
        self._warmup_done = False

        try:
            import gc
            import torch
            gc.collect()
            if torch.cuda.is_available():
                torch.cuda.empty_cache()
        except ImportError:
            pass

    @abstractmethod
    def _load_model(self):
        pass

    @abstractmethod
    def _preprocess(self, input_data: Any) -> Any:
        pass

    @abstractmethod
    def _infer(self, preprocessed: Any) -> Any:
        pass

    @abstractmethod
    def _postprocess(self, raw_output: Any, original_input: Any) -> list[dict]:
        pass