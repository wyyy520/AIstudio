"""gRPC Server - AIEngine gRPC service implementation."""

import json
import logging
from concurrent import futures

import grpc

from proto import aiengine_pb2
from proto import aiengine_pb2_grpc
from manager.model_manager import ModelManager, ModelType
from task_router import TaskRouter

logger = logging.getLogger(__name__)

ENGINE_VERSION = "0.1.0"


class AIEngineService(aiengine_pb2_grpc.AIEngineServicer):
    """gRPC service implementing the AIEngine proto definition."""

    def __init__(self):
        self.model_manager = ModelManager()
        self.task_router = TaskRouter(model_manager=self.model_manager)
        logger.info("AIEngine service initialized")

    def ExecuteTask(self, request, context):
        """Execute an AI task."""
        task_id = request.task_id or str(__import__("uuid").uuid4())
        task_type = request.task_type
        input_data = request.input
        config = dict(request.config)

        logger.info(f"Executing task: {task_id} (type={task_type})")

        result = self.task_router.route(task_type, input_data, config)

        return aiengine_pb2.TaskResponse(
            task_id=task_id,
            status=result.get("status", "error"),
            result=result.get("result", ""),
            metadata=result.get("metadata", {}),
        )

    def GetModelStatus(self, request, context):
        """Get model status."""
        model_name = request.model_name
        info = self.model_manager.get_model_info(model_name)

        if info is None:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(f"Model '{model_name}' not found")
            return aiengine_pb2.ModelStatusResponse()

        return aiengine_pb2.ModelStatusResponse(
            model_name=info.name,
            status=info.state.value,
            device=info.device,
            memory_usage=info.memory_usage,
        )

    def LoadModel(self, request, context):
        """Load a model."""
        model_type_str = request.model_type.lower()
        type_map = {
            "llm": ModelType.LLM,
            "vision": ModelType.VISION,
            "timeseries": ModelType.TIMESERIES,
            "embedding": ModelType.EMBEDDING,
        }
        model_type = type_map.get(model_type_str)
        if model_type is None:
            return aiengine_pb2.LoadModelResponse(
                success=False,
                message=f"Unknown model type: {request.model_type}",
            )

        self.model_manager.register_model(
            name=request.model_name,
            model_type=model_type,
            path=request.model_path,
            options=dict(request.options),
        )
        success = self.model_manager.load_model(request.model_name)

        return aiengine_pb2.LoadModelResponse(
            success=success,
            message=f"Model '{request.model_name}' {'loaded' if success else 'failed to load'}",
        )

    def UnloadModel(self, request, context):
        """Unload a model."""
        success = self.model_manager.unload_model(request.model_name)
        return aiengine_pb2.UnloadModelResponse(
            success=success,
            message=f"Model '{request.model_name}' {'unloaded' if success else 'not found or not loaded'}",
        )

    def HealthCheck(self, request, context):
        """Health check."""
        return aiengine_pb2.HealthCheckResponse(
            status="healthy",
            version=ENGINE_VERSION,
        )


def create_server(port: int = 50051, max_workers: int = 10) -> grpc.server:
    """Create and configure the gRPC server."""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=max_workers))
    aiengine_pb2_grpc.AIEngineServicerToServer.add_to_server(
        AIEngineService(), server
    )
    server.add_insecure_port(f"[::]:{port}")
    return server
