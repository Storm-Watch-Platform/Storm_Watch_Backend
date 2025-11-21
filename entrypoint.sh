#!/bin/bash

# 1️⃣ Start FastAPI AI-service background (private)
uvicorn ai-service.ai_service:app --host 127.0.0.1 --port 8001 &

# 2️⃣ Start Go Gin server (public)
exec /app/main
