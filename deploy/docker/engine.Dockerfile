# =============================================================================
# AIStudio Python Engine - Docker Build
# =============================================================================
FROM python:3.11-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1-mesa-glx \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY Engine/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY Engine/ .

RUN mkdir -p /app/models /app/datasets /app/logs

EXPOSE 8082

ENV ENGINE_HOST=0.0.0.0
ENV ENGINE_PORT=8082
ENV PYTHONUNBUFFERED=1

CMD ["python", "server.py", "--host", "0.0.0.0", "--port", "8082"]