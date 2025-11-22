FROM python:3.11-slim

WORKDIR /app

# Cài libs hệ thống tối thiểu (chủ yếu cho numpy, sklearn)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential gcc && \
    rm -rf /var/lib/apt/lists/*

# Copy code
COPY . /app

# Cài dependencies
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Biến môi trường (nếu cần)
ENV MODEL_DIR=/app/models
ENV CONFIDENCE_MIN=0.65

# Expose port trong container
EXPOSE 8000

# Lệnh chạy server
CMD ["uvicorn", "ai_service:app", "--host", "0.0.0.0", "--port", "8000"]
