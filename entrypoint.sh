#!/bin/bash

# Start FastAPI AI-service nội bộ
/opt/venv/bin/uvicorn ai-service.ai_service:app --host 127.0.0.1 --port 8001 &

# Start Go Gin server (public port)
exec /app/main
