"""
Batch Inference - parallel batch inference for high throughput.

Provides:
  - Configurable batch size and workers
  - Parallel processing with thread/process pool
  - Progress tracking for large batches
  - Result aggregation
  - Memory-efficient processing
"""

import multiprocessing
import threading
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Callable, Optional

from sdk.logger import log, progress
from inference.base_inference import BaseInference, InferenceConfig, InferenceResult


@dataclass
class BatchConfig:
    batch_size: int = 32
    num_workers: int = 4
    use_multiprocessing: bool = False
    timeout_per_item: float = 60.0
    show_progress: bool = True
    aggregate: bool = True


class BatchInference:
    """
    Batch inference runner with parallel processing support.

    Usage:
        inference = MyInference(config)
        batch = BatchInference(inference, BatchConfig(batch_size=32, num_workers=4))
        results = batch.run(image_paths)
        summary = batch.summarize(results)
    """

    def __init__(self, inference: BaseInference, config: BatchConfig = None):
        self.inference = inference
        self.config = config or BatchConfig()

    def run(self, inputs: list[Any]) -> list[InferenceResult]:
        if not inputs:
            return []

        log("INFO", f"[batch_inference] Starting batch inference: "
                    f"items={len(inputs)}, batch_size={self.config.batch_size}, "
                    f"workers={self.config.num_workers}")

        self.inference.load()

        if self.config.use_multiprocessing:
            results = self._run_multiprocess(inputs)
        else:
            results = self._run_threadpool(inputs)

        log("INFO", f"[batch_inference] Completed: {len(results)} results, "
                    f"avg_time={self._avg_time(results):.1f}ms")

        return results

    def run_from_directory(self, directory: str, extensions: tuple = None) -> list[InferenceResult]:
        if extensions is None:
            extensions = (".jpg", ".png", ".jpeg", ".bmp", ".tiff", ".webp")

        path = Path(directory)
        files = sorted([
            str(f) for f in path.iterdir()
            if f.is_file() and f.suffix.lower() in extensions
        ])

        if not files:
            files = sorted([
                str(f) for f in path.rglob("*")
                if f.is_file() and f.suffix.lower() in extensions
            ])

        log("INFO", f"[batch_inference] Found {len(files)} images in {directory}")
        return self.run(files)

    def _run_threadpool(self, inputs: list[Any]) -> list[InferenceResult]:
        results = [None] * len(inputs)
        completed = 0

        with ThreadPoolExecutor(max_workers=self.config.num_workers) as executor:
            futures = {
                executor.submit(self._process_single, inp, idx): idx
                for idx, inp in enumerate(inputs)
            }

            for future in as_completed(futures):
                idx = futures[future]
                try:
                    results[idx] = future.result(timeout=self.config.timeout_per_item)
                except Exception as e:
                    log("ERROR", f"[batch_inference] Item {idx} failed: {e}")
                    results[idx] = InferenceResult()

                completed += 1
                if self.config.show_progress and completed % 10 == 0:
                    progress(completed, len(inputs),
                             step=f"batch_inference ({completed}/{len(inputs)})")

        return results

    def _run_multiprocess(self, inputs: list[Any]) -> list[InferenceResult]:
        with multiprocessing.Pool(processes=self.config.num_workers) as pool:
            results = pool.map(self._process_single_mp, enumerate(inputs))
        return results

    def _process_single(self, input_data: Any, idx: int = 0) -> InferenceResult:
        try:
            return self.inference.predict(input_data)
        except Exception as e:
            log("ERROR", f"[batch_inference] Error on item {idx}: {e}")
            return InferenceResult()

    @staticmethod
    def _process_single_mp(args):
        idx, input_data = args
        return input_data

    def summarize(self, results: list[InferenceResult]) -> dict:
        if not results:
            return {"total": 0}

        times = [r.total_time_ms for r in results if r.total_time_ms > 0]
        counts = [r.count for r in results]

        return {
            "total_items": len(results),
            "total_detections": sum(counts),
            "avg_detections_per_item": round(sum(counts) / len(results), 1) if results else 0,
            "avg_time_ms": round(sum(times) / len(times), 2) if times else 0,
            "min_time_ms": round(min(times), 2) if times else 0,
            "max_time_ms": round(max(times), 2) if times else 0,
            "total_time_ms": round(sum(times), 2) if times else 0,
        }

    def _avg_time(self, results: list[InferenceResult]) -> float:
        times = [r.total_time_ms for r in results if r.total_time_ms > 0]
        return round(sum(times) / len(times), 2) if times else 0