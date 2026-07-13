"""
Realtime Inference - low-latency inference for streaming/live applications.

Provides:
  - Persistent model loading for low latency
  - Streaming input support (camera, video, websocket)
  - Frame skipping for performance
  - Async result callbacks
  - FPS tracking
"""

import queue
import threading
import time
from dataclasses import dataclass, field
from typing import Any, Callable, Optional

from sdk.logger import log
from inference.base_inference import BaseInference, InferenceConfig, InferenceResult


@dataclass
class RealtimeConfig:
    max_fps: int = 30
    frame_skip: int = 0
    queue_size: int = 10
    async_mode: bool = True
    display: bool = False
    save_output: bool = False
    output_dir: str = "Storage/predictions"


class RealtimeInference:
    """
    Low-latency real-time inference runner.

    Usage:
        inference = MyInference(config)
        realtime = RealtimeInference(inference, RealtimeConfig(max_fps=30))

        def on_result(result):
            print(f"Detected {result.count} objects in {result.total_time_ms}ms")

        realtime.on_result = on_result
        realtime.start()

        for frame in camera_stream:
            realtime.feed(frame)

        realtime.stop()
    """

    def __init__(self, inference: BaseInference, config: RealtimeConfig = None):
        self.inference = inference
        self.config = config or RealtimeConfig()

        self._frame_queue: queue.Queue = queue.Queue(maxsize=self.config.queue_size)
        self._result_queue: queue.Queue = queue.Queue()
        self._running = False
        self._worker_thread: Optional[threading.Thread] = None

        self._frame_count: int = 0
        self._processed_count: int = 0
        self._skipped_count: int = 0
        self._start_time: float = 0.0
        self._fps: float = 0.0
        self._last_fps_update: float = 0.0

        self.on_result: Optional[Callable[[InferenceResult], None]] = None
        self.on_error: Optional[Callable[[Exception], None]] = None

    def start(self):
        if self._running:
            return

        self.inference.load()
        self._running = True
        self._start_time = time.time()
        self._last_fps_update = self._start_time

        if self.config.async_mode:
            self._worker_thread = threading.Thread(
                target=self._process_loop, daemon=True
            )
            self._worker_thread.start()

        log("INFO", f"[realtime] Started (async={self.config.async_mode}, "
                    f"max_fps={self.config.max_fps}, queue={self.config.queue_size})")

    def stop(self):
        self._running = False
        if self._worker_thread:
            self._worker_thread.join(timeout=5.0)
            self._worker_thread = None

        elapsed = time.time() - self._start_time
        fps = self._processed_count / elapsed if elapsed > 0 else 0
        log("INFO", f"[realtime] Stopped: processed={self._processed_count}, "
                    f"skipped={self._skipped_count}, avg_fps={fps:.1f}")

    def feed(self, frame: Any) -> Optional[InferenceResult]:
        self._frame_count += 1

        if self.config.frame_skip > 0 and self._frame_count % (self.config.frame_skip + 1) != 0:
            self._skipped_count += 1
            return None

        if self.config.async_mode:
            try:
                self._frame_queue.put_nowait(frame)
            except queue.Full:
                self._skipped_count += 1
            return None
        else:
            return self._process_frame(frame)

    def get_result(self, timeout: float = 1.0) -> Optional[InferenceResult]:
        try:
            return self._result_queue.get(timeout=timeout)
        except queue.Empty:
            return None

    @property
    def fps(self) -> float:
        now = time.time()
        elapsed = now - self._last_fps_update
        if elapsed >= 1.0:
            recent = self._processed_count - getattr(self, "_last_count", 0)
            self._fps = recent / elapsed if elapsed > 0 else 0
            self._last_fps_update = now
            setattr(self, "_last_count", self._processed_count)
        return self._fps

    @property
    def stats(self) -> dict:
        elapsed = time.time() - self._start_time if self._start_time > 0 else 0
        return {
            "running": self._running,
            "frames_received": self._frame_count,
            "frames_processed": self._processed_count,
            "frames_skipped": self._skipped_count,
            "fps": round(self.fps, 1),
            "avg_fps": round(self._processed_count / elapsed, 1) if elapsed > 0 else 0,
            "queue_size": self._frame_queue.qsize(),
            "uptime": round(elapsed, 1),
        }

    def _process_loop(self):
        frame_interval = 1.0 / self.config.max_fps if self.config.max_fps > 0 else 0
        last_frame_time = 0.0

        while self._running:
            try:
                frame = self._frame_queue.get(timeout=0.1)
            except queue.Empty:
                continue

            now = time.time()
            if frame_interval > 0 and (now - last_frame_time) < frame_interval:
                self._skipped_count += 1
                continue

            last_frame_time = now
            self._process_frame(frame)

    def _process_frame(self, frame: Any) -> InferenceResult:
        try:
            result = self.inference.predict(frame)
            self._processed_count += 1

            if self.on_result:
                try:
                    self.on_result(result)
                except Exception:
                    pass

            if self.config.async_mode:
                try:
                    self._result_queue.put_nowait(result)
                except queue.Full:
                    pass

            return result
        except Exception as e:
            if self.on_error:
                try:
                    self.on_error(e)
                except Exception:
                    pass
            return InferenceResult()