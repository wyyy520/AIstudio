"""AI Studio Engine - Main entry point."""

import argparse
import logging
import signal
import sys

from grpc.server import create_server, ENGINE_VERSION

logger = logging.getLogger("aiengine")


def setup_logging(level: str = "INFO"):
    logging.basicConfig(
        level=getattr(logging, level.upper(), logging.INFO),
        format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )


def main():
    parser = argparse.ArgumentParser(description="AI Studio Engine")
    parser.add_argument("--port", type=int, default=50051, help="gRPC server port")
    parser.add_argument("--workers", type=int, default=10, help="Max worker threads")
    parser.add_argument("--log-level", default="INFO", help="Log level")
    args = parser.parse_args()

    setup_logging(args.log_level)

    logger.info(f"Starting AI Studio Engine v{ENGINE_VERSION}")
    logger.info(f"Listening on port {args.port}")

    server = create_server(port=args.port, max_workers=args.workers)

    def shutdown(signum, frame):
        logger.info("Shutting down engine...")
        server.stop(grace=5)
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)

    server.start()
    logger.info("Engine is ready")
    server.wait_for_termination()


if __name__ == "__main__":
    main()
