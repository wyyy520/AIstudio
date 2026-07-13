"""
Result Exporter - outputs training results in various formats.

Provides:
  - JSON/YAML/CSV result export
  - Training summary reports
  - Per-epoch metric history
  - Model comparison tables
  - Integration with downstream reporting
"""

import csv
import json
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any, Optional

from sdk.logger import log


class ExportFormat(Enum):
    JSON = "json"
    YAML = "yaml"
    CSV = "csv"
    SUMMARY = "summary"


@dataclass
class TrainingReport:
    model_name: str
    output_dir: str
    status: str
    epochs_completed: int
    total_epochs: int
    best_epoch: int
    best_loss: float
    best_metrics: dict = field(default_factory=dict)
    final_metrics: dict = field(default_factory=dict)
    training_time_seconds: float = 0.0
    model_path: str = ""
    history: list[dict] = field(default_factory=list)
    config: dict = field(default_factory=dict)
    environment: dict = field(default_factory=dict)
    timestamp: str = ""


class ResultExporter:
    """
    Exports training results in multiple formats.

    Usage:
        exporter = ResultExporter()
        report = exporter.build_report(
            model_name="yolo",
            output_dir="Storage/models",
            history=epoch_results,
        )
        exporter.export(report, format=ExportFormat.JSON)
        exporter.export(report, format=ExportFormat.CSV)
    """

    def __init__(self, output_dir: str = "Storage/results"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)

    def build_report(self, model_name: str, output_dir: str,
                     status: str = "completed",
                     epochs_completed: int = 0,
                     total_epochs: int = 0,
                     best_epoch: int = 0,
                     best_loss: float = 0.0,
                     best_metrics: dict = None,
                     final_metrics: dict = None,
                     training_time_seconds: float = 0.0,
                     model_path: str = "",
                     history: list = None,
                     config: dict = None,
                     environment: dict = None) -> TrainingReport:
        import time as time_module
        return TrainingReport(
            model_name=model_name,
            output_dir=output_dir,
            status=status,
            epochs_completed=epochs_completed,
            total_epochs=total_epochs,
            best_epoch=best_epoch,
            best_loss=best_loss,
            best_metrics=best_metrics or {},
            final_metrics=final_metrics or {},
            training_time_seconds=training_time_seconds,
            model_path=model_path,
            history=history or [],
            config=config or {},
            environment=environment or {},
            timestamp=time_module.strftime("%Y-%m-%dT%H:%M:%SZ", time_module.gmtime()),
        )

    def export(self, report: TrainingReport,
               format: ExportFormat = ExportFormat.JSON,
               filename: str = None) -> str:
        if filename is None:
            filename = f"{report.model_name}_results"

        if format == ExportFormat.JSON:
            return self._export_json(report, filename)
        elif format == ExportFormat.YAML:
            return self._export_yaml(report, filename)
        elif format == ExportFormat.CSV:
            return self._export_csv(report, filename)
        elif format == ExportFormat.SUMMARY:
            return self._export_summary(report, filename)
        else:
            raise ValueError(f"Unsupported format: {format}")

    def _export_json(self, report: TrainingReport, filename: str) -> str:
        filepath = self.output_dir / f"{filename}.json"
        data = {
            "model_name": report.model_name,
            "output_dir": report.output_dir,
            "status": report.status,
            "training": {
                "epochs_completed": report.epochs_completed,
                "total_epochs": report.total_epochs,
                "best_epoch": report.best_epoch,
                "best_loss": report.best_loss,
                "best_metrics": report.best_metrics,
                "final_metrics": report.final_metrics,
                "time_seconds": report.training_time_seconds,
                "time_formatted": self._format_time(report.training_time_seconds),
            },
            "model_path": report.model_path,
            "history": report.history,
            "config": report.config,
            "environment": report.environment,
            "timestamp": report.timestamp,
        }

        with open(filepath, "w") as f:
            json.dump(data, f, indent=2, ensure_ascii=False)

        log("INFO", f"[exporter] JSON results written: {filepath}")
        return str(filepath)

    def _export_yaml(self, report: TrainingReport, filename: str) -> str:
        filepath = self.output_dir / f"{filename}.yaml"
        try:
            import yaml

            data = {
                "model_name": report.model_name,
                "status": report.status,
                "training": {
                    "epochs_completed": report.epochs_completed,
                    "total_epochs": report.total_epochs,
                    "best_epoch": report.best_epoch,
                    "best_loss": round(report.best_loss, 6),
                    "best_metrics": report.best_metrics,
                    "time_seconds": report.training_time_seconds,
                },
                "model_path": report.model_path,
                "timestamp": report.timestamp,
            }

            with open(filepath, "w") as f:
                yaml.dump(data, f, default_flow_style=False, allow_unicode=True)
        except ImportError:
            with open(filepath, "w") as f:
                f.write(f"model_name: {report.model_name}\n")
                f.write(f"status: {report.status}\n")
                f.write(f"best_loss: {report.best_loss}\n")
                f.write(f"best_epoch: {report.best_epoch}\n")

        log("INFO", f"[exporter] YAML results written: {filepath}")
        return str(filepath)

    def _export_csv(self, report: TrainingReport, filename: str) -> str:
        filepath = self.output_dir / f"{filename}.csv"

        if not report.history:
            log("WARN", "[exporter] No history data to export as CSV")
            return ""

        with open(filepath, "w", newline="") as f:
            writer = csv.writer(f)
            header = ["epoch", "loss"]
            extra_keys = []
            for entry in report.history:
                if "metrics" in entry and entry["metrics"]:
                    for k in entry["metrics"].keys():
                        if k not in extra_keys:
                            extra_keys.append(k)
            header.extend(extra_keys)
            writer.writerow(header)

            for entry in report.history:
                row = [entry.get("epoch", ""), entry.get("loss", "")]
                for k in extra_keys:
                    row.append(entry.get("metrics", {}).get(k, ""))
                writer.writerow(row)

        log("INFO", f"[exporter] CSV results written: {filepath}")
        return str(filepath)

    def _export_summary(self, report: TrainingReport, filename: str) -> str:
        filepath = self.output_dir / f"{filename}_summary.txt"

        lines = [
            "=" * 60,
            f"  Training Report: {report.model_name}",
            "=" * 60,
            f"  Status:      {report.status}",
            f"  Epochs:      {report.epochs_completed}/{report.total_epochs}",
            f"  Best Epoch:  {report.best_epoch}",
            f"  Best Loss:   {report.best_loss:.6f}",
            f"  Time:        {self._format_time(report.training_time_seconds)}",
            f"  Model:       {report.model_path}",
            f"  Timestamp:   {report.timestamp}",
            "-" * 60,
        ]

        if report.best_metrics:
            lines.append("  Best Metrics:")
            for k, v in report.best_metrics.items():
                if isinstance(v, (int, float)):
                    lines.append(f"    {k:20s}: {v:.4f}")
                else:
                    lines.append(f"    {k:20s}: {v}")

        if report.final_metrics:
            lines.append("  Final Metrics:")
            for k, v in report.final_metrics.items():
                if isinstance(v, (int, float)):
                    lines.append(f"    {k:20s}: {v:.4f}")
                else:
                    lines.append(f"    {k:20s}: {v}")

        lines.append("=" * 60)

        with open(filepath, "w") as f:
            f.write("\n".join(lines))

        log("INFO", f"[exporter] Summary written: {filepath}")
        return str(filepath)

    def _format_time(self, seconds: float) -> str:
        if seconds < 60:
            return f"{seconds:.1f}s"
        if seconds < 3600:
            return f"{seconds / 60:.1f}m"
        return f"{seconds / 3600:.1f}h"