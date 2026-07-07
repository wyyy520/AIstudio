"""Task Router - routes tasks to appropriate handlers."""

import json
import logging
import uuid
from typing import Any, Dict, Optional

from handlers.vision.handler import VisionHandler
from handlers.nlp.handler import NLPHandler
from handlers.timeseries.handler import TimeseriesHandler
from manager.model_manager import ModelManager

logger = logging.getLogger(__name__)


class TaskRouter:
    """Routes AI tasks to the appropriate handler based on task_type."""

    def __init__(self, model_manager: Optional[ModelManager] = None):
        self.model_manager = model_manager or ModelManager()
        self._handlers = {
            "vision": VisionHandler(model_manager=self.model_manager),
            "nlp": NLPHandler(model_manager=self.model_manager),
            "timeseries": TimeseriesHandler(model_manager=self.model_manager),
        }

    def route(self, task_type: str, input_data: str, config: Dict[str, str]) -> Dict[str, Any]:
        """Route a task to its handler and return the result."""
        category = task_type.split(".")[0] if "." in task_type else task_type

        handler = self._handlers.get(category)
        if handler is None:
            return {
                "status": "error",
                "result": json.dumps({"error": f"No handler for category: {category}"}),
                "metadata": {"task_type": task_type},
            }

        try:
            result = handler.handle(task_type, input_data, config)
            return {
                "status": "success",
                "result": json.dumps(result),
                "metadata": {"task_type": task_type, "handler": category},
            }
        except Exception as e:
            logger.error(f"Task {task_type} failed: {e}")
            return {
                "status": "error",
                "result": json.dumps({"error": str(e)}),
                "metadata": {"task_type": task_type, "handler": category},
            }

    def register_handler(self, category: str, handler: Any) -> None:
        """Register a custom handler for a category."""
        self._handlers[category] = handler
        logger.info(f"Registered handler for category: {category}")

    def supported_task_types(self) -> list:
        """Return all supported task type prefixes."""
        return list(self._handlers.keys())
