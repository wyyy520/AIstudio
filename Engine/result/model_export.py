"""
Model Export - exports trained models to various deployment formats.

Provides:
  - PyTorch -> ONNX conversion
  - PyTorch -> TorchScript conversion
  - PyTorch -> TensorRT conversion
  - FP16/INT8 quantization
  - Model optimization for deployment
  - Export metadata tracking
"""

import json
import time
from enum import Enum
from pathlib import Path
from typing import Any, Optional

from sdk.logger import log


class ExportTarget(Enum):
    ONNX = "onnx"
    TORCHSCRIPT = "torchscript"
    TENSORRT = "tensorrt"
    COREML = "coreml"
    TFLITE = "tflite"
    OPENVINO = "openvino"


class ModelExporter:
    """
    Exports trained models to deployment formats.

    Usage:
        exporter = ModelExporter()
        exporter.export(model, "model.pt", ExportTarget.ONNX)
        exporter.export(model, "model.pt", ExportTarget.TORCHSCRIPT)
        exporter.export_to_tensorrt("model.onnx", "model.trt")
    """

    def __init__(self, output_dir: str = "Storage/exports"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self._export_history: list[dict] = []

    def export(self, model: Any, model_path: str,
               target: ExportTarget, **kwargs) -> Optional[str]:
        log("INFO", f"[export] Exporting model to {target.value}: {model_path}")

        if target == ExportTarget.ONNX:
            return self._export_onnx(model, model_path, **kwargs)
        elif target == ExportTarget.TORCHSCRIPT:
            return self._export_torchscript(model, model_path, **kwargs)
        elif target == ExportTarget.TENSORRT:
            return self._export_tensorrt(model, model_path, **kwargs)
        elif target == ExportTarget.COREML:
            return self._export_coreml(model, model_path, **kwargs)
        elif target == ExportTarget.TFLITE:
            return self._export_tflite(model, model_path, **kwargs)
        elif target == ExportTarget.OPENVINO:
            return self._export_openvino(model, model_path, **kwargs)
        else:
            log("ERROR", f"[export] Unsupported target: {target}")
            return None

    def _export_onnx(self, model: Any, model_path: str, **kwargs) -> Optional[str]:
        try:
            import torch

            model_path_obj = Path(model_path)
            output_path = self.output_dir / f"{model_path_obj.stem}.onnx"

            model.eval()
            device = next(model.parameters()).device

            img_size = kwargs.get("img_size", 640)
            batch_size = kwargs.get("batch_size", 1)
            dummy_input = torch.randn(batch_size, 3, img_size, img_size, device=device)

            dynamic_axes = kwargs.get("dynamic_axes", {
                "images": {0: "batch"},
                "output": {0: "batch"},
            })

            input_names = kwargs.get("input_names", ["images"])
            output_names = kwargs.get("output_names", ["output"])

            torch.onnx.export(
                model,
                dummy_input,
                str(output_path),
                verbose=False,
                opset_version=kwargs.get("opset_version", 12),
                input_names=input_names,
                output_names=output_names,
                dynamic_axes=dynamic_axes,
            )

            self._try_onnx_simplify(output_path)

            file_size = output_path.stat().st_size / (1024 * 1024)
            self._record_export(model_path, str(output_path), "onnx", file_size)
            log("INFO", f"[export] ONNX model saved: {output_path} ({file_size:.1f}MB)")
            return str(output_path)

        except ImportError as e:
            log("ERROR", f"[export] PyTorch not available: {e}")
            return None
        except Exception as e:
            log("ERROR", f"[export] ONNX export failed: {e}")
            return None

    def _try_onnx_simplify(self, onnx_path: Path):
        try:
            import onnx
            from onnxsim import simplify
            model = onnx.load(str(onnx_path))
            model_simplified, check = simplify(model)
            if check:
                onnx.save(model_simplified, str(onnx_path))
                log("INFO", f"[export] ONNX model simplified: {onnx_path}")
        except ImportError:
            pass
        except Exception:
            pass

    def _export_torchscript(self, model: Any, model_path: str, **kwargs) -> Optional[str]:
        try:
            import torch

            model_path_obj = Path(model_path)
            output_path = self.output_dir / f"{model_path_obj.stem}.torchscript.pt"

            model.eval()
            device = next(model.parameters()).device

            img_size = kwargs.get("img_size", 640)
            dummy_input = torch.randn(1, 3, img_size, img_size, device=device)

            method = kwargs.get("method", "trace")
            if method == "trace":
                traced = torch.jit.trace(model, dummy_input)
            else:
                traced = torch.jit.script(model)

            traced.save(str(output_path))

            file_size = output_path.stat().st_size / (1024 * 1024)
            self._record_export(model_path, str(output_path), "torchscript", file_size)
            log("INFO", f"[export] TorchScript model saved: {output_path} ({file_size:.1f}MB)")
            return str(output_path)

        except ImportError as e:
            log("ERROR", f"[export] PyTorch not available: {e}")
            return None
        except Exception as e:
            log("ERROR", f"[export] TorchScript export failed: {e}")
            return None

    def _export_tensorrt(self, model: Any, model_path: str, **kwargs) -> Optional[str]:
        try:
            model_path_obj = Path(model_path)
            output_path = self.output_dir / f"{model_path_obj.stem}.engine"

            is_onnx = model_path_obj.suffix == ".onnx"

            if is_onnx:
                try:
                    import tensorrt as trt
                    logger = trt.Logger(trt.Logger.WARNING)
                    builder = trt.Builder(logger)
                    network = builder.create_network(
                        1 << int(trt.NetworkDefinitionCreationFlag.EXPLICIT_BATCH)
                    )
                    parser = trt.OnnxParser(network, logger)

                    with open(model_path, "rb") as f:
                        if not parser.parse(f.read()):
                            for i in range(parser.num_errors):
                                log("ERROR", f"[export] TRT parse error: {parser.get_error(i)}")
                            return None

                    config = builder.create_builder_config()
                    config.max_workspace_size = kwargs.get("workspace_size", 1 << 30)

                    fp16 = kwargs.get("fp16", False)
                    if fp16 and builder.platform_has_fast_fp16:
                        config.set_flag(trt.BuilderFlag.FP16)

                    engine = builder.build_engine(network, config)
                    if engine is None:
                        return None

                    with open(output_path, "wb") as f:
                        f.write(engine.serialize())

                    file_size = output_path.stat().st_size / (1024 * 1024)
                    self._record_export(model_path, str(output_path), "tensorrt", file_size)
                    log("INFO", f"[export] TensorRT engine saved: {output_path} ({file_size:.1f}MB)")
                    return str(output_path)

                except ImportError:
                    log("WARN", "[export] tensorrt not installed, "
                               "attempting torch2trt fallback")
                    try:
                        from torch2trt import torch2trt
                        import torch
                        model.eval()
                        model = model.cuda()
                        x = torch.ones(1, 3, 224, 224).cuda()
                        model_trt = torch2trt(model, [x])
                        torch.save(model_trt.state_dict(), output_path)
                        log("INFO", f"[export] TRT (torch2trt) saved: {output_path}")
                        return str(output_path)
                    except ImportError:
                        log("ERROR", "[export] torch2trt not installed")
                        return None

        except Exception as e:
            log("ERROR", f"[export] TensorRT export failed: {e}")
            return None

    def _export_coreml(self, model: Any, model_path: str, **kwargs) -> Optional[str]:
        try:
            import torch
            model_path_obj = Path(model_path)
            output_path = self.output_dir / f"{model_path_obj.stem}.mlmodel"

            model.eval()
            dummy_input = torch.randn(1, 3, 224, 224)

            traced = torch.jit.trace(model, dummy_input)

            try:
                import coremltools as ct
                mlmodel = ct.convert(
                    traced,
                    inputs=[ct.TensorType(shape=dummy_input.shape)],
                )
                mlmodel.save(str(output_path))
                log("INFO", f"[export] CoreML model saved: {output_path}")
                return str(output_path)
            except ImportError:
                log("WARN", "[export] coremltools not installed, exporting ONNX first")
                onnx_path = self._export_onnx(model, model_path, **kwargs)
                if onnx_path:
                    try:
                        import coremltools as ct
                        mlmodel = ct.converters.onnx.convert(model=onnx_path)
                        mlmodel.save(str(output_path))
                        log("INFO", f"[export] CoreML (from ONNX) saved: {output_path}")
                        return str(output_path)
                    except ImportError:
                        pass
                return None

        except Exception as e:
            log("ERROR", f"[export] CoreML export failed: {e}")
            return None

    def _export_tflite(self, model: Any, model_path: str, **kwargs) -> Optional[str]:
        try:
            onnx_path = self._export_onnx(model, model_path, **kwargs)
            if onnx_path is None:
                return None

            model_path_obj = Path(model_path)
            output_path = self.output_dir / f"{model_path_obj.stem}.tflite"

            try:
                import onnx
                from onnx_tf.backend import prepare
                onnx_model = onnx.load(onnx_path)
                tf_rep = prepare(onnx_model)
                tf_rep.export_graph(str(self.output_dir / f"{model_path_obj.stem}_tf"))

                import tensorflow as tf
                converter = tf.lite.TFLiteConverter.from_saved_model(
                    str(self.output_dir / f"{model_path_obj.stem}_tf")
                )
                tflite_model = converter.convert()
                with open(output_path, "wb") as f:
                    f.write(tflite_model)

                log("INFO", f"[export] TFLite model saved: {output_path}")
                return str(output_path)
            except ImportError:
                log("WARN", "[export] ONNX->TFLite conversion tools not installed")
                return None

        except Exception as e:
            log("ERROR", f"[export] TFLite export failed: {e}")
            return None

    def _export_openvino(self, model: Any, model_path: str, **kwargs) -> Optional[str]:
        try:
            onnx_path = self._export_onnx(model, model_path, **kwargs)
            if onnx_path is None:
                return None

            model_path_obj = Path(model_path)
            output_dir = self.output_dir / f"{model_path_obj.stem}_openvino"
            output_dir.mkdir(parents=True, exist_ok=True)

            try:
                from openvino.tools import mo
                from openvino.runtime import Core
                mo.convert_model(onnx_path, output_dir=str(output_dir))
                log("INFO", f"[export] OpenVINO model saved: {output_dir}")
                return str(output_dir)
            except ImportError:
                log("WARN", "[export] OpenVINO tools not installed")
                return None

        except Exception as e:
            log("ERROR", f"[export] OpenVINO export failed: {e}")
            return None

    def quantize(self, model_path: str, method: str = "fp16") -> Optional[str]:
        path = Path(model_path)
        if not path.exists():
            log("ERROR", f"[export] Model not found: {model_path}")
            return None

        output_path = self.output_dir / f"{path.stem}_{method}.pt"

        try:
            import torch

            if method == "fp16":
                model = torch.load(path, map_location="cpu", weights_only=False)
                if hasattr(model, "half"):
                    model = model.half()
                    torch.save(model, output_path)
                    log("INFO", f"[export] FP16 model saved: {output_path}")
                    return str(output_path)

            elif method == "int8":
                try:
                    import torch.quantization as quant
                    model = torch.load(path, map_location="cpu", weights_only=False)
                    if hasattr(model, "eval"):
                        model.eval()
                    model.qconfig = quant.get_default_qconfig("fbgemm")
                    quant.prepare(model, inplace=True)
                    quant.convert(model, inplace=True)
                    torch.save(model, output_path)
                    log("INFO", f"[export] INT8 model saved: {output_path}")
                    return str(output_path)
                except Exception as e:
                    log("WARN", f"[export] INT8 quantization not supported: {e}")

            log("WARN", f"[export] Quantization method '{method}' not supported")
            return None

        except ImportError:
            log("ERROR", "[export] PyTorch not available for quantization")
            return None

    def _record_export(self, source: str, output: str, format: str, size_mb: float):
        self._export_history.append({
            "source": source,
            "output": output,
            "format": format,
            "size_mb": round(size_mb, 2),
            "timestamp": time.time(),
        })

    def get_export_history(self) -> list[dict]:
        return self._export_history