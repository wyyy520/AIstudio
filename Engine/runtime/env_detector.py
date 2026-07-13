"""
Environment Detector - detects Python and PyTorch availability.

Returns a plain dict. No GPU management, no subprocess calls.
"""

from __future__ import annotations

import platform
import sys
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
            "installed": True,
        }
    except ImportError:
        return {
            "version": None,
            "cuda_available": False,
            "installed": False,
        }


def get_full_status() -> dict[str, Any]:
    """Get the complete environment status (plain dict)."""
    return {
        "os": platform.system(),
        "os_version": platform.version(),
        "architecture": platform.machine(),
        "python": detect_python(),
        "pytorch": detect_pytorch(),
    }


def main():
    """Standalone entry point for environment detection."""
    import json
    status = get_full_status()
    json.dump(status, sys.stdout, indent=2, ensure_ascii=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
