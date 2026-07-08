"""
Environment Detector - detects Python, CUDA, PyTorch, and GPU status.

Outputs JSON lines to stdout so the Go backend can read the environment status.
Can be run standalone: python -m runtime.env_detector
"""

import json
import sys
import platform
import subprocess
from typing import Any


def detect_python() -> dict[str, Any]:
    """Detect Python version and path."""
    return {
        "version": platform.python_version(),
        "path": sys.executable,
        "implementation": platform.python_implementation(),
    }


def detect_pytorch() -> dict[str, Any]:
    """Detect PyTorch installation and CUDA availability."""
    try:
        import torch
        return {
            "version": torch.__version__,
            "cuda_available": torch.cuda.is_available(),
            "cuda_version": torch.version.cuda if torch.cuda.is_available() else None,
            "gpu_count": torch.cuda.device_count() if torch.cuda.is_available() else 0,
            "gpu_names": [torch.cuda.get_device_name(i) for i in range(torch.cuda.device_count())] if torch.cuda.is_available() else [],
            "installed": True,
        }
    except ImportError:
        return {
            "version": None,
            "cuda_available": False,
            "cuda_version": None,
            "gpu_count": 0,
            "gpu_names": [],
            "installed": False,
        }


def detect_cuda() -> dict[str, Any]:
    """Detect CUDA toolkit via nvidia-smi."""
    result = {
        "cuda_version": None,
        "driver_version": None,
        "gpus": [],
    }
    try:
        output = subprocess.check_output(
            ["nvidia-smi", "--query-gpu=name,memory.total,driver_version",
             "--format=csv,noheader,nounits"],
            stderr=subprocess.DEVNULL,
            timeout=10,
        )
        lines = output.decode("utf-8").strip().split("\n")
        for line in lines:
            parts = [p.strip() for p in line.split(",")]
            if len(parts) >= 3:
                result["gpus"].append({
                    "name": parts[0],
                    "memory_mb": parts[1],
                })
                if not result["driver_version"]:
                    result["driver_version"] = parts[2]
    except (FileNotFoundError, subprocess.CalledProcessError,
            subprocess.TimeoutExpired):
        pass

    # Try to get CUDA version from nvcc
    try:
        output = subprocess.check_output(
            ["nvcc", "--version"],
            stderr=subprocess.DEVNULL,
            timeout=10,
        )
        text = output.decode("utf-8")
        for line in text.split("\n"):
            if "release" in line.lower():
                parts = line.split("release")
                if len(parts) > 1:
                    result["cuda_version"] = parts[1].strip().split(",")[0].strip()
                    break
    except (FileNotFoundError, subprocess.CalledProcessError,
            subprocess.TimeoutExpired):
        pass

    return result


def detect_ultralytics() -> dict[str, Any]:
    """Detect ultralytics (YOLO) installation."""
    try:
        import ultralytics
        return {
            "version": ultralytics.__version__,
            "installed": True,
        }
    except ImportError:
        return {
            "version": None,
            "installed": False,
        }


def get_full_status() -> dict[str, Any]:
    """Get the complete environment status."""
    return {
        "os": platform.system(),
        "os_version": platform.version(),
        "architecture": platform.machine(),
        "python": detect_python(),
        "pytorch": detect_pytorch(),
        "cuda": detect_cuda(),
        "ultralytics": detect_ultralytics(),
    }


def main():
    """Standalone entry point for environment detection."""
    status = get_full_status()
    json.dump(status, sys.stdout, indent=2, ensure_ascii=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()